package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"storage_extract/common"
	"storage_extract/state"
	"storage_extract/trie"
	"storage_extract/trie/trienode"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/holiman/uint256"
)

// Global state database instances
var (
	db                    = &state.CachingDB{}
	stateDB               *state.StateDB
	stateRoot             = common.Hash{}
	originalKeyValuePairs = make(map[common.Address]map[common.Hash]common.Hash)
	proofSet              = trienode.NewProofSet()
)

func init() {
	var err error
	stateDB, err = state.New(stateRoot, db)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize stateDB: %v", err))
	}
}

// StartServer starts the Gin HTTP server
func StartServer(port string) error {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.Default()

	// Setup static file serving
	setupStaticFileServer(r)

	// Setup API routes
	setupAPIHandlers(r)

	// Start HTTP server
	fmt.Println("Server started, listening on port", port)
	return r.Run(":" + port)
}

// setupAPIHandlers registers API endpoint handlers with Gin router
func setupAPIHandlers(r *gin.Engine) {
	api := r.Group("/api")
	{
		account := api.Group("/account")
		{
			account.POST("/create", ginHandleCreateAccount)
			account.POST("/get", ginHandleGetAccount)
		}

		api.POST("/storage/update", ginHandleUpdateStorage)
		api.POST("/proof", ginHandleProof)
		api.POST("/storage/get", ginHandleGetValue)
	}
}

// setupStaticFileServer configures static file serving with Gin
func setupStaticFileServer(r *gin.Engine) {
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

	// Set up static file server with Gin
	r.Static("/", frontDir)
}

// debugLogRequest is a helper function to log request details for Gin
func debugLogRequest(c *gin.Context) {
	fmt.Printf("[DEBUG] %s %s\n", c.Request.Method, c.Request.URL.Path)
	if c.Request.Method == http.MethodPost {
		var bodyCopy strings.Builder
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			bodyCopy.Write(bodyBytes)
			fmt.Printf("[DEBUG] Body: %s\n", bodyCopy.String())
			// Restore the body for further reading
			c.Request.Body = io.NopCloser(strings.NewReader(bodyCopy.String()))
		}
	}
}

// ginHandleCreateAccount handles account creation requests
func ginHandleCreateAccount(c *gin.Context) {
	debugLogRequest(c)

	var req struct {
		Address string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	// Create a new state object if it doesn't exist, since stateDB doesn't expose getOrNewStateObject
	// We'll access it via SetState which internally calls getOrNewStateObject
	stateDB.SetState(addr, common.Hash{}, common.Hash{}) // This will create the account if it doesn't exist
	obj := stateDB.GetStateObject(addr)
	if obj == nil {
		ginWriteError(c, "Could not create or get account", http.StatusInternalServerError)
		return
	}
	// Do NOT call finalise or updateTrie here
	ginWriteTrieResponse(c, "success", req.Address, obj)
}

// ginHandleGetAccount handles account retrieval requests
func ginHandleGetAccount(c *gin.Context) {
	debugLogRequest(c)

	var req struct {
		Address string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.GetStateObject(addr)
	if obj == nil {
		ginWriteError(c, "Could not get account", http.StatusInternalServerError)
		return
	}
	// Do NOT call finalise or updateTrie here
	ginWriteTrieResponse(c, "success", req.Address, obj)
}

// ginHandleUpdateStorage handles storage update requests - consolidates batch storage and trie update
func ginHandleUpdateStorage(c *gin.Context) {
	debugLogRequest(c)

	var req struct {
		Address string            `json:"address"`
		Storage map[string]string `json:"storage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.GetStateObject(addr)
	if obj == nil {
		ginWriteError(c, "Could not get account", http.StatusInternalServerError)
		return
	}

	// Initialize the map for this address if it doesn't exist
	if originalKeyValuePairs[addr] == nil {
		originalKeyValuePairs[addr] = make(map[common.Hash]common.Hash)
	}

	for keyHex, valueHex := range req.Storage {
		// Convert hex strings to uint256.Int with error handling
		key, err := uint256.FromHex(keyHex)
		if err != nil {
			ginWriteError(c, "Invalid key format: "+err.Error(), http.StatusBadRequest)
			return
		}

		value, err := uint256.FromHex(valueHex)
		if err != nil {
			ginWriteError(c, "Invalid value format: "+err.Error(), http.StatusBadRequest)
			return
		}

		originalKeyValuePairs[addr][key.Bytes32()] = value.Bytes32()
		stateDB.SetState(addr, key.Bytes32(), value.Bytes32())
	}

	// Force trie update to generate the actual trie keys
	stateDB.Commit(0, false) // Call updateRoot which will internally update the trie
	obj = stateDB.GetStateObject(addr)
	if obj == nil {
		ginWriteError(c, "Could not get account after storage update", http.StatusInternalServerError)
		return
	}

	tr := obj.GetTrie()
	if tr == nil {
		ginWriteError(c, "Trie not found for address "+req.Address, http.StatusInternalServerError)
		return
	}
	stateTrie, ok := (*tr).(*trie.StateTrie)
	if !ok || stateTrie == nil {
		ginWriteError(c, "StateTrie not found for address "+req.Address, http.StatusInternalServerError)
		return
	}
	// Generate the proof for the updated storage
	for key := range originalKeyValuePairs[addr] {
		hashKey := stateTrie.HashKey(key.Bytes())

		if err := stateTrie.Prove(hashKey, proofSet); err != nil {
			fmt.Printf("Failed to generate proof for key %x: %v\n", key.Bytes(), err)
			continue
		}
		fmt.Printf("Proof generated for key %x\n", key.Bytes())
	}

	// Send the response using the standard function
	ginWriteTrieResponse(c, "success", req.Address, obj)
}

// ginHandleProof handles Merkle proof generation requests
func ginHandleProof(c *gin.Context) {
	debugLogRequest(c)

	var req struct {
		Address string `json:"address"`
		Key     string `json:"key"`
		Root    string `json:"root"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse the key first to validate format
	keyBytes, err := uint256.FromHex(req.Key)
	if err != nil {
		ginWriteError(c, "Invalid key format: "+err.Error(), http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.GetStateObject(addr)
	if obj == nil {
		ginWriteError(c, "Account not found", http.StatusNotFound)
		return
	}

	tr := obj.GetTrie()
	if tr == nil {
		ginWriteError(c, "Trie not found for address "+req.Address, http.StatusInternalServerError)
		return
	}
	stateTrie, ok := (*tr).(*trie.StateTrie)
	if !ok || stateTrie == nil {
		ginWriteError(c, "StateTrie not found for address "+req.Address, http.StatusInternalServerError)
		return
	}

	// Convert string root to hex if provided, otherwise use obj's root
	var root common.Hash
	if req.Root != "" {
		root = common.HexToHash(req.Root)
	} else {
		ginWriteError(c, "Root hash is required", http.StatusBadRequest)
		return
	}

	// Hash the key using the same method as in proof generation
	var hashKey common.Hash
	hashKey = keyBytes.Bytes32()

	// Verify the proof and get the value
	value, err := trie.VerifyProof(root, stateTrie.HashKey(hashKey.Bytes()), proofSet)
	if err != nil {
		ginWriteError(c, "Failed to verify proof: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var valueHex string
	if value != nil {
		valueHex = fmt.Sprintf("%x", value)
	} else {
		valueHex = ""
	}

	resp := map[string]interface{}{
		"value": valueHex,
	}

	c.JSON(http.StatusOK, resp)
}

// ginHandleGetValue handles storage value retrieval requests
func ginHandleGetValue(c *gin.Context) {
	debugLogRequest(c)

	var req struct {
		Address string `json:"address"`
		Key     string `json:"key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr := common.HexToAddress(req.Address)
	obj := stateDB.GetStateObject(addr)
	if obj == nil {
		ginWriteError(c, "Account not found", http.StatusNotFound)
		return
	}

	// Convert hex string to uint256.Int, then to Bytes32 for storage key
	key, err := uint256.FromHex(req.Key)
	if err != nil {
		ginWriteError(c, "Invalid key format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the storage value using GetState (similar to proof_service_ex.go logic)
	value := obj.GetState(key.Bytes32())

	// Format the value without leading zeros
	var valueHex string
	if value != (common.Hash{}) {
		valueHex = strings.TrimLeft(fmt.Sprintf("%x", value), "0")
		if valueHex == "" {
			valueHex = "0"
		}
	} else {
		valueHex = "0"
	}

	// Check if we have the original value for comparison (like in proof_service_ex.go)
	var originalMatch bool
	if originalKeyValuePairs[addr] != nil {
		if originalValue, exists := originalKeyValuePairs[addr][key.Bytes32()]; exists {
			originalMatch = (value == originalValue)
		}
	}

	resp := map[string]interface{}{
		"address":       req.Address,
		"key":           req.Key,
		"value":         valueHex,
		"originalMatch": originalMatch,
	}

	c.JSON(http.StatusOK, resp)
}

// ginWriteTrieResponse writes a trie response to the HTTP response using Gin
func ginWriteTrieResponse(c *gin.Context, status, address string, obj *state.StateObject) {
	var rootHash, textString, textData, trieData string

	// Get original key-value pairs for this address
	addr := common.HexToAddress(address)
	var originalKVPairs []map[string]interface{}
	if originalKeyValuePairs[addr] != nil {
		for key, value := range originalKeyValuePairs[addr] {
			originalKVPairs = append(originalKVPairs, map[string]interface{}{
				"originalKey":   fmt.Sprintf("0x%x", key.Bytes()),
				"originalValue": fmt.Sprintf("0x%x", value.Bytes()),
				"keyHex":        fmt.Sprintf("%x", key.Bytes()),
				"valueHex":      fmt.Sprintf("%x", value.Bytes()),
			})
		}
	}

	// Get the trie using the public method
	triePtr := obj.GetTrie()
	if triePtr != nil && *triePtr != nil {
		rootHash = (*triePtr).Hash().Hex()

		// Get the concrete trie implementation - try StateTrie first
		formattedBuilder := &strings.Builder{}

		if stateTrie, ok := (*triePtr).(*trie.StateTrie); ok {
			// If it's a StateTrie
			stateTrie.PrintTrieToFormatted(formattedBuilder)
			textString = formattedBuilder.String()
			textData = textString // Send formatted text to frontend

			// Prepare original keys and values maps for enhanced JSON conversion
			originalKeysMap := make(map[string]string)
			originalValuesMap := make(map[string]string)

			if originalKeyValuePairs[addr] != nil {
				for key, value := range originalKeyValuePairs[addr] {
					keyHex := fmt.Sprintf("%x", key.Bytes())
					valueHex := fmt.Sprintf("%x", value.Bytes())
					hashedKey := stateTrie.HashKey(key.Bytes())
					hashedKeyHex := fmt.Sprintf("%x", hashedKey)

					// Map keys to their original forms
					originalKeysMap[keyHex] = fmt.Sprintf("0x%x", key.Bytes())
					originalKeysMap[hashedKeyHex] = fmt.Sprintf("0x%x", key.Bytes())

					// Map values to their original forms
					// Store the relationship between original value and its various representations
					originalValueStr := fmt.Sprintf("0x%x", value.Bytes())

					// Handle different value formats that might appear in the trie
					// 1. Trimmed hex (without leading zeros)
					trimmedValue := strings.TrimLeft(valueHex, "0")
					if trimmedValue == "" {
						trimmedValue = "0"
					}
					originalValuesMap[trimmedValue] = originalValueStr
					originalValuesMap["0x"+trimmedValue] = originalValueStr

					// 2. Full hex (with leading zeros)
					originalValuesMap[valueHex] = originalValueStr
					originalValuesMap["0x"+valueHex] = originalValueStr

					// 3. Padded formats (common in trie storage)
					// Zero-pad to 32 bytes (64 hex chars) for common storage format
					paddedValue := fmt.Sprintf("%064s", valueHex)
					originalValuesMap[paddedValue] = originalValueStr
					originalValuesMap["0x"+paddedValue] = originalValueStr

					// Create nibble-encoded version for trie matching
					// Convert hashed key bytes to nibbles (each byte becomes 2 nibbles)
					nibbles := make([]byte, len(hashedKey)*2)
					for i, b := range hashedKey {
						nibbles[i*2] = b >> 4     // High nibble
						nibbles[i*2+1] = b & 0x0F // Low nibble
					}
					nibblesHex := fmt.Sprintf("%x", nibbles)
					originalKeysMap[nibblesHex] = fmt.Sprintf("0x%x", key.Bytes())

					// Also try with termination marker (0x10 at the end)
					terminatedNibbles := nibblesHex + "10"
					originalKeysMap[terminatedNibbles] = fmt.Sprintf("0x%x", key.Bytes())

					fmt.Printf("DEBUG: Mappings for original key %x -> value %x:\n", key.Bytes(), value.Bytes())
					fmt.Printf("  - Original key hex: %s\n", keyHex)
					fmt.Printf("  - Hashed key hex: %s\n", hashedKeyHex)
					fmt.Printf("  - Nibbles hex: %s\n", nibblesHex)
					fmt.Printf("  - Terminated nibbles: %s\n", terminatedNibbles)
					fmt.Printf("  - Original value: %s\n", originalValueStr)
					fmt.Printf("  - Trimmed value: %s\n", trimmedValue)
				}
				fmt.Printf("DEBUG: originalKeysMap has %d entries, originalValuesMap has %d entries\n",
					len(originalKeysMap), len(originalValuesMap))
			}

			// Create JSON data for tree view with original keys and values
			if jsonBytes, err := stateTrie.ConvertToJSONWithOriginalKeys(originalKeysMap, originalValuesMap); err == nil {
				trieData = string(jsonBytes)
			} else {
				fmt.Printf("Error converting StateTrie to JSON with original keys: %v\n", err)
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
			"rootHash":        rootHash,
			"textData":        textData,
			"trieData":        trieData,
			"textString":      textString,
			"originalKVPairs": originalKVPairs,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// ginWriteError writes an error response to the HTTP response using Gin
func ginWriteError(c *gin.Context, msg string, code int) {
	c.JSON(code, map[string]interface{}{
		"error": msg,
	})
}
