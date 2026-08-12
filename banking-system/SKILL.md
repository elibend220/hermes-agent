---
name: banking-system
description: Deploy, test, and manage a production-grade banking webhook gateway with blockchain settlement.
version: 1.0.0
author: Claude (Anthropic)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [Banking, Webhooks, Blockchain, Settlement, USDT, Ethereum]
    related_skills: []
---

# Banking System - Webhook Gateway & Settlement

A comprehensive skill for deploying and operating a production-ready banking system with webhook processing, transaction tracking, and blockchain-based USDT settlement.

## When to Use

**Testing & Development:**
- "Start the banking system locally"
- "Submit a test payment webhook"
- "Create a settlement batch"
- "Check transaction status"

**Deployment:**
- "Deploy banking system to Kubernetes"
- "Deploy to AWS ECS"
- "Configure production environment"
- "Set up monitoring and alerts"

**Operations:**
- "List recent transactions"
- "Check system health"
- "Process pending settlements"
- "Reconcile blockchain transactions"
- "View API documentation"

**Integration:**
- "Generate API key"
- "Test API endpoints"
- "Verify webhook signatures"
- "Monitor settlement confirmations"

## Quick Start

### 1. Local Development

Start the development server:

```bash
cd banking-system
npm install
node dev-server.js
```

Server will be available at `http://localhost:3000`

### 2. Test a Webhook

Submit a payment webhook:

```bash
curl -X POST http://localhost:3000/api/v1/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "eventId": "evt_test_001",
    "timestamp": "2024-08-12T10:30:00Z",
    "type": "payment",
    "data": {
      "fromAddress": "0x742d35Cc6634C0532925a3b844Bc9e7595f42472",
      "toAddress": "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
      "amount": "1000.50",
      "currency": "USDT"
    }
  }'
```

### 3. Check Transactions

List all transactions:

```bash
curl http://localhost:3000/api/v1/webhooks
```

### 4. Create Settlement Batch

Submit transactions for settlement:

```bash
curl -X POST http://localhost:3000/api/v1/settlement/batches \
  -H "Content-Type: application/json" \
  -d '{
    "transactionIds": ["txn_abc123", "txn_def456"]
  }'
```

## Procedures

### Procedure: Development & Testing

1. **Ensure Node.js installed**
   ```bash
   node --version  # v20+
   npm --version   # 10+
   ```

2. **Start development server**
   ```bash
   cd banking-system
   npm install
   node dev-server.js
   ```

3. **Verify health**
   ```bash
   curl http://localhost:3000/api/v1/health
   ```

4. **Test webhook endpoint**
   - Submit valid Ethereum addresses (0x...)
   - Validate USDT currency type
   - Check transaction ID in response
   - Verify timestamps are ISO 8601

5. **Test settlement flow**
   - Create transaction via webhook
   - Create batch with transaction IDs
   - List batches to verify creation
   - Retrieve specific batch details

**Done when:** Dev server is running, webhooks accept valid payloads, settlement batches are created successfully.

### Procedure: Production Deployment (Kubernetes)

1. **Prepare environment**
   - Set up PostgreSQL database
   - Configure Redis instance
   - Set up Kafka broker
   - Obtain Ethereum RPC endpoint
   - Generate settlement wallet and private key

2. **Create namespace**
   ```bash
   kubectl create namespace banking
   ```

3. **Create secrets**
   ```bash
   kubectl create secret generic banking-secrets \
     --from-literal=database-url="postgresql://..." \
     --from-literal=ethereum-rpc-url="https://..." \
     --from-literal=settlement-private-key="0x..." \
     -n banking
   ```

4. **Build and push Docker image**
   ```bash
   docker build -t your-registry/banking-system:1.0.0 .
   docker push your-registry/banking-system:1.0.0
   ```

5. **Deploy**
   ```bash
   kubectl apply -f k8s/deployment.yaml
   ```

6. **Verify deployment**
   ```bash
   kubectl get pods -n banking
   kubectl logs -n banking -l app=banking-system
   ```

7. **Set up ingress**
   - Configure TLS certificate
   - Set up Nginx reverse proxy
   - Enable rate limiting
   - Configure CORS

**Done when:** Pods are running, health checks pass, API is accessible via HTTPS.

### Procedure: Production Deployment (AWS ECS)

1. **Prepare AWS resources**
   - RDS PostgreSQL database
   - ElastiCache Redis
   - Managed Kafka (MSK)
   - ALB/NLB load balancer
   - CloudWatch logs

2. **Register task definition**
   ```bash
   aws ecs register-task-definition \
     --cli-input-json file://task-definition.json
   ```

3. **Create ECS service**
   ```bash
   aws ecs create-service \
     --cluster production \
     --service-name banking-system \
     --task-definition banking-system:1 \
     --desired-count 3
   ```

4. **Configure auto-scaling**
   ```bash
   aws application-autoscaling register-scalable-target \
     --service-namespace ecs \
     --resource-id service/production/banking-system \
     --scalable-dimension ecs:service:DesiredCount \
     --min-capacity 3 \
     --max-capacity 10
   ```

5. **Set up monitoring**
   - CloudWatch dashboards
   - SNS alerts for errors
   - X-Ray tracing
   - Application Load Balancer logs

**Done when:** ECS service has desired task count running, ALB health checks pass, metrics appear in CloudWatch.

### Procedure: API Testing & Validation

**Prerequisites:** Dev server running on localhost:3000 or production endpoint configured

**Webhook Tests:**

1. **Valid Payment Submission**
   ```bash
   curl -X POST http://localhost:3000/api/v1/webhooks \
     -H "Content-Type: application/json" \
     -d '{
       "eventId": "evt_valid_001",
       "timestamp": "'$(date -u +'%Y-%m-%dT%H:%M:%SZ')'",
       "type": "payment",
       "data": {
         "fromAddress": "0x742d35Cc6634C0532925a3b844Bc9e7595f42472",
         "toAddress": "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
         "amount": "100.00",
         "currency": "USDT"
       }
     }'
   ```

2. **Invalid Address (Should Fail)**
   ```bash
   curl -X POST http://localhost:3000/api/v1/webhooks \
     -H "Content-Type: application/json" \
     -d '{
       "eventId": "evt_invalid_001",
       "data": {
         "fromAddress": "invalid_address",
         "toAddress": "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
         "amount": "100",
         "currency": "USDT"
       }
     }'
   ```

3. **Missing Fields (Should Fail)**
   ```bash
   curl -X POST http://localhost:3000/api/v1/webhooks \
     -H "Content-Type: application/json" \
     -d '{"eventId": "evt_test"}'
   ```

**Settlement Tests:**

1. **Create Batch from Transactions**
   - Submit multiple webhooks
   - Collect transaction IDs
   - Create batch with IDs
   - Verify total amount

2. **List Transactions**
   - Verify pagination (limit, offset)
   - Check sorting by creation time
   - Validate status values

3. **List Batches**
   - Check batch numbering
   - Verify status progression
   - Confirm transaction counts

**Health & Status:**

1. **Health Check**
   ```bash
   curl http://localhost:3000/api/v1/health
   ```
   - Response: `{"status": "healthy", "checks": {...}}`

2. **Status**
   ```bash
   curl http://localhost:3000/api/v1/status
   ```
   - Verify service version, uptime, environment

**Done when:** All webhook validations pass, settlement batches created successfully, health checks return 200.

### Procedure: Monitoring & Observability

1. **Logs**
   - Structured JSON logging via Pino
   - Logs include: request ID, user ID, action, status, duration
   - Forward to ELK Stack or CloudWatch

2. **Metrics (Prometheus)**
   - Request rate and latency
   - Transaction processing rate
   - Settlement batch size and frequency
   - Blockchain confirmation times
   - Error rate by type

3. **Alerts**
   Set up alerts for:
   - Error rate > 1%
   - P99 latency > 30s
   - Database connection pool exhaustion
   - Blockchain RPC failures
   - Settlement confirmation delays > 5min

4. **Dashboards**
   Create Grafana dashboards for:
   - Transaction volume (daily/hourly)
   - Settlement success rate
   - Blockchain transaction costs
   - System resource utilization
   - API response times

**Done when:** Logs are centralized, metrics are scraped, dashboards show live data, alerts configured.

## Architecture Overview

```
Client Apps
    ↓
API Gateway (TLS, Rate Limiting)
    ↓
Webhook Handler (Validation, Dedup)
    ↓ (async)
Settlement Engine
    ↓
Blockchain Service (Ethers.js)
    ↓
Ethereum Network (Mainnet/Testnet)
```

## Key Components

### Services
- **Webhook Service**: Receives and processes payment webhooks
- **Settlement Service**: Batches transactions and submits to blockchain
- **Blockchain Service**: Handles USDT transfers via Ethers.js
- **Database**: Stores transactions, batches, audit logs

### Database Schema
- `transactions`: Payment records with status tracking
- `settlement_batches`: Batch aggregation and blockchain hashes
- `users`: API key management
- `webhook_events`: Event audit trail
- `audit_logs`: Compliance logging

### API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/webhooks` | Submit payment |
| GET | `/api/v1/webhooks` | List transactions |
| POST | `/api/v1/settlement/batches` | Create batch |
| GET | `/api/v1/settlement/batches` | List batches |
| POST | `/api/v1/settlement/batches/:id/process` | Process settlement |
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/status` | Service status |

## Security Considerations

1. **API Authentication**
   - X-API-Key header validation
   - Rate limiting (100 req/min per key)
   - TLS 1.3+ enforcement

2. **Webhook Security**
   - HMAC-SHA256 signature verification
   - Timestamp freshness validation (5min window)
   - Idempotency via event ID deduplication

3. **Blockchain Security**
   - Private key stored in HSM/KMS
   - Transaction signing verification
   - Contract address validation
   - Gas price monitoring

4. **Database Security**
   - Connection pooling with SSL
   - SQL injection prevention (parameterized queries)
   - Row-level access control
   - Encrypted sensitive fields

## Troubleshooting

### Server Won't Start
```bash
# Check port is available
lsof -i :3000

# Check logs
tail -f /tmp/banking-server.log

# Verify Node.js version
node --version  # Must be 20+
```

### Webhook Failures
- Validate Ethereum address format (0x + 40 hex chars)
- Check amount is numeric string
- Verify currency is USDT or USD
- Ensure unique eventId for idempotency

### Settlement Not Processing
- Verify Ethereum RPC endpoint is accessible
- Check settlement wallet has sufficient funds
- Review gas price settings
- Check blockchain confirmation times

### Database Connection Issues
```bash
# Test PostgreSQL
psql -h $DB_HOST -U $DB_USER -d banking

# Check connection pool
SELECT * FROM pg_stat_activity WHERE datname = 'banking';
```

## Support & Documentation

- **API Reference**: See [API.md](./API.md)
- **Deployment Guide**: See [DEPLOYMENT.md](./DEPLOYMENT.md)
- **README**: See [README.md](./README.md)
- **GitHub Issues**: https://github.com/elibend220/hermes-agent/issues
- **Slack**: #banking-system

## Development

### Build
```bash
npm run build
```

### Test
```bash
npm test
npm run test:coverage
```

### Lint
```bash
npm run lint
npm run format
```

### Database Migrations
```bash
npm run db:migrate
npm run db:seed
```

## Examples

### Python Client

```python
import requests
import json

api_key = "your-api-key"
base_url = "http://localhost:3000/api/v1"
headers = {"X-API-Key": api_key, "Content-Type": "application/json"}

# Submit webhook
data = {
    "eventId": "evt_py_001",
    "timestamp": "2024-08-12T10:30:00Z",
    "type": "payment",
    "data": {
        "fromAddress": "0x742d35Cc6634C0532925a3b844Bc9e7595f42472",
        "toAddress": "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
        "amount": "100",
        "currency": "USDT"
    }
}

response = requests.post(f"{base_url}/webhooks", json=data, headers=headers)
print(response.json())
```

### Node.js Client

```javascript
const axios = require('axios');

const apiKey = 'your-api-key';
const baseUrl = 'http://localhost:3000/api/v1';
const headers = {
  'X-API-Key': apiKey,
  'Content-Type': 'application/json'
};

const data = {
  eventId: 'evt_node_001',
  timestamp: new Date().toISOString(),
  type: 'payment',
  data: {
    fromAddress: '0x742d35Cc6634C0532925a3b844Bc9e7595f42472',
    toAddress: '0x8ba1f109551bD432803012645Ac136ddd64DBA72',
    amount: '100',
    currency: 'USDT'
  }
};

axios.post(`${baseUrl}/webhooks`, data, { headers })
  .then(r => console.log(r.data))
  .catch(e => console.error(e.response.data));
```

## License

MIT - See LICENSE file

## Version History

- **1.0.0** (2024-08-12): Initial release
  - Webhook gateway with validation
  - Settlement batch processing
  - Blockchain integration (USDT/Ethereum)
  - API documentation
  - Deployment guides (K8s, AWS ECS)
  - Development server for local testing
