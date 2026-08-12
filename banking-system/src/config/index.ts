import dotenv from 'dotenv';

dotenv.config();

export const config = {
  server: {
    port: parseInt(process.env.PORT || '3000'),
    nodeEnv: process.env.NODE_ENV || 'development',
  },
  database: {
    url: process.env.DATABASE_URL || '',
    poolSize: parseInt(process.env.DATABASE_POOL_SIZE || '20'),
  },
  redis: {
    url: process.env.REDIS_URL || 'redis://localhost:6379',
  },
  kafka: {
    brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(','),
    clientId: process.env.KAFKA_CLIENT_ID || 'banking-system',
  },
  blockchain: {
    ethereumRpcUrl: process.env.ETHEREUM_RPC_URL || '',
    chainId: parseInt(process.env.ETHEREUM_CHAIN_ID || '1'),
    usdtAddress: process.env.USDT_CONTRACT_ADDRESS || '',
    settlementWallet: process.env.SETTLEMENT_WALLET_ADDRESS || '',
    settlementPrivateKey: process.env.SETTLEMENT_PRIVATE_KEY || '',
  },
  auth: {
    apiKeySecret: process.env.API_KEY_SECRET || '',
    jwtSecret: process.env.JWT_SECRET || '',
    jwtExpiry: process.env.JWT_EXPIRY || '24h',
  },
  security: {
    allowedOrigins: (process.env.ALLOWED_ORIGINS || '').split(','),
    rateLimitWindowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS || '60000'),
    rateLimitMaxRequests: parseInt(process.env.RATE_LIMIT_MAX_REQUESTS || '100'),
  },
  monitoring: {
    logLevel: process.env.LOG_LEVEL || 'info',
    enableMetrics: process.env.ENABLE_METRICS === 'true',
    metricsPort: parseInt(process.env.METRICS_PORT || '9090'),
  },
};

export default config;
