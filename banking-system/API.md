# Banking System API Documentation

## Base URL

```
https://api.banking-system.local/api/v1
```

## Authentication

All requests must include the following header:

```
X-API-Key: your-api-key-here
Content-Type: application/json
```

## Response Format

### Success Response

```json
{
  "success": true,
  "data": {
    // Response data
  },
  "timestamp": "2024-08-12T10:30:00Z"
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description",
    "statusCode": 400
  }
}
```

## Webhooks API

### 1. Submit Webhook

**Endpoint**: `POST /webhooks`

**Description**: Submit a new webhook for payment processing

**Request Body**:

```json
{
  "eventId": "evt_unique_identifier",
  "timestamp": "2024-08-12T10:30:00Z",
  "type": "payment",
  "data": {
    "fromAddress": "0x1234567890123456789012345678901234567890",
    "toAddress": "0x0987654321098765432109876543210987654321",
    "amount": "1000.50",
    "currency": "USDT",
    "reference": "ORD-2024-001",
    "description": "Payment for services rendered"
  }
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| eventId | string | Yes | Unique event identifier for idempotency |
| timestamp | string | Yes | ISO 8601 timestamp of the event |
| type | string | Yes | Event type (payment, settlement, reconciliation) |
| data.fromAddress | string | Yes | Source Ethereum address (0x...) |
| data.toAddress | string | Yes | Destination Ethereum address (0x...) |
| data.amount | string | Yes | Amount in USDT/USD |
| data.currency | string | Yes | Currency type (USDT or USD) |
| data.reference | string | No | Reference number or invoice ID |
| data.description | string | No | Transaction description |

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "txn_abc123def456",
    "eventId": "evt_unique_identifier",
    "fromAddress": "0x1234567890123456789012345678901234567890",
    "toAddress": "0x0987654321098765432109876543210987654321",
    "amount": "1000.50",
    "currency": "USDT",
    "status": "received",
    "createdAt": "2024-08-12T10:30:00Z"
  },
  "timestamp": "2024-08-12T10:30:00Z"
}
```

**Status Codes**:

- `201 Created`: Webhook processed successfully
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Invalid API key
- `409 Conflict`: Duplicate event ID
- `500 Internal Server Error`: Server error

**Example cURL**:

```bash
curl -X POST https://api.banking-system.local/api/v1/webhooks \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "eventId": "evt_20240812_001",
    "timestamp": "2024-08-12T10:30:00Z",
    "type": "payment",
    "data": {
      "fromAddress": "0x742d35Cc6634C0532925a3b844Bc9e7595f42472",
      "toAddress": "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
      "amount": "100.00",
      "currency": "USDT"
    }
  }'
```

### 2. Get Transaction

**Endpoint**: `GET /webhooks/:transactionId`

**Description**: Retrieve a specific transaction by ID

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "txn_abc123def456",
    "eventId": "evt_unique_identifier",
    "status": "confirmed",
    "amount": "1000.50",
    "currency": "USDT",
    "fromAddress": "0x1234567890123456789012345678901234567890",
    "toAddress": "0x0987654321098765432109876543210987654321",
    "blockchainTxHash": "0xabcd1234...",
    "createdAt": "2024-08-12T10:30:00Z",
    "updatedAt": "2024-08-12T10:31:00Z"
  },
  "timestamp": "2024-08-12T10:30:00Z"
}
```

**Example cURL**:

```bash
curl -X GET https://api.banking-system.local/api/v1/webhooks/txn_abc123def456 \
  -H "X-API-Key: your-api-key"
```

### 3. List Transactions

**Endpoint**: `GET /webhooks`

**Description**: Retrieve all transactions for the authenticated user

**Query Parameters**:

| Parameter | Type | Default | Max | Description |
|-----------|------|---------|-----|-------------|
| limit | integer | 50 | 100 | Number of results to return |
| offset | integer | 0 | - | Number of results to skip |

**Response**:

```json
{
  "success": true,
  "data": {
    "transactions": [
      {
        "id": "txn_abc123def456",
        "eventId": "evt_unique_identifier",
        "status": "confirmed",
        "amount": "1000.50",
        "currency": "USDT",
        "createdAt": "2024-08-12T10:30:00Z"
      }
    ],
    "limit": 50,
    "offset": 0,
    "count": 1
  },
  "timestamp": "2024-08-12T10:30:00Z"
}
```

**Example cURL**:

```bash
curl -X GET "https://api.banking-system.local/api/v1/webhooks?limit=25&offset=0" \
  -H "X-API-Key: your-api-key"
```

## Settlement API

### 1. Create Settlement Batch

**Endpoint**: `POST /settlement/batches`

**Description**: Create a new settlement batch from multiple transactions

**Request Body**:

```json
{
  "transactionIds": [
    "txn_abc123def456",
    "txn_def456ghi789",
    "txn_ghi789jkl012"
  ]
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "batch_xyz789",
    "batchNumber": 1,
    "status": "pending",
    "transactionCount": 3,
    "totalAmount": "3050.75",
    "createdAt": "2024-08-12T10:30:00Z"
  },
  "timestamp": "2024-08-12T10:30:00Z"
}
```

**Example cURL**:

```bash
curl -X POST https://api.banking-system.local/api/v1/settlement/batches \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "transactionIds": ["txn_abc123def456", "txn_def456ghi789"]
  }'
```

### 2. Process Settlement Batch

**Endpoint**: `POST /settlement/batches/:batchId/process`

**Description**: Submit batch to blockchain for settlement

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "batch_xyz789",
    "status": "completed",
    "blockchainTxHash": "0xabcd1234ef567890...",
    "completedAt": "2024-08-12T10:31:00Z"
  },
  "timestamp": "2024-08-12T10:31:00Z"
}
```

**Example cURL**:

```bash
curl -X POST https://api.banking-system.local/api/v1/settlement/batches/batch_xyz789/process \
  -H "X-API-Key: your-api-key"
```

### 3. Get Settlement Batch

**Endpoint**: `GET /settlement/batches/:batchId`

**Description**: Retrieve details of a specific settlement batch

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "batch_xyz789",
    "batchNumber": 1,
    "status": "completed",
    "transactionCount": 3,
    "totalAmount": "3050.75",
    "blockchainTxHash": "0xabcd1234ef567890...",
    "createdAt": "2024-08-12T10:30:00Z",
    "completedAt": "2024-08-12T10:31:00Z"
  },
  "timestamp": "2024-08-12T10:31:00Z"
}
```

### 4. List Settlement Batches

**Endpoint**: `GET /settlement/batches`

**Description**: Retrieve all settlement batches

**Response**:

```json
{
  "success": true,
  "data": {
    "batches": [
      {
        "id": "batch_xyz789",
        "batchNumber": 1,
        "status": "completed",
        "totalAmount": "3050.75"
      }
    ],
    "count": 1
  },
  "timestamp": "2024-08-12T10:31:00Z"
}
```

### 5. Reconcile Settlement

**Endpoint**: `POST /settlement/batches/:batchId/reconcile`

**Description**: Verify blockchain confirmation and update transaction statuses

**Response**:

```json
{
  "success": true,
  "data": {
    "batchId": "batch_xyz789",
    "reconciled": true
  },
  "timestamp": "2024-08-12T10:32:00Z"
}
```

## Health Check API

### 1. Health Status

**Endpoint**: `GET /health`

**Description**: Get system health status

**Response**:

```json
{
  "status": "healthy",
  "checks": {
    "database": true,
    "blockchain": true,
    "timestamp": "2024-08-12T10:30:00Z"
  }
}
```

### 2. Service Status

**Endpoint**: `GET /status`

**Description**: Get service information and uptime

**Response**:

```json
{
  "service": "banking-system",
  "version": "1.0.0",
  "environment": "production",
  "timestamp": "2024-08-12T10:30:00Z",
  "uptime": 86400
}
```

## Transaction Status Values

- `received`: Webhook received and queued for processing
- `pending`: Awaiting blockchain confirmation
- `confirmed`: Blockchain transaction confirmed
- `failed`: Transaction failed

## Settlement Batch Status Values

- `pending`: Batch created, awaiting processing
- `processing`: Batch submitted to blockchain
- `completed`: All transactions settled
- `failed`: Settlement failed

## Rate Limiting

Requests are limited to:

- **100 requests per minute** per API key
- Rate limit headers included in response:
  - `X-RateLimit-Limit`: 100
  - `X-RateLimit-Remaining`: Number of remaining requests
  - `X-RateLimit-Reset`: Unix timestamp of reset time

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_API_KEY | 401 | Invalid or missing API key |
| VALIDATION_ERROR | 400 | Request validation failed |
| DUPLICATE_EVENT | 409 | Event ID already processed |
| NOT_FOUND | 404 | Resource not found |
| INSUFFICIENT_BALANCE | 402 | Blockchain insufficient funds |
| TRANSACTION_FAILED | 500 | Blockchain transaction failed |
| INTERNAL_ERROR | 500 | Server error |

## Webhooks Best Practices

1. **Idempotency**: Always use unique `eventId` values
2. **Retry Logic**: Implement exponential backoff for failed requests
3. **Validation**: Verify blockchain addresses before submission
4. **Monitoring**: Track response times and error rates
5. **Logging**: Log all requests and responses for debugging

## Code Examples

### Python

```python
import requests
import json

api_key = "your-api-key"
headers = {
    "X-API-Key": api_key,
    "Content-Type": "application/json"
}

data = {
    "eventId": "evt_20240812_001",
    "timestamp": "2024-08-12T10:30:00Z",
    "type": "payment",
    "data": {
        "fromAddress": "0x742d35Cc6634C0532925a3b844Bc9e7595f42472",
        "toAddress": "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
        "amount": "100.00",
        "currency": "USDT"
    }
}

response = requests.post(
    "https://api.banking-system.local/api/v1/webhooks",
    headers=headers,
    json=data
)

print(response.json())
```

### Node.js

```javascript
const axios = require('axios');

const apiKey = 'your-api-key';
const headers = {
  'X-API-Key': apiKey,
  'Content-Type': 'application/json'
};

const data = {
  eventId: 'evt_20240812_001',
  timestamp: new Date().toISOString(),
  type: 'payment',
  data: {
    fromAddress: '0x742d35Cc6634C0532925a3b844Bc9e7595f42472',
    toAddress: '0x8ba1f109551bD432803012645Ac136ddd64DBA72',
    amount: '100.00',
    currency: 'USDT'
  }
};

axios.post('https://api.banking-system.local/api/v1/webhooks', data, { headers })
  .then(response => console.log(response.data))
  .catch(error => console.error(error.response.data));
```

## Support

For API support and issues:
- Email: api-support@banking-system.local
- Docs: https://docs.banking-system.local
- Status: https://status.banking-system.local
