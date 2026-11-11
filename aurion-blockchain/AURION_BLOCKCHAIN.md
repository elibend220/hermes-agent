# AURION BLOCKCHAIN - COMPLETE TECHNICAL SPECIFICATION v1.0

**Sovereign Consortium Distributed Ledger with Deterministic Instant Finality**

---

## 🏛️ CONSENSUS: Proof-of-Sovereignty (PoSov) + Byzantine Fault Tolerance

### Validator Architecture: 4 Institutional Nodes (3-of-4 Multisig Required)

```
AURION VALIDATOR QUORUM:
├─ Royal Central Bank (RCB)
├─ Sovereign Treasury Vaults (STV)
├─ Judicial Verification Node (JVN)
└─ Independent Custody Auditor (ICA)

Required Consensus: 3-of-4 signatures for block finality
Timeout: 15 seconds (deterministic, no network gossip delays)
Finality: IMMEDIATE (no chain reorg possible)
```

### Proof-of-Sovereignty (PoSov) Mechanism

**Block Validation Loop (1-1.5 second block time):**

```
1. LEADER ELECTION (Deterministic Round-Robin)
   Round 0: RCB proposes block
   Round 1: STV proposes block
   Round 2: JVN proposes block
   Round 3: ICA proposes block
   (Repeats cyclically)

2. RESERVE ATTESTATION CHECK (MUST PASS)
   ├─ Verify 3-of-4 vault HSM signatures
   ├─ Check gold bar serial numbers against merkle root
   ├─ Query commodity telemetry (energy kWh, grain silos)
   ├─ Verify infrastructure yield reports
   └─ If ANY reserve check fails → REJECT BLOCK

3. TRANSACTION VALIDATION
   ├─ Check nonces (prevent double-spend)
   ├─ Verify signatures (Ed25519)
   ├─ Validate balances
   └─ Process merkle roots

4. MULTISIG FINALIZATION
   ├─ Leader signs block (1/4)
   ├─ Request signatures from other 3 validators
   ├─ Require 3 total signatures (3/4 = super-majority)
   ├─ Emit FINALIZED event (irreversible)
   └─ Propagate to sentry/edge nodes

5. TIMEOUT HANDLING
   └─ If leader offline > 5 seconds, skip to next leader
```

**Advantages over generic PoW/PoS:**
- ✅ **No 51% attacks** (only authorized institutions)
- ✅ **Immediate finality** (no probabilistic confirmation)
- ✅ **No chain reorganizations** (BFT guarantees)
- ✅ **Reserve validation on every block**
- ✅ **Deterministic, auditable validator set**

---

## 📦 BLOCK STRUCTURE

### Block Header (256 bytes)

```protobuf
message BlockHeader {
  uint32 version = 1;                    // 4 bytes
  uint64 height = 2;                     // 8 bytes (block number)
  uint64 timestamp = 3;                  // 8 bytes (Unix seconds)
  bytes32 previous_hash = 4;             // 32 bytes (SHA3-256)
  bytes32 merkle_root = 5;               // 32 bytes (tx merkle tree)
  bytes32 reserve_root = 6;              // 32 bytes (reserve commitment)
  bytes32 state_root = 7;                // 32 bytes (account state)
  
  // MULTISIG SIGNATURES (3-of-4 quorum)
  repeated bytes signature = 8;          // N × 65 bytes (Ed25519)
  repeated string signer = 9;            // N validator IDs
  
  uint64 gas_used = 10;                  // 8 bytes
  uint64 gas_limit = 11;                 // 8 bytes
}
```

### Block Payload

```protobuf
message Block {
  BlockHeader header = 1;
  
  // TRANSACTIONS (variable, max 4 MB)
  repeated Transaction transactions = 2;
  
  // RESERVE ATTESTATIONS (per 100 blocks)
  repeated ReserveAttestation reserve_proofs = 3;
  
  // VALIDATOR ACTIVITY LOG
  message ValidatorActivity {
    string validator_id = 1;
    uint64 blocks_proposed = 2;
    uint64 blocks_attested = 3;
    uint64 last_active_height = 4;
  }
  repeated ValidatorActivity validator_activity = 4;
}
```

### Block Time & Capacity

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Block Time | 1-1.5s | Fast settlement for retail + gov |
| Max Block Size | 4 MB | ~3,000+ TPS @ 1.3s/block |
| Target Gas Limit | 80 million | Prevents memory DoS |
| Min Block Time | 1.0s | Prevent timestamp attacks |

---

## 💳 TRANSACTION TYPES (5 Native Types)

### 1. Transfer Tx (Regular payments)

```protobuf
message TransactionTransfer {
  string tx_id = 1;                  // Unique ID (SHA3(sender+nonce))
  string from = 2;                   // Sender address (256-bit)
  string to = 3;                     // Recipient address
  uint64 amount = 4;                 // Mils (1 AUR = 100,000,000 mils)
  uint64 fee = 5;                    // Transaction fee in mils (burned)
  uint32 nonce = 6;                  // Prevent replays
  uint64 timestamp = 7;              // Unix seconds
  bool is_private = 8;               // ZK-SNARK privacy flag
  bytes zk_proof = 9;                // If is_private=true
  bytes signature = 10;              // Ed25519(from, data)
}
```

### 2. Reserve Mint Tx (Only Royal Treasury)

```protobuf
message TransactionReserveMint {
  string tx_id = 1;
  string vault_id = 2;               // e.g., "VAULT_GOLD_SUBTERRANEAN_01"
  string asset_class = 3;            // "GOLD_BULLION" | "ENERGY_RESERVES" | "GRAIN" | "MINERALS"
  uint64 quantity = 4;               // Amount deposited (in native units)
  string unit = 5;                   // "troy_oz" | "kwh" | "metric_tons" | "kg"
  uint32 purity = 6;                 // Gold purity (e.g., 9995 = 99.95%)
  uint64 aurions_to_mint = 7;        // AUR issued (in mils)
  
  // 3-of-4 vault HSM multisig attestation
  repeated bytes vault_signature = 8;
  repeated string vault_signer_id = 9;
  
  bytes merkle_proof = 10;           // Proof of serial #s in commitment
  uint64 timestamp = 11;
}
```

### 3. Sovereign Tax Tx (Mandatory civic payments)

```protobuf
message TransactionSovereignTax {
  string tx_id = 1;
  string taxpayer_id = 2;            // Entity submitting tax
  uint64 tax_amount = 3;             // Mils owed
  string tax_type = 4;               // "INCOME" | "CUSTOMS" | "REGISTRATION" | "JUDICIAL"
  string destination_vault = 5;      // Royal Treasury vault ID
  
  // ZERO-KNOWLEDGE PRIVACY
  bool is_anonymous = 6;             // If true, amount/sender concealed
  bytes zk_proof = 7;                // zk-SNARK: amount valid, sender=taxpayer, dest=treasury
  
  uint64 burn_amount = 8;            // 0.025% auto-burn for stability
  uint64 timestamp = 9;
  bytes signature = 10;              // If not anonymous
}
```

### 4. Physical Redemption Tx (Burn for gold/resources)

```protobuf
message TransactionPhysicalRedemption {
  string tx_id = 1;
  string redeemer = 2;               // Who's requesting redemption
  uint64 aurions_to_burn = 3;        // Mils to burn
  string physical_asset = 4;         // "GOLD_BULLION" | "COMMODITY" | "ENERGY_CREDITS"
  string delivery_address = 5;       // Warehouse/vault for delivery
  
  bytes redemption_proof = 6;        // Attestation from physical custodian
  uint64 timestamp = 7;
  bytes signature = 8;
}
```

### 5. Reserve Attestation Tx (Proof updates, every 100 blocks)

```protobuf
message TransactionReserveAttestation {
  string tx_id = 1;
  uint64 block_height = 2;           // Attestation at which height
  
  message GoldAttestation {
    uint64 total_bars = 1;
    uint64 total_ounces = 2;
    bytes merkle_root = 3;           // SHA3(all_serial_numbers)
    repeated string vault_locations = 4;
    repeated bytes custodian_signature = 5;  // 3-of-4 multisig
  }
  GoldAttestation gold = 3;
  
  message CommodityAttestation {
    uint64 energy_kwh = 1;
    uint64 grain_metric_tons = 2;
    map<string, uint64> critical_minerals_kg = 3;
    bytes telemetry_hash = 4;        // SHA3(IoT sensor data)
    uint64 timestamp_telemetry = 5;
  }
  CommodityAttestation commodities = 4;
  
  message InfrastructureAttestation {
    map<string, uint64> port_monthly_yield = 1;    // (location -> AUR)
    map<string, uint64> solar_monthly_output = 2;  // (location -> mils)
    map<string, uint64> property_annual_revenue = 3;
  }
  InfrastructureAttestation infrastructure = 5;
  
  message ForeignClearingAttestation {
    uint64 total_liquid_reserves = 1;  // In AUR equivalents
    map<string, uint64> currency_breakdown = 2;  // (currency_code -> mils)
  }
  ForeignClearingAttestation foreign_clearing = 6;
  
  // Validator consensus
  repeated bytes validator_signature = 7;
  repeated string validator_id = 8;
}
```

---

## 🔐 CRYPTOGRAPHY & SIGNING

### Ed25519 (Primary Signature Algorithm)

```
Private Key: 32 bytes (secret seed)
Public Key: 32 bytes (derived)
Signature: 64 bytes
Total per signature: 96 bytes
```

### Multisig Threshold Signatures (M-of-N)

For reserve deposits and block finalization:
- **Block finalization**: 3-of-4 validators (RCB, STV, JVN, ICA)
- **Reserve attestation**: 3-of-4 vault custodians
- **Threshold scheme**: Shamir Secret Sharing (SSS) or BLS aggregation

### Zero-Knowledge Proofs (ZK-SNARKs for Privacy)

For anonymous sovereign tax transactions:

```
Circuit: TaxAnonymity
  Input (private):
    - taxpayer_id
    - tax_amount
    - nonce
  
  Input (public):
    - commitment_hash
    - treasury_destination
  
  Constraint:
    - SHA3(taxpayer_id || tax_amount || nonce) == commitment_hash
    - treasury_destination == ROYAL_TREASURY_VAULT
    - tax_amount >= min_tax_threshold
  
  Proof generation: ~500ms (SnarkJS + circom)
  Proof size: ~288 bytes
```

### Merkle Trees (State & Reserve Commitments)

```
Account State Merkle:
├─ Hash all 2^N account balances
├─ Proof of inclusion: O(log N) branches
└─ Used for light client verification

Reserve Merkle:
├─ Hash all gold bar serial numbers
├─ Hash all commodity inventory records
├─ Hash all infrastructure yields
└─ Root committed in every block
```

---

## 💾 LEDGER STATE MANAGEMENT

### Account State

```protobuf
message AccountState {
  string address = 1;                // 256-bit account ID
  uint64 balance = 2;                // Mils (AUR × 100,000,000)
  uint32 nonce = 3;                  // Tx count (prevent replay)
  
  enum AccountType {
    USER = 0;
    VAULT = 1;
    VALIDATOR = 2;
    TREASURY = 3;
  }
  AccountType account_type = 4;
  
  enum PrivacyLevel {
    TRANSPARENT = 0;
    HYBRID = 1;
    PRIVATE = 2;
  }
  PrivacyLevel privacy = 5;
  
  uint64 created_at = 6;
  uint64 last_activity = 7;
  map<string, bytes> metadata = 8;   // Custom fields per account type
}
```

### Reserve State (Immutable Reference)

```protobuf
message ReserveState {
  uint64 updated_at = 1;             // Last attestation timestamp
  
  message GoldReserve {
    uint64 total_bars = 1;
    uint64 total_ounces = 2;         // troy ounces
    bytes merkle_root = 3;           // All serial numbers committed
    repeated string vault_locations = 4;
    uint32 min_purity_percentage = 5; // 9950 = 99.50%
  }
  GoldReserve gold = 2;
  
  message CommodityReserve {
    uint64 energy_kwh = 3;           // Strategic energy storage
    uint64 grain_metric_tons = 4;    // Grain silos
    map<string, uint64> minerals_kg = 5;  // Critical minerals
  }
  CommodityReserve commodities = 3;
  
  message InfrastructureReserve {
    repeated string port_ids = 4;
    repeated string solar_grid_ids = 5;
    repeated string property_ids = 6;
    uint64 total_monthly_yield_aur = 7;  // Mils
  }
  InfrastructureReserve infrastructure = 4;
  
  uint64 foreign_clearing_mils = 5;
  uint64 total_aurions_issued = 6;
  uint64 max_aurions_allowed = 7;    // Cap based on backing formula
  
  enum ReserveHealthStatus {
    HEALTHY = 0;
    WARNING = 1;
    CRITICAL = 2;
  }
  ReserveHealthStatus health_status = 8;
}
```

### State Persistence Strategy

**Archive Nodes** (Judicial/audit use):
- Keep full history from genesis
- No pruning
- ~500 GB per year

**Validator Nodes** (Fast operation):
- State Trie: Current balances only
- Recent history: Last 100,000 blocks (~1.15 days @ 1.3s/block)
- Prune older than 100,000 blocks
- ~50 GB typical size

---

## 🌍 NETWORK TOPOLOGY

### Node Types

```
INTERNET PUBLIC
    │
    └─→ SENTRY/EDGE NODES (2-3)
         ├─ Rate limiting (100 req/s per IP)
         ├─ DDoS mitigation
         ├─ Transaction mempool
         └─ JSON-RPC endpoint
    
    └─→ INTERNAL VALIDATOR NETWORK (4 nodes)
         ├─ RCB Validator (consensus participant)
         ├─ STV Validator (reserve attestation)
         ├─ JVN Validator (legal audit)
         └─ ICA Validator (independent auditor)
```

### Validator Network Isolation

```
PoSov Validators: Air-gapped internal network
├─ Private key custody: Hardware Security Modules (HSM)
├─ Communication: Encrypted gRPC over mTLS
├─ Network: VPN/private WAN only (no public internet)
└─ Admin access: Quorum approval required (3-of-4)
```

---

## ⚙️ SMART CONTRACTS (Execution Layer)

### Technology: WebAssembly (WASM) + Rust Runtime

**Why WASM, not Solidity/EVM:**
- ✅ Formal verifiability (Rust's type system)
- ✅ No reentrancy bugs
- ✅ Deterministic execution
- ✅ Native performance
- ✅ Memory safety guarantees

### Whitelisted Contract Deployment

Smart contracts are NOT user-deployed. Only authorized modules:

1. **Land Registry Contract** (Property titles)
2. **Sovereign Bonds Contract** (Debt instruments)
3. **Export License Contract** (Trade authorization)
4. **Infrastructure Dividend Contract** (Yield distribution)

Deployment requires 3-of-4 validator approval.

---

## 📊 TRANSACTION FEES & ECONOMIC MODEL

### Fee Structure

```
Standard Transfer: 1 mil (0.00000001 AUR)
Reserve Mint: FREE (protocol-incentivized)
Tax Payment: Embedded in TxType_SovereignTax
Burn Component: 0.025% automatic burn per tax tx

Gas Model (optional):
├─ Simple Transfer: 21,000 gas
├─ Reserve Mint: 150,000 gas
├─ Tax Payment: 50,000 gas
└─ Smart Contract: Measured WASM ops
```

### Deflationary Mechanism

Every tax payment automatically burns 0.025% of the amount:

```
Tax Amount: 1,000,000 mils (0.01 AUR)
Burn Amount: 250 mils
Royal Treasury Receives: 999,750 mils
```

Over time: **deflation** → **increasing purchasing power of AUR**

---

## 📡 CROSS-BORDER INTEROPERABILITY

### ISO 20022 Gateway

For international settlements (remittances, trade settlements):

```
INCOMING SWIFT MESSAGE (pacs.008 format):
├─ Parse MT103 payment instruction
├─ Extract amount, sender bank, recipient
├─ Create atomic Aurion transaction
├─ Sign with gateway validator key
└─ Settle on Aurion blockchain

OUTGOING SETTLEMENT:
├─ Query Aurion ledger
├─ Generate ISO 20022 response
├─ Route through SWIFT network
└─ Notify corresponding bank
```

---

## 🚀 IMPLEMENTATION ROADMAP

### Phase 1: Core Infrastructure (Weeks 1-2)

- [ ] Protobuf definitions (all messages)
- [ ] Go node implementation (consensus engine)
- [ ] Ed25519 signing + multisig verification
- [ ] Merkle tree implementation
- [ ] Block validation logic

### Phase 2: Reserve Layer (Weeks 3-4)

- [ ] Vault HSM integration (gold attestation)
- [ ] Commodity telemetry ingestion (IoT)
- [ ] Infrastructure yield tracking
- [ ] Foreign clearing reconciliation
- [ ] Proof-of-Reserve validator

### Phase 3: Privacy (Weeks 5-6)

- [ ] Circom circuit design (tax anonymity)
- [ ] SnarkJS integration
- [ ] ZK proof generation/verification
- [ ] Private transaction pool

### Phase 4: APIs & Monitoring (Weeks 7-8)

- [ ] JSON-RPC server
- [ ] WebSocket subscriptions
- [ ] Monitoring dashboard
- [ ] Log aggregation
- [ ] Alerting

### Phase 5: Integration (Weeks 9-10)

- [ ] Banking system bridge
- [ ] Telegram notifications
- [ ] Aurion CLI wallet
- [ ] Light client SDK

---

## 📁 PROJECT STRUCTURE

```
aurion-blockchain/
├── AURION_BLOCKCHAIN.md (this file)
├── ARCHITECTURE.md
├── go.mod
├── go.sum
├── main.go
│
├── proto/
│   ├── block.proto
│   ├── transaction.proto
│   ├── ledger.proto
│   ├── consensus.proto
│   └── reserve.proto
│
├── pkg/
│   ├── block/
│   │   ├── block.go
│   │   ├── validator.go
│   │   └── merkle.go
│   │
│   ├── consensus/
│   │   ├── posov.go            # Proof-of-Sovereignty
│   │   ├── bft.go              # Byzantine Fault Tolerance
│   │   ├── finality.go         # 3-of-4 finalization
│   │   └── validator_set.go
│   │
│   ├── crypto/
│   │   ├── ed25519.go
│   │   ├── multisig.go
│   │   ├── merkle.go
│   │   └── zk.go               # ZK-SNARK interface
│   │
│   ├── reserve/
│   │   ├── gold_vault.go       # Multisig attestation
│   │   ├── telemetry.go        # Commodity IoT
│   │   ├── infrastructure.go   # Yield tracking
│   │   └── foreign_clearing.go # Settlement
│   │
│   ├── ledger/
│   │   ├── state.go            # Account state DB
│   │   ├── reserve.go          # Reserve commitment
│   │   └── history.go          # Historical queries
│   │
│   ├── api/
│   │   ├── rpc.go              # JSON-RPC server
│   │   ├── websocket.go        # Real-time feeds
│   │   └── middleware.go       # Auth, rate limit
│   │
│   └── config/
│       └── config.go
│
├── tests/
│   ├── block_test.go
│   ├── consensus_test.go
│   ├── reserve_test.go
│   └── integration_test.go
│
└── docs/
    ├── API.md
    ├── DEPLOYMENT.md
    └── RESERVE_PROTOCOL.md
```

---

## 🔧 NEXT IMMEDIATE STEPS

1. **Generate Protobuf definitions** (proto/ directory)
2. **Implement Genesis Block engine** in Go
3. **Build block validator** (signature checking, reserve verification)
4. **Create consensus loop** (PoSov + BFT)
5. **Wire Merkle trees** (state + reserve)

Ready to code? 🚀
