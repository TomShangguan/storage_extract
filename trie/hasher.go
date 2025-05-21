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
