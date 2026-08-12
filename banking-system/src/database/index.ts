import { Pool } from 'pg';
import config from '../config/index.js';
import logger from '../logger/index.js';

export class Database {
  private pool: Pool;

  constructor() {
    this.pool = new Pool({
      connectionString: config.database.url,
      max: config.database.poolSize,
    });

    this.pool.on('error', (err) => {
      logger.error({ err }, 'Unexpected error on idle client');
    });
  }

  async connect() {
    try {
      const client = await this.pool.connect();
      client.release();
      logger.info('Database connected successfully');
    } catch (err) {
      logger.error({ err }, 'Failed to connect to database');
      throw err;
    }
  }

  async query(text: string, params?: any[]) {
    const start = Date.now();
    try {
      const result = await this.pool.query(text, params);
      const duration = Date.now() - start;
      logger.debug({ query: text, duration }, 'Query executed');
      return result;
    } catch (err) {
      logger.error({ err, query: text }, 'Query failed');
      throw err;
    }
  }

  async close() {
    await this.pool.end();
    logger.info('Database connection pool closed');
  }
}

export const db = new Database();
export default db;
