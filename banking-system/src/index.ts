import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import pinoHttp from 'pino-http';
import config from './config/index.js';
import logger from './logger/index.js';
import db from './database/index.js';
import { errorHandler } from './middleware/auth.js';
import webhookRoutes from './routes/webhooks.js';
import settlementRoutes from './routes/settlement.js';
import healthRoutes from './routes/health.js';

const app = express();

// Security middleware
app.use(helmet());
app.use(cors({
  origin: config.security.allowedOrigins,
  credentials: true,
}));

// Logging middleware
app.use(pinoHttp({ logger }));

// Body parsing
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ limit: '10mb', extended: true }));

// API Routes
app.use('/api/v1/webhooks', webhookRoutes);
app.use('/api/v1/settlement', settlementRoutes);
app.use('/api/v1', healthRoutes);

// Root endpoint
app.get('/', (req, res) => {
  return res.json({
    service: 'Banking System - Webhook Gateway',
    version: '1.0.0',
    documentation: 'https://docs.banking-system.local/api',
    endpoints: {
      webhooks: '/api/v1/webhooks',
      settlement: '/api/v1/settlement',
      health: '/api/v1/health',
      status: '/api/v1/status',
    },
  });
});

// 404 handler
app.use((req, res) => {
  return res.status(404).json({
    success: false,
    error: { code: 'NOT_FOUND', message: 'Endpoint not found', statusCode: 404 },
  });
});

// Error handling
app.use(errorHandler);

async function startServer() {
  try {
    // Connect to database
    await db.connect();
    logger.info('Database connection established');

    // Start listening
    app.listen(config.server.port, () => {
      logger.info({ port: config.server.port }, 'Server started successfully');
      console.log(`🏦 Banking System running on port ${config.server.port}`);
    });
  } catch (err) {
    logger.error({ err }, 'Failed to start server');
    process.exit(1);
  }
}

// Graceful shutdown
process.on('SIGINT', async () => {
  logger.info('Shutting down gracefully');
  await db.close();
  process.exit(0);
});

process.on('SIGTERM', async () => {
  logger.info('Termination signal received');
  await db.close();
  process.exit(0);
});

startServer().catch((err) => {
  logger.error({ err }, 'Failed to initialize server');
  process.exit(1);
});

export default app;
