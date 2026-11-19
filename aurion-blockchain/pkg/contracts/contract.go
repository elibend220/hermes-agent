package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Contract represents a deployed smart contract
type Contract struct {
	Address       string            // Contract address (hash of deployer + nonce)
	CodeHash      string            // SHA256 of WASM bytecode
	Bytecode      []byte            // Compiled WASM code
	DeployedAt    uint64            // Block height when deployed
	DeployedBy    string            // Account that deployed it
	State         map[string][]byte // Contract storage (key-value)
	Balance       uint64            // Contract AUR balance (in mils)
	Nonce         uint64            // Call counter for ordering
	CreatedAt     time.Time         // Deployment timestamp
	GasLimit      uint64            // Per-call gas limit
	GasPrice      uint64            // Price per gas unit
	Active        bool              // Is contract active
}

// ContractCall represents a call to a contract function
type ContractCall struct {
	ContractAddr  string            // Which contract to call
	Caller        string            // Who's calling
	Function      string            // Function name
	Arguments     [][]byte          // Function arguments
	Value         uint64            // AUR to send (in mils)
	GasLimit      uint64            // Gas allowed for execution
	Nonce         uint64            // Call ordering
	CallID        string            // Unique call identifier
}

// ContractResult represents the result of execution
type ContractResult struct {
	Success      bool              // Did execution succeed
	Output       []byte            // Return value
	GasUsed      uint64            // Actual gas consumed
	Error        string            // Error message if failed
	StateChanges map[string][]byte // Modified state
	Events       []*ContractEvent  // Emitted events
	BlockHeight  uint64            // Block where executed
}

// ContractEvent represents a logged event from contract execution
type ContractEvent struct {
	ContractAddr string   // Which contract emitted
	EventName    string   // Event type
	Topics       []string // Indexed parameters
	Data         []byte   // Non-indexed data
	BlockHeight  uint64   // Block height
	TxIndex      uint32   // Transaction index in block
}

// ContractStore manages contract lifecycle
type ContractStore struct {
	mu        sync.RWMutex
	contracts map[string]*Contract // address -> contract
	codeByHash map[string][]byte   // codehash -> bytecode
	callLogs  []*ContractCall      // Call history
	results   map[string]*ContractResult // callid -> result
	events    []*ContractEvent     // All events
}

// NewContractStore creates a new contract store
func NewContractStore() *ContractStore {
	return &ContractStore{
		contracts: make(map[string]*Contract),
		codeByHash: make(map[string][]byte),
		callLogs:  make([]*ContractCall, 0),
		results:   make(map[string]*ContractResult),
		events:    make([]*ContractEvent, 0),
	}
}

// DeployContract stores a newly deployed contract
func (cs *ContractStore) DeployContract(
	deployer string,
	bytecode []byte,
	blockHeight uint64,
) (*Contract, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Validate bytecode
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("bytecode cannot be empty")
	}

	if len(bytecode) > 1_000_000 { // 1MB max
		return nil, fmt.Errorf("bytecode too large: %d bytes", len(bytecode))
	}

	// Compute contract address: SHA256(deployer || nonce || timestamp)
	hash := sha256.New()
	hash.Write([]byte(deployer))
	hash.Write([]byte(fmt.Sprintf("%d", blockHeight)))
	hash.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	address := hex.EncodeToString(hash.Sum(nil))[:40] // 20 bytes = 40 hex chars

	// Compute code hash
	codeHash := sha256.Sum256(bytecode)
	codeHashStr := hex.EncodeToString(codeHash[:])

	contract := &Contract{
		Address:    address,
		CodeHash:   codeHashStr,
		Bytecode:   bytecode,
		DeployedAt: blockHeight,
		DeployedBy: deployer,
		State:      make(map[string][]byte),
		Balance:    0,
		Nonce:      0,
		CreatedAt:  time.Now(),
		GasLimit:   10_000_000, // 10M gas per call
		GasPrice:   1,          // 1 mil per gas unit
		Active:     true,
	}

	cs.contracts[address] = contract
	cs.codeByHash[codeHashStr] = bytecode

	fmt.Printf("✅ Contract deployed: %s (%.6s)\n", address, codeHashStr)
	return contract, nil
}

// GetContract retrieves a contract by address
func (cs *ContractStore) GetContract(address string) *Contract {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.contracts[address]
}

// RecordCall logs a contract call
func (cs *ContractStore) RecordCall(call *ContractCall) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.callLogs = append(cs.callLogs, call)
}

// RecordResult stores the execution result
func (cs *ContractStore) RecordResult(callID string, result *ContractResult) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.results[callID] = result

	// Update contract state if successful
	if result.Success && result.StateChanges != nil {
		if contract, ok := cs.contracts[result.StateChanges["__contract_addr__"]]; ok {
			for k, v := range result.StateChanges {
				if k != "__contract_addr__" {
					contract.State[k] = v
				}
			}
			contract.Nonce++
		}
	}

	// Store events
	if result.Events != nil {
		cs.events = append(cs.events, result.Events...)
	}
}

// GetResult retrieves a call result
func (cs *ContractStore) GetResult(callID string) *ContractResult {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.results[callID]
}

// GetEvents returns events matching criteria
func (cs *ContractStore) GetEvents(contractAddr string, eventName string) []*ContractEvent {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var matching []*ContractEvent
	for _, e := range cs.events {
		if (contractAddr == "" || e.ContractAddr == contractAddr) &&
			(eventName == "" || e.EventName == eventName) {
			matching = append(matching, e)
		}
	}
	return matching
}

// GetState returns contract state value
func (cs *ContractStore) GetState(contractAddr string, key string) []byte {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	contract := cs.contracts[contractAddr]
	if contract == nil {
		return nil
	}
	return contract.State[key]
}

// SetState updates contract state (internal use)
func (cs *ContractStore) SetState(contractAddr string, key string, value []byte) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	contract := cs.contracts[contractAddr]
	if contract == nil {
		return fmt.Errorf("contract not found: %s", contractAddr)
	}

	contract.State[key] = value
	return nil
}

// ContractCount returns total deployed contracts
func (cs *ContractStore) ContractCount() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.contracts)
}

// CallCount returns total calls made
func (cs *ContractStore) CallCount() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.callLogs)
}

// EventCount returns total events logged
func (cs *ContractStore) EventCount() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.events)
}

// PrintContractStatus displays formatted contract info
func PrintContractStatus(cs *ContractStore) {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     📜 AURION SMART CONTRACTS STATUS                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fmt.Printf("\n📊 STATISTICS:\n")
	fmt.Printf("  Deployed Contracts:    %d\n", cs.ContractCount())
	fmt.Printf("  Total Calls:           %d\n", cs.CallCount())
	fmt.Printf("  Total Events:          %d\n", cs.EventCount())

	fmt.Printf("\n⚙️  LIMITS:\n")
	fmt.Printf("  Max Bytecode Size:     1,000,000 bytes (1 MB)\n")
	fmt.Printf("  Default Gas Limit:     10,000,000 gas\n")
	fmt.Printf("  Gas Price:             1 mil per gas\n")

	fmt.Printf("\n✅ Smart contract engine operational\n")
}
