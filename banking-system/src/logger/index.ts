import pino from 'pino';
import config from '../config/index.js';

export const logger = pino({
  level: config.monitoring.logLevel,
  transport:
    config.server.nodeEnv === 'development'
      ? {
          target: 'pino-pretty',
          options: {
            colorize: true,
            ignore: 'pid,hostname',
            singleLine: false,
          },
        }
      : undefined,
  base: {
    service: 'banking-system',
    environment: config.server.nodeEnv,
  },
});

export default logger;
