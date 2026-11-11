package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the complete Aurion node configuration
type Config struct {
	Version    string         `json:"version"`
	Network    string         `json:"network"`
	Node       NodeConfig    `json:"node"`
	Blockchain BlockchainConfig `json:"blockchain"`
	Reserve    ReserveConfig  `json:"reserve"`
	RPC        RPCConfig      `json:"rpc"`
}

// NodeConfig contains node-level settings
type NodeConfig struct {
	ID              string `json:"id"`               // Node identifier
	Type            string `json:"type"`             // "VALIDATOR" | "SENTRY" | "ARCHIVE"
	HomeDir         string `json:"home_dir"`
	DataDir         string `json:"data_dir"`
	LogLevel        string `json:"log_level"`        // "debug", "info", "warn", "error"
	MaxPeers        int    `json:"max_peers"`
	ListenAddr      string `json:"listen_addr"`
}

// BlockchainConfig contains blockchain parameters
type BlockchainConfig struct {
	BlockTime          int    `json:"block_time_ms"`      // 1000-1500 ms
	MaxBlockSize       int    `json:"max_block_size"`     // 4 MB
	MaxGasLimit        uint64 `json:"max_gas_limit"`      // 80 million
	TransactionFee     uint64 `json:"transaction_fee"`    // 1 mil
	BurnRate           float64 `json:"burn_rate"`         // 0.025% per tax
	ReserveCheckInterval int  `json:"reserve_check_interval_blocks"` // Every 100
}

// ReserveConfig contains reserve verification settings
type ReserveConfig struct {
	GoldBacking       float64 `json:"gold_backing_pct"`       // 40%
	CommodityBacking  float64 `json:"commodity_backing_pct"`  // 30%
	InfrastructureBacking float64 `json:"infrastructure_backing_pct"` // 20%
	ForeignBacking    float64 `json:"foreign_backing_pct"`    // 10%
	MinimumRatio      float64 `json:"minimum_ratio"`          // 1.0
	TelemetryInterval int    `json:"telemetry_interval_sec"` // Commodity check freq
}

// RPCConfig contains RPC server settings
type RPCConfig struct {
	Enabled      bool   `json:"enabled"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	MaxConnections int  `json:"max_connections"`
	RateLimit    int    `json:"rate_limit_req_sec"` // 100 req/s
}

// DefaultConfig returns the default Aurion configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0.0",
		Network: "MAINNET", // MAINNET | TESTNET | DEVNET
		Node: NodeConfig{
			ID:        "aurion-node-1",
			Type:      "VALIDATOR",
			HomeDir:   os.ExpandEnv("$HOME/.aurion"),
			DataDir:   os.ExpandEnv("$HOME/.aurion/data"),
			LogLevel:  "info",
			MaxPeers:  32,
			ListenAddr: "0.0.0.0:26656",
		},
		Blockchain: BlockchainConfig{
			BlockTime:        1300,      // 1.3 seconds (middle of 1-1.5 range)
			MaxBlockSize:     4194304,   // 4 MB
			MaxGasLimit:      80000000,
			TransactionFee:   1,         // 1 mil
			BurnRate:         0.00025,   // 0.025%
			ReserveCheckInterval: 100,
		},
		Reserve: ReserveConfig{
			GoldBacking:       40.0,
			CommodityBacking:  30.0,
			InfrastructureBacking: 20.0,
			ForeignBacking:    10.0,
			MinimumRatio:      1.0,
			TelemetryInterval: 300, // 5 minutes
		},
		RPC: RPCConfig{
			Enabled:      true,
			Host:         "127.0.0.1",
			Port:         26657,
			MaxConnections: 256,
			RateLimit:    100,
		},
	}
}

// LoadConfig loads configuration from environment or config file
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	// Try to load from config file
	configPath := os.Getenv("AURION_CONFIG")
	if configPath == "" {
		configPath = filepath.Join(cfg.Node.HomeDir, "config.json")
	}

	if _, err := os.Stat(configPath); err == nil {
		// File exists, load it
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	if nodeID := os.Getenv("AURION_NODE_ID"); nodeID != "" {
		cfg.Node.ID = nodeID
	}
	if nodeType := os.Getenv("AURION_NODE_TYPE"); nodeType != "" {
		cfg.Node.Type = nodeType
	}
	if network := os.Getenv("AURION_NETWORK"); network != "" {
		cfg.Network = network
	}

	return cfg, nil
}

// SaveConfig writes configuration to file
func SaveConfig(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig checks configuration validity
func ValidateConfig(cfg *Config) error {
	// Check block time
	if cfg.Blockchain.BlockTime < 1000 || cfg.Blockchain.BlockTime > 1500 {
		return fmt.Errorf("block_time must be 1000-1500 ms, got %d", cfg.Blockchain.BlockTime)
	}

	// Check backing percentages
	total := cfg.Reserve.GoldBacking + cfg.Reserve.CommodityBacking +
		cfg.Reserve.InfrastructureBacking + cfg.Reserve.ForeignBacking
	if total != 100.0 {
		return fmt.Errorf("backing percentages must sum to 100%%, got %.1f%%", total)
	}

	// Check minimum ratio
	if cfg.Reserve.MinimumRatio < 1.0 {
		return fmt.Errorf("minimum ratio must be >= 1.0, got %.2f", cfg.Reserve.MinimumRatio)
	}

	// Check RPC settings
	if cfg.RPC.Port < 1024 || cfg.RPC.Port > 65535 {
		return fmt.Errorf("RPC port must be 1024-65535, got %d", cfg.RPC.Port)
	}

	return nil
}
