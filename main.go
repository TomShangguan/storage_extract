package main

import (
	"fmt"
	"storage_extract/common"
	"storage_extract/state"
	"storage_extract/trie/test"
)

func main() {
	// // Parse command line arguments
	// mode := flag.String("mode", "server", "Run mode: 'server' (web interface) or 'test' (command line test)")
	// port := flag.String("port", "8080", "Server port to use when running in server mode")
	// flag.Parse()

	// // Choose the appropriate mode
	// switch *mode {
	// case "server":
	// 	// Start the web server
	// 	fmt.Printf("Starting Ethereum Storage Visualizer on port %s...\n", *port)
	// 	if err := state.StartServer(*port); err != nil {
	// 		fmt.Printf("Server error: %v\n", err)
	// 	}

	// case "test":
	// 	// Run the command line test
	// 	fmt.Println("Running command line storage test...")
	// 	runTestMode()

	// default:
	// 	fmt.Printf("Unknown mode: %s\n", *mode)
	// 	fmt.Println("Use -mode=server for web interface or -mode=test for command line test")
	// }
	test.Proof_Service_Try()
}

// runTestMode executes the original test functionality
func runTestMode() {
	// 1. Create a contract address
	contractAddr := common.HexToAddress("0x8f5b2b7E299d1eA21b4912A4Cd1339fB157D5362")
	fmt.Printf("Creating contract test: Address %x\n", contractAddr)

	// 2. Create StateDB and CachingDB
	db := &state.CachingDB{}
	stateRoot := common.Hash{}
	stateDB, err := state.New(stateRoot, db)
	if err != nil {
		fmt.Printf("Failed to create StateDB: %v\n", err)
		return
	}

	// 3. Insert several key-value pairs
	testStorage := []struct {
		key   string
		value string
	}{
		// Test different lengths and formats of key-value pairs
		{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "0xaaaaaaaabbbbbbbbccccccccddddddddeeeeeeeeffffffffaaaaaaaabbbbbbbb"},
		{"0x0000000000000000000000000000000000000000000000000000000000000001", "0x0000000000000000000000000000000000000000000000000000000000000123"},
		{"0xabcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		// Add a special value - zero value
		{"0x4444444444444444444444444444444444444444444444444444444444444444", "0x0000000000000000000000000000000000000000000000000000000000000000"},
		// Add a pair of keys with shared prefix, testing branch node
		{"0x7777777700000000000000000000000000000000000000000000000000000000", "0x1111111111111111111111111111111111111111111111111111111111111111"},
		{"0x7777777711111111111111111111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222222222222222222222222222"},
	}

	fmt.Println("\n======= Starting Storage Key-Value Pairs =======")
	for i, kv := range testStorage {
		key := common.HexToHash(kv.key)
		value := common.HexToHash(kv.value)

		fmt.Printf("Storage Item #%d:\n", i+1)
		fmt.Printf("  Key: %x\n", key)
		fmt.Printf("  Value: %x\n", value)

		prevValue := stateDB.SetState(contractAddr, key, value)
		fmt.Printf("  Previous Value: %x\n\n", prevValue)
	}

	// 4. Calculate intermediate root hash to update MPT
	fmt.Println("\n======= Calculating Intermediate Root Hash and Updating MPT =======")
	rootHash := stateDB.IntermediateRoot(false)
	fmt.Printf("\nState Root Hash: %x\n", rootHash)

	fmt.Println("\n======= Test Complete =======")
}
