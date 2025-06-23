package trienode

import (
	"fmt"
	"maps"
	"storage_extract/common"
)

// Node is a wrapper which contains the encoded blob of the trie node and its
// node hash. It is general enough that can be used to represent trie node
// corresponding to different trie implementations.
type Node struct {
	Hash common.Hash // Node hash, empty for deleted node
	Blob []byte      // Encoded node blob, nil for the deleted node
}

// New constructs a node with provided node information.
func New(hash common.Hash, blob []byte) *Node {
	return &Node{Hash: hash, Blob: blob}
}

// IsDeleted returns the indicator if the node is marked as deleted.
func (n *Node) IsDeleted() bool {
	return len(n.Blob) == 0
}

// leaf represents a trie leaf node
type leaf struct {
	Blob   []byte      // raw blob of leaf
	Parent common.Hash // the hash of parent node
}

// NodeSet contains a set of nodes collected during the commit operation.
// Each node is keyed by path. It's not thread-safe to use.
type NodeSet struct {
	Owner   common.Hash
	Leaves  []*leaf
	Nodes   map[string]*Node
	updates int // the count of updated and inserted nodes
	deletes int // the count of deleted nodes
}

// NewNodeSet initializes a node set. The owner is zero for the account trie and
// the owning account address hash for storage tries.
func NewNodeSet(owner common.Hash) *NodeSet {
	return &NodeSet{
		Owner: owner,
		Nodes: make(map[string]*Node),
	}
}

// MergedNodeSet represents a merged node set for a group of tries.
type MergedNodeSet struct {
	Sets map[common.Hash]*NodeSet
}

// NewMergedNodeSet initializes an empty merged set.
func NewMergedNodeSet() *MergedNodeSet {
	return &MergedNodeSet{Sets: make(map[common.Hash]*NodeSet)}
}

// Merge merges the provided dirty nodes of a trie into the set. The assumption
// is held that no duplicated set belonging to the same trie will be merged twice.
func (set *MergedNodeSet) Merge(other *NodeSet) error {
	subset, present := set.Sets[other.Owner]
	if present {
		fmt.Println("Owner of the node set, ", other.Owner, " already exist, merging with the presenting nodeset")
		return subset.Merge(other.Owner, other.Nodes)
	}
	fmt.Println("Adding the nodeset of account: ", other.Owner, " to the MergedNodeSet")
	set.Sets[other.Owner] = other
	return nil
}

func (set *NodeSet) AddNode(path []byte, n *Node) {
	if n.IsDeleted() {
		set.deletes += 1
	} else {
		set.updates += 1
	}
	set.Nodes[string(path)] = n
}

// AddLeaf adds the provided leaf node into set. (NOT USED FOR CURRENT IMPLEMENTATION)
func (set *NodeSet) AddLeaf(parent common.Hash, blob []byte) {
	set.Leaves = append(set.Leaves, &leaf{Blob: blob, Parent: parent})
}

// Size returns the number of dirty nodes in set.
func (set *NodeSet) Size() (int, int) {
	return set.updates, set.deletes
}

// MergeSet merges this 'set' with 'other'. It assumes that the sets are disjoint,
// and thus does not deduplicate data (count deletes, dedup leaves etc).
func (set *NodeSet) MergeSet(other *NodeSet) error {
	if set.Owner != other.Owner {
		return fmt.Errorf("nodesets belong to different owner are not mergeable %x-%x", set.Owner, other.Owner)
	}
	maps.Copy(set.Nodes, other.Nodes)

	set.deletes += other.deletes
	set.updates += other.updates

	// Since we assume the sets are disjoint, we can safely append leaves
	// like this without deduplication.
	set.Leaves = append(set.Leaves, other.Leaves...)
	return nil
}

// Merge adds a set of nodes into the set.
func (set *NodeSet) Merge(owner common.Hash, nodes map[string]*Node) error {
	if set.Owner != owner {
		return fmt.Errorf("nodesets belong to different owner are not mergeable %x-%x", set.Owner, owner)
	}
	for path, node := range nodes {
		prev, ok := set.Nodes[path]
		if ok {
			// overwrite happens, revoke the counter
			if prev.IsDeleted() {
				set.deletes -= 1
			} else {
				set.updates -= 1
			}
		}
		if node.IsDeleted() {
			set.deletes += 1
		} else {
			set.updates += 1
		}
		set.Nodes[path] = node
	}
	return nil
}
