package trie

import (
	"fmt"
	"storage_extract/common"

	"github.com/ethereum/go-ethereum/rlp"
)

// StateTrie wraps a trie with key hashing. In a stateTrie trie, all
// access operations hash the key using keccak256. This prevents
// calling code from creating long chains of nodes that
// increase the access time.
//
// Contrary to a regular trie, a StateTrie can only be created with
// New and must have an attached database. The database also stores
// the preimage of each key if preimage recording is enabled.
//
// StateTrie is not safe for concurrent use.
type StateTrie struct {
	trie       Trie
	hashKeyBuf [common.HashLength]byte // buffer for hashKey (hash of key)
	// secKeyCache      map[string][]byte Not Implemented for now (may not be needed)
}

// NewStateTrie creates a trie with an existing root node from a backing database.
// If root is the zero hash or the sha3 hash of an empty string, the
// trie is initially empty.
// TODO: The current implementation doesn't not use db database.NodeDatabase as a parameter.
// Original function: github.com/ethereum/go-ethereum/trie/secure_trie.go line 77
func NewStateTrie(id *ID) (*StateTrie, error) {
	trie, err := New(id)
	if err != nil {
		return nil, err
	}

	tr := &StateTrie{trie: *trie}

	// Preimage logic is not implemented in the current code.
	return tr, nil

}

// UpdateStorage associates key with value in the trie. Subsequent calls to
// Get will return value. If value has length zero, any existing value
// is deleted from the trie and calls to Get will return nil.
// The value bytes must not be modified by the caller while they are
// stored in the trie.
// If a node is not found in the database, a MissingNodeError is returned.
// Original function: github.com/ethereum/go-ethereum/trie/secure_trie.go line 174
func (t *StateTrie) UpdateStorage(_ common.Address, key, value []byte) error {
	hk := t.hashKey(key)
	v, _ := rlp.EncodeToBytes(value)
	fmt.Println("UpdateStorage key:", hk, "value:", v)
	err := t.trie.Update(hk, v)
	if err != nil {
		return err
	}
	// TODO: Logic for secKeyCache

	return nil
}

// Hash returns the root hash of StateTrie. It does not write to the
// database and can be used even if the trie doesn't have one.
// Original function: github.com/ethereum/go-ethereum/trie/secure_trie.go line 271
func (t *StateTrie) Hash() common.Hash {
	return t.trie.Hash()
}

// hashKey returns the hash of key as an ephemeral buffer.
// The caller must not hold onto the return value because it will become
// invalid on the next call to hashKey or secKey.
// Original function: github.com/ethereum/go-ethereum/trie/secure_trie.go line 299
func (t *StateTrie) hashKey(key []byte) []byte {
	h := newHasher(false)
	h.sha.Reset()
	h.sha.Write(key)
	h.sha.Read(t.hashKeyBuf[:])
	returnHasherToPool(h)
	return t.hashKeyBuf[:]
}

func (t *StateTrie) PrintTrie() {
	t.trie.PrintTrie()
}

//------------------------------------------------------------------------------------------------------------------------
// Below are the additional methods that are not part of the original code but used in the test code snippet.

func (t *StateTrie) HashKey(key []byte) []byte {
	return t.hashKey(key)
}
