package trie

import (
	"bytes"
	"fmt"
	"storage_extract/common"
	"storage_extract/types"
)

// Trie is a Merkle Patricia Trie. Use New to create a trie that sits on
// top of a database. Whenever trie performs a commit operation, the generated
// nodes will be gathered and returned in a set. Once the trie is committed,
// it's not usable anymore. Callers have to re-create the trie with new root
// based on the updated trie database.
type Trie struct {
	root  node
	owner common.Hash

	// Keep track of the number leaves which have been inserted since the last
	// hashing operation. This number will not directly map to the number of
	// actually unhashed nodes.
	unhashed int

	uncommitted int // uncommitted is the number of updates since last commit.
}

// newFlag returns the cache flag value for a newly created node.
func (t *Trie) newFlag() nodeFlag {
	return nodeFlag{dirty: true}
}

// New creates the trie instance with provided trie id and the read-only
// database.
// TODO: The current implementation doesn't use the database as a parameter.
// Thus, it doesn't support:
// 1. provide a reader
// 2. read trie node from the database to initialize the root node
func New(id *ID) (*Trie, error) {
	trie := &Trie{
		owner: id.Owner,
	}
	return trie, nil
}

// Different from the original code, missing the commit check logic
func (t *Trie) Update(key, value []byte) error {
	return t.update(key, value)
}

// The current implementation use a dummy function func (t *Trie) Update(key, value []byte) error
// Original function: github.com/ethereum/go-ethereum/trie/trie.go line 314
func (t *Trie) update(key, value []byte) error {
	t.unhashed++
	t.uncommitted++
	k := keybytesToHex(key)
	if len(value) != 0 {
		_, n, err := t.insert(t.root, nil, k, valueNode(value))
		if err != nil {
			return err
		}
		t.root = n
	}

	// TODO: Implement the logic for delete

	return nil
}

func (t *Trie) insert(n node, prefix, key []byte, value node) (bool, node, error) {
	// Base case: if key is empty, we've reached where value should be inserted
	if len(key) == 0 {
		if v, ok := n.(valueNode); ok {
			// If there's already a value, only return true if new value is different
			return !bytes.Equal(v, value.(valueNode)), value, nil
		}
		return true, value, nil
	}
	switch n := n.(type) {
	case *shortNode:
		// Find the length of common prefix between path and key
		matchlen := prefixLen(key, n.Key)

		if matchlen == len(n.Key) {
			// The entire path matches, recurse with remaining key
			// Example:
			//   Existing: shortNode("hello") -> node1
			//   Inserting: "hello-world" -> value2
			//   matchlen = 5 (full match of "hello")
			//   Action: Recurse with remaining path "-world"
			dirty, nn, err := t.insert(n.Val, append(prefix, key[:matchlen]...), key[matchlen:], value)
			if !dirty || err != nil {
				return false, n, err
			}
			return true, &shortNode{n.Key, nn, t.newFlag()}, nil
		}
		// Otherwise branch out at the index where they differ.
		// Example:
		//   Existing: shortNode("hello") -> value1
		//   Inserting: "help" -> value2
		//   matchlen = 3 (matched "hel")
		//   Action: Create branch (full) node at 'l' vs 'p'
		branch := &fullNode{flags: t.newFlag()}
		var err error
		_, branch.Children[n.Key[matchlen]], err = t.insert(nil, append(prefix, n.Key[:matchlen+1]...), n.Key[matchlen+1:], n.Val)
		if err != nil {
			return false, nil, err
		}
		_, branch.Children[key[matchlen]], err = t.insert(nil, append(prefix, key[:matchlen+1]...), key[matchlen+1:], value)
		if err != nil {
			return false, nil, err
		}
		// Replace this shortNode with the branch if it occurs at index 0. (No common prefix)
		if matchlen == 0 {
			return true, branch, nil
		}
		// New branch node is created as a child of the original short node.
		// Result structure for "hello"/"help" example:
		//   shortNode("hel")
		//      └── branch
		//          ├── 'l' -> shortNode("o") -> value1
		//          └── 'p' -> value2

		// TODO: tracer logic to track the changes of the trie

		// Replace it with a short node leading up to the branch.
		return true, &shortNode{key[:matchlen], branch, t.newFlag()}, nil
	case *fullNode:
		// fullNode has exactly 16 children, one for each hex character (0-f)
		// Structure of fullNode:
		//   [0] -> node0     [1] -> node1 ... [e] -> nodeE    [f] -> nodeF

		// Use first hex char of key as index into children array
		dirty, nn, err := t.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.flags = t.newFlag()
		n.Children[key[0]] = nn
		return true, n, nil
	case nil:
		// Inserting into an empty trie or reaching an empty slot in a branch node
		// Example 1 - Empty trie:
		//   Before: root = nil
		//   Insert: "abc" -> value1
		//   After:  root = shortNode("abc" -> value1)
		//
		// Example 2 - Empty slot in branch node:
		//   Before: fullNode
		//           └── [7] -> shortNode("xyz") -> value1
		//   Insert: "5abc" -> value2
		//   After:  fullNode
		//           ├── [5] -> shortNode("abc") -> value2
		//           └── [7] -> shortNode("xyz") -> value1

		// TODO: tracer logic to track the changes of the trie
		return true, &shortNode{key, value, t.newFlag()}, nil
	// case hashNode: is ignored for now as we don't hit part of the trie that's not loaded

	default:
		panic(fmt.Sprintf("%T: invalid node: %v", n, n))
	}
}

// Hash returns the root hash of the trie. It does not write to the
// database and can be used even if the trie doesn't have one.
// Original function: github.com/ethereum/go-ethereum/trie/trie.go line 609
func (t *Trie) Hash() common.Hash {
	hash, cached := t.hashRoot()
	t.root = cached
	return common.BytesToHash(hash.(hashNode))
}

// hashRoot calculates the root hash of the given trie
// Oringinal function: github.com/ethereum/go-ethereum/trie/trie.go line 663
func (t *Trie) hashRoot() (node, node) {
	if t.root == nil {
		return hashNode(types.EmptyRootHash.Bytes()), nil
	}
	// If the number of changes is below 100, we let one thread handle it
	h := newHasher(t.unhashed >= 100)
	defer func() {
		returnHasherToPool(h)
		t.unhashed = 0
	}()
	hashed, cached := h.hash(t.root, true)
	return hashed, cached
}
