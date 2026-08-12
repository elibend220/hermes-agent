import { Router, Response } from 'express';
import db from '../database/index.js';
import blockchain from '../blockchain/index.js';
import logger from '../logger/index.js';

const router = Router();

router.get('/health', async (req, res: Response) => {
  try {
    const checks = {
      database: false,
      blockchain: false,
      timestamp: new Date().toISOString(),
    };

    // Check database
    try {
      const result = await db.query('SELECT NOW()');
      checks.database = !!result.rows[0];
    } catch (err) {
      logger.error({ err }, 'Database health check failed');
    }

    // Check blockchain
    try {
      const gasPrice = await blockchain.getGasPrice();
      checks.blockchain = !!gasPrice;
    } catch (err) {
      logger.error({ err }, 'Blockchain health check failed');
    }

    const allHealthy = checks.database && checks.blockchain;
    const statusCode = allHealthy ? 200 : 503;

    return res.status(statusCode).json({
      status: allHealthy ? 'healthy' : 'degraded',
      checks,
    });
  } catch (err) {
    logger.error({ err }, 'Health check error');
    return res.status(500).json({
      status: 'unhealthy',
      error: 'Health check failed',
    });
  }
});

router.get('/status', async (req, res: Response) => {
  try {
    return res.json({
      service: 'banking-system',
      version: '1.0.0',
      environment: process.env.NODE_ENV || 'development',
      timestamp: new Date().toISOString(),
      uptime: process.uptime(),
    });
  } catch (err) {
    logger.error({ err }, 'Status check error');
    return res.status(500).json({
      error: 'Status check failed',
    });
  }
});

export default router;
