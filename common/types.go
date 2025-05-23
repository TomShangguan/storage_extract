package common

import "github.com/ethereum/go-ethereum/common/hexutil"

// Lengths of hashes and addresses in bytes.
const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the address
	AddressLength = 20
)

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte

// Hash related functions

// BytesToHash sets b to hash.
// If b is larger than len(h), b will be cropped from the left.
// Original function: github.com/ethereum/go-ethereum/common/types.go line 60
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// BytesToHash converts a byte slice to a Hash.
// Original function: github.com/ethereum/go-ethereum/common/types.go line 72
func HexToHash(s string) Hash { return BytesToHash(FromHex(s)) }

// SetBytes sets the hash to the value of b.
// If b is larger than len(h), b will be cropped from the left.
// Original function: github.com/ethereum/go-ethereum/common/types.go line 147
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// Bytes gets the byte representation of the underlying hash.
func (h Hash) Bytes() []byte { return h[:] }

// Hex converts a hash to a hex string.
func (h Hash) Hex() string { return hexutil.Encode(h[:]) }

// /////////////////////////////////////////////////////////////////////////
// Address represents the 20 byte address of an Ethereum account.
type Address [AddressLength]byte

// Address related functions

// Bytes gets the string representation of the underlying address.
func (a Address) Bytes() []byte { return a[:] }

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func HexToAddress(s string) Address { return BytesToAddress(FromHex(s)) }

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a), b will be cropped from the left.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}
