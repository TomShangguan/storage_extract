package state

import (
	"storage_extract/common"
	"storage_extract/trie/trienode"
	"storage_extract/types"
)

// stateUpdate represents a collection of changes made to the Ethereum state.
// TODO: Add support for accounts and contract code updates in the future.
type stateUpdate struct {
	originRoot     common.Hash                               // hash of the state before applying mutation
	root           common.Hash                               // hash of the state after applying mutation
	storages       map[common.Hash]map[common.Hash][]byte    // storages stores mutated slots in 'prefix-zero-trimmed' RLP format
	storagesOrigin map[common.Address]map[common.Hash][]byte // storagesOrigin stores the original values of mutated slots in 'prefix-zero-trimmed' RLP format
	nodes          *trienode.MergedNodeSet                   // Aggregated dirty nodes caused by state changes
}

// accountUpdate represents an operation for updating an Ethereum account.
type accountUpdate struct {
	address common.Address // address is the unique account identifier
	data    []byte         // data is the slim-RLP encoded account data.
	origin  []byte         // origin is the original value of account data in slim-RLP encoding.
	//code           *contractCode          // code represents mutated contract code; nil means it's not modified.
	storages       map[common.Hash][]byte // storages stores mutated slots in prefix-zero-trimmed RLP format.
	storagesOrigin map[common.Hash][]byte // storagesOrigin stores the original values of mutated slots in prefix-zero-trimmed RLP format.
}

// newStateUpdate constructs a state update object, representing the differences
// between two states by performing state execution. It aggregates the given
// account deletions and account updates to form a comprehensive state update.
func newStateUpdate(originRoot common.Hash, root common.Hash, updates map[common.Hash]*accountUpdate, nodes *trienode.MergedNodeSet) *stateUpdate {
	var (
		storages       = make(map[common.Hash]map[common.Hash][]byte)
		storagesOrigin = make(map[common.Address]map[common.Hash][]byte)
	)
	// TODO:
	// Due to the fact that some accounts could be destructed and resurrected
	// within the same block, the deletions must be aggregated first.
	// Aggregate account updates then.
	for addrHash, op := range updates {
		addr := op.address
		// TODO:
		// Add support for contract code updates in the future.
		// Aggregate the account changes. The original account value will only
		// be tracked if it's not present yet.

		// Aggregate the storage changes. The original storage slot value will
		// only be tracked if it's not present yet.
		if len(op.storages) > 0 {
			storages[addrHash] = op.storages
		}
		if len(op.storagesOrigin) > 0 {
			origin := storagesOrigin[addr]
			if origin == nil {
				storagesOrigin[addr] = op.storagesOrigin
				continue
			}
			for key, slot := range op.storagesOrigin {
				if _, found := origin[key]; !found {
					origin[key] = slot
				}
			}
			storagesOrigin[addr] = origin
		}
	}
	return &stateUpdate{
		originRoot: types.TrieRootHash(originRoot),
		root:       types.TrieRootHash(root),
		// destructs:      destructs,
		// accounts:       accounts,
		// accountsOrigin: accountsOrigin,
		storages:       storages,
		storagesOrigin: storagesOrigin,
		// codes:          codes,
		nodes: nodes,
	}
}
