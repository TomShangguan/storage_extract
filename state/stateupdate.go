package state

import "storage_extract/common"

type stateUpdate struct {
	originRoot common.Hash // hash of the state before applying mutation
	root       common.Hash // hash of the state after applying mutation
}
