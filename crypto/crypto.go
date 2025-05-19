/*
Package crypto implements the cryptographic functions used in Ethereum, specifically
the Keccak-256 hashing. This is a simplified version of Ethereum's crypto package
that focuses on the essential hashing operations.

Key features:
  - KeccakState interface that wraps sha3.state
  - Keccak256Hash function for computing Ethereum-style hashes
  - Direct state reading capability for better performance

The package uses the legacy Keccak implementation (not SHA3-256) to maintain
compatibility with Ethereum's hashing scheme. This is important because Ethereum
chose to use the original Keccak submission rather than the final FIPS-202
standard (SHA3).

Example usage:
    hash := crypto.Keccak256Hash([]byte("data"))
*/

package crypto

import (
	"hash"
	"storage_extract/common"

	"golang.org/x/crypto/sha3"
)

// KeccakState wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// NewKeccakState creates a new KeccakState
// Oringinal function: crypto/crypto.go line 71
func NewKeccakState() KeccakState {
	return sha3.NewLegacyKeccak256().(KeccakState)
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
// Original function: crypto/crypto.go line 84
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	b := make([]byte, 32)
	d := NewKeccakState()
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return h
}
