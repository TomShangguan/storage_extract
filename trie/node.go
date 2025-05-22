package trie

import "github.com/ethereum/go-ethereum/rlp"

type node interface {
	cache() (hashNode, bool)
	encode(w rlp.EncoderBuffer)
}

type (
	fullNode struct {
		Children [17]node
		flags    nodeFlag
	}
	shortNode struct {
		Key   []byte
		Val   node
		flags nodeFlag
	}
	hashNode  []byte
	valueNode []byte
)

// nodeFlag contains caching-related metadata about a node.
type nodeFlag struct {
	hash  hashNode // cached hash of the node (may be nil)
	dirty bool     // whether the node has changes that must be written to the database
}

// nilValueNode is used when collapsing internal trie nodes for hashing, since
// unset children need to serialize correctly.
var nilValueNode = valueNode(nil)

func (n *fullNode) copy() *fullNode   { copy := *n; return &copy }
func (n *shortNode) copy() *shortNode { copy := *n; return &copy }

func (n *fullNode) cache() (hashNode, bool)  { return n.flags.hash, n.flags.dirty }
func (n *shortNode) cache() (hashNode, bool) { return n.flags.hash, n.flags.dirty }
func (n hashNode) cache() (hashNode, bool)   { return nil, true }
func (n valueNode) cache() (hashNode, bool)  { return nil, true }
