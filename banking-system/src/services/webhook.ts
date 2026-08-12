import { v4 as uuidv4 } from 'uuid';
import db from '../database/index.js';
import logger from '../logger/index.js';
import blockchain from '../blockchain/index.js';
import { WebhookPayload, Transaction } from '../types/index.js';

export class WebhookService {
  async processWebhook(payload: WebhookPayload, userId: string): Promise<Transaction> {
    const txId = uuidv4();

    try {
      logger.info({ eventId: payload.eventId, txId, userId }, 'Processing webhook');

      // Validate addresses
      const isFromValid = await blockchain.validateAddress(payload.data.fromAddress);
      const isToValid = await blockchain.validateAddress(payload.data.toAddress);

      if (!isFromValid || !isToValid) {
        throw new Error('Invalid blockchain address');
      }

      // Store initial transaction record
      const result = await db.query(
        `INSERT INTO transactions (id, event_id, user_id, from_address, to_address, amount, currency, status, created_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
         RETURNING *`,
        [txId, payload.eventId, userId, payload.data.fromAddress, payload.data.toAddress, payload.data.amount, payload.data.currency, 'received', new Date()],
      );

      logger.info({ txId, eventId: payload.eventId }, 'Transaction received');

      // Emit event for async processing
      await this.emitTransactionEvent(result.rows[0]);

      return result.rows[0];
    } catch (err) {
      logger.error({ err, eventId: payload.eventId, txId }, 'Webhook processing failed');

      // Store failed transaction
      await db.query(
        `INSERT INTO transactions (id, event_id, user_id, from_address, to_address, amount, currency, status, created_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
        [txId, payload.eventId, userId, payload.data.fromAddress, payload.data.toAddress, payload.data.amount, payload.data.currency, 'failed', new Date()],
      );

      throw err;
    }
  }

  async getTransaction(txId: string, userId: string): Promise<Transaction> {
    const result = await db.query('SELECT * FROM transactions WHERE id = $1 AND user_id = $2', [txId, userId]);

    if (result.rows.length === 0) {
      throw new Error('Transaction not found');
    }

    return result.rows[0];
  }

  async getTransactionsByUserId(userId: string, limit: number = 50, offset: number = 0): Promise<Transaction[]> {
    const result = await db.query(
      'SELECT * FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3',
      [userId, limit, offset],
    );

    return result.rows;
  }

  async getTransactionsByStatus(status: string, limit: number = 100): Promise<Transaction[]> {
    const result = await db.query(
      'SELECT * FROM transactions WHERE status = $1 ORDER BY created_at DESC LIMIT $2',
      [status, limit],
    );

    return result.rows;
  }

  private async emitTransactionEvent(transaction: Transaction): Promise<void> {
    // This will be implemented with Kafka/Redis pub-sub
    logger.info({ txId: transaction.id }, 'Transaction event emitted for processing');
  }
}

export const webhookService = new WebhookService();
export default webhookService;
