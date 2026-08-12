import { Request, Response, NextFunction } from 'express';
import crypto from 'crypto';
import config from '../config/index.js';
import logger from '../logger/index.js';
import db from '../database/index.js';

export interface AuthRequest extends Request {
  userId?: string;
  apiKey?: string;
}

export async function validateApiKey(req: AuthRequest, res: Response, next: NextFunction) {
  try {
    const apiKey = req.headers['x-api-key'] as string;

    if (!apiKey) {
      logger.warn('Missing API key');
      return res.status(401).json({
        success: false,
        error: { code: 'INVALID_API_KEY', message: 'API key is required', statusCode: 401 },
      });
    }

    // Hash the API key for lookup
    const hashedKey = crypto.createHash('sha256').update(apiKey).digest('hex');

    const result = await db.query('SELECT id, email FROM users WHERE api_key_hash = $1 AND is_active = true', [
      hashedKey,
    ]);

    if (result.rows.length === 0) {
      logger.warn({ apiKey: apiKey.substring(0, 10) }, 'Invalid API key');
      return res.status(401).json({
        success: false,
        error: { code: 'INVALID_API_KEY', message: 'Invalid API key', statusCode: 401 },
      });
    }

    req.userId = result.rows[0].id;
    req.apiKey = apiKey;
    next();
  } catch (err) {
    logger.error({ err }, 'API key validation error');
    return res.status(500).json({
      success: false,
      error: { code: 'AUTH_ERROR', message: 'Authentication failed', statusCode: 500 },
    });
  }
}

export async function validateWebhookSignature(req: AuthRequest, res: Response, next: NextFunction) {
  try {
    const signature = req.headers['x-webhook-signature'] as string;
    const timestamp = req.headers['x-webhook-timestamp'] as string;

    if (!signature || !timestamp) {
      logger.warn('Missing webhook signature or timestamp');
      return res.status(401).json({
        success: false,
        error: { code: 'INVALID_SIGNATURE', message: 'Webhook signature is required', statusCode: 401 },
      });
    }

    // Verify timestamp is not older than 5 minutes
    const requestTime = parseInt(timestamp);
    const currentTime = Math.floor(Date.now() / 1000);
    if (Math.abs(currentTime - requestTime) > 300) {
      logger.warn('Webhook timestamp too old');
      return res.status(401).json({
        success: false,
        error: { code: 'INVALID_TIMESTAMP', message: 'Webhook timestamp is too old', statusCode: 401 },
      });
    }

    // Verify signature using HMAC-SHA256
    const body = JSON.stringify(req.body);
    const message = `${timestamp}.${body}`;
    const expectedSignature = crypto.createHmac('sha256', config.auth.apiKeySecret).update(message).digest('hex');

    if (signature !== expectedSignature) {
      logger.warn('Invalid webhook signature');
      return res.status(401).json({
        success: false,
        error: { code: 'INVALID_SIGNATURE', message: 'Invalid webhook signature', statusCode: 401 },
      });
    }

    next();
  } catch (err) {
    logger.error({ err }, 'Webhook signature validation error');
    return res.status(500).json({
      success: false,
      error: { code: 'AUTH_ERROR', message: 'Authentication failed', statusCode: 500 },
    });
  }
}

export function errorHandler(err: any, req: Request, res: Response, next: NextFunction) {
  logger.error({ err }, 'Unhandled error');
  return res.status(500).json({
    success: false,
    error: { code: 'INTERNAL_ERROR', message: 'Internal server error', statusCode: 500 },
  });
}
