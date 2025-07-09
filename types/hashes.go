package types

import (
	"fmt"
	"storage_extract/common"
)

var (
	// EmptyRootHash is the hash of an empty state trie.
	EmptyRootHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

// TrieRootHash returns the hash itself if it's non-empty or the predefined
// emptyHash one instead.
func TrieRootHash(hash common.Hash) common.Hash {
	if hash == (common.Hash{}) {
		fmt.Println("Zero trie root hash!")
		return EmptyRootHash
	}
	return hash
}
