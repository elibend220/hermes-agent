---
name: aurion-blockchain
description: Manage Kingdom of Joseph sovereign blockchain with Proof-of-Sovereignty consensus.
version: 1.0.0
author: Claude (Anthropic)
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [Banking, Blockchain, Sovereign, Settlement, Monetary, Digital-Asset]
    related_skills: [banking-system]
---

# Aurion Blockchain Skill

Operate the Aurion sovereign blockchain for the Kingdom of Joseph. Manage consensus, verify reserves, monitor node health, and process transactions with cryptographic certainty.

## When to Use

**System Administration:**
- "Start Aurion node with validator RCB"
- "Check blockchain consensus state"
- "Verify reserve backing ratio"
- "Monitor validator uptime"

**Operations:**
- "Submit tax settlement transaction"
- "Create reserve mint transaction"
- "Check gold vault attestation"
- "Get blockchain health status"

**Monitoring:**
- "Show reserve status"
- "List recent blocks"
- "Get transaction history for address"
- "Display consensus statistics"

## Prerequisites

- Aurion blockchain node running (`aurion-node` binary)
- Go 1.21+ installed
- Configuration file at `~/.aurion/config.json`
- Access to validator HSM or private keys

## How to Run

### 1. Start the Blockchain Node

```bash
# Start validator node (RCB)
hermes aurion start --validator=RCB --network=DEVNET

# Start archive node (for historical queries)
hermes aurion start --type=ARCHIVE --network=MAINNET
```

### 2. Check Consensus State

```bash
# Get current consensus info
hermes aurion consensus status

# Get next leader (3 rounds ahead)
hermes aurion consensus next-leader --rounds=3
```

### 3. Verify Reserves

```bash
# Show current reserve status
hermes aurion reserve status

# Get backing ratio (should be >= 1.0)
hermes aurion reserve ratio

# Verify gold vault attestation
hermes aurion reserve verify gold --block-height=<height>
```

### 4. Manage Transactions

```bash
# Submit transfer transaction
hermes aurion tx transfer \
  --from=0x<sender> \
  --to=0x<recipient> \
  --amount=1000000000 \
  --private=false

# Create reserve mint (Royal Treasury only)
hermes aurion tx mint \
  --vault=VAULT_GOLD_01 \
  --asset=GOLD_BULLION \
  --quantity=1000 \
  --unit=troy_oz \
  --purity=9995

# Submit sovereign tax
hermes aurion tx tax \
  --taxpayer=ENTITY_XYZ \
  --amount=50000000 \
  --type=INCOME
```

### 5. Monitor Health

```bash
# Real-time blockchain metrics
hermes aurion monitor

# Get node health
hermes aurion health

# List recent blocks
hermes aurion blocks --limit=10
```

## Procedure: Deploy and Initialize

1. **Build the node**
   ```bash
   cd /home/user/hermes-agent/aurion-blockchain
   go build -o aurion-node cmd/aurion-node/main.go
   ```

2. **Generate Protobuf code**
   ```bash
   protoc --go_out=. --go-grpc_out=. proto/*.proto
   ```

3. **Create configuration**
   ```bash
   mkdir -p ~/.aurion
   cat > ~/.aurion/config.json << 'EOF'
   {
     "version": "1.0.0",
     "network": "DEVNET",
     "node": {
       "id": "validator-rcb",
       "type": "VALIDATOR"
     },
     "blockchain": {
       "block_time_ms": 1300,
       "max_block_size": 4194304
     }
   }
   EOF
   ```

4. **Start the node**
   ```bash
   hermes aurion start --validator=RCB
   ```

5. **Verify health**
   ```bash
   hermes aurion health
   ```

**Done when:** Node is running, consensus shows height > 0, reserve backing ratio ≥ 1.0

## Procedure: Submit Settlement Transaction

1. **Verify reserve status**
   ```bash
   hermes aurion reserve status
   ```

2. **Create transaction**
   ```bash
   hermes aurion tx tax \
     --taxpayer=KINGDOM_TREASURY \
     --amount=100000000 \
     --type=CUSTOMS \
     --anonymous=false
   ```

3. **Wait for inclusion**
   ```bash
   hermes aurion wait-tx --tx-id=<tx_id> --timeout=30s
   ```

4. **Verify finalization**
   ```bash
   hermes aurion tx status --tx-id=<tx_id>
   ```

**Done when:** Transaction shows status=FINALIZED

## Procedure: Verify Physical Reserves

1. **Get gold attestation**
   ```bash
   hermes aurion reserve verify gold
   ```

2. **Check commodity telemetry**
   ```bash
   hermes aurion reserve telemetry
   ```

3. **Validate infrastructure yields**
   ```bash
   hermes aurion reserve infrastructure
   ```

4. **Confirm foreign clearing**
   ```bash
   hermes aurion reserve foreign
   ```

5. **Calculate backing ratio**
   ```bash
   hermes aurion reserve ratio
   ```

If ratio < 1.0, halt new minting immediately.

**Done when:** All reserves verified, ratio ≥ 1.0, status=HEALTHY

## Procedure: Monitor Consensus Health

1. **Check validator set**
   ```bash
   hermes aurion consensus validators
   ```

2. **Get current leader**
   ```bash
   hermes aurion consensus leader
   ```

3. **Check voting progress**
   ```bash
   hermes aurion consensus votes
   ```

4. **Monitor block production**
   ```bash
   hermes aurion monitor
   ```

**Done when:** All validators active, blocks produced every 1-1.5 seconds, no timeouts

## Quick Reference

### Consensus Parameters

| Parameter | Value |
|-----------|-------|
| Block Time | 1-1.5 seconds |
| Finality | 3 minutes (6 blocks) |
| Quorum Size | 3-of-4 validators |
| Timeout | 15 seconds per round |
| Validators | RCB, STV, JVN, ICA |

### Reserve Requirements

| Component | Allocation | Backing |
|-----------|-----------|---------|
| Gold Bullion | 40% | Physical bars (99.5% min) |
| Commodities | 30% | Energy, grain, minerals |
| Infrastructure | 20% | Ports, solar, properties |
| Foreign Clearing | 10% | Settlement balances |

### Transaction Fees

| Type | Fee |
|------|-----|
| Transfer | 1 mil (0.00000001 AUR) |
| Tax Payment | Included in amount |
| Reserve Mint | Free (protocol-incentivized) |
| Physical Redemption | Burn amount |

## Troubleshooting

### Node won't start
```bash
# Check configuration
cat ~/.aurion/config.json

# Verify Go version
go version  # Must be 1.21+

# Check logs
tail -f ~/.aurion/logs/aurion.log
```

### Consensus stuck
```bash
# Check validator health
hermes aurion consensus validators

# Check leader election
hermes aurion consensus leader --round=5

# Force new round
hermes aurion consensus skip-round
```

### Reserve verification fails
```bash
# Get detailed reserve status
hermes aurion reserve status --verbose

# Check gold vault
hermes aurion reserve verify gold --detailed

# Verify backing ratio calculation
hermes aurion reserve ratio --show-calculation
```

### Transaction not finalizing
```bash
# Check mempool
hermes aurion mempool

# Verify transaction
hermes aurion tx status --tx-id=<id>

# Check block inclusion
hermes aurion blocks --filter=tx_id:<id>
```

## Verification

**Health Check:**
```bash
hermes aurion health
```

Expected output:
```
✅ Node Status: HEALTHY
✅ Consensus Height: 1234
✅ Reserve Ratio: 1.15 (HEALTHY)
✅ Validators Active: 4/4
✅ Block Time: 1.3s
```

## Integration with Banking System

The Aurion blockchain serves as the settlement engine for the banking system:

```
Banking Webhook
  ↓
Settlement Batch
  ↓
hermes aurion tx tax  (Sovereign Tax Transaction)
  ↓
AURION BLOCKCHAIN (Finalized in 3 min)
  ↓
Reserve Verified + Telegram Notification
```

## Examples

### Example 1: Tax Collection Settlement

```bash
# 1. Create tax settlement transaction
hermes aurion tx tax \
  --taxpayer=KINGDOM_CUSTOMS \
  --amount=1000000000 \
  --type=CUSTOMS \
  --destination=ROYAL_TREASURY

# 2. Monitor inclusion in block
hermes aurion monitor

# 3. Wait for finalization (3 minutes)
sleep 180

# 4. Verify in ledger
hermes aurion tx status --tx-id=<tx_id>
```

### Example 2: Gold Reserve Deposit

```bash
# 1. Submit reserve mint
hermes aurion tx mint \
  --vault=VAULT_GOLD_SUBTERRANEAN_01 \
  --asset=GOLD_BULLION \
  --quantity=10000 \
  --unit=troy_oz \
  --purity=9995

# 2. Watch for minting
hermes aurion monitor

# 3. Verify AUR issued
hermes aurion reserve status

# 4. Check backing ratio
hermes aurion reserve ratio
```

### Example 3: Physical Redemption

```bash
# 1. Burn AUR for physical gold
hermes aurion tx redeem \
  --redeemer=0x<address> \
  --amount=100000000 \
  --asset=GOLD_BULLION \
  --delivery=VAULT_GOLD_SUBTERRANEAN_01

# 2. Wait for finalization
hermes aurion wait-tx --tx-id=<tx_id>

# 3. Coordinate with vault keeper
# Vault keeper authorizes withdrawal and updates reserve state
```

## Support & Documentation

- **Technical Spec**: `/home/user/hermes-agent/aurion-blockchain/AURION_BLOCKCHAIN.md`
- **Architecture**: `/home/user/hermes-agent/aurion-blockchain/ARCHITECTURE.md`
- **API Reference**: `/home/user/hermes-agent/aurion-blockchain/README.md`
- **GitHub Issues**: https://github.com/elibend220/hermes-agent/issues

---

**Aurion Blockchain v1.0** | Kingdom of Joseph | 🏛️ Sovereign Digital Ledger
