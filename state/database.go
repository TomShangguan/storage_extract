package state

import (
	"fmt"
	"storage_extract/common"
	"storage_extract/crypto"
	"storage_extract/trie"
)

// Database wraps access to tries and contract code.
type Database interface {

	// OpenStorageTrie opens the storage trie of an account.
	// TODO: Currently, one parameter is missing: trie Trie (used to check Verkle trie, so not used for now)
	OpenStorageTrie(stateRoot common.Hash, address common.Address, root common.Hash) (Trie, error)
}

// Trie is a Ethereum Merkle Patricia trie.
type Trie interface {
	// UpdateStorage associates key with value in the trie. If value has length zero,
	// any existing value is deleted from the trie. The value bytes must not be modified
	// by the caller while they are stored in the trie. If a node was not found in the
	// database, a trie.MissingNodeError is returned.
	// Implementation in secure_trie.go
	UpdateStorage(addr common.Address, key, value []byte) error

	// Hash returns the root hash of the trie. It does not write to the database and
	// can be used even if the trie doesn't have one.
	Hash() common.Hash

	// PrintTrie prints the structure of the trie in a human-readable format.
	// It recursively traverses the trie and displays each node with proper indentation.
	// Notice: This function is not included in the original code.
	PrintTrie()
}

// CachingDB is an implementation of Database interface.
type CachingDB struct {
	// disk          ethdb.KeyValueStore TODO: provide an mock underlying keyvalue store db
}

// OpenStorageTrie opens the storage trie of an account.
func (db *CachingDB) OpenStorageTrie(stateRoot common.Hash, address common.Address, root common.Hash) (Trie, error) {
	// Verkle trie case ignored for now
	// TODO: Implement db.triedb paramter for the trie
	fmt.Println("Opening storage trie for address:", (address.Bytes()), "with state root:", stateRoot.Hex(), "and root:", root.Hex())
	tr, err := trie.NewStateTrie(trie.StorageTrieID(stateRoot, crypto.Keccak256Hash(address.Bytes()), root))
	if err != nil {
		return nil, err
	}
	return tr, nil
}
