package trie

// TODO: Replace the RLP package with local implementation to avoid dependency on go-ethereum

import (
	"storage_extract/crypto"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

type hasher struct {
	sha      crypto.KeccakState
	tmp      []byte
	encbuf   rlp.EncoderBuffer
	parallel bool // Whether to use parallel threads when hashing
}

// hasherPool holds pureHashers
var hasherPool = sync.Pool{
	New: func() interface{} {
		return &hasher{
			tmp:    make([]byte, 0, 550), // cap is as large as a full fullNode.
			sha:    crypto.NewKeccakState(),
			encbuf: rlp.NewEncoderBuffer(nil),
		}
	},
}

func returnHasherToPool(h *hasher) {
	hasherPool.Put(h)
}

func newHasher(parallel bool) *hasher {
	h := hasherPool.Get().(*hasher)
	h.parallel = parallel
	return h
}

// hash collapses a node down into a hash node, also returning a copy of the
// original node initialized with the computed hash to replace the original one.
// Original function: github.com/ethereum/go-ethereum/trie/secure_trie.go line 58
func (h *hasher) hash(n node, force bool) (hashed node, cached node) {
	// Return the cached hash if it's available
	if hash, _ := n.cache(); hash != nil {
		return hash, n
	}
	// Trie not processed yet, walk the children
	switch n := n.(type) {
	case *shortNode:
		collapsed, cached := h.hashShortNodeChildren(n)
		hashed := h.shortnodeToHash(collapsed, force)
		// We need to retain the possibly _not_ hashed node, in case it was too
		// small to be hashed
		if hn, ok := hashed.(hashNode); ok {
			cached.flags.hash = hn
		} else {
			cached.flags.hash = nil
		}
		return hashed, cached
	case *fullNode:
		collapsed, cached := h.hashFullNodeChildren(n)
		hashed = h.fullnodeToHash(collapsed, force)
		if hn, ok := hashed.(hashNode); ok {
			cached.flags.hash = hn
		} else {
			cached.flags.hash = nil
		}
		return hashed, cached
	default:
		// Value and hash nodes don't have children, so they're left as were
		return n, n
	}
}

// hashShortNodeChildren collapses the short node. The returned collapsed node
// holds a live reference to the Key, and must not be modified.
// Original function: github.com/ethereum/go-ethereum/trie/secure_trie.go line 93
func (h *hasher) hashShortNodeChildren(n *shortNode) (collapsed, cached *shortNode) {
	// Hash the short node's child, caching the newly hashed subtree
	collapsed, cached = n.copy(), n.copy()
	// Previously, we did copy this one. We don't seem to need to actually
	// do that, since we don't overwrite/reuse keys
	// cached.Key = common.CopyBytes(n.Key)
	collapsed.Key = hexToCompact(n.Key)
	// Unless the child is a valuenode or hashnode, hash it
	switch n.Val.(type) {
	case *fullNode, *shortNode:
		collapsed.Val, cached.Val = h.hash(n.Val, false)
	}
	return collapsed, cached
}

func (h *hasher) hashFullNodeChildren(n *fullNode) (collapsed *fullNode, cached *fullNode) {
	// Hash the full node's children, caching the newly hashed subtrees
	cached = n.copy()
	collapsed = n.copy()
	if h.parallel {
		var wg sync.WaitGroup
		wg.Add(16)
		for i := 0; i < 16; i++ {
			go func(i int) {
				hasher := newHasher(false)
				if child := n.Children[i]; child != nil {
					collapsed.Children[i], cached.Children[i] = hasher.hash(child, false)
				} else {
					collapsed.Children[i] = nilValueNode
				}
				returnHasherToPool(hasher)
				wg.Done()
			}(i)
		}
		wg.Wait()
	} else {
		for i := 0; i < 16; i++ {
			if child := n.Children[i]; child != nil {
				collapsed.Children[i], cached.Children[i] = h.hash(child, false)
			} else {
				collapsed.Children[i] = nilValueNode
			}
		}
	}
	return collapsed, cached
}

// shortnodeToHash creates a hashNode from a shortNode. The supplied shortnode
// should have hex-type Key, which will be converted (without modification)
// into compact form for RLP encoding.
// If the rlp data is smaller than 32 bytes, `nil` is returned.
func (h *hasher) shortnodeToHash(n *shortNode, force bool) node {
	n.encode(h.encbuf)
	enc := h.encodedBytes()

	if len(enc) < 32 && !force {
		return n // Nodes smaller than 32 bytes are stored inside their parent
	}
	return h.hashData(enc)
}

// fullnodeToHash is used to create a hashNode from a fullNode, (which
// may contain nil values)
func (h *hasher) fullnodeToHash(n *fullNode, force bool) node {
	n.encode(h.encbuf)
	enc := h.encodedBytes()

	if len(enc) < 32 && !force {
		return n // Nodes smaller than 32 bytes are stored inside their parent
	}
	return h.hashData(enc)
}

// encodedBytes returns the result of the last encoding operation on h.encbuf.
// This also resets the encoder buffer.
//
// All node encoding must be done like this:
//
//	node.encode(h.encbuf)
//	enc := h.encodedBytes()
//
// This convention exists because node.encode can only be inlined/escape-analyzed when
// called on a concrete receiver type.
func (h *hasher) encodedBytes() []byte {
	h.tmp = h.encbuf.AppendToBytes(h.tmp[:0])
	h.encbuf.Reset(nil)
	return h.tmp
}

// hashData hashes the provided data
func (h *hasher) hashData(data []byte) hashNode {
	n := make(hashNode, 32)
	h.sha.Reset()
	h.sha.Write(data)
	h.sha.Read(n)
	return n
}

// proofHash is used to construct trie proofs, and returns the 'collapsed'
// node (for later RLP encoding) as well as the hashed node -- unless the
// node is smaller than 32 bytes, in which case it will be returned as is.
// This method does not do anything on value- or hash-nodes.
func (h *hasher) proofHash(original node) (collapsed, hashed node) {
	switch n := original.(type) {
	case *shortNode:
		sn, _ := h.hashShortNodeChildren(n)
		return sn, h.shortnodeToHash(sn, false)
	case *fullNode:
		fn, _ := h.hashFullNodeChildren(n)
		return fn, h.fullnodeToHash(fn, false)
	default:
		// Value and hash nodes don't have children, so they're left as were
		return n, n
	}
}
