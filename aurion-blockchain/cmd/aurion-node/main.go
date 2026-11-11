package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/elibend220/aurion-blockchain/pkg/block"
	"github.com/elibend220/aurion-blockchain/pkg/config"
	pb_consensus "github.com/elibend220/aurion-blockchain/pb/pb_consensus"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Print startup banner
	printBanner()

	// Initialize validator set (4 institutional validators)
	validatorSet := initializeValidators()

	// Create block validator
	validator := block.NewBlockValidator(validatorSet)

	log.Println("✅ Aurion Blockchain Node initialized")
	log.Printf("   Version: %s", cfg.Version)
	log.Printf("   Network: %s", cfg.Network)
	log.Printf("   Validators: %d", len(validatorSet.Validators))

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main node loop (placeholder)
	log.Println("🔗 Aurion Blockchain Node is running...")
	log.Println("   Block Time: 1-1.5 seconds")
	log.Println("   Consensus: Proof-of-Sovereignty (PoSov)")
	log.Println("   Finality: Immediate (3-of-4 multisig)")

	// Keep node running until shutdown signal
	<-sigChan
	log.Println("\n✅ Shutting down gracefully...")
	log.Println("   Validators: ", validatorSet.Validators)
}

// printBanner displays the Aurion blockchain banner
func printBanner() {
	fmt.Println("")
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     🏛️  AURION BLOCKCHAIN - SOVEREIGN LEDGER            ║")
	fmt.Println("║     Kingdom of Joseph - Royal Central Bank             ║")
	fmt.Println("║     Asset-Backed Parity & Digital Reserve Protocol     ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println("")
}

// initializeValidators creates the 4 institutional validators
func initializeValidators() *pb_consensus.ValidatorSet {
	return &pb_consensus.ValidatorSet{
		Validators: []*pb_consensus.ValidatorInfo{
			{
				ValidatorId: "RCB",
				PublicKey:   "0x" + "a".repeat(64), // Placeholder
				Role:        pb_consensus.ValidatorInfo_CENTRAL_BANK,
				IsActive:    true,
			},
			{
				ValidatorId: "STV",
				PublicKey:   "0x" + "b".repeat(64),
				Role:        pb_consensus.ValidatorInfo_TREASURY,
				IsActive:    true,
			},
			{
				ValidatorId: "JVN",
				PublicKey:   "0x" + "c".repeat(64),
				Role:        pb_consensus.ValidatorInfo_JUDICIAL,
				IsActive:    true,
			},
			{
				ValidatorId: "ICA",
				PublicKey:   "0x" + "d".repeat(64),
				Role:        pb_consensus.ValidatorInfo_AUDITOR,
				IsActive:    true,
			},
		},
		QuorumSize: 3, // 3-of-4
	}
}

// String repeater helper
func (s string) repeat(n int) string {
	var result string
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
