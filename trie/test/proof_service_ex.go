package test

import (
	"fmt"
	"storage_extract/common"
	"storage_extract/state"
	"storage_extract/types"

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
		{key: *uint256.NewInt(0), value: *uint256.NewInt(0)},
		{key: *uint256.NewInt(12), value: *uint256.NewInt(12)},
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
		key := kv.key
		value := obj.GetState(key.Bytes32())
		fmt.Printf("Storage Item:\n")
		fmt.Printf("  Key: %v\n", key.Hex())
		fmt.Printf("  Value: %v\n", value.Hex())
	}

	fmt.Println("Proof Service Test")

}
