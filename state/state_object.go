package state

import (
	"fmt"
	"storage_extract/common"
	"storage_extract/crypto"
	"storage_extract/types"
)

// Storage represents a map of storage keys to their values.
type Storage map[common.Hash]common.Hash

// StateObject represents an Ethereum account in the state database.
// Differ from the original code, the struct is public to allow access for testing.
type StateObject struct {
	db       *StateDB
	address  common.Address      // address of ethereum account
	addrHash common.Hash         // hash of the address
	origin   *types.StateAccount // original state account
	data     types.StateAccount  // Account data with all mutations applied in the scope of block

	trie Trie // storage trie, which becomes non-nil on first access

	dirtyStorage   Storage // dirty storage changes
	pendingStorage Storage // Storage entries that have been modified within the current block

	// uncommittedStorage tracks a set of storage entries that have been modified
	// but not yet committed since the "last commit operation", along with their
	// original values before mutation.
	//
	// Specifically, the commit will be performed after each transaction before
	// the byzantium fork, therefore the map is already reset at the transaction
	// boundary; however post the byzantium fork, the commit will only be performed
	// at the end of block, this set essentially tracks all the modifications
	// made within the block.

	uncommittedStorage Storage
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
		db:                 db,
		address:            addr,
		addrHash:           crypto.Keccak256Hash(addr[:]),
		origin:             origin,
		data:               *acct,
		dirtyStorage:       make(Storage),
		uncommittedStorage: make(Storage),
		pendingStorage:     make(Storage),
	}
}

// getTrie returns the associated storage trie. The trie will be opened if it's
// not loaded previously. An error will be returned if trie can't be loaded.
//
// TODO: Support read trie from the database and triedb parameter.
// Original function: github.com/ethereum/go-ethereum/core/state/state_object.go line 124
func (s *StateObject) getTrie() (Trie, error) {
	if s.trie == nil {
		fmt.Println("Opening storage trie for address:", s.address)
		tr, err := s.db.db.OpenStorageTrie(s.db.originalRoot, s.address, s.data.Root)
		if err != nil {
			return nil, err
		}
		s.trie = tr
	}
	return s.trie, nil
}

// GetState retrieves a value associated with the given storage key.
// Original function: github.com/ethereum/go-ethereum/core/state/state_object.go line 154
func (s *StateObject) GetState(key common.Hash) common.Hash {
	value, _ := s.getState(key)
	return value
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
	if value, pending := s.pendingStorage[key]; pending {
		return value
	}
	// TODO: follow the original code and read from the database, the current implementation only read from the pendingStorage
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
		fmt.Println("Setting storage for key:", key, "Value:", value, "Origin:", origin, "No change detected.")
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
		fmt.Println("Setting storage for key:", key, "Value:", value, "Origin:", origin)
		delete(s.dirtyStorage, key)
		return
	}
	fmt.Println("Setting storage for key:", key, "Value:", value, "Origin:", origin)
	s.dirtyStorage[key] = value
}

// finalise moves all dirty storage slots into the pending area to be hashed or
// committed later. It is invoked at the end of every transaction.
// Original function: github.com/ethereum/go-ethereum/core/state/state_object.go line 245
func (s *StateObject) finalise() {
	//TODO: Prefetch the slot from the database
	for key, value := range s.dirtyStorage {
		if origin, exist := s.uncommittedStorage[key]; exist && origin == value {
			// The slot is reverted to its original value, delete the entry
			// to avoid thrashing the data structures.
			fmt.Println("Reverting storage for key:", key, "Value:", value, "Origin:", origin)
			delete(s.uncommittedStorage, key)
		} else if exist {
			// The slot is modified to another value and the slot has been
			// tracked for commit, do nothing here.
		} else {
			// The slot is different from its original value and hasn't been
			// tracked for commit yet.
			// TODO: Get committed state from the database
			s.uncommittedStorage[key] = common.Hash{}
		}
		// Aggregate the dirty storage slots into the pending area. It might
		// be possible that the value of tracked slot here is same with the
		// one in originStorage (e.g. the slot was modified in tx_a and then
		// modified back in tx_b). We can't blindly remove it from pending
		// map as the dirty slot might have been committed already
		// and entry is necessary to modify the value back.
		fmt.Println("Finalising storage for key:", key, "Value:", value)
		s.pendingStorage[key] = value

	}
	// TODO: Prefetch logic
	if len(s.dirtyStorage) > 0 {
		s.dirtyStorage = make(Storage) // Reset the dirty storage
	}

	// s.newContract = false (may not needed)
}

// updateTrie is responsible for persisting cached storage changes into the
// object's storage trie. In case the storage trie is not yet loaded, this
// function will load the trie automatically. If any issues arise during the
// loading or updating of the trie, an error will be returned. Furthermore,
// this function will return the mutated storage trie, or nil if there is no
// storage change at all.
// It assumes all the dirty storage slots have been finalized (moved to pendingStorage) before.
// Original function: github.com/ethereum/go-ethereum/core/state/state_object.go line 295
func (s *StateObject) updateTrie() (Trie, error) {
	// The logic is different from the original code where it checks witness of db

	if len(s.uncommittedStorage) == 0 {
		// Short circuit if nothing changed, don't bother with hashing anything
		return s.trie, nil
	}
	var err error
	tr, err := s.getTrie()
	if err != nil {
		// TODO: handle error (db)
		return nil, err
	}
	// 	var (
	// 	deletions []common.Hash
	// 	used      = make([]common.Hash, 0, len(s.uncommittedStorage))
	// ) (Not used in the current implementation)

	// The logic of handling updates is different from the original code.
	// The process of checking whether the value is same as the original value is ignored for now.
	for key, origin := range s.uncommittedStorage {
		value, exist := s.pendingStorage[key]
		fmt.Println("Updating storage for key:", key, "Origin:", origin, "Value:", value, "Exist:", exist)
		if value == origin {
			continue
		}
		if !exist {
			continue
		}
		if (value != common.Hash{}) {
			if err := tr.UpdateStorage(s.address, key[:], common.TrimLeftZeroes(value[:])); err != nil {
				// TODO: handle error (s.db.setError(err))
				return nil, err
			}
		}
	}
	// TODO: handle deletions
	s.uncommittedStorage = make(Storage) // empties the commit markers
	return tr, nil
}

// updateRoot flushes all cached storage mutations to trie, recalculating the
// new storage trie root.
// Original function: github.com/ethereum/go-ethereum/core/state/state_object.go line 382
func (s *StateObject) updateRoot() {
	// Flush cached storage mutations into trie, short circuit if any error
	// is occurred or there is no change in the trie.
	tr, err := s.updateTrie()
	if err != nil || tr == nil {
		return
	}

	// Print the trie structure for visualization and debugging
	// This will show the complete structure of the MPT after updates
	if tr != nil {
		fmt.Printf("\n==== Storage Trie for Account %x ====\n", s.address)
		tr.PrintTrie()
	}
	s.data.Root = tr.Hash()
}

//------------------------------------------------------------------------------------------------------------------------
// Below are the additional methods that are not part of the original code but used in the test code snippet.

// GetAddress returns the address of the state object.
func (s *StateObject) GetTrie() *Trie {
	return &s.trie
}

// GetAddress returns the address of the state object.
func (s *StateObject) GetRoot() common.Hash {
	return s.data.Root
}
