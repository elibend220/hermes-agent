# Deployment Guide

## Production Deployment

### Prerequisites

- Kubernetes cluster (v1.24+) or Docker Swarm
- PostgreSQL 16 (managed or self-hosted)
- Redis 7+ (managed or self-hosted)
- Apache Kafka 7.5+ (managed or self-hosted)
- SSL/TLS certificate
- Ethereum JSON-RPC provider (Alchemy, Infura)
- Private key for settlement wallet (secured in vault)

### Environment Setup

Create production `.env` file:

```bash
# Server
NODE_ENV=production
PORT=3000

# Database (use managed RDS/Cloud SQL)
DATABASE_URL=postgresql://user:pass@prod-db.example.com:5432/banking
DATABASE_POOL_SIZE=30

# Redis (use managed ElastiCache/Redis Cloud)
REDIS_URL=redis://prod-cache.example.com:6379

# Kafka (use managed MSK/Confluent)
KAFKA_BROKERS=kafka1.example.com:9092,kafka2.example.com:9092,kafka3.example.com:9092

# Blockchain
ETHEREUM_RPC_URL=https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY
ETHEREUM_CHAIN_ID=1
USDT_CONTRACT_ADDRESS=0xdAC17F958D2ee523a2206206994597C13D831ec7
SETTLEMENT_WALLET_ADDRESS=0x... # Your settlement wallet
SETTLEMENT_PRIVATE_KEY=0x... # Secured in KMS/Vault

# Security
API_KEY_SECRET=generate-random-256-bit-key
JWT_SECRET=generate-random-256-bit-key
ALLOWED_ORIGINS=https://yourdomain.com

# Monitoring
LOG_LEVEL=info
ENABLE_METRICS=true
METRICS_PORT=9090
```

### Docker Registry

Push to your registry:

```bash
# Build
docker build -t banking-system:1.0.0 .

# Tag
docker tag banking-system:1.0.0 your-registry/banking-system:1.0.0

# Push
docker push your-registry/banking-system:1.0.0
```

### Kubernetes Deployment

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: banking-system
  namespace: banking
spec:
  replicas: 3
  selector:
    matchLabels:
      app: banking-system
  template:
    metadata:
      labels:
        app: banking-system
    spec:
      containers:
      - name: app
        image: your-registry/banking-system:1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 3000
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: NODE_ENV
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: banking-secrets
              key: database-url
        - name: ETHEREUM_RPC_URL
          valueFrom:
            secretKeyRef:
              name: banking-secrets
              key: ethereum-rpc-url
        - name: SETTLEMENT_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: banking-secrets
              key: settlement-private-key
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 3000
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/status
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1024Mi
      serviceAccountName: banking-system

---
apiVersion: v1
kind: Service
metadata:
  name: banking-system
  namespace: banking
spec:
  type: LoadBalancer
  selector:
    app: banking-system
  ports:
  - port: 443
    targetPort: 3000
    protocol: TCP
    name: https
  - port: 9090
    targetPort: 9090
    protocol: TCP
    name: metrics

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: banking-system
  namespace: banking
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: banking-system
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

Deploy:

```bash
# Create namespace
kubectl create namespace banking

# Create secrets
kubectl create secret generic banking-secrets \
  --from-literal=database-url=$DATABASE_URL \
  --from-literal=ethereum-rpc-url=$ETHEREUM_RPC_URL \
  --from-literal=settlement-private-key=$SETTLEMENT_PRIVATE_KEY \
  -n banking

# Deploy
kubectl apply -f k8s/deployment.yaml
```

### AWS ECS Deployment

Create `task-definition.json`:

```json
{
  "family": "banking-system",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "banking-system",
      "image": "your-registry/banking-system:1.0.0",
      "portMappings": [
        {
          "containerPort": 3000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "NODE_ENV",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:banking-db-url"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/banking-system",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

Deploy:

```bash
aws ecs register-task-definition --cli-input-json file://task-definition.json
aws ecs create-service \
  --cluster production \
  --service-name banking-system \
  --task-definition banking-system:1 \
  --desired-count 3
```

### Database Migration

Run migrations in production:

```bash
npm run db:migrate

# Or with Docker:
docker run \
  -e DATABASE_URL=$DATABASE_URL \
  your-registry/banking-system:1.0.0 \
  npm run db:migrate
```

### SSL/TLS Configuration

Configure with Nginx reverse proxy:

```nginx
upstream banking_system {
  server 127.0.0.1:3000;
}

server {
  listen 443 ssl http2;
  server_name api.banking-system.com;

  ssl_certificate /etc/letsencrypt/live/banking-system.com/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/banking-system.com/privkey.pem;

  ssl_protocols TLSv1.3 TLSv1.2;
  ssl_ciphers HIGH:!aNULL:!MD5;
  ssl_prefer_server_ciphers on;

  client_max_body_size 10M;

  location / {
    proxy_pass http://banking_system;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # Rate limiting
    limit_req zone=api burst=10 nodelay;
  }

  location /api/v1/health {
    proxy_pass http://banking_system;
    access_log off;
  }
}

limit_req_zone $binary_remote_addr zone=api:10m rate=100r/m;
```

### Health Checks & Monitoring

Prometheus scrape config:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'banking-system'
    static_configs:
      - targets: ['localhost:9090']
```

Create Grafana dashboard for:
- Request rate and latency
- Transaction processing rate
- Blockchain confirmation time
- Database query performance
- Error rate and types

### Logging & Alerting

Send logs to centralized system:

```bash
# Send to ELK Stack
docker run -d \
  -e LOGZ_IO_TOKEN=your_token \
  logz/filebeat:latest

# Or use CloudWatch
aws logs create-log-group --log-group-name /banking/production
```

Set up alerts for:
- High error rate (> 1%)
- Slow transactions (> 30s)
- Blockchain confirmation delays
- Database connection pool exhaustion
- Memory/CPU usage > 80%

### Backup Strategy

PostgreSQL:

```bash
# Daily backup
pg_dump -U user -h prod-db.example.com banking | \
  gzip > /backups/banking_$(date +%Y%m%d).sql.gz

# Or use AWS RDS automated backups
aws rds modify-db-instance \
  --db-instance-identifier banking-prod \
  --backup-retention-period 30
```

### Disaster Recovery

1. **Database failover**: Multi-AZ RDS with automatic failover
2. **Cache failover**: Redis Sentinel or Cluster
3. **Message queue**: Kafka replication (min.insync.replicas=2)
4. **DNS failover**: Route53 health checks
5. **Backup restore**: Test weekly restore procedures

### Security Hardening

1. Network policies (firewall rules)
2. VPC isolation (private subnets for databases)
3. Secrets rotation (90-day key rotation)
4. WAF rules (rate limiting, SQL injection prevention)
5. Certificate pinning for blockchain calls
6. API key rotation

### Zero-Downtime Deployment

```bash
# 1. Build new image
docker build -t banking:v2 .

# 2. Update deployment (Kubernetes handles rolling update)
kubectl set image deployment/banking-system \
  banking-system=your-registry/banking:v2 \
  --record

# 3. Monitor rollout
kubectl rollout status deployment/banking-system

# 4. Rollback if needed
kubectl rollout undo deployment/banking-system
```

### Performance Tuning

1. **Database**: Enable pg_stat_statements, tune query plans
2. **Redis**: Increase maxmemory, use eviction policies
3. **Kafka**: Tune batch size, compression settings
4. **Node.js**: Enable clustering, adjust heap size
5. **Network**: Use CDN for static content

### Compliance & Audit

- Enable CloudTrail for API audit
- Enable VPC Flow Logs for network monitoring
- Export transaction logs daily
- Implement PCI-DSS requirements
- Regular penetration testing

## Troubleshooting

### Database Connection Issues

```bash
# Check connection
psql -h $DB_HOST -U $DB_USER -d banking -c "SELECT 1"

# Check connection pool
SELECT * FROM pg_stat_activity WHERE datname = 'banking';
```

### Blockchain Transaction Failures

```bash
# Check blockchain connection
curl https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}'

# Check wallet balance
# Use web3.js or etherscan
```

### Performance Issues

```bash
# Profile Node.js
node --prof app.js

# Analyze profile
node --prof-process isolate-*.log > profile.txt
```

## Support

- CloudFormation templates in `/infrastructure`
- Terraform configs in `/terraform`
- Ansible playbooks in `/ansible`
