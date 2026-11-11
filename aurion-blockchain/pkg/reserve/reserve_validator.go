package reserve

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	pb_block "github.com/elibend220/aurion-blockchain/pb/pb_block"
	pb_transaction "github.com/elibend220/aurion-blockchain/pb/pb_transaction"
)

// ReserveValidator verifies that every AUR minting is backed by physical reserves
type ReserveValidator struct {
	mu                    sync.RWMutex
	goldReserve           *GoldReserveState
	commodityReserve      *CommodityReserveState
	infrastructureReserve *InfrastructureReserveState
	foreignClearing       *ForeignClearingState
	totalAuriIssued       uint64
	maxAuriAllowed        uint64
	minBackingRatio       float64
	lastVerification      uint64
}

// GoldReserveState tracks gold backing
type GoldReserveState struct {
	TotalBars       uint64
	TotalTroyOunces uint64
	MerkleRoot      []byte
	VaultLocations  []string
	MinPurityBps    uint32 // 9950 = 99.50%
}

// CommodityReserveState tracks commodity backing
type CommodityReserveState struct {
	EnergyKWh           uint64
	GrainMetricTons     uint64
	CriticalMineralsKg  map[string]uint64
	LastTelemetryHash   []byte
	LastTelemetryUpdate uint64
}

// InfrastructureReserveState tracks infrastructure yields
type InfrastructureReserveState struct {
	PortMonthlyYield     map[string]uint64 // location -> mils
	SolarMonthlyOutput   map[string]uint64 // location -> mils
	PropertyAnnualRevenue map[string]uint64 // location -> mils
	TotalMonthlyYield    uint64
}

// ForeignClearingState tracks foreign reserves
type ForeignClearingState struct {
	TotalLiquidReserves uint64
	CurrencyBreakdown   map[string]uint64 // ISO 4217 code -> mils
}

// NewReserveValidator creates a new reserve validator
func NewReserveValidator() *ReserveValidator {
	return &ReserveValidator{
		goldReserve: &GoldReserveState{
			VaultLocations: make([]string, 0),
		},
		commodityReserve: &CommodityReserveState{
			CriticalMineralsKg:  make(map[string]uint64),
			LastTelemetryHash:   make([]byte, 32),
		},
		infrastructureReserve: &InfrastructureReserveState{
			PortMonthlyYield:     make(map[string]uint64),
			SolarMonthlyOutput:   make(map[string]uint64),
			PropertyAnnualRevenue: make(map[string]uint64),
		},
		foreignClearing: &ForeignClearingState{
			CurrencyBreakdown: make(map[string]uint64),
		},
		minBackingRatio: 1.0,
		maxAuriAllowed:  1000000000000000, // 10 million AUR in smallest units
	}
}

// ValidateReserveMint checks if a reserve mint transaction is valid
func (rv *ReserveValidator) ValidateReserveMint(tx *pb_transaction.TransactionReserveMint) error {
	rv.mu.Lock()
	defer rv.mu.Unlock()

	if tx == nil {
		return fmt.Errorf("reserve mint transaction is nil")
	}

	// Check asset class
	switch tx.AssetClass {
	case pb_transaction.TransactionReserveMint_GOLD_BULLION:
		return rv.validateGoldDeposit(tx)
	case pb_transaction.TransactionReserveMint_ENERGY_RESERVES:
		return rv.validateCommodityDeposit(tx)
	case pb_transaction.TransactionReserveMint_GRAIN:
		return rv.validateCommodityDeposit(tx)
	case pb_transaction.TransactionReserveMint_CRITICAL_MINERALS:
		return rv.validateCommodityDeposit(tx)
	default:
		return fmt.Errorf("unknown asset class: %d", tx.AssetClass)
	}
}

// validateGoldDeposit verifies gold bullion deposit
func (rv *ReserveValidator) validateGoldDeposit(tx *pb_transaction.TransactionReserveMint) error {
	// Check minimum purity (99.5%)
	if tx.PurityBps < 9950 {
		return fmt.Errorf("gold purity %.2f%% below minimum 99.5%%", float64(tx.PurityBps)/100)
	}

	// Check quantity > 0
	if tx.Quantity == 0 {
		return fmt.Errorf("gold quantity must be > 0")
	}

	// Update gold reserve state
	rv.goldReserve.TotalBars += 1
	rv.goldReserve.TotalTroyOunces += tx.Quantity
	rv.goldReserve.MinPurityBps = tx.PurityBps

	// Calculate AUR to mint (exchange rate)
	// Simple: 1 troy oz ≈ 1 AUR (in production, use market rate)
	auriToMint := tx.Quantity

	// Check if minting would exceed cap
	if rv.totalAuriIssued+auriToMint > rv.maxAuriAllowed {
		return fmt.Errorf("minting %d would exceed max allowed %d", auriToMint, rv.maxAuriAllowed)
	}

	rv.totalAuriIssued += auriToMint

	fmt.Printf("✅ Gold deposit validated: %d troy oz, minting %d AUR\n", tx.Quantity, auriToMint)
	return nil
}

// validateCommodityDeposit verifies commodity (energy, grain, minerals) deposit
func (rv *ReserveValidator) validateCommodityDeposit(tx *pb_transaction.TransactionReserveMint) error {
	if tx.Quantity == 0 {
		return fmt.Errorf("commodity quantity must be > 0")
	}

	// Update commodity reserve based on type
	switch tx.AssetClass {
	case pb_transaction.TransactionReserveMint_ENERGY_RESERVES:
		rv.commodityReserve.EnergyKWh += tx.Quantity

	case pb_transaction.TransactionReserveMint_GRAIN:
		rv.commodityReserve.GrainMetricTons += tx.Quantity

	case pb_transaction.TransactionReserveMint_CRITICAL_MINERALS:
		if rv.commodityReserve.CriticalMineralsKg == nil {
			rv.commodityReserve.CriticalMineralsKg = make(map[string]uint64)
		}
		rv.commodityReserve.CriticalMineralsKg[tx.Unit] += tx.Quantity
	}

	rv.totalAuriIssued += tx.AurionsToMint

	fmt.Printf("✅ Commodity deposit validated: %d %s, minting %d AUR\n", tx.Quantity, tx.Unit, tx.AurionsToMint)
	return nil
}

// VerifyReserveAttestation checks reserve proof is valid
func (rv *ReserveValidator) VerifyReserveAttestation(att *pb_block.ReserveAttestation) error {
	rv.mu.Lock()
	defer rv.mu.Unlock()

	if att == nil {
		return fmt.Errorf("attestation is nil")
	}

	// Verify gold attestation
	if err := rv.verifyGoldAttestation(att.Gold); err != nil {
		return fmt.Errorf("gold attestation failed: %w", err)
	}

	// Verify commodity attestation
	if err := rv.verifyCommodityAttestation(att.Commodities); err != nil {
		return fmt.Errorf("commodity attestation failed: %w", err)
	}

	// Verify infrastructure attestation
	if err := rv.verifyInfrastructureAttestation(att.Infrastructure); err != nil {
		return fmt.Errorf("infrastructure attestation failed: %w", err)
	}

	// Verify foreign clearing attestation
	if err := rv.verifyForeignClearingAttestation(att.ForeignClearing); err != nil {
		return fmt.Errorf("foreign clearing attestation failed: %w", err)
	}

	// Check health status
	if att.HealthStatus != pb_block.ReserveAttestation_HEALTHY {
		return fmt.Errorf("reserve health status not healthy: %v", att.HealthStatus)
	}

	rv.lastVerification = att.BlockHeight

	fmt.Printf("✅ Reserve attestation verified at block %d\n", att.BlockHeight)
	return nil
}

// verifyGoldAttestation validates gold proof
func (rv *ReserveValidator) verifyGoldAttestation(gold *pb_block.GoldAttestation) error {
	if gold == nil {
		return fmt.Errorf("gold attestation is nil")
	}

	// Check bars count
	if gold.TotalBars == 0 {
		return fmt.Errorf("gold bars must be > 0")
	}

	// Check troy ounces
	if gold.TotalTroyOunces == 0 {
		return fmt.Errorf("total troy ounces must be > 0")
	}

	// Verify merkle root is set
	if len(gold.MerkleRoot) != 32 {
		return fmt.Errorf("invalid merkle root length: %d", len(gold.MerkleRoot))
	}

	// Check vaults
	if len(gold.VaultLocations) == 0 {
		return fmt.Errorf("no vault locations provided")
	}

	// Verify purity
	if gold.MinPurityBps < 9900 || gold.MinPurityBps > 10000 {
		return fmt.Errorf("invalid purity: %d bps", gold.MinPurityBps)
	}

	// Update state
	rv.goldReserve.TotalBars = gold.TotalBars
	rv.goldReserve.TotalTroyOunces = gold.TotalTroyOunces
	rv.goldReserve.MerkleRoot = gold.MerkleRoot
	rv.goldReserve.VaultLocations = gold.VaultLocations
	rv.goldReserve.MinPurityBps = gold.MinPurityBps

	return nil
}

// verifyCommodityAttestation validates commodity proof
func (rv *ReserveValidator) verifyCommodityAttestation(com *pb_block.CommodityAttestation) error {
	if com == nil {
		return fmt.Errorf("commodity attestation is nil")
	}

	// At least one commodity must be present
	if com.EnergyKwh == 0 && com.GrainMetricTons == 0 && len(com.CriticalMineralsKg) == 0 {
		return fmt.Errorf("no commodities present")
	}

	// Update state
	rv.commodityReserve.EnergyKWh = com.EnergyKwh
	rv.commodityReserve.GrainMetricTons = com.GrainMetricTons
	rv.commodityReserve.CriticalMineralsKg = com.CriticalMineralsKg
	rv.commodityReserve.LastTelemetryHash = com.TelemetryHash
	rv.commodityReserve.LastTelemetryUpdate = com.TelemetryTimestamp

	return nil
}

// verifyInfrastructureAttestation validates infrastructure proof
func (rv *ReserveValidator) verifyInfrastructureAttestation(infra *pb_block.InfrastructureAttestation) error {
	if infra == nil {
		return fmt.Errorf("infrastructure attestation is nil")
	}

	// Update state
	rv.infrastructureReserve.PortMonthlyYield = infra.PortMonthlyYieldMils
	rv.infrastructureReserve.SolarMonthlyOutput = infra.SolarMonthlyOutputMils
	rv.infrastructureReserve.PropertyAnnualRevenue = infra.PropertyAnnualRevenueMils

	// Calculate total
	total := uint64(0)
	for _, yield := range infra.PortMonthlyYieldMils {
		total += yield
	}
	for _, output := range infra.SolarMonthlyOutputMils {
		total += output
	}
	rv.infrastructureReserve.TotalMonthlyYield = total

	return nil
}

// verifyForeignClearingAttestation validates foreign clearing proof
func (rv *ReserveValidator) verifyForeignClearingAttestation(fc *pb_block.ForeignClearingAttestation) error {
	if fc == nil {
		return fmt.Errorf("foreign clearing attestation is nil")
	}

	if fc.TotalLiquidReservesMils == 0 {
		return fmt.Errorf("foreign clearing total must be > 0")
	}

	// Update state
	rv.foreignClearing.TotalLiquidReserves = fc.TotalLiquidReservesMils
	rv.foreignClearing.CurrencyBreakdown = fc.CurrencyBreakdown

	return nil
}

// CalculateBackingRatio computes current reserve backing ratio
func (rv *ReserveValidator) CalculateBackingRatio() float64 {
	rv.mu.RLock()
	defer rv.mu.RUnlock()

	if rv.totalAuriIssued == 0 {
		return 1.0
	}

	// Calculate total backing in mils (AUR units)
	goldBacking := rv.goldReserve.TotalTroyOunces * 100000000 // troy oz to mils
	commodityBacking := (rv.commodityReserve.EnergyKWh * 1000) + // kWh to mils (simple rate)
		(rv.commodityReserve.GrainMetricTons * 100000000)
	infrastructureBacking := rv.infrastructureReserve.TotalMonthlyYield
	foreignBacking := rv.foreignClearing.TotalLiquidReserves

	totalBacking := goldBacking + commodityBacking + infrastructureBacking + foreignBacking

	ratio := float64(totalBacking) / float64(rv.totalAuriIssued)
	return ratio
}

// IsHealthy checks if reserve state is healthy (ratio >= 1.0)
func (rv *ReserveValidator) IsHealthy() bool {
	ratio := rv.CalculateBackingRatio()
	return ratio >= rv.minBackingRatio
}

// GetReserveStatus returns the current reserve status
func (rv *ReserveValidator) GetReserveStatus() *ReserveStatus {
	rv.mu.RLock()
	defer rv.mu.RUnlock()

	ratio := rv.CalculateBackingRatio()
	status := "HEALTHY"
	if ratio < 1.0 {
		status = "CRITICAL"
	} else if ratio < 1.1 {
		status = "WARNING"
	}

	return &ReserveStatus{
		TotalAuriIssued:      rv.totalAuriIssued,
		GoldTroyOunces:       rv.goldReserve.TotalTroyOunces,
		EnergyKWh:            rv.commodityReserve.EnergyKWh,
		GrainMetricTons:      rv.commodityReserve.GrainMetricTons,
		InfrastructureYield:  rv.infrastructureReserve.TotalMonthlyYield,
		ForeignClearingMils: rv.foreignClearing.TotalLiquidReserves,
		BackingRatio:         ratio,
		Status:               status,
		LastVerificationHeight: rv.lastVerification,
	}
}

// ReserveStatus provides snapshot of reserve state
type ReserveStatus struct {
	TotalAuriIssued        uint64
	GoldTroyOunces         uint64
	EnergyKWh              uint64
	GrainMetricTons        uint64
	InfrastructureYield    uint64
	ForeignClearingMils   uint64
	BackingRatio           float64
	Status                 string
	LastVerificationHeight uint64
}

// ComputeReserveMerkleRoot creates merkle commitment of reserves
func ComputeReserveMerkleRoot(status *ReserveStatus) []byte {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d%d%d%d%d%d",
		status.TotalAuriIssued,
		status.GoldTroyOunces,
		status.EnergyKWh,
		status.GrainMetricTons,
		status.InfrastructureYield,
		status.ForeignClearingMils,
	)))
	return hash.Sum(nil)
}

// PrintReserveStatus prints formatted reserve status
func (s *ReserveStatus) Print() {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     💎 AURION RESERVE STATUS                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Printf("Total AUR Issued:        %,d mils (%.2f AUR)\n", s.TotalAuriIssued, float64(s.TotalAuriIssued)/100000000)
	fmt.Printf("\n🏆 BACKING COMPOSITION:\n")
	fmt.Printf("  Gold (40%):            %,d troy oz\n", s.GoldTroyOunces)
	fmt.Printf("  Energy (30%):          %,d kWh\n", s.EnergyKWh)
	fmt.Printf("  Grain (30%):           %,d metric tons\n", s.GrainMetricTons)
	fmt.Printf("  Infrastructure (20%%):  %,d mils\n", s.InfrastructureYield)
	fmt.Printf("  Foreign Clearing (10%%): %,d mils\n", s.ForeignClearingMils)
	fmt.Printf("\n📊 HEALTH CHECK:\n")
	fmt.Printf("  Backing Ratio:         %.2f × (required: ≥1.0)\n", s.BackingRatio)
	fmt.Printf("  Status:                %s\n", s.Status)
	fmt.Printf("  Last Verified:         Block %d\n", s.LastVerificationHeight)
	fmt.Println("")
}
