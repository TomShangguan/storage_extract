package pathdb

import "storage_extract/common"

// StateSetWithOrigin wraps the state set with additional original values of the
// mutated states.
type StateSetWithOrigin struct {

	// AccountOrigin represents the account data before the state transition,
	// corresponding to both the accountData and destructSet. It's keyed by the
	// account address. The nil value means the account was not present before.
	accountOrigin map[common.Address][]byte

	// StorageOrigin represents the storage data before the state transition,
	// corresponding to storageData and deleted slots of destructSet. It's keyed
	// by the account address and slot key hash. The nil value means the slot was
	// not present.
	storageOrigin map[common.Address]map[common.Hash][]byte

	// Memory size of the state data (accountOrigin and storageOrigin)
	size uint64
}

// NewStateSetWithOrigin constructs the state set with the provided data.
func NewStateSetWithOrigin(accountOrigin map[common.Address][]byte, storageOrigin map[common.Address]map[common.Hash][]byte) *StateSetWithOrigin {
	// Don't panic for the lazy callers, initialize the nil maps instead.
	if accountOrigin == nil {
		accountOrigin = make(map[common.Address][]byte)
	}
	if storageOrigin == nil {
		storageOrigin = make(map[common.Address]map[common.Hash][]byte)
	}
	// Count the memory size occupied by the set. Note that each slot key here
	// uses 2*common.HashLength to keep consistent with the calculation method
	// of stateSet.
	var size int
	for _, data := range accountOrigin {
		size += common.HashLength + len(data)
	}
	for _, slots := range storageOrigin {
		for _, data := range slots {
			size += 2*common.HashLength + len(data)
		}
	}
	return &StateSetWithOrigin{
		accountOrigin: accountOrigin,
		storageOrigin: storageOrigin,
		size:          uint64(size),
	}
}
