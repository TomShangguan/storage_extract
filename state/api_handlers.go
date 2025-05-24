package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"storage_extract/common"
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
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.getOrNewStateObject(addr)
	if obj == nil {
		writeError(w, "Could not create or get account", http.StatusInternalServerError)
		return
	}
	obj.finalise()
	obj.updateTrie()
	writeTrieResponse(w, "success", req.Address, obj)
}

// handleGetAccount handles account retrieval requests
func handleGetAccount(w http.ResponseWriter, r *http.Request) {
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
	obj.finalise()
	obj.updateTrie()
	writeTrieResponse(w, "success", req.Address, obj)
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
	obj.finalise()
	obj.updateTrie()
	writeTrieResponse(w, "success", req.Address, obj)
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
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.getOrNewStateObject(addr)
	if obj == nil {
		writeError(w, "Could not get account", http.StatusInternalServerError)
		return
	}
	obj.finalise()
	obj.updateTrie()
	writeTrieResponse(w, "success", req.Address, obj)
}

// writeTrieResponse writes a trie response to the HTTP response
func writeTrieResponse(w http.ResponseWriter, status, address string, obj *StateObject) {
	w.Header().Set("Content-Type", "application/json")
	var rootHash, textData, trieData string
	if obj.trie != nil {
		rootHash = obj.trie.Hash().Hex()
		textBuilder := &strings.Builder{}
		fmt.Fprintf(textBuilder, "==== Storage Trie for Account %x ====\n", obj.address)
		fmt.Fprintf(textBuilder, "Root Hash: %x\n", obj.trie.Hash())
		fmt.Fprintf(textBuilder, "\nHierarchy:\n")
		if printer, ok := any(obj.trie).(interface{ PrintTrieTo(w *strings.Builder) }); ok {
			printer.PrintTrieTo(textBuilder)
		} else {
			obj.trie.PrintTrie()
		}
		textData = textBuilder.String()
		if converter, ok := any(obj.trie).(interface{ ConvertToJSON() ([]byte, error) }); ok {
			if jsonBytes, err := converter.ConvertToJSON(); err == nil {
				trieData = string(jsonBytes)
			}
		}
	} else {
		rootHash = "-"
		textData = "No trie data available."
		trieData = ""
	}
	resp := map[string]interface{}{
		"status":  status,
		"address": address,
		"trie": map[string]interface{}{
			"rootHash": rootHash,
			"textData": textData,
			"trieData": trieData,
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
