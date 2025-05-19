package state

import "storage_extract/common"

type StateDB struct {
	stateObjects map[common.Address]*StateObject
}

// SetState sets the state of the given address and key to the given value.
// It retrieves the state object for the address, and if it doesn't exist, it creates a new one.
// Original function: github.com/ethereum/go-ethereum/core/state/statedb.go line 450
func (s *StateDB) SetState(addr common.Address, key, value common.Hash) common.Hash {
	if stateObject := s.getOrNewStateObject(addr); stateObject != nil {
		return stateObject.SetState()
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
	// add the object to the stateObjects map
	// This is different from the original code,
	// where calls another function (setStateObject) to add the object to the stateObjects map.
	// Here, directly add it to the map.
	s.stateObjects[addr] = obj
	return obj
}
