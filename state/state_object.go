package state

import (
	"storage_extract/common"
	"storage_extract/crypto"
	"storage_extract/types"
)

// Storage represents a map of storage keys to their values.
type Storage map[common.Hash]common.Hash

// StateObject represents an Ethereum account in the state database.
// Differ from the original code, the struct is public to allow access for testing.
type StateObject struct {
	db           *StateDB
	address      common.Address      // address of ethereum account
	addrHash     common.Hash         // hash of the address
	origin       *types.StateAccount // original state account
	dirtyStorage Storage             // dirty storage changes
}

// newObject creates a new state object with the given address and account.
// If the account is nil, it creates a new empty state account.
// The address is hashed using Keccak256 to create the addrHash.
func newObject(db *StateDB, addr common.Address, acct *types.StateAccount) *StateObject {
	origin := acct
	if acct == nil {
		acct = types.NewEmptyStateAccount()
	}
	return &StateObject{
		db:           db,
		address:      addr,
		addrHash:     crypto.Keccak256Hash(addr[:]),
		origin:       origin,
		dirtyStorage: make(Storage),
	}
}

// getState retrieves a value associated with the given storage key, along with
// its original value.
// Original function: github.com/ethereum/go-ethereum/core/state/state_object.go line 160
func (s *StateObject) getState(key common.Hash) (common.Hash, common.Hash) {
	origin := s.GetCommittedState(key)
	value, dirty := s.dirtyStorage[key]
	if dirty {
		return value, origin
		// If the value is dirty, return the dirty value and the original value.
		// The dirty value is the one that has been modified but not yet committed.
	}
	return origin, origin // If the value is not dirty, return the original value
	// Why two origins?
	// The first origin is the previous value from the committed state.
	// If the value isn't dirty, the second origin is the same as the first.
}

// GetCommittedState retrieves the value associated with the specific key
// without any mutations caused in the current execution.
// Orignal function: github.com/ethereum/go-ethereum/core/state/state_object.go line 170
func (s *StateObject) GetCommittedState(key common.Hash) common.Hash {
	// TODO: follow the original code and read from the database
	// The current implementation only plays a role of a placeholder.
	return common.Hash{}
}

// SetState updates a value in account storage.
// It returns the previous value
// Original function: github.com/ethereum/go-ethereum/core/state/state_object.go line 214
func (s *StateObject) SetState(key, value common.Hash) common.Hash {
	// If the new value is the same as old, don't set. Otherwise, track only the
	// dirty changes, supporting reverting all of it back to no change.
	prev, origin := s.getState(key)
	if prev == value {
		return prev
	}
	s.db.journal.storageChange(s.address, key, prev, origin)
	s.setState(key, value, origin)
	return prev
}

// setState updates a value in account dirty storage. The dirtiness will be
// removed if the value being set equals to the original value.
// Original function: github.com/ethereum/go-ethereum/core/state/state_object.go line 232
func (s *StateObject) setState(key common.Hash, value common.Hash, origin common.Hash) {
	// Storage slot is set back to its original value, undo the dirty marker
	if value == origin {
		delete(s.dirtyStorage, key)
		return
	}
	s.dirtyStorage[key] = value
}
