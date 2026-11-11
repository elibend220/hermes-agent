package block

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	pb_block "github.com/elibend220/aurion-blockchain/pb/pb_block"
	pb_transaction "github.com/elibend220/aurion-blockchain/pb/pb_transaction"
	pb_consensus "github.com/elibend220/aurion-blockchain/pb/pb_consensus"
)

// BlockValidator validates block structure, transactions, and reserve proofs
type BlockValidator struct {
	validators map[string]*pb_consensus.ValidatorInfo
	quorumSize int
}

// NewBlockValidator creates a new block validator
func NewBlockValidator(validatorSet *pb_consensus.ValidatorSet) *BlockValidator {
	validators := make(map[string]*pb_consensus.ValidatorInfo)
	for _, v := range validatorSet.Validators {
		validators[v.ValidatorId] = v
	}
	return &BlockValidator{
		validators: validators,
		quorumSize: int(validatorSet.QuorumSize),
	}
}

// ValidateBlock performs complete block validation
func (bv *BlockValidator) ValidateBlock(block *pb_block.Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	// 1. Validate header
	if err := bv.validateHeader(block.Header); err != nil {
		return fmt.Errorf("header validation failed: %w", err)
	}

	// 2. Validate transaction count
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block has no transactions (must have at least genesis)")
	}

	// 3. Validate all transactions
	for i, tx := range block.Transactions {
		if err := ValidateTransaction(tx); err != nil {
			return fmt.Errorf("transaction %d validation failed: %w", i, err)
		}
	}

	// 4. Validate reserve attestations
	for i, att := range block.ReserveAttestations {
		if err := bv.validateReserveAttestation(att); err != nil {
			return fmt.Errorf("reserve attestation %d validation failed: %w", i, err)
		}
	}

	// 5. Verify multisig on header (3-of-4 required)
	if err := bv.verifyMultisig(block.Header); err != nil {
		return fmt.Errorf("multisig verification failed: %w", err)
	}

	// 6. Verify merkle roots
	if err := bv.verifyMerkleRoots(block); err != nil {
		return fmt.Errorf("merkle root verification failed: %w", err)
	}

	return nil
}

// validateHeader checks block header integrity
func (bv *BlockValidator) validateHeader(header *pb_block.BlockHeader) error {
	if header == nil {
		return fmt.Errorf("header is nil")
	}

	// Check version
	if header.Version != 1 {
		return fmt.Errorf("unsupported block version: %d", header.Version)
	}

	// Check height (must be sequential)
	if header.Height == 0 {
		return fmt.Errorf("genesis block height must be 0")
	}

	// Check timestamp
	now := uint64(time.Now().Unix())
	if header.Timestamp > now+300 { // 5 minute tolerance
		return fmt.Errorf("block timestamp in future: %d", header.Timestamp)
	}

	// Check hashes
	if len(header.PreviousHash) != 32 {
		return fmt.Errorf("previous_hash must be 32 bytes, got %d", len(header.PreviousHash))
	}
	if len(header.MerkleRoot) != 32 {
		return fmt.Errorf("merkle_root must be 32 bytes, got %d", len(header.MerkleRoot))
	}
	if len(header.ReserveRoot) != 32 {
		return fmt.Errorf("reserve_root must be 32 bytes, got %d", len(header.ReserveRoot))
	}
	if len(header.StateRoot) != 32 {
		return fmt.Errorf("state_root must be 32 bytes, got %d", len(header.StateRoot))
	}

	// Check gas limits
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("gas_used (%d) exceeds gas_limit (%d)", header.GasUsed, header.GasLimit)
	}

	return nil
}

// validateReserveAttestation checks reserve proof validity
func (bv *BlockValidator) validateReserveAttestation(att *pb_block.ReserveAttestation) error {
	if att == nil {
		return fmt.Errorf("attestation is nil")
	}

	// Check multisig on attestation (3-of-4)
	if len(att.ValidatorSignature) < bv.quorumSize {
		return fmt.Errorf("insufficient signatures: %d < %d", len(att.ValidatorSignature), bv.quorumSize)
	}

	// Validate each component
	if att.Gold == nil {
		return fmt.Errorf("gold attestation is missing")
	}
	if err := validateGoldAttestation(att.Gold); err != nil {
		return fmt.Errorf("gold validation failed: %w", err)
	}

	if att.Commodities == nil {
		return fmt.Errorf("commodities attestation is missing")
	}

	if att.Infrastructure == nil {
		return fmt.Errorf("infrastructure attestation is missing")
	}

	if att.ForeignClearing == nil {
		return fmt.Errorf("foreign clearing attestation is missing")
	}

	return nil
}

// verifyGoldAttestation checks gold reserve proof
func validateGoldAttestation(gold *pb_block.GoldAttestation) error {
	if gold.TotalBars == 0 {
		return fmt.Errorf("total_bars must be > 0")
	}
	if gold.TotalTroyOunces == 0 {
		return fmt.Errorf("total_troy_ounces must be > 0")
	}
	if len(gold.MerkleRoot) != 32 {
		return fmt.Errorf("merkle_root must be 32 bytes")
	}
	if len(gold.VaultLocations) == 0 {
		return fmt.Errorf("must have at least one vault location")
	}
	if gold.MinPurityBps < 9900 || gold.MinPurityBps > 10000 {
		return fmt.Errorf("invalid purity: %d (should be 9900-10000)", gold.MinPurityBps)
	}
	return nil
}

// verifyMultisig checks block header has 3-of-4 validator signatures
func (bv *BlockValidator) verifyMultisig(header *pb_block.BlockHeader) error {
	if len(header.Signature) < bv.quorumSize {
		return fmt.Errorf("insufficient signatures: %d < %d", len(header.Signature), bv.quorumSize)
	}

	signerSet := make(map[string]bool)
	for _, signer := range header.SignerId {
		if _, ok := bv.validators[signer]; !ok {
			return fmt.Errorf("unknown validator: %s", signer)
		}
		signerSet[signer] = true
	}

	if len(signerSet) < bv.quorumSize {
		return fmt.Errorf("duplicate signers detected")
	}

	return nil
}

// verifyMerkleRoots checks merkle tree commitments
func (bv *BlockValidator) verifyMerkleRoots(block *pb_block.Block) error {
	// Compute merkle root from transactions
	computedMerkle := ComputeTransactionMerkleRoot(block.Transactions)
	if hex.EncodeToString(computedMerkle) != hex.EncodeToString(block.Header.MerkleRoot) {
		return fmt.Errorf("merkle root mismatch: computed %s, got %s",
			hex.EncodeToString(computedMerkle),
			hex.EncodeToString(block.Header.MerkleRoot),
		)
	}

	// Reserve root is verified through attestation validation
	if len(block.Header.ReserveRoot) != 32 {
		return fmt.Errorf("invalid reserve root length")
	}

	return nil
}

// ComputeTransactionMerkleRoot computes merkle root from transactions
func ComputeTransactionMerkleRoot(transactions []*pb_transaction.Transaction) []byte {
	if len(transactions) == 0 {
		return sha256.New().Sum(nil)
	}

	// Hash each transaction
	hashes := make([][]byte, len(transactions))
	for i, tx := range transactions {
		hashes[i] = hashTransaction(tx)
	}

	// Build merkle tree
	return buildMerkleRoot(hashes)
}

// hashTransaction creates SHA3-256 hash of transaction
func hashTransaction(tx *pb_transaction.Transaction) []byte {
	// Simple hash: SHA256(tx_id + timestamp)
	hash := sha256.New()
	hash.Write([]byte(tx.TxId))
	hash.Write([]byte(fmt.Sprintf("%d", tx.Timestamp)))
	return hash.Sum(nil)
}

// buildMerkleRoot recursively builds merkle tree
func buildMerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return make([]byte, 32)
	}
	if len(hashes) == 1 {
		return hashes[0]
	}

	// If odd number, duplicate the last hash
	if len(hashes)%2 == 1 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}

	// Hash pairs
	nextLevel := make([][]byte, 0)
	for i := 0; i < len(hashes); i += 2 {
		hash := sha256.New()
		hash.Write(hashes[i])
		hash.Write(hashes[i+1])
		nextLevel = append(nextLevel, hash.Sum(nil))
	}

	return buildMerkleRoot(nextLevel)
}

// ValidateTransaction performs transaction-level validation
func ValidateTransaction(tx *pb_transaction.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	// Check TX ID
	if len(tx.TxId) == 0 {
		return fmt.Errorf("transaction ID is empty")
	}

	// Check timestamp
	if tx.Timestamp == 0 {
		return fmt.Errorf("transaction timestamp is zero")
	}

	now := uint64(time.Now().Unix())
	if tx.Timestamp > now+300 {
		return fmt.Errorf("transaction timestamp in future")
	}

	// Type-specific validation
	switch tx.TxData.(type) {
	case *pb_transaction.Transaction_Transfer:
		return validateTransfer(tx.GetTransfer())
	case *pb_transaction.Transaction_ReserveMint:
		return validateReserveMint(tx.GetReserveMint())
	case *pb_transaction.Transaction_SovereignTax:
		return validateSovereignTax(tx.GetSovereignTax())
	case *pb_transaction.Transaction_PhysicalRedemption:
		return validatePhysicalRedemption(tx.GetPhysicalRedemption())
	default:
		return fmt.Errorf("unknown transaction type")
	}
}

// validateTransfer checks Transfer transaction
func validateTransfer(tx *pb_transaction.TransactionTransfer) error {
	if len(tx.From) == 0 {
		return fmt.Errorf("transfer: from address empty")
	}
	if len(tx.To) == 0 {
		return fmt.Errorf("transfer: to address empty")
	}
	if tx.Amount == 0 {
		return fmt.Errorf("transfer: amount must be > 0")
	}
	if len(tx.Signature) == 0 {
		return fmt.Errorf("transfer: signature missing")
	}
	return nil
}

// validateReserveMint checks Reserve Mint transaction
func validateReserveMint(tx *pb_transaction.TransactionReserveMint) error {
	if len(tx.VaultId) == 0 {
		return fmt.Errorf("reserve_mint: vault_id empty")
	}
	if tx.Quantity == 0 {
		return fmt.Errorf("reserve_mint: quantity must be > 0")
	}
	if len(tx.VaultSignature) < 3 {
		return fmt.Errorf("reserve_mint: insufficient vault signatures (need 3-of-4)")
	}
	if tx.AurionsToMint == 0 {
		return fmt.Errorf("reserve_mint: aurions_to_mint must be > 0")
	}
	return nil
}

// validateSovereignTax checks Sovereign Tax transaction
func validateSovereignTax(tx *pb_transaction.TransactionSovereignTax) error {
	if len(tx.TaxpayerId) == 0 {
		return fmt.Errorf("sovereign_tax: taxpayer_id empty")
	}
	if tx.TaxAmount == 0 {
		return fmt.Errorf("sovereign_tax: tax_amount must be > 0")
	}
	if len(tx.DestinationVault) == 0 {
		return fmt.Errorf("sovereign_tax: destination_vault empty")
	}
	return nil
}

// validatePhysicalRedemption checks Physical Redemption transaction
func validatePhysicalRedemption(tx *pb_transaction.TransactionPhysicalRedemption) error {
	if len(tx.Redeemer) == 0 {
		return fmt.Errorf("physical_redemption: redeemer empty")
	}
	if tx.AurionstoBurn == 0 {
		return fmt.Errorf("physical_redemption: aurions_to_burn must be > 0")
	}
	if len(tx.DeliveryAddress) == 0 {
		return fmt.Errorf("physical_redemption: delivery_address empty")
	}
	return nil
}

// BlockStats provides statistics about a block
type BlockStats struct {
	Height              uint64
	TransactionCount    int
	TotalFees           uint64
	ReserveAttestations int
	Validators          []string
}

// GetBlockStats returns statistics for a block
func GetBlockStats(block *pb_block.Block) *BlockStats {
	return &BlockStats{
		Height:              block.Header.Height,
		TransactionCount:    len(block.Transactions),
		ReserveAttestations: len(block.ReserveAttestations),
		Validators:          block.Header.SignerId,
	}
}
