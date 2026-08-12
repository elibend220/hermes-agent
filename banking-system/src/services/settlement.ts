import { v4 as uuidv4 } from 'uuid';
import db from '../database/index.js';
import logger from '../logger/index.js';
import blockchain from '../blockchain/index.js';
import { SettlementBatch, Transaction } from '../types/index.js';

export class SettlementService {
  async createSettlementBatch(transactionIds: string[]): Promise<SettlementBatch> {
    const batchId = uuidv4();

    try {
      logger.info({ batchId, transactionCount: transactionIds.length }, 'Creating settlement batch');

      // Get batch number
      const batchNumberResult = await db.query(
        'SELECT COALESCE(MAX(batch_number), 0) + 1 as batch_number FROM settlement_batches',
      );
      const batchNumber = batchNumberResult.rows[0].batch_number;

      // Get total amount from transactions
      const transactionsResult = await db.query(
        'SELECT SUM(CAST(amount AS DECIMAL)) as total_amount FROM transactions WHERE id = ANY($1) AND status = $2',
        [transactionIds, 'received'],
      );

      const totalAmount = transactionsResult.rows[0].total_amount || '0';

      // Create settlement batch record
      const batchResult = await db.query(
        `INSERT INTO settlement_batches (id, batch_number, status, transaction_count, total_amount, created_at)
         VALUES ($1, $2, $3, $4, $5, $6)
         RETURNING *`,
        [batchId, batchNumber, 'pending', transactionIds.length, totalAmount, new Date()],
      );

      logger.info({ batchId, totalAmount }, 'Settlement batch created');
      return batchResult.rows[0];
    } catch (err) {
      logger.error({ err, batchId }, 'Settlement batch creation failed');
      throw err;
    }
  }

  async processSettlementBatch(batchId: string): Promise<SettlementBatch> {
    try {
      logger.info({ batchId }, 'Processing settlement batch');

      // Update batch status to processing
      await db.query('UPDATE settlement_batches SET status = $1 WHERE id = $2', ['processing', batchId]);

      // Get batch details
      const batchResult = await db.query('SELECT * FROM settlement_batches WHERE id = $1', [batchId]);
      const batch = batchResult.rows[0];

      if (!batch) {
        throw new Error('Batch not found');
      }

      // Get transactions for this batch
      const txResult = await db.query(
        'SELECT DISTINCT to_address, SUM(CAST(amount AS DECIMAL)) as total FROM transactions WHERE batch_id = $1 GROUP BY to_address',
        [batchId],
      );

      let totalTxHash: string | undefined;

      // Process settlements
      for (const row of txResult.rows) {
        try {
          const txHash = await blockchain.transferUsdt(row.to_address, row.total);
          totalTxHash = txHash;

          // Update transactions with blockchain hash
          await db.query('UPDATE transactions SET blockchain_tx_hash = $1, status = $2 WHERE batch_id = $3 AND to_address = $4', [
            txHash,
            'pending',
            batchId,
            row.to_address,
          ]);

          logger.info({ batchId, recipient: row.to_address, txHash }, 'Settlement transaction sent');
        } catch (err) {
          logger.error({ err, batchId, recipient: row.to_address }, 'Settlement transaction failed');
        }
      }

      // Update batch as completed
      const completedBatch = await db.query(
        'UPDATE settlement_batches SET status = $1, blockchain_tx_hash = $2, completed_at = $3 WHERE id = $4 RETURNING *',
        ['completed', totalTxHash, new Date(), batchId],
      );

      logger.info({ batchId }, 'Settlement batch completed');
      return completedBatch.rows[0];
    } catch (err) {
      logger.error({ err, batchId }, 'Settlement batch processing failed');

      // Mark batch as failed
      await db.query('UPDATE settlement_batches SET status = $1 WHERE id = $2', ['failed', batchId]);
      throw err;
    }
  }

  async getSettlementBatch(batchId: string): Promise<SettlementBatch> {
    const result = await db.query('SELECT * FROM settlement_batches WHERE id = $1', [batchId]);

    if (result.rows.length === 0) {
      throw new Error('Settlement batch not found');
    }

    return result.rows[0];
  }

  async getPendingBatches(): Promise<SettlementBatch[]> {
    const result = await db.query('SELECT * FROM settlement_batches WHERE status = $1 ORDER BY created_at ASC', ['pending']);
    return result.rows;
  }

  async getSettlementHistory(limit: number = 50): Promise<SettlementBatch[]> {
    const result = await db.query('SELECT * FROM settlement_batches ORDER BY created_at DESC LIMIT $1', [limit]);
    return result.rows;
  }

  async reconcileSettlement(batchId: string): Promise<boolean> {
    try {
      logger.info({ batchId }, 'Reconciling settlement');

      const batchResult = await db.query('SELECT * FROM settlement_batches WHERE id = $1', [batchId]);
      const batch = batchResult.rows[0];

      if (!batch.blockchain_tx_hash) {
        throw new Error('No blockchain transaction hash found');
      }

      // Check blockchain confirmation
      const txStatus = await blockchain.getTransactionStatus(batch.blockchain_tx_hash);

      if (txStatus.status === 'confirmed') {
        await db.query(
          'UPDATE transactions SET status = $1 WHERE batch_id = $2',
          ['confirmed', batchId],
        );

        logger.info({ batchId }, 'Settlement reconciliation completed');
        return true;
      }

      return false;
    } catch (err) {
      logger.error({ err, batchId }, 'Settlement reconciliation failed');
      throw err;
    }
  }
}

export const settlementService = new SettlementService();
export default settlementService;
