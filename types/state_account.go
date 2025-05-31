package types

import "storage_extract/common"

type StateAccount struct {
	Root common.Hash
}

// NewEmptyStateAccount creates a new empty state account with a zero root hash.
func NewEmptyStateAccount() *StateAccount {
	return &StateAccount{
		Root: EmptyRootHash,
	}
}
