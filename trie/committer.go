/*
The commit process is a critical operation in Ethereum's MPT that transforms an in-memory
trie structure into a persistent, cryptographically-verifiable form.

=== OVERVIEW ===

The commit operation serves multiple purposes:
1. Converts dirty (modified) nodes into their hash representations
2. Collects all modified nodes into a NodeSet for database persistence
3. Optimizes storage by embedding small nodes directly in their parents
4. Maintains cryptographic integrity through recursive hashing

=== COMMIT WORKFLOW ===

1. **Cache Check**: Before processing any node, the committer checks:
   - Does the node already have a cached hash?
   - Is the node marked as dirty (modified)?
   - If clean hash exists, return it immediately (optimization)

2. **Node Type Processing**: The commit process handles different node types:

   a) **shortNode Processing**:
      - Creates a collapsed copy for storage
      - Recursively commits child nodes if they are fullNodes
      - Converts hex keys to compact format for RLP encoding
      - Stores the node and returns hash or embedded node

   b) **fullNode Processing**:
      - Commits all 17 children through `commitChildren()`
      - Supports parallel processing for performance
      - Creates collapsed copy with committed children
      - Stores the complete node structure

   c) **hashNode Processing**:
      - Already committed nodes are returned as-is
      - No further processing required

3. **Children Commitment** (`commitChildren`):
   - Processes the 16 hex branches (0-F) and optional value slot
   - Handles three child types:
     * nil: Skipped (empty slot)
     * hashNode: Used directly (already committed)
     * Other nodes: Recursively committed
   - Parallel mode: Creates independent committers for concurrent processing
   - Serial mode: Sequential processing with shared committer

4. **Storage Decision** (`store`):
   The storage process implements a critical optimization:

   a) **Small Node Embedding** (< 32 bytes):
      - Nodes with no cached hash are considered "small"
      - These nodes are embedded directly in their parent
      - No separate database storage required
      - Reduces storage overhead and improves access speed

   b) **Large Node Hashing** (≥ 32 bytes):
      - Nodes are RLP-encoded and hashed
      - Hash becomes the node's identifier
      - Full node data stored in database
      - Parent stores only the 32-byte hash reference

5. **NodeSet Management**:
   - All committed nodes are collected in a NodeSet
   - NodeSet tracks ownership (trie owner address)
   - Supports merging multiple NodeSets (for parallel processing)
   - Provides interface for database persistence

=== PARALLEL PROCESSING ===

The committer supports parallel processing for fullNode children:
- Each child gets an independent committer and NodeSet
- Results are safely merged using mutex protection
- Significantly improves performance for large tries
- Maintains consistency through proper synchronization

*/

package trie

import (
	"fmt"
	"storage_extract/common"
	"storage_extract/trie/trienode"
	"sync"
)

// committer is the tool used for the trie Commit operation. The committer will
// capture all dirty nodes during the commit process and keep them cached in
// insertion order.
type committer struct {
	nodes *trienode.NodeSet
	//tracer      *tracer
	collectLeaf bool
}

// newCommitter creates a new committer or picks one from the pool.
// Original function: github.com/ethereum/go-ethereum/trie/committer.go line 38
func newCommitter(nodeset *trienode.NodeSet, collectLeaf bool) *committer {
	return &committer{
		nodes: nodeset,
		//tracer:      tracer,
		collectLeaf: collectLeaf,
	}
}

// Commit collapses a node down into a hash node.
func (c *committer) Commit(n node, parallel bool) hashNode {
	fmt.Println("Committer.go - Commit")
	return c.commit(nil, n, parallel).(hashNode)
}

// commit collapses a node down into a hash node and returns it.
// Original function: github.com/ethereum/go-ethereum/trie/committer.go line 58
func (c *committer) commit(path []byte, n node, parallel bool) node {
	// if this path is clean, use available cached data
	fmt.Println("committing node, owner: ", c.nodes.Owner, " path: ", path, " node: ", n)
	hash, dirty := n.cache()
	if hash != nil && !dirty {
		return hash
	}
	// Commit children, then parent, and remove the dirty flag.
	switch cn := n.(type) {
	case *shortNode:
		// Commit child
		collapsed := cn.copy()

		// If the child is fullNode, recursively commit,
		// otherwise it can only be hashNode or valueNode.
		if _, ok := cn.Val.(*fullNode); ok {
			collapsed.Val = c.commit(append(path, cn.Key...), cn.Val, false)
		}
		// The key needs to be copied, since we're adding it to the
		// modified nodeset.
		collapsed.Key = hexToCompact(cn.Key)
		hashedNode := c.store(path, collapsed)
		if hn, ok := hashedNode.(hashNode); ok {
			return hn
		}
		return collapsed
	case *fullNode:
		hashedKids := c.commitChildren(path, cn, parallel)
		collapsed := cn.copy()
		collapsed.Children = hashedKids

		hashedNode := c.store(path, collapsed)
		if hn, ok := hashedNode.(hashNode); ok {
			return hn
		}
		return collapsed
	case hashNode:
		return cn
	default:
		// nil, valuenode shouldn't be committed
		panic(fmt.Sprintf("%T: invalid node: %v", n, n))
	}
}

// commitChildren commits the children of the given fullnode
// Original function: github.com/ethereum/go-ethereum/trie/committer.go line 98
func (c *committer) commitChildren(path []byte, n *fullNode, parallel bool) [17]node {
	var (
		wg       sync.WaitGroup
		nodesMu  sync.Mutex
		children [17]node
	)
	for i := 0; i < 16; i++ {
		child := n.Children[i]
		if child == nil {
			continue
		}
		// If it's the hashed child, save the hash value directly.
		// Note: it's impossible that the child in range [0, 15]
		// is a valueNode.
		if hn, ok := child.(hashNode); ok {
			children[i] = hn
			continue
		}
		// Commit the child recursively and store the "hashed" value.
		// Note the returned node can be some embedded nodes, so it's
		// possible the type is not hashNode.
		if !parallel {
			children[i] = c.commit(append(path, byte(i)), child, false)
		} else {
			wg.Add(1)
			go func(index int) {
				p := append(path, byte(index))
				childSet := trienode.NewNodeSet(c.nodes.Owner)
				childCommitter := newCommitter(childSet, c.collectLeaf)
				children[index] = childCommitter.commit(p, child, false)
				nodesMu.Lock()
				c.nodes.MergeSet(childSet)
				nodesMu.Unlock()
				wg.Done()
			}(i)
		}
	}
	if parallel {
		wg.Wait()
	}
	// For the 17th child, it's possible the type is valuenode.
	if n.Children[16] != nil {
		children[16] = n.Children[16]
	}
	return children
}

// store hashes the node n and adds it to the modified nodeset. If leaf collection
// is enabled, leaf nodes will be tracked in the modified nodeset as well.
// Original function: github.com/ethereum/go-ethereum/trie/committer.go line 147
func (c *committer) store(path []byte, n node) node {
	// Larger nodes are replaced by their hash and stored in the database.
	var hash, _ = n.cache()

	// This was not generated - must be a small node stored in the parent.
	// In theory, we should check if the node is leaf here (embedded node
	// usually is leaf node). But small value (less than 32bytes) is not
	// our target (leaves in account trie only).
	if hash == nil {
		// The node is embedded in its parent, in other words, this node
		// will not be stored in the database independently, mark it as
		// deleted only if the node was existent in database before.
		// TODO: Logic of Tracer
		return n
	}
	// Collect the dirty node to nodeset for return.
	//fmt.Println("Owner: ", c.nodes.Owner, "Added node at path %x with hash %x to nodeset (before hash).", path, hash)
	nhash := common.BytesToHash(hash)
	c.nodes.AddNode(path, trienode.New(nhash, nodeToBytes(n)))
	//fmt.Println("Owner: ", c.nodes.Owner, "Added node at path %x with hash %x to nodeset.", path, nhash.Hex())
	// Collect the corresponding leaf node if it's required. We don't check
	// full node since it's impossible to store value in fullNode. The key
	// length of leaves should be exactly same.
	if c.collectLeaf {
		if sn, ok := n.(*shortNode); ok {
			if val, ok := sn.Val.(valueNode); ok {
				c.nodes.AddLeaf(nhash, val)
			}
		}
	}
	return hash
}
