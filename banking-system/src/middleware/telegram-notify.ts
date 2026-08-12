import { Request, Response, NextFunction } from 'express';
import { createTelegramService } from '../services/telegram.js';
import logger from '../logger/index.js';

const telegramService = createTelegramService();

export async function notifyTransactionCreated(req: Request, res: Response, next: NextFunction) {
  // Capture response to notify about transaction
  const originalJson = res.json;

  res.json = function (data: any) {
    if (data.success && data.data?.id && res.statusCode === 201) {
      const tx = data.data;
      if (telegramService) {
        telegramService
          .notifyTransaction(tx.id, tx.fromAddress, tx.toAddress, tx.amount, tx.currency)
          .catch((err) => logger.error({ err }, 'Telegram notification failed'));
      }
    }
    return originalJson.call(this, data);
  };

  next();
}

export async function notifySettlementCreated(req: Request, res: Response, next: NextFunction) {
  const originalJson = res.json;

  res.json = function (data: any) {
    if (data.success && data.data?.id && res.statusCode === 201 && data.data.batchNumber) {
      const batch = data.data;
      if (telegramService) {
        telegramService
          .notifySettlement(batch.id, batch.transactionCount, batch.totalAmount)
          .catch((err) => logger.error({ err }, 'Telegram notification failed'));
      }
    }
    return originalJson.call(this, data);
  };

  next();
}

export async function notifySettlementConfirmed(req: Request, res: Response, next: NextFunction) {
  const originalJson = res.json;

  res.json = function (data: any) {
    if (
      data.success &&
      data.data?.id &&
      data.data?.blockchainTxHash &&
      data.data?.status === 'completed'
    ) {
      const batch = data.data;
      if (telegramService) {
        telegramService
          .notifySettlementConfirmed(batch.id, batch.blockchainTxHash, batch.totalAmount)
          .catch((err) => logger.error({ err }, 'Telegram notification failed'));
      }
    }
    return originalJson.call(this, data);
  };

  next();
}

export default { notifyTransactionCreated, notifySettlementCreated, notifySettlementConfirmed };
