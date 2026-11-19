package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/elibend220/aurion-blockchain/pkg/contracts"
	"github.com/gorilla/mux"
)

// APIServer handles JSON-RPC and REST endpoints
type APIServer struct {
	store    *contracts.ContractStore
	executor *contracts.ContractExecutor
	port     int
	router   *mux.Router
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// NewAPIServer creates a new API server
func NewAPIServer(port int) *APIServer {
	store := contracts.NewContractStore()
	executor := contracts.NewContractExecutor(store)

	server := &APIServer{
		store:    store,
		executor: executor,
		port:     port,
		router:   mux.NewRouter(),
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all API endpoints
func (as *APIServer) setupRoutes() {
	// JSON-RPC endpoint (POST)
	as.router.HandleFunc("/rpc", as.handleRPC).Methods("POST")

	// REST endpoints
	as.router.HandleFunc("/api/contract/deploy", as.handleDeploy).Methods("POST")
	as.router.HandleFunc("/api/contract/call", as.handleCall).Methods("POST")
	as.router.HandleFunc("/api/contract/state/{address}/{key}", as.handleGetState).Methods("GET")
	as.router.HandleFunc("/api/contract/{address}", as.handleGetContract).Methods("GET")
	as.router.HandleFunc("/api/contracts", as.handleListContracts).Methods("GET")
	as.router.HandleFunc("/api/events", as.handleGetEvents).Methods("GET")
	as.router.HandleFunc("/api/status", as.handleStatus).Methods("GET")

	// Health check
	as.router.HandleFunc("/health", as.handleHealth).Methods("GET")
}

// handleRPC processes JSON-RPC 2.0 requests
func (as *APIServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   "Invalid JSON",
			ID:      nil,
		})
		return
	}

	var result interface{}
	var errMsg interface{}

	switch req.Method {
	case "contract_deploy":
		result, errMsg = as.rpcDeploy(req.Params)
	case "contract_call":
		result, errMsg = as.rpcCall(req.Params)
	case "contract_getState":
		result, errMsg = as.rpcGetState(req.Params)
	case "contract_getContract":
		result, errMsg = as.rpcGetContract(req.Params)
	case "contract_listContracts":
		result, errMsg = as.rpcListContracts()
	case "contract_gasEstimate":
		result, errMsg = as.rpcEstimateGas(req.Params)
	default:
		errMsg = fmt.Sprintf("Method not found: %s", req.Method)
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		Error:   errMsg,
		ID:      req.ID,
	}

	json.NewEncoder(w).Encode(response)
}

// rpcDeploy handles contract_deploy RPC method
func (as *APIServer) rpcDeploy(params json.RawMessage) (interface{}, interface{}) {
	var deployReq struct {
		Deployer string `json:"deployer"`
		Bytecode string `json:"bytecode"` // Hex-encoded
		BlockHeight uint64 `json:"blockHeight"`
	}

	if err := json.Unmarshal(params, &deployReq); err != nil {
		return nil, "Invalid params"
	}

	// Decode hex bytecode
	bytecode := make([]byte, len(deployReq.Bytecode)/2)
	for i := 0; i < len(bytecode); i++ {
		fmt.Sscanf(deployReq.Bytecode[i*2:i*2+2], "%x", &bytecode[i])
	}

	contract, err := as.store.DeployContract(deployReq.Deployer, bytecode, deployReq.BlockHeight)
	if err != nil {
		return nil, err.Error()
	}

	return map[string]interface{}{
		"address":      contract.Address,
		"codeHash":     contract.CodeHash,
		"deployedAt":   contract.DeployedAt,
		"deployedBy":   contract.DeployedBy,
	}, nil
}

// rpcCall handles contract_call RPC method
func (as *APIServer) rpcCall(params json.RawMessage) (interface{}, interface{}) {
	var callReq struct {
		ContractAddr string   `json:"contractAddr"`
		Caller       string   `json:"caller"`
		Function     string   `json:"function"`
		Arguments    []string `json:"arguments"`
		Value        uint64   `json:"value"`
		GasLimit     uint64   `json:"gasLimit"`
	}

	if err := json.Unmarshal(params, &callReq); err != nil {
		return nil, "Invalid params"
	}

	// Decode arguments
	args := make([][]byte, len(callReq.Arguments))
	for i, arg := range callReq.Arguments {
		args[i] = []byte(arg)
	}

	call := &contracts.ContractCall{
		ContractAddr: callReq.ContractAddr,
		Caller:       callReq.Caller,
		Function:     callReq.Function,
		Arguments:    args,
		Value:        callReq.Value,
		GasLimit:     callReq.GasLimit,
		CallID:       fmt.Sprintf("%s_%d", callReq.ContractAddr, as.store.CallCount()),
	}

	result, err := as.executor.ExecuteContract(call, 0)
	if err != nil {
		return nil, err.Error()
	}

	return map[string]interface{}{
		"success":    result.Success,
		"gasUsed":    result.GasUsed,
		"error":      result.Error,
		"output":     result.Output,
		"events":     len(result.Events),
	}, nil
}

// rpcGetState handles contract_getState RPC method
func (as *APIServer) rpcGetState(params json.RawMessage) (interface{}, interface{}) {
	var req struct {
		ContractAddr string `json:"contractAddr"`
		Key          string `json:"key"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, "Invalid params"
	}

	value := as.store.GetState(req.ContractAddr, req.Key)
	return map[string]interface{}{
		"key":   req.Key,
		"value": string(value),
	}, nil
}

// rpcGetContract handles contract_getContract RPC method
func (as *APIServer) rpcGetContract(params json.RawMessage) (interface{}, interface{}) {
	var req struct {
		Address string `json:"address"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, "Invalid params"
	}

	contract := as.store.GetContract(req.Address)
	if contract == nil {
		return nil, "Contract not found"
	}

	return map[string]interface{}{
		"address":    contract.Address,
		"codeHash":   contract.CodeHash,
		"deployedAt": contract.DeployedAt,
		"deployedBy": contract.DeployedBy,
		"balance":    contract.Balance,
		"nonce":      contract.Nonce,
		"active":     contract.Active,
	}, nil
}

// rpcListContracts handles contract_listContracts RPC method
func (as *APIServer) rpcListContracts() (interface{}, interface{}) {
	return map[string]interface{}{
		"count": as.store.ContractCount(),
	}, nil
}

// rpcEstimateGas handles contract_gasEstimate RPC method
func (as *APIServer) rpcEstimateGas(params json.RawMessage) (interface{}, interface{}) {
	var callReq struct {
		ContractAddr string   `json:"contractAddr"`
		Function     string   `json:"function"`
		Arguments    []string `json:"arguments"`
	}

	if err := json.Unmarshal(params, &callReq); err != nil {
		return nil, "Invalid params"
	}

	call := &contracts.ContractCall{
		ContractAddr: callReq.ContractAddr,
		Function:     callReq.Function,
		Arguments:    make([][]byte, len(callReq.Arguments)),
	}

	estimate := as.executor.EstimateGas(call)
	return map[string]interface{}{
		"estimatedGas": estimate,
	}, nil
}

// REST endpoint handlers
func (as *APIServer) handleDeploy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Deployer    string `json:"deployer"`
		Bytecode    string `json:"bytecode"`
		BlockHeight uint64 `json:"blockHeight"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	bytecode := make([]byte, len(req.Bytecode)/2)
	for i := 0; i < len(bytecode); i++ {
		fmt.Sscanf(req.Bytecode[i*2:i*2+2], "%x", &bytecode[i])
	}

	contract, err := as.store.DeployContract(req.Deployer, bytecode, req.BlockHeight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(contract)
}

func (as *APIServer) handleCall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req contracts.ContractCall
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := as.executor.ExecuteContract(&req, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func (as *APIServer) handleGetState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	value := as.store.GetState(vars["address"], vars["key"])
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":   vars["key"],
		"value": string(value),
	})
}

func (as *APIServer) handleGetContract(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	contract := as.store.GetContract(vars["address"])
	if contract == nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(contract)
}

func (as *APIServer) handleListContracts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": as.store.ContractCount(),
	})
}

func (as *APIServer) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	contractAddr := r.URL.Query().Get("contract")
	eventName := r.URL.Query().Get("event")

	events := as.store.GetEvents(contractAddr, eventName)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  len(events),
		"events": events,
	})
}

func (as *APIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "operational",
		"contracts": as.store.ContractCount(),
		"calls":     as.store.CallCount(),
		"events":    as.store.EventCount(),
	})
}

func (as *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Start begins listening for requests
func (as *APIServer) Start() error {
	addr := fmt.Sprintf(":%d", as.port)
	fmt.Printf("\n🌐 JSON-RPC API Server starting on %s\n", addr)
	fmt.Printf("   POST /rpc - JSON-RPC 2.0 endpoint\n")
	fmt.Printf("   POST /api/contract/deploy - Deploy contract\n")
	fmt.Printf("   POST /api/contract/call - Call contract\n")
	fmt.Printf("   GET  /api/status - API status\n\n")

	return http.ListenAndServe(addr, as.router)
}

func main() {
	port := 8545
	if p := os.Getenv("AURION_API_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	server := NewAPIServer(port)

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     🌐 AURION JSON-RPC API SERVER v1.0               ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
