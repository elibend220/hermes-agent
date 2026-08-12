# 🏦 Banking System - Webhook Gateway

A production-grade banking webhook gateway with blockchain settlement capabilities, designed for high-volume payment processing with cryptographic security and real-time settlement.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Applications                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│              API Gateway & Authentication                    │
│         (API Key Validation, Rate Limiting, TLS)             │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                 Webhook Handler Service                      │
│    (Payload Validation, Deduplication, Event Logging)       │
└──────────────────────┬──────────────────────────────────────┘
                       │
            ┌──────────┴──────────┐
            │                     │
    ┌───────▼────────┐   ┌───────▼────────┐
    │  Redis Cache   │   │  Message Queue │
    │  (Dedup, TTL)  │   │   (Kafka/SQS)  │
    └────────────────┘   └────────────────┘
            │                     │
            └──────────┬──────────┘
                       │
        ┌──────────────▼──────────────────┐
        │  Settlement Processing Service  │
        │  (Batch Creation, Scheduling)   │
        └──────────────┬──────────────────┘
                       │
        ┌──────────────▼──────────────────┐
        │  Blockchain Transaction Service │
        │  (USDT Transfers, Validation)   │
        └──────────────┬──────────────────┘
                       │
        ┌──────────────▼──────────────────┐
        │   Ethereum Network / Smart      │
        │   Contracts (Mainnet/Testnet)   │
        └─────────────────────────────────┘
```

## Key Features

- **High-Volume Processing**: Handle thousands of webhook requests per minute
- **Blockchain Integration**: Direct USDT (ERC-20) settlement via Ethereum
- **Idempotency**: Duplicate detection and prevention
- **Real-time Settlement**: Instant or batched cryptocurrency settlement
- **Audit Trail**: Complete transaction history and compliance logging
- **Webhook Security**: HMAC signature verification and timestamp validation
- **Rate Limiting**: Per-user request throttling
- **Enterprise Reliability**: 99.99% uptime architecture with load balancing
- **Monitoring**: Prometheus metrics, structured logging, health checks

## Technology Stack

- **Runtime**: Node.js 20+ (TypeScript)
- **Framework**: Express.js
- **Database**: PostgreSQL 16
- **Cache**: Redis
- **Message Queue**: Apache Kafka
- **Blockchain**: Ethers.js v6
- **Logging**: Pino
- **Validation**: Joi
- **Security**: Helmet, CORS, HMAC-SHA256

## Getting Started

### Prerequisites

- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16
- Redis 7+
- Apache Kafka 7.5+
- Ethereum JSON-RPC endpoint (Alchemy, Infura)

### Installation

1. Clone the repository:
```bash
git clone <repo-url>
cd banking-system
```

2. Install dependencies:
```bash
npm install
```

3. Create `.env` file:
```bash
cp .env.example .env
```

4. Configure environment variables in `.env`

### Development

Start all services with Docker Compose:

```bash
docker-compose up -d
```

Run database migrations:

```bash
npm run db:migrate
```

Start the development server:

```bash
npm run dev
```

The API will be available at `http://localhost:3000`

### Production Deployment

Build the application:

```bash
npm run build
```

Run in production:

```bash
NODE_ENV=production npm start
```

Or use Docker:

```bash
docker build -t banking-system:latest .
docker run -d \
  -e DATABASE_URL=postgresql://... \
  -e ETHEREUM_RPC_URL=https://... \
  -p 3000:3000 \
  banking-system:latest
```

## API Documentation

### Authentication

All endpoints require an `X-API-Key` header:

```bash
curl -H "X-API-Key: your-api-key" http://localhost:3000/api/v1/webhooks
```

### Webhook Endpoint

**POST** `/api/v1/webhooks`

Process a payment webhook:

```json
{
  "eventId": "evt_123456",
  "timestamp": "2024-08-12T10:30:00Z",
  "type": "payment",
  "data": {
    "fromAddress": "0x1234567890123456789012345678901234567890",
    "toAddress": "0x0987654321098765432109876543210987654321",
    "amount": "1000.50",
    "currency": "USDT",
    "reference": "ORD-2024-001",
    "description": "Payment for services"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "txn_abc123",
    "eventId": "evt_123456",
    "status": "received",
    "createdAt": "2024-08-12T10:30:00Z"
  },
  "timestamp": "2024-08-12T10:30:00Z"
}
```

### Settlement Endpoints

**POST** `/api/v1/settlement/batches`

Create a settlement batch:

```json
{
  "transactionIds": ["txn_123", "txn_124", "txn_125"]
}
```

**POST** `/api/v1/settlement/batches/:batchId/process`

Process a settlement batch and submit to blockchain:

```json
{}
```

**GET** `/api/v1/settlement/batches/:batchId`

Get batch details and status

**POST** `/api/v1/settlement/batches/:batchId/reconcile`

Verify blockchain confirmation

### Health Checks

**GET** `/api/v1/health`

System health status

**GET** `/api/v1/status`

Service status and uptime

## Database Schema

### transactions
- `id` (UUID): Primary key
- `event_id` (VARCHAR): Webhook event identifier
- `user_id` (UUID): User reference
- `from_address` (VARCHAR): Source blockchain address
- `to_address` (VARCHAR): Destination blockchain address
- `amount` (VARCHAR): Transaction amount
- `currency` (VARCHAR): Currency type (USDT/USD)
- `status` (VARCHAR): Transaction state
- `blockchain_tx_hash` (VARCHAR): Ethereum transaction hash
- `created_at` (TIMESTAMP): Creation timestamp

### settlement_batches
- `id` (UUID): Primary key
- `batch_number` (BIGINT): Sequential batch number
- `status` (VARCHAR): Batch processing status
- `transaction_count` (INTEGER): Number of transactions
- `total_amount` (VARCHAR): Total settlement amount
- `blockchain_tx_hash` (VARCHAR): Settlement transaction hash
- `created_at` (TIMESTAMP): Batch creation time
- `completed_at` (TIMESTAMP): Completion timestamp

## Security Considerations

1. **API Key Management**:
   - Store hashed API keys in database
   - Rotate keys regularly
   - Use strong, cryptographically random keys

2. **Webhook Validation**:
   - Verify HMAC-SHA256 signatures
   - Check timestamp freshness (5-minute window)
   - Implement idempotency using event IDs

3. **Blockchain Security**:
   - Store private keys in secure vaults (AWS KMS, HashiCorp Vault)
   - Never commit keys to version control
   - Implement transaction signing and verification
   - Use hardware wallets for high-value operations

4. **Network Security**:
   - Enforce TLS 1.3+ for all connections
   - Implement rate limiting and DDoS protection
   - Use VPC/private networks for service communication
   - Enable audit logging for all operations

## Monitoring & Observability

### Logs

Structured JSON logging with Pino:

```bash
tail -f application.log | jq '.'
```

### Metrics

Prometheus metrics available at `/metrics` (when enabled)

### Health Checks

Kubernetes-compatible health checks:

```bash
curl http://localhost:3000/api/v1/health
```

## Error Handling

All errors follow a consistent format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "statusCode": 400
  }
}
```

### Common Error Codes

- `INVALID_API_KEY` (401): Authentication failed
- `INVALID_SIGNATURE` (401): Webhook signature verification failed
- `VALIDATION_ERROR` (400): Request validation failed
- `INSUFFICIENT_BALANCE` (402): Blockchain transfer insufficient funds
- `TRANSACTION_FAILED` (500): Blockchain transaction failed
- `INTERNAL_ERROR` (500): Server error

## Performance Considerations

- **Database Connection Pooling**: 20 connections by default
- **Redis Caching**: Deduplication and rate limiting
- **Kafka Batching**: Asynchronous event processing
- **Request Timeout**: 30 seconds
- **Rate Limit**: 100 requests per minute per API key

## Testing

Run test suite:

```bash
npm test
```

Run with coverage:

```bash
npm run test:coverage
```

## Contributing

See CONTRIBUTING.md for guidelines

## License

MIT

## Support

- **Documentation**: https://docs.banking-system.local
- **Status Page**: https://status.banking-system.local
- **Support Email**: support@banking-system.local
