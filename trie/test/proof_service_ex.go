package test

import (
	"fmt"
	"storage_extract/common"
	"storage_extract/state"
	"storage_extract/trie"
	"storage_extract/types"

	"storage_extract/trie/trienode"

	"github.com/holiman/uint256"
)

func init_stateDB() *state.StateDB {
	db := &state.CachingDB{}
	stateRoot := types.EmptyRootHash
	stateDB, err := state.New(stateRoot, db)
	if err != nil {
		panic("Failed to initialize stateDB: " + err.Error())
	}
	return stateDB
}

func Proof_Service_Try() {
	originalKeyValuePairs := map[common.Address]map[common.Hash]common.Hash{}
	stateDB := init_stateDB()

	// List of test storage key-value pairs in format of uint256.Int
	testStorage := []struct {
		key   uint256.Int
		value uint256.Int
	}{
		{key: *uint256.NewInt(1234), value: *uint256.NewInt(5678)},
		// {key: *uint256.NewInt(1), value: *uint256.NewInt(1)},
		// {key: *uint256.NewInt(12), value: *uint256.NewInt(12)},
	}
	contractAddr := common.HexToAddress("0x8f5b2b7E299d1eA21b4912A4Cd1339fB157D5362")
	originalKeyValuePairs[contractAddr] = map[common.Hash]common.Hash{}
	for i, kv := range testStorage {
		key := kv.key
		value := kv.value
		fmt.Printf("Storage Item #%d:\n", i+1)
		fmt.Printf("  Key: %v\n", key)
		fmt.Printf("  Value: %v\n", value)
		originalKeyValuePairs[contractAddr][key.Bytes32()] = value.Bytes32()
		prevValue := stateDB.SetState(contractAddr, key.Bytes32(), value.Bytes32())
		fmt.Printf("  Previous Value: %x\n\n", prevValue)
	}
	stateDB.Commit(0, false)
	fmt.Println("StateDB initialized and storage items set successfully.")
	obj := stateDB.GetStateObject(contractAddr)
	for _, kv := range testStorage {
		var ori_value uint256.Int
		key := kv.key
		value := obj.GetState(key.Bytes32())
		fmt.Printf("Storage Item:\n")
		fmt.Printf("  Key: %v\n", key)
		fmt.Printf("  Value: %v\n", value)
		ori_value.SetBytes(originalKeyValuePairs[contractAddr][key.Bytes32()].Bytes())

		if value != originalKeyValuePairs[contractAddr][key.Bytes32()] {
			fmt.Printf("  Value mismatch! Expected: %x, Got: %x\n", originalKeyValuePairs[contractAddr][key.Bytes32()].Bytes(), value)
		} else {
			fmt.Printf("  Value matches the original key-value pair: %s \n", &ori_value)
		}
	}

	fmt.Println("Proof Service Test")
	proofSet := trienode.NewProofSet()
	tr := obj.GetTrie()
	if tr == nil {
		fmt.Printf("Trie not found for address %x\n", contractAddr.Bytes())
	}

	stateTrie := (*tr).(*trie.StateTrie)
	if stateTrie == nil {
		fmt.Printf("StateTrie not found for address %x\n", contractAddr.Bytes())
	}
	for key, value := range originalKeyValuePairs[contractAddr] {
		fmt.Printf("Generating proof for key: %x, value: %x\n", key.Bytes(), value.Bytes())

		// Type assertion to convert from the interface to a concrete type
		hashKey := stateTrie.HashKey(key.Bytes())
		if err := stateTrie.Prove(hashKey, proofSet); err != nil {
			fmt.Printf("Failed to generate proof for key %x: %v\n", key.Bytes(), err)
			continue
		}
		fmt.Printf("Proof generated for key %x\n", key.Bytes())
	}

	fmt.Println("Proof generation completed. Verifying proofs...")
	for key, value := range originalKeyValuePairs[contractAddr] {
		fmt.Printf("Verifying proof for key: %x, value: %x\n", key.Bytes(), value.Bytes())
		hashKey := stateTrie.HashKey(key.Bytes())
		valueFromProf, err := trie.VerifyProof(obj.GetRoot(), hashKey, proofSet)
		if err != nil {
			fmt.Printf("Failed to verify proof for key %x: %v\n", key.Bytes(), err)
			continue
		}
		fmt.Println("Proof verification successful. Value from proof: ", valueFromProf)
	}
}
