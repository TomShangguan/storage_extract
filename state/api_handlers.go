package state

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"storage_extract/common"
	"storage_extract/trie"
	"storage_extract/trie/trienode"
	"strings"
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
	http.HandleFunc("/api/account/get", handleGetAccount)
	http.HandleFunc("/api/storage/batch", handleBatchStorage)
	http.HandleFunc("/api/trie/update", handleUpdateTrie)
	http.HandleFunc("/api/proof", handleProof)
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

// debugLogRequest is a helper function to log request details
func debugLogRequest(r *http.Request) {
	fmt.Printf("[DEBUG] %s %s\n", r.Method, r.URL.Path)
	if r.Method == http.MethodPost {
		var bodyCopy strings.Builder
		if r.Body != nil {
			bodyBytes, _ := io.ReadAll(r.Body)
			bodyCopy.Write(bodyBytes)
			fmt.Printf("[DEBUG] Body: %s\n", bodyCopy.String())
			// Restore the body for further reading
			r.Body = io.NopCloser(strings.NewReader(bodyCopy.String()))
		}
	}
}

// handleCreateAccount handles account creation requests
func handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	debugLogRequest(r)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.getOrNewStateObject(addr)
	if obj == nil {
		writeError(w, "Could not create or get account", http.StatusInternalServerError)
		return
	}
	// Do NOT call finalise or updateTrie here
	writeTrieResponse(w, "success", req.Address, obj)
}

// handleGetAccount handles account retrieval requests
func handleGetAccount(w http.ResponseWriter, r *http.Request) {
	debugLogRequest(r)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.getOrNewStateObject(addr)
	if obj == nil {
		writeError(w, "Could not get account", http.StatusInternalServerError)
		return
	}
	// Do NOT call finalise or updateTrie here
	writeTrieResponse(w, "success", req.Address, obj)
}

// handleBatchStorage handles batch storage updates
func handleBatchStorage(w http.ResponseWriter, r *http.Request) {
	debugLogRequest(r)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string            `json:"address"`
		Storage map[string]string `json:"storage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.getOrNewStateObject(addr)
	if obj == nil {
		writeError(w, "Could not get account", http.StatusInternalServerError)
		return
	}

	for keyHex, valueHex := range req.Storage {
		key := common.HexToHash(keyHex)
		value := common.HexToHash(valueHex)
		stateDB.SetState(addr, key, value)
	}

	// Force trie update to generate the actual trie keys
	obj.updateRoot() // Call updateRoot which will internally update the trie

	// Send the response using the standard function
	writeTrieResponse(w, "success", req.Address, obj)
}

// handleUpdateTrie handles trie update requests
func handleUpdateTrie(w http.ResponseWriter, r *http.Request) {
	debugLogRequest(r)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	println("updateTrie called")
	// Commit all state_object tries
	_, _ = stateDB.Commit(0, false) // Placeholder for now

	addr := common.HexToAddress(req.Address)
	obj := stateDB.getOrNewStateObject(addr)
	if obj == nil {
		writeError(w, "Could not get account after commit", http.StatusInternalServerError)
		return
	}
	// Now the trie is updated, so return the new root
	writeTrieResponse(w, "success", req.Address, obj)
}

// handleProof handles Merkle proof generation requests
func handleProof(w http.ResponseWriter, r *http.Request) {
	debugLogRequest(r)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Address string `json:"address"`
		Key     string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	addr := common.HexToAddress(req.Address)
	obj := stateDB.getOrNewStateObject(addr)
	if obj == nil || obj.trie == nil {
		writeError(w, "Account or trie not found", http.StatusNotFound)
		return
	}
	stateTrie, ok := obj.trie.(*trie.StateTrie)
	if !ok {
		writeError(w, "Trie type not supported for proof", http.StatusInternalServerError)
		return
	}
	// Prepare proof set
	proofDb := trienode.NewProofSet()
	keyBytes := common.Hex2Bytes(req.Key)
	// Generate the proof
	err := stateTrie.Prove(keyBytes, proofDb)
	if err != nil {
		writeError(w, "Failed to generate proof: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Get the root hash
	rootHash := stateTrie.Hash()
	// Verify the proof and get the value
	value, err := trie.VerifyProof(rootHash, keyBytes, proofDb)
	if err != nil {
		writeError(w, "Failed to verify proof: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var valueHex string
	if value != nil {
		valueHex = fmt.Sprintf("%x", value)
	} else {
		valueHex = ""
	}
	resp := map[string]interface{}{
		"rootHash": rootHash.Hex(),
		"value":    valueHex,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// writeTrieResponse writes a trie response to the HTTP response
func writeTrieResponse(w http.ResponseWriter, status, address string, obj *StateObject) {
	w.Header().Set("Content-Type", "application/json")
	var rootHash, textString, textData, trieData string

	if obj.trie != nil {
		rootHash = obj.trie.Hash().Hex()

		// Get the concrete trie implementation - try StateTrie first
		formattedBuilder := &strings.Builder{}

		if stateTrie, ok := obj.trie.(*trie.StateTrie); ok {
			// If it's a StateTrie
			stateTrie.PrintTrieToFormatted(formattedBuilder)
			textString = formattedBuilder.String()
			textData = textString // Send formatted text to frontend

			// Create JSON data for tree view with original keys
			if jsonBytes, err := stateTrie.ConvertToJSON(); err == nil {
				trieData = string(jsonBytes)
			} else {
				fmt.Printf("Error converting StateTrie to JSON: %v\n", err)
				trieData = ""
			}
		} else {
			// Not a StateTrie - basic fallback
			textString = "Trie visualization not supported for this trie type."
			textData = textString
			trieData = ""
		}
	} else {
		rootHash = "-"
		textString = "No trie data available."
		textData = "No trie data available."
		trieData = ""
	}
	resp := map[string]interface{}{
		"status":  status,
		"address": address,
		"trie": map[string]interface{}{
			"rootHash":   rootHash,
			"textData":   textData,
			"trieData":   trieData,
			"textString": textString,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response to the HTTP response
func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": msg,
	})
}
