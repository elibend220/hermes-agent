import sqlite3 from 'sqlite3';
import { open, Database as SqliteDB } from 'sqlite';
import logger from '../logger/index.js';

let db: SqliteDB | null = null;

export class SqliteDatabase {
  async connect() {
    try {
      db = await open({
        filename: './banking.db',
        driver: sqlite3.Database,
      });

      logger.info('SQLite database connected');

      // Enable foreign keys
      await db.exec('PRAGMA foreign_keys = ON');

      // Initialize schema
      await this.initializeSchema();
    } catch (err) {
      logger.error({ err }, 'Failed to connect to SQLite database');
      throw err;
    }
  }

  private async initializeSchema() {
    if (!db) throw new Error('Database not connected');

    const schema = `
      CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        wallet_address TEXT UNIQUE NOT NULL,
        api_key_hash TEXT UNIQUE NOT NULL,
        is_active INTEGER DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
      );

      CREATE TABLE IF NOT EXISTS transactions (
        id TEXT PRIMARY KEY,
        event_id TEXT NOT NULL,
        user_id TEXT NOT NULL,
        batch_id TEXT,
        from_address TEXT NOT NULL,
        to_address TEXT NOT NULL,
        amount TEXT NOT NULL,
        currency TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'pending',
        blockchain_tx_hash TEXT,
        metadata TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id)
      );

      CREATE TABLE IF NOT EXISTS settlement_batches (
        id TEXT PRIMARY KEY,
        batch_number INTEGER UNIQUE NOT NULL,
        status TEXT NOT NULL DEFAULT 'pending',
        transaction_count INTEGER NOT NULL,
        total_amount TEXT NOT NULL,
        blockchain_tx_hash TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        completed_at DATETIME
      );

      CREATE TABLE IF NOT EXISTS webhook_events (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        event_id TEXT UNIQUE NOT NULL,
        event_type TEXT NOT NULL,
        payload TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'pending',
        error_message TEXT,
        processed_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id)
      );

      CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
      CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
      CREATE INDEX IF NOT EXISTS idx_settlement_batches_status ON settlement_batches(status);
      CREATE INDEX IF NOT EXISTS idx_users_api_key_hash ON users(api_key_hash);
    `;

    for (const statement of schema.split(';')) {
      if (statement.trim()) {
        await db.exec(statement);
      }
    }

    logger.info('Database schema initialized');
  }

  async query(text: string, params?: any[]) {
    if (!db) throw new Error('Database not connected');

    try {
      // For SELECT queries
      if (text.trim().toUpperCase().startsWith('SELECT')) {
        const rows = await db.all(text, params);
        return { rows };
      }

      // For INSERT/UPDATE/DELETE queries
      if (text.includes('RETURNING')) {
        const cleanText = text.split('RETURNING')[0].trim();
        await db.run(cleanText, params);
        // Re-query for the inserted row
        const idParam = params?.[0];
        const selectQuery = 'SELECT * FROM transactions WHERE id = ? LIMIT 1';
        const rows = await db.all(selectQuery, [idParam]);
        return { rows };
      }

      const result = await db.run(text, params);
      return { changes: result.changes, lastID: result.lastID };
    } catch (err) {
      logger.error({ err, query: text }, 'Query failed');
      throw err;
    }
  }

  async close() {
    if (db) {
      await db.close();
      logger.info('SQLite database closed');
    }
  }
}

export const sqliteDb = new SqliteDatabase();
export default sqliteDb;
