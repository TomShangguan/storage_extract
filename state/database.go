package state

import (
	"storage_extract/common"
	"storage_extract/crypto"
)

// Database wraps access to tries and contract code.
type Database interface {

	// OpenStorageTrie opens the storage trie of an account.
	OpenStorageTrie(stateRoot common.Hash, address common.Address, root common.Hash, trie Trie) (Trie, error)
}

// Trie is a Ethereum Merkle Patricia trie.
type Trie interface {
}

// CachingDB is an implementation of Database interface.
type CachingDB struct {
	// disk          ethdb.KeyValueStore TODO: provide an mock underlying keyvalue store db
}

// OpenStorageTrie opens the storage trie of an account.
func (db *CachingDB) OpenStorageTrie(stateRoot common.Hash, address common.Address, root common.Hash, trie Trie) (Trie, error) {
	// Verkle trie case ignored for now
	// TODO: Tuesday 2025/5/20
	tr, err := trie.NewStateTrie(trie.StorageTrieID(stateRoot, crypto.Keccak256Hash(address.Bytes()), root), db.triedb)
	if err != nil {
		return nil, err
	}
	return tr, nil
}
