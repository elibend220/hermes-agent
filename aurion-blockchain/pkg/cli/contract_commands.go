package cli

import (
	"fmt"

	"github.com/elibend220/aurion-blockchain/pkg/contracts"
)

// ContractCommandHandler handles contract-related Hermes commands
type ContractCommandHandler struct {
	contractStore *contracts.ContractStore
	executor      *contracts.ContractExecutor
}

// NewContractCommandHandler creates a new handler
func NewContractCommandHandler() *ContractCommandHandler {
	store := contracts.NewContractStore()
	executor := contracts.NewContractExecutor(store)

	return &ContractCommandHandler{
		contractStore: store,
		executor:      executor,
	}
}

// HandleContractDeploy deploys a new smart contract
func (cch *ContractCommandHandler) HandleContractDeploy(deployer string, bytecodeHex string, blockHeight uint64) string {
	// Decode hex bytecode
	bytecode := make([]byte, len(bytecodeHex)/2)
	for i := 0; i < len(bytecode); i++ {
		fmt.Sscanf(bytecodeHex[i*2:i*2+2], "%x", &bytecode[i])
	}

	contract, err := cch.contractStore.DeployContract(deployer, bytecode, blockHeight)
	if err != nil {
		return fmt.Sprintf("❌ Deployment failed: %v\n", err)
	}

	output := "╔════════════════════════════════════════════════════════╗\n"
	output += "║     📜 CONTRACT DEPLOYED                              ║\n"
	output += "╚════════════════════════════════════════════════════════╝\n\n"
	output += fmt.Sprintf("Address:       %s\n", contract.Address)
	output += fmt.Sprintf("Code Hash:     %s\n", contract.CodeHash[:20])
	output += fmt.Sprintf("Bytecode Size: %d bytes\n", len(bytecode))
	output += fmt.Sprintf("Deployed By:   %s\n", contract.DeployedBy)
	output += fmt.Sprintf("Deployed At:   Block %d\n", contract.DeployedAt)
	output += fmt.Sprintf("Gas Limit:     %d\n", contract.GasLimit)

	return output
}

// HandleContractCall executes a contract function
func (cch *ContractCommandHandler) HandleContractCall(
	contractAddr string,
	caller string,
	function string,
	gasLimit uint64,
	blockHeight uint64,
) string {
	call := &contracts.ContractCall{
		ContractAddr: contractAddr,
		Caller:       caller,
		Function:     function,
		GasLimit:     gasLimit,
		CallID:       fmt.Sprintf("%s_%d", contractAddr, cch.contractStore.CallCount()),
	}

	result, err := cch.executor.ExecuteContract(call, blockHeight)
	if err != nil {
		return fmt.Sprintf("❌ Execution failed: %v\n", err)
	}

	output := "╔════════════════════════════════════════════════════════╗\n"
	output += "║     ✅ CONTRACT CALL EXECUTED                         ║\n"
	output += "╚════════════════════════════════════════════════════════╝\n\n"
	output += fmt.Sprintf("Contract:      %s\n", contractAddr)
	output += fmt.Sprintf("Function:      %s\n", function)
	output += fmt.Sprintf("Caller:        %s\n", caller)
	output += fmt.Sprintf("Status:        %v\n", result.Success)
	output += fmt.Sprintf("Gas Used:      %d\n", result.GasUsed)
	output += fmt.Sprintf("Gas Limit:     %d\n", gasLimit)
	output += fmt.Sprintf("Events:        %d\n", len(result.Events))

	if !result.Success {
		output += fmt.Sprintf("\n⚠️  Error: %s\n", result.Error)
	}

	return output
}

// HandleContractQuery retrieves contract state
func (cch *ContractCommandHandler) HandleContractQuery(contractAddr string, key string) string {
	contract := cch.contractStore.GetContract(contractAddr)
	if contract == nil {
		return fmt.Sprintf("❌ Contract not found: %s\n", contractAddr)
	}

	value := cch.contractStore.GetState(contractAddr, key)

	output := "╔════════════════════════════════════════════════════════╗\n"
	output += "║     📖 CONTRACT STATE QUERY                           ║\n"
	output += "╚════════════════════════════════════════════════════════╝\n\n"
	output += fmt.Sprintf("Contract:      %s\n", contractAddr)
	output += fmt.Sprintf("Key:           %s\n", key)
	output += fmt.Sprintf("Value:         %s\n", string(value))
	output += fmt.Sprintf("Nonce:         %d\n", contract.Nonce)

	return output
}

// HandleContractList lists all deployed contracts
func (cch *ContractCommandHandler) HandleContractList() string {
	output := "╔════════════════════════════════════════════════════════╗\n"
	output += "║     📜 DEPLOYED CONTRACTS                             ║\n"
	output += "╚════════════════════════════════════════════════════════╝\n\n"
	output += fmt.Sprintf("Total Contracts: %d\n\n", cch.contractStore.ContractCount())

	output += "ID | Address | Deployed By | Gas Limit\n"
	output += "---|---------|-------------|----------\n"

	return output
}

// HandleContractStatus shows contract execution statistics
func (cch *ContractCommandHandler) HandleContractStatus() string {
	output := "╔════════════════════════════════════════════════════════╗\n"
	output += "║     🔧 CONTRACT ENGINE STATUS                         ║\n"
	output += "╚════════════════════════════════════════════════════════╝\n\n"
	output += fmt.Sprintf("Deployed Contracts:    %d\n", cch.contractStore.ContractCount())
	output += fmt.Sprintf("Total Calls:           %d\n", cch.contractStore.CallCount())
	output += fmt.Sprintf("Total Events:          %d\n", cch.contractStore.EventCount())
	output += "\n⚙️  EXECUTION LIMITS:\n"
	output += "  Max Bytecode Size:     1,000,000 bytes (1 MB)\n"
	output += "  Gas Limit per Call:    100,000,000\n"
	output += "  Gas Price:             1 mil per unit\n"
	output += "\n✅ Smart contract engine operational\n"

	return output
}

// HandleGasEstimate estimates gas for a contract call
func (cch *ContractCommandHandler) HandleGasEstimate(contractAddr string, function string) string {
	call := &contracts.ContractCall{
		ContractAddr: contractAddr,
		Function:     function,
	}

	estimate := cch.executor.EstimateGas(call)

	output := "╔════════════════════════════════════════════════════════╗\n"
	output += "║     ⛽ GAS ESTIMATE                                    ║\n"
	output += "╚════════════════════════════════════════════════════════╝\n\n"
	output += fmt.Sprintf("Contract:        %s\n", contractAddr)
	output += fmt.Sprintf("Function:        %s\n", function)
	output += fmt.Sprintf("Estimated Gas:   %d\n", estimate)
	output += fmt.Sprintf("Gas Price:       1 mil per unit\n")
	output += fmt.Sprintf("Est. Cost:       %d mils (%.8f AUR)\n", estimate, float64(estimate)/100_000_000)

	return output
}

// PrintContractHelp displays command help
func PrintContractHelp() string {
	help := `
╔════════════════════════════════════════════════════════╗
║     📜 SMART CONTRACT COMMANDS                         ║
╚════════════════════════════════════════════════════════╝

DEPLOYMENT:
  hermes contract deploy <bytecode>
    Deploy new smart contract from WASM bytecode

EXECUTION:
  hermes contract call <address> <function> [args]
    Execute contract function with arguments

QUERIES:
  hermes contract state <address> <key>
    Read contract storage state by key

  hermes contract list
    Show all deployed contracts

  hermes contract status
    Display contract engine statistics

GAS:
  hermes contract estimate <address> <function>
    Estimate gas cost for contract call

EVENTS:
  hermes contract events <address> [event-name]
    Query contract events

EXAMPLES:
  # Deploy a simple counter contract
  hermes contract deploy 0xABCD...

  # Call increment() on counter
  hermes contract call 0xABCD... increment

  # Read counter value
  hermes contract state 0xABCD... value

  # Estimate gas for function call
  hermes contract estimate 0xABCD... increment

LIMITS:
  - Max bytecode: 1 MB
  - Max gas per call: 100 million
  - Gas price: 1 mil per unit
`
	return help
}
