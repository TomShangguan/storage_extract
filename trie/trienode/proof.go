package trienode

import (
	"errors"
	"sync"

	"storage_extract/common"
)

// ProofSet stores a set of trie nodes. It implements trie.Database and can also
// act as a cache for another trie.Database.
type ProofSet struct {
	nodes map[string][]byte
	order []string

	dataSize int
	lock     sync.RWMutex
}

func NewProofSet() *ProofSet {
	return &ProofSet{
		nodes: make(map[string][]byte),
	}
}

// Put stores a new node in the set
func (db *ProofSet) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	if _, ok := db.nodes[string(key)]; ok {
		return nil
	}
	keystr := string(key)

	db.nodes[keystr] = common.CopyBytes(value)
	db.order = append(db.order, keystr)
	db.dataSize += len(value)

	return nil
}

// Get returns a stored node
func (db *ProofSet) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if entry, ok := db.nodes[string(key)]; ok {
		return entry, nil
	}
	return nil, errors.New("not found")
}

// Has returns true if the node set contains the given key
func (db *ProofSet) Has(key []byte) (bool, error) {
	_, err := db.Get(key)
	return err == nil, nil
}
