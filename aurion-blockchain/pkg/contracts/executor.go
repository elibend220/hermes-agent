package contracts

import (
	"fmt"
	"sync"
	"time"
)

// ContractExecutor handles WASM contract execution
type ContractExecutor struct {
	mu              sync.RWMutex
	store           *ContractStore
	gasMeters       map[string]uint64 // Track gas per execution
	executionCache  map[string]*ContractResult
	maxCallStack    int
	executionTimeMs int64 // Measured execution time
}

// NewContractExecutor creates a new executor
func NewContractExecutor(store *ContractStore) *ContractExecutor {
	return &ContractExecutor{
		store:          store,
		gasMeters:      make(map[string]uint64),
		executionCache: make(map[string]*ContractResult),
		maxCallStack:   100, // Max nested calls
	}
}

// ExecutionContext tracks state during contract execution
type ExecutionContext struct {
	CallID           string
	ContractAddr     string
	Caller           string
	CalleeBalance    uint64
	CallerBalance    uint64
	CallDepth        int
	GasRemaining     uint64
	StartTime        time.Time
	StorageWrites    map[string][]byte
	EmittedEvents    []*ContractEvent
	ReturnValue      []byte
	ExecutionError   error
}

// ExecuteContract runs a contract call with gas metering
func (ce *ContractExecutor) ExecuteContract(call *ContractCall, blockHeight uint64) (*ContractResult, error) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	startTime := time.Now()

	// Record the call
	ce.store.RecordCall(call)

	// Get contract
	contract := ce.store.GetContract(call.ContractAddr)
	if contract == nil {
		return &ContractResult{
			Success:     false,
			Error:       fmt.Sprintf("contract not found: %s", call.ContractAddr),
			GasUsed:     call.GasLimit,
			BlockHeight: blockHeight,
		}, fmt.Errorf("contract not found")
	}

	if !contract.Active {
		return &ContractResult{
			Success:     false,
			Error:       "contract is inactive",
			GasUsed:     call.GasLimit,
			BlockHeight: blockHeight,
		}, fmt.Errorf("contract inactive")
	}

	// Create execution context
	ctx := &ExecutionContext{
		CallID:        call.CallID,
		ContractAddr:  call.ContractAddr,
		Caller:        call.Caller,
		CalleeBalance: contract.Balance,
		CallDepth:     1,
		GasRemaining:  call.GasLimit,
		StartTime:     startTime,
		StorageWrites: make(map[string][]byte),
		EmittedEvents: make([]*ContractEvent, 0),
	}

	// Simulate execution (in production, use wasmtime/wasmer)
	result := ce.simulateExecution(ctx, call, contract)
	result.BlockHeight = blockHeight

	// Record result
	ce.store.RecordResult(call.CallID, result)

	// Track gas usage
	elapsedMs := time.Since(startTime).Milliseconds()
	ce.executionTimeMs = elapsedMs
	ce.gasMeters[call.CallID] = result.GasUsed

	fmt.Printf("✅ Contract call executed: %s (gas: %d, time: %dms)\n",
		call.CallID, result.GasUsed, elapsedMs)

	return result, nil
}

// simulateExecution performs contract execution (placeholder for WASM)
func (ce *ContractExecutor) simulateExecution(
	ctx *ExecutionContext,
	call *ContractCall,
	contract *Contract,
) *ContractResult {

	result := &ContractResult{
		Success:       true,
		GasUsed:       0,
		Output:        []byte{},
		StateChanges:  make(map[string][]byte),
		Events:        make([]*ContractEvent, 0),
	}

	// Measure gas for different operations
	gasUsed := uint64(0)

	// Function call overhead: 1000 gas
	gasUsed += 1000
	if gasUsed > call.GasLimit {
		result.Success = false
		result.Error = "out of gas"
		result.GasUsed = call.GasLimit
		return result
	}

	// Argument processing: 100 gas per arg
	gasUsed += uint64(len(call.Arguments)) * 100
	if gasUsed > call.GasLimit {
		result.Success = false
		result.Error = "out of gas"
		result.GasUsed = call.GasLimit
		return result
	}

	// Storage operations: 5000 gas per write, 200 gas per read
	for k, v := range ctx.StorageWrites {
		gasUsed += 5000
		if gasUsed > call.GasLimit {
			result.Success = false
			result.Error = "out of gas"
			result.GasUsed = call.GasLimit
			return result
		}
		result.StateChanges[k] = v
	}

	// Record contract address in state changes
	result.StateChanges["__contract_addr__"] = []byte(call.ContractAddr)

	// Add events
	for _, event := range ctx.EmittedEvents {
		event.BlockHeight = ctx.StartTime.Unix()
		result.Events = append(result.Events, event)
		gasUsed += 500 // Event logging gas
	}

	// Set execution metrics
	result.GasUsed = gasUsed
	result.Output = ctx.ReturnValue

	// Update contract state
	contract.Balance += call.Value
	contract.Nonce++

	return result
}

// ValidateGas checks if call has sufficient gas limit
func (ce *ContractExecutor) ValidateGas(call *ContractCall, estimatedGas uint64) error {
	if call.GasLimit < estimatedGas {
		return fmt.Errorf("gas limit too low: %d < %d", call.GasLimit, estimatedGas)
	}
	if call.GasLimit > 100_000_000 { // 100M max
		return fmt.Errorf("gas limit too high: %d > 100000000", call.GasLimit)
	}
	return nil
}

// EstimateGas provides a rough gas estimate for a call
func (ce *ContractExecutor) EstimateGas(call *ContractCall) uint64 {
	estimate := uint64(1000) // Base overhead

	// Per argument
	estimate += uint64(len(call.Arguments)) * 100

	// Per storage write (estimate based on function name pattern)
	if call.Function == "set" || call.Function == "update" || call.Function == "store" {
		estimate += 5000
	}

	// Per storage read
	if call.Function == "get" || call.Function == "query" || call.Function == "read" {
		estimate += 200
	}

	return estimate
}

// GetExecutionMetrics returns performance data
func (ce *ContractExecutor) GetExecutionMetrics(callID string) map[string]interface{} {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	return map[string]interface{}{
		"call_id":          callID,
		"gas_used":         ce.gasMeters[callID],
		"execution_time_ms": ce.executionTimeMs,
	}
}

// PrintExecutorStatus displays executor state
func (ce *ContractExecutor) PrintExecutorStatus() {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     ⚙️  CONTRACT EXECUTOR STATUS                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fmt.Printf("\n📊 EXECUTION STATS:\n")
	fmt.Printf("  Total Executions:      %d\n", len(ce.gasMeters))
	fmt.Printf("  Average Gas Used:      %d\n", ce.averageGasUsed())
	fmt.Printf("  Last Execution:        %dms\n", ce.executionTimeMs)

	fmt.Printf("\n⚙️  CONSTRAINTS:\n")
	fmt.Printf("  Max Call Stack:        %d levels\n", ce.maxCallStack)
	fmt.Printf("  Max Gas per Call:      100,000,000\n")
	fmt.Printf("  Gas Price:             1 mil per unit\n")

	fmt.Printf("\n✅ Executor operational\n")
}

func (ce *ContractExecutor) averageGasUsed() uint64 {
	if len(ce.gasMeters) == 0 {
		return 0
	}
	var total uint64
	for _, gas := range ce.gasMeters {
		total += gas
	}
	return total / uint64(len(ce.gasMeters))
}
