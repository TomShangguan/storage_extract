package state

import "storage_extract/common"

type journalEntry interface {
	dirtied() *common.Address // dirtied returns the Ethereum address modified by this journal entry.
}

type storageChange struct {
	account   common.Address
	key       common.Hash
	prevvalue common.Hash
	origvalue common.Hash
}

// journal contains the list of state modifications applied since the last state commit.
type journal struct {
	entries []journalEntry         // Current changes tracked by the journal
	dirties map[common.Address]int // Dirty accounts and the number of changes
}

// newJournal creates a new journal instance.
func newJournal() *journal {
	return &journal{
		dirties: make(map[common.Address]int),
	}
}

// append inserts a new modification entry to the end of the change journal.
func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
	if addr := entry.dirtied(); addr != nil {
		j.dirties[*addr]++
	}
}

func (j *journal) storageChange(addr common.Address, key, prev, origin common.Hash) {
	j.append(storageChange{
		account:   addr,
		key:       key,
		prevvalue: prev,
		origvalue: origin,
	})
}

func (ch storageChange) dirtied() *common.Address {
	return &ch.account
}
