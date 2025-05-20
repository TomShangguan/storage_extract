package state

import (
	"storage_extract/common"
)

type StateDB struct {
	stateObjects map[common.Address]*StateObject
	journal      *journal
}

// SetState sets the state of the given address and key to the given value.
// It retrieves the state object for the address, and if it doesn't exist, it creates a new one.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 450
func (s *StateDB) SetState(addr common.Address, key, value common.Hash) common.Hash {
	if stateObject := s.getOrNewStateObject(addr); stateObject != nil {
		return stateObject.SetState(key, value)
	}
	return common.Hash{}
}

// getStateObject retrieves the state object for the given address.
// Orginal function: github.com/ethereum/go-ethereum/core/state/statedb.go line 573
func (s *StateDB) getStateObject(addr common.Address) *StateObject {
	if obj := s.stateObjects[addr]; obj != nil {
		return obj
	}
	// TODO: The current implementation only checks the stateObjects map.
	// in the original code, it reads from the database if the object is not found.
	// If the account found in the database, it will be added to the prefetch list and insert into the stateObjects map.
	return nil
}

// getOrNewStateObject retrieves a state object or create a new state object if nil.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 613
func (s *StateDB) getOrNewStateObject(addr common.Address) *StateObject {
	obj := s.getStateObject(addr)
	if obj == nil {
		obj = s.createObject(addr)
	}
	return obj
}

// createObject creates a new state object. The assumption is held there is no
// existing account with the given address, otherwise it will be silently overwritten.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 622
func (s *StateDB) createObject(addr common.Address) *StateObject {
	obj := newObject(s, addr, nil)
	//TODO: s.journal.createObject(addr)
	// Current implementation: add the object to the stateObjects map
	// This is different from the original code,
	// where calls another function (setStateObject) to add the object to the stateObjects map.
	// Here, directly add it to the map.
	s.stateObjects[addr] = obj
	return obj
}

// Finalise finalises the state by removing the destructed objects and clears
// the journal as well as the refunds. Finalise, however, will not push any updates
// into the tries just yet. Only IntermediateRoot or Commit will do that.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 730
func (s *StateDB) Finalise(deleteEmptyObjects bool) {
	//TODO: Prefetch the addresses
	for addr := range s.journal.dirties {
		obj, exist := s.stateObjects[addr]
		if !exist {
			continue
		}
		// TODO: delete logic
		obj.finalise()
		// TODO: s.markUpdate(addr) may not needed
	}
	// TODO: Clear the journal
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 774
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.Finalise(deleteEmptyObjects)
	// TODO: Intermediate processing
	return common.Hash{}
}

// commit gathers the state mutations accumulated along with the associated
// trie changes, resetting all internal flags with the new state as the base.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 1100
func (s *StateDB) commit(deleteEmptyObjects bool) (*stateUpdate, error) {
	// TODO: Error check of db before executing the commit
	s.IntermediateRoot(deleteEmptyObjects)
	// TODO: Intermediate processing
	return nil, nil
}

// commitAndFlush is a wrapper of commit which also commits the state mutations
// to the configured data stores.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 1260
func (s *StateDB) commitAndFlush(block uint64, deleteEmptyObjects bool) (*stateUpdate, error) {
	ret, err := s.commit(deleteEmptyObjects)
	// TODO: Intermediate processing
	return ret, err
}

// Commit writes the state mutations into the configured data stores.
// Once the state is committed, tries cached in stateDB (including account
// trie, storage tries) will no longer be functional. A new state instance
// must be created with new root and updated database for accessing post-
// commit states.
// The associated block number of the state transition is also provided
// for more chain context.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 1317
// TODO: Current implementation doesn't support deleteEmptyObjects. Now, it's a placeholder.
func (s *StateDB) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	// Placeholder
	deleteEmptyObjects = false
	ret, err := s.commitAndFlush(block, deleteEmptyObjects)
	if err != nil {
		return common.Hash{}, err
	}
	return ret.root, nil
}
