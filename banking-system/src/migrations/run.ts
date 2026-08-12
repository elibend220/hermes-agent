import { readFileSync } from 'fs';
import { join } from 'path';
import db from '../database/index.js';
import logger from '../logger/index.js';

async function runMigrations() {
  try {
    logger.info('Starting database migrations');

    const migrationPath = join(process.cwd(), 'src/migrations/001_init.sql');
    const migration = readFileSync(migrationPath, 'utf-8');

    await db.query(migration);

    logger.info('Migrations completed successfully');
  } catch (err) {
    logger.error({ err }, 'Migration failed');
    throw err;
  }
}

runMigrations()
  .then(() => {
    logger.info('Database migrations finished');
    process.exit(0);
  })
  .catch((err) => {
    logger.error({ err }, 'Failed to run migrations');
    process.exit(1);
  });
