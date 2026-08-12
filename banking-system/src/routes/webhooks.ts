import { Router, Response } from 'express';
import Joi from 'joi';
import { AuthRequest, validateApiKey } from '../middleware/auth.js';
import webhookService from '../services/webhook.js';
import { WebhookPayload, SuccessResponse } from '../types/index.js';
import logger from '../logger/index.js';

const router = Router();

const webhookSchema = Joi.object({
  eventId: Joi.string().required(),
  timestamp: Joi.string().required(),
  type: Joi.string().valid('payment', 'settlement', 'reconciliation').required(),
  data: Joi.object({
    fromAddress: Joi.string().required(),
    toAddress: Joi.string().required(),
    amount: Joi.string().required(),
    currency: Joi.string().valid('USDT', 'USD').required(),
    reference: Joi.string().optional(),
    description: Joi.string().optional(),
  }).required(),
});

router.post('/', validateApiKey, async (req: AuthRequest, res: Response) => {
  try {
    const { error, value } = webhookSchema.validate(req.body);

    if (error) {
      logger.warn({ error: error.message }, 'Webhook validation failed');
      return res.status(400).json({
        success: false,
        error: { code: 'VALIDATION_ERROR', message: error.message, statusCode: 400 },
      });
    }

    const payload: WebhookPayload = value;
    const transaction = await webhookService.processWebhook(payload, req.userId!);

    const response: SuccessResponse<any> = {
      success: true,
      data: transaction,
      timestamp: new Date().toISOString(),
    };

    return res.status(201).json(response);
  } catch (err: any) {
    logger.error({ err }, 'Webhook processing error');
    return res.status(500).json({
      success: false,
      error: { code: 'WEBHOOK_ERROR', message: err.message || 'Webhook processing failed', statusCode: 500 },
    });
  }
});

router.get('/:eventId', validateApiKey, async (req: AuthRequest, res: Response) => {
  try {
    const transaction = await webhookService.getTransaction(req.params.eventId, req.userId!);

    const response: SuccessResponse<any> = {
      success: true,
      data: transaction,
      timestamp: new Date().toISOString(),
    };

    return res.json(response);
  } catch (err: any) {
    logger.error({ err }, 'Failed to fetch transaction');
    return res.status(404).json({
      success: false,
      error: { code: 'NOT_FOUND', message: 'Transaction not found', statusCode: 404 },
    });
  }
});

router.get('/', validateApiKey, async (req: AuthRequest, res: Response) => {
  try {
    const limit = Math.min(parseInt(req.query.limit as string) || 50, 100);
    const offset = parseInt(req.query.offset as string) || 0;

    const transactions = await webhookService.getTransactionsByUserId(req.userId!, limit, offset);

    const response: SuccessResponse<any> = {
      success: true,
      data: {
        transactions,
        limit,
        offset,
        count: transactions.length,
      },
      timestamp: new Date().toISOString(),
    };

    return res.json(response);
  } catch (err: any) {
    logger.error({ err }, 'Failed to fetch transactions');
    return res.status(500).json({
      success: false,
      error: { code: 'FETCH_ERROR', message: 'Failed to fetch transactions', statusCode: 500 },
    });
  }
});

export default router;
