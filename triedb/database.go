package triedb

import (
	"errors"
	"storage_extract/common"
	"storage_extract/ethdb"
	"storage_extract/trie/trienode"
	"storage_extract/triedb/pathdb"
)

// backend defines the methods needed to access/update trie nodes in different
// state scheme.
type backend interface {
}

// Database is the wrapper of the underlying backend which is shared by different
// types of node backend as an entrypoint. It's responsible for all interactions
// relevant with trie nodes and node preimages.
type Database struct {
	disk    ethdb.Database
	backend backend // The backend for managing trie nodes
}

// Update performs a state transition by committing dirty nodes contained in the
// given set in order to update state from the specified parent to the specified
// root. The held pre-images accumulated up to this point will be flushed in case
// the size exceeds the threshold.
//
// The passed in maps(nodes, states) will be retained to avoid copying everything.
// Therefore, these maps must not be changed afterwards.
func (db *Database) Update(root common.Hash, parent common.Hash, block uint64, nodes *trienode.MergedNodeSet, states *StateSet) error {
	switch b := db.backend.(type) {
	case *pathdb.Database:
		return b.Update(root, parent, block, nodes, states.internal())
	}

	return errors.New("unsupported backend type for trie database update")
}
