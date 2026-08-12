import { Router, Response } from 'express';
import Joi from 'joi';
import { AuthRequest, validateApiKey } from '../middleware/auth.js';
import settlementService from '../services/settlement.js';
import { SuccessResponse } from '../types/index.js';
import logger from '../logger/index.js';

const router = Router();

const createBatchSchema = Joi.object({
  transactionIds: Joi.array().items(Joi.string()).required(),
});

router.post('/batches', validateApiKey, async (req: AuthRequest, res: Response) => {
  try {
    const { error, value } = createBatchSchema.validate(req.body);

    if (error) {
      logger.warn({ error: error.message }, 'Settlement batch creation validation failed');
      return res.status(400).json({
        success: false,
        error: { code: 'VALIDATION_ERROR', message: error.message, statusCode: 400 },
      });
    }

    const batch = await settlementService.createSettlementBatch(value.transactionIds);

    const response: SuccessResponse<any> = {
      success: true,
      data: batch,
      timestamp: new Date().toISOString(),
    };

    return res.status(201).json(response);
  } catch (err: any) {
    logger.error({ err }, 'Settlement batch creation failed');
    return res.status(500).json({
      success: false,
      error: { code: 'BATCH_ERROR', message: err.message || 'Failed to create batch', statusCode: 500 },
    });
  }
});

router.post('/batches/:batchId/process', validateApiKey, async (req: AuthRequest, res: Response) => {
  try {
    const batch = await settlementService.processSettlementBatch(req.params.batchId);

    const response: SuccessResponse<any> = {
      success: true,
      data: batch,
      timestamp: new Date().toISOString(),
    };

    return res.json(response);
  } catch (err: any) {
    logger.error({ err }, 'Settlement batch processing failed');
    return res.status(500).json({
      success: false,
      error: { code: 'PROCESSING_ERROR', message: err.message || 'Failed to process batch', statusCode: 500 },
    });
  }
});

router.get('/batches/:batchId', validateApiKey, async (req: AuthRequest, res: Response) => {
  try {
    const batch = await settlementService.getSettlementBatch(req.params.batchId);

    const response: SuccessResponse<any> = {
      success: true,
      data: batch,
      timestamp: new Date().toISOString(),
    };

    return res.json(response);
  } catch (err: any) {
    logger.error({ err }, 'Failed to fetch settlement batch');
    return res.status(404).json({
      success: false,
      error: { code: 'NOT_FOUND', message: 'Batch not found', statusCode: 404 },
    });
  }
});

router.get('/batches', validateApiKey, async (req: AuthRequest, res: Response) => {
  try {
    const batches = await settlementService.getSettlementHistory();

    const response: SuccessResponse<any> = {
      success: true,
      data: {
        batches,
        count: batches.length,
      },
      timestamp: new Date().toISOString(),
    };

    return res.json(response);
  } catch (err: any) {
    logger.error({ err }, 'Failed to fetch settlement batches');
    return res.status(500).json({
      success: false,
      error: { code: 'FETCH_ERROR', message: 'Failed to fetch batches', statusCode: 500 },
    });
  }
});

router.post('/batches/:batchId/reconcile', validateApiKey, async (req: AuthRequest, res: Response) => {
  try {
    const isReconciled = await settlementService.reconcileSettlement(req.params.batchId);

    const response: SuccessResponse<any> = {
      success: true,
      data: {
        batchId: req.params.batchId,
        reconciled: isReconciled,
      },
      timestamp: new Date().toISOString(),
    };

    return res.json(response);
  } catch (err: any) {
    logger.error({ err }, 'Settlement reconciliation failed');
    return res.status(500).json({
      success: false,
      error: { code: 'RECONCILIATION_ERROR', message: err.message || 'Failed to reconcile', statusCode: 500 },
    });
  }
});

export default router;
