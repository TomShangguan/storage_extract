package trie

import (
	"bytes"
	"fmt"
	"storage_extract/common"
	"storage_extract/ethdb"
)

// Prove constructs a merkle proof for key. The result contains all encoded nodes
// on the path to the value at key. The value itself is also included in the last
// node and can be retrieved by verifying the proof.
//
// If the trie does not contain a value for key, the returned proof contains all
// nodes of the longest existing prefix of the key (at least the root node), ending
// with the node that proves the absence of the key.
// Original function: github.com/ethereum/go-ethereum/trie/proof.go line 36
func (t *Trie) Prove(key []byte, proofDb ethdb.KeyValueWriter) error {
	// TODO: The logic of trie commit is not implemented yet.
	// Short circuit if the trie is already committed and not usable.
	// if t.committed {
	// 	return ErrCommitted
	// }
	// Collect all nodes on the path to key.
	var (
		prefix []byte
		nodes  []node
		tn     = t.root
	)
	key = keybytesToHex(key)
	for len(key) > 0 && tn != nil {
		switch n := tn.(type) {
		case *shortNode:
			if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
				// The trie doesn't contain the key.
				fmt.Println("Key not found in short node:", key)
				tn = nil
			} else {
				tn = n.Val
				prefix = append(prefix, n.Key...)
				key = key[len(n.Key):]
			}
			nodes = append(nodes, n)
		case *fullNode:
			tn = n.Children[key[0]]
			prefix = append(prefix, key[0])
			key = key[1:]
			nodes = append(nodes, n)
		default:
			panic(fmt.Sprintf("%T: invalid node: %v", tn, tn))
		}
		// TODO: The case of handling the hash node is not implemented yet.
	}
	hasher := newHasher(false)
	defer returnHasherToPool(hasher)

	for i, n := range nodes {
		var hn node
		n, hn = hasher.proofHash(n)
		if hash, ok := hn.(hashNode); ok || i == 0 {
			// If the node's database encoding is a hash (or is the
			// root node), it becomes a proof element.
			enc := nodeToBytes(n)
			if !ok {
				hash = hasher.hashData(enc)
			}
			proofDb.Put(hash, enc)
		}
	}

	return nil
}

// Prove constructs a merkle proof for key. The result contains all encoded nodes
// on the path to the value at key. The value itself is also included in the last
// node and can be retrieved by verifying the proof.
//
// If the trie does not contain a value for key, the returned proof contains all
// nodes of the longest existing prefix of the key (at least the root node), ending
// with the node that proves the absence of the key.
// Original function: github.com/ethereum/go-ethereum/trie/proof.go line 110
func (t *StateTrie) Prove(key []byte, proofDb ethdb.KeyValueWriter) error {
	return t.trie.Prove(key, proofDb)
}

// VerifyProof checks merkle proofs. The given proof must contain the value for
// key in a trie with the given root hash. VerifyProof returns an error if the
// proof contains invalid trie nodes or the wrong value.
// Original function: github.com/ethereum/go-ethereum/trie/proof.go line 117
func VerifyProof(rootHash common.Hash, key []byte, proofDb ethdb.KeyValueReader) (value []byte, err error) {
	key = keybytesToHex(key)
	fmt.Println("VerifyProof key:", key)
	wantHash := rootHash
	for i := 0; ; i++ {
		buf, _ := proofDb.Get(wantHash[:])
		if buf == nil {
			return nil, fmt.Errorf("proof node %d (hash %064x) missing", i, wantHash)
		}
		n, err := decodeNode(wantHash[:], buf)
		if err != nil {
			return nil, fmt.Errorf("bad proof node %d: %v", i, err)
		}
		keyrest, cld := get(n, key, true)
		switch cld := cld.(type) {
		case nil:
			fmt.Println("Child node type is nil")
			// The trie doesn't contain the key.
			return nil, nil
		case hashNode:
			key = keyrest
			copy(wantHash[:], cld)
		case valueNode:
			return cld, nil
		}
	}
}

// get returns the child of the given node. Return nil if the
// node with specified key doesn't exist at all.
//
// There is an additional flag `skipResolved`. If it's set then
// all resolved nodes won't be returned.
// Original function: github.com/ethereum/go-ethereum/trie/proof.go line 588
func get(tn node, key []byte, skipResolved bool) ([]byte, node) {
	for {
		switch n := tn.(type) {
		case *shortNode:
			fmt.Println("Short node found with key:", n.Key)
			if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
				return nil, nil
			}
			tn = n.Val
			key = key[len(n.Key):]
			if !skipResolved {
				return key, tn
			}
		case *fullNode:
			tn = n.Children[key[0]]
			key = key[1:]
			if !skipResolved {
				return key, tn
			}
		case hashNode:
			return key, n
		case nil:
			return key, nil
		case valueNode:
			return nil, n
		default:
			panic(fmt.Sprintf("%T: invalid node: %v", tn, tn))
		}
	}
}
