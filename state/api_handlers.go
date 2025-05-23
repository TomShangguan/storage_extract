package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"storage_extract/common"
)

// Global state database instances
var (
	db        = &CachingDB{}
	stateDB   *StateDB
	stateRoot = common.Hash{}
)

func init() {
	var err error
	stateDB, err = New(stateRoot, db)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize stateDB: %v", err))
	}
}

// StartServer starts the HTTP server
func StartServer(port string) error {
	// Setup API routes
	setupAPIHandlers()

	// Setup static file serving
	setupStaticFileServer()

	// Start HTTP server
	fmt.Println("Server started, listening on port", port)
	return http.ListenAndServe(":"+port, nil)
}

// setupAPIHandlers registers API endpoint handlers
func setupAPIHandlers() {
	http.HandleFunc("/api/account/create", handleCreateAccount)
	http.HandleFunc("/api/storage/set", handleSetStorage)
	http.HandleFunc("/api/storage/batch", handleBatchStorage)
	http.HandleFunc("/api/trie/update", handleUpdateTrie)
}

// setupStaticFileServer configures static file serving
func setupStaticFileServer() {
	// First check if front directory exists
	frontDir := "./front"
	if _, err := os.Stat(frontDir); os.IsNotExist(err) {
		// If front directory doesn't exist, try frontend directory
		frontDir = "./frontend"
		if _, err := os.Stat(frontDir); os.IsNotExist(err) {
			// If neither directory exists, print warning
			fmt.Println("WARNING: Neither 'front' nor 'frontend' directory exists!")
			fmt.Println("Please create either 'front' or 'frontend' directory with HTML/CSS/JS files.")
			return
		}
	}

	// Print the frontend directory being used
	absPath, _ := filepath.Abs(frontDir)
	fmt.Printf("Serving frontend files from: %s\n", absPath)

	// Set up static file server
	fs := http.FileServer(http.Dir(frontDir))
	http.Handle("/", fs)
}

// handleCreateAccount handles account creation requests
func handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	stateDB.getOrNewStateObject(addr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"address": req.Address,
	})
}

// handleSetStorage handles storage key-value setting requests
func handleSetStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
		Key     string `json:"key"`
		Value   string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	key := common.HexToHash(req.Key)
	value := common.HexToHash(req.Value)

	prevValue := stateDB.SetState(addr, key, value)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":        "success",
		"address":       req.Address,
		"key":           req.Key,
		"value":         req.Value,
		"previousValue": prevValue.Hex(),
	})
}

// handleBatchStorage handles batch storage updates
func handleBatchStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string            `json:"address"`
		Storage map[string]string `json:"storage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)

	for keyHex, valueHex := range req.Storage {
		key := common.HexToHash(keyHex)
		value := common.HexToHash(valueHex)
		stateDB.SetState(addr, key, value)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":         "success",
		"address":        req.Address,
		"itemsProcessed": fmt.Sprintf("%d", len(req.Storage)),
	})
}

// handleUpdateTrie handles trie update requests
func handleUpdateTrie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate intermediate root hash to update MPT
	rootHash := stateDB.IntermediateRoot(false)

	// Get account object
	addr := common.HexToAddress(req.Address)
	obj := stateDB.getStateObject(addr)

	// Prepare response with visualization data if available
	response := map[string]interface{}{
		"rootHash": rootHash.Hex(),
		"textData": "Trie updated successfully. Root hash calculated.",
	}

	// If object exists and has storage trie, add trie visualization data
	if obj != nil && obj.trie != nil {
		// TODO: Add trie visualization data
		// This would require implementing interfaces for trie visualization
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
