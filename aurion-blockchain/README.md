# 🏛️ AURION BLOCKCHAIN

**Sovereign Digital Ledger with Cryptographically-Verified Asset Backing**

Kingdom of Joseph | Royal Central Bank | Supreme Treasury  
Currency Code: `AUR` (Aurion) | Protocol Version: 1.0

---

## 📖 Overview

**Aurion** is a sovereign blockchain implementing the Kingdom of Joseph's monetary framework. Every Aurion token is backed by verified physical reserves:

- **40%** Gold Bullion (subterranean vaults, multisig attestation)
- **30%** Strategic Commodities (energy kWh, grain, critical minerals)
- **20%** Infrastructure Assets (ports, solar grids, properties)
- **10%** Foreign Clearing (liquid settlement instruments)

### Core Properties

✅ **Non-Inflationary**: AUR minted only upon verified physical deposits  
✅ **Immediate Finality**: Deterministic 3-of-4 multisig consensus (no chain reorganizations)  
✅ **Mandatory Utility**: 100% of sovereign taxes, customs, & registration fees settle in AUR  
✅ **Hybrid Privacy**: ZK-SNARKs for domestic transactions, transparent auditing for international flows  
✅ **Institutional Consensus**: 4 Royal Validators (RCB, STV, JVN, ICA)  

---

## ⚙️ Technical Specification

### Consensus: Proof-of-Sovereignty (PoSov) + Byzantine Fault Tolerance

```
VALIDATOR QUORUM (4 Institutional Nodes):
├─ Royal Central Bank (RCB)           → Monetary policy
├─ Sovereign Treasury Vaults (STV)    → Reserve custody
├─ Judicial Verification Node (JVN)   → Legal audit
└─ Independent Custody Auditor (ICA)  → External verification

BLOCK FINALIZATION: 3-of-4 multisig required
TIMEOUT: 15 seconds (deterministic round-robin)
FINALITY: IMMEDIATE (no probabilistic confirmation)
```

### Block Parameters

| Parameter | Value | Purpose |
|-----------|-------|---------|
| Block Time | 1-1.5 seconds | Fast settlement for retail + gov |
| Max Block Size | 4 MB | ~3,000+ TPS capacity |
| Transaction Fee | 1 mil (0.00000001 AUR) | Minimal friction |
| Burn Rate | 0.025% per tax tx | Deflationary mechanism |
| Reserve Check | Every 100 blocks | Continuous asset validation |

### Transaction Types (5 Native)

1. **Transfer** - User-to-user payments (optional ZK privacy)
2. **Reserve Mint** - Physical deposit → AUR minting (RCB only)
3. **Sovereign Tax** - Mandatory civic payments (auto-burn 0.025%)
4. **Physical Redemption** - Burn AUR for gold/commodities
5. **Reserve Attestation** - Cryptographic proof updates

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Protobuf compiler (`protoc`)
- Ed25519 cryptography library

### Installation

```bash
# Clone the repository
git clone https://github.com/elibend220/hermes-agent.git
cd aurion-blockchain

# Install dependencies
go mod download

# Generate Protobuf Go code
protoc --go_out=. --go-grpc_out=. proto/*.proto

# Build the node
go build -o aurion-node cmd/aurion-node/main.go
```

### Configuration

Create `~/.aurion/config.json`:

```json
{
  "version": "1.0.0",
  "network": "DEVNET",
  "node": {
    "id": "validator-rcb",
    "type": "VALIDATOR",
    "home_dir": "/home/user/.aurion",
    "log_level": "info"
  },
  "blockchain": {
    "block_time_ms": 1300,
    "max_block_size": 4194304,
    "transaction_fee": 1
  },
  "reserve": {
    "gold_backing_pct": 40.0,
    "commodity_backing_pct": 30.0,
    "infrastructure_backing_pct": 20.0,
    "foreign_backing_pct": 10.0,
    "minimum_ratio": 1.0
  },
  "rpc": {
    "enabled": true,
    "host": "127.0.0.1",
    "port": 26657
  }
}
```

### Running a Node

```bash
# Start validator node
export AURION_NODE_ID=validator-rcb
export AURION_NETWORK=DEVNET
./aurion-node

# Or with custom config
export AURION_CONFIG=/path/to/config.json
./aurion-node
```

---

## 🔗 Integration with Banking System

The Aurion blockchain replaces Ethereum settlement in the Hermes-Agent banking system:

```
User Payment Webhook
    ↓
Banking System (Webhook Gateway)
    ↓
Settlement Batch Creation
    ↓
AURION BLOCKCHAIN ← Direct settlement on sovereign ledger
    ↓
Reserve Verification (every block)
    ↓
Telegram Notification (Hebrew)
```

### Bridge Adapter

```typescript
// banking-system/src/services/aurion-settlement.ts
export async function settleOnAurionBlockchain(batch: SettlementBatch) {
  const tx = {
    type: "COMMERCIAL_SETTLEMENT",
    amount: batch.totalAmount,
    txId: batch.id,
    timestamp: Date.now(),
  };
  
  return await aurionclient.submitTransaction(tx);
}
```

---

## 💾 Project Structure

```
aurion-blockchain/
├── AURION_BLOCKCHAIN.md          # Complete technical spec
├── ARCHITECTURE.md               # Design decisions
├── README.md                     # This file
├── go.mod, go.sum
│
├── proto/                        # Protobuf definitions
│   ├── block.proto              # Block structure
│   ├── transaction.proto        # 5 transaction types
│   ├── consensus.proto          # PoSov + BFT
│   └── ledger.proto             # Account state
│
├── cmd/
│   └── aurion-node/main.go      # Node entry point
│
├── pkg/
│   ├── block/
│   │   ├── block.go             # Validation engine
│   │   ├── validator.go         # Signature verification
│   │   └── merkle.go            # Merkle tree proofs
│   │
│   ├── consensus/
│   │   ├── posov.go             # Proof-of-Sovereignty
│   │   ├── bft.go               # Byzantine Fault Tolerance
│   │   └── validator_set.go     # Validator quorum
│   │
│   ├── crypto/
│   │   ├── ed25519.go           # EdDSA signing
│   │   ├── multisig.go          # Threshold signatures
│   │   └── zk.go                # ZK-SNARK interface
│   │
│   ├── reserve/
│   │   ├── gold_vault.go        # Multisig attestation
│   │   ├── telemetry.go         # Commodity IoT ingestion
│   │   └── infrastructure.go    # Yield tracking
│   │
│   ├── ledger/
│   │   ├── state.go             # Account state DB
│   │   └── history.go           # Historical queries
│   │
│   ├── api/
│   │   ├── rpc.go               # JSON-RPC server
│   │   └── websocket.go         # Real-time subscriptions
│   │
│   └── config/
│       └── config.go            # Configuration loading
│
└── tests/
    ├── block_test.go
    ├── consensus_test.go
    └── reserve_test.go
```

---

## 🔐 Security Model

### Cryptography

- **Signing**: Ed25519 (fast, deterministic)
- **Multisig**: 3-of-4 Threshold Signatures (Shamir Secret Sharing or BLS)
- **Privacy**: ZK-SNARKs (tax anonymity)
- **Merkle Trees**: SHA3-256 commitments

### Key Custody

- **Validator Keys**: Hardware Security Modules (HSM), air-gapped
- **Vault Signatures**: Multi-authority attestation (3-of-4 custodians)
- **Network**: Encrypted gRPC over mTLS (internal only)

### Reserve Verification

Every block validates:
1. Gold bar serial numbers (merkle root)
2. Commodity telemetry (IoT sensor data)
3. Infrastructure yields (port revenue, solar output)
4. Foreign clearing balances (international settlement)

Failure → automatic halt of new AUR minting

---

## 📊 Performance Targets

| Metric | Target |
|--------|--------|
| TPS (Transactions/sec) | 3,000+ |
| Block Time | 1-1.5s |
| Finality | 3 minutes (6 blocks) |
| Reserve Check Latency | 2 seconds |
| Network Sync Time | 15 minutes (full node) |
| ZK Proof Generation | 500ms |

---

## 🛠️ Development Roadmap

### Phase 1: Core (Weeks 1-2) ✅
- [x] Block structure + validation
- [x] Protobuf definitions
- [x] Ed25519 signing
- [x] Configuration system

### Phase 2: Consensus (Weeks 3-4)
- [ ] Proof-of-Sovereignty (PoSov)
- [ ] BFT consensus loop
- [ ] 3-of-4 multisig finalization
- [ ] Block proposal/vote mechanism

### Phase 3: Reserve Layer (Weeks 5-6)
- [ ] Gold vault verification
- [ ] Commodity telemetry ingestion
- [ ] Infrastructure yield tracking
- [ ] Proof-of-Reserve validator

### Phase 4: Privacy (Weeks 7-8)
- [ ] Circom circuit design
- [ ] SnarkJS integration
- [ ] Private transaction pool
- [ ] ZK proof verification

### Phase 5: APIs (Weeks 9-10)
- [ ] JSON-RPC server
- [ ] WebSocket subscriptions
- [ ] Banking system integration
- [ ] Monitoring dashboard

---

## 📡 Network Architecture

```
                    PUBLIC INTERNET
                          │
        ┌─────────────────┴──────────────────┐
        │                                    │
    SENTRY/EDGE NODES              SENTRY/EDGE NODES
    ├─ DDoS protection             ├─ JSON-RPC endpoint
    ├─ Rate limiting               ├─ WebSocket streams
    └─ Mempool                     └─ Tx broadcast
        │                              │
        └──────────────┬───────────────┘
                       │
        ┌──────────────┴──────────────┐
        │ VALIDATOR NETWORK (Private) │
        │                             │
        ├─ RCB (Royal Central Bank)  │
        ├─ STV (Treasury Vaults)      │
        ├─ JVN (Judicial Audit)      │
        └─ ICA (Independent Auditor) │
        
        (Encrypted gRPC, air-gapped)
```

---

## 🔗 API Reference

### JSON-RPC Methods (Planned)

```
# Get block
aurion_getBlock(height: uint64) -> Block

# Get account balance
aurion_getBalance(address: string) -> uint64

# Get reserve state
aurion_getReserveState() -> ReserveState

# Submit transaction
aurion_submitTransaction(tx: Transaction) -> TxReceipt

# Subscribe to block headers
aurion_subscribeBlocks() -> WebSocket stream
```

---

## 📚 Documentation

- [AURION_BLOCKCHAIN.md](./AURION_BLOCKCHAIN.md) - Complete technical specification
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Design decisions and rationale
- API Reference (forthcoming)
- Deployment Guide (forthcoming)
- Reserve Protocol Deep Dive (forthcoming)

---

## 🤝 Integration with Hermes-Agent

This blockchain serves the **Hermes-Agent Banking System**:

1. **Payment Webhooks** → Settlement on Aurion ledger
2. **Reserve Verification** → Continuous on-chain audits
3. **Telegram Notifications** → Real-time status updates
4. **Tax Collection** → Automatic sovereign settlements

See `/home/user/hermes-agent/banking-system` for integration code.

---

## 📝 License

MIT - See LICENSE file

## Version History

- **1.0.0** (2026-11-11): Initial technical specification
  - Proof-of-Sovereignty consensus
  - 4 institutional validators
  - Asset-backed parity mechanism
  - Immediate deterministic finality
  - 5 native transaction types
  - Zero-knowledge privacy layer

---

## 🚀 Next Steps

1. Compile Protobuf definitions
2. Implement consensus engine (PoSov + BFT)
3. Wire reserve verification layer
4. Build JSON-RPC API
5. Integrate with banking system

**Ready to build a sovereign monetary infrastructure? 🏛️**

```bash
./aurion-node --network=DEVNET --log-level=debug
```
