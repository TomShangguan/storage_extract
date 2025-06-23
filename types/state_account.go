package types

import (
	"storage_extract/common"

	"github.com/ethereum/go-ethereum/rlp"
)

type StateAccount struct {
	Root common.Hash
}

// NewEmptyStateAccount creates a new empty state account with a zero root hash.
func NewEmptyStateAccount() *StateAccount {
	return &StateAccount{
		Root: EmptyRootHash,
	}
}

// SlimAccount is a modified version of an Account, where the root is replaced
// with a byte slice. This format can be used to represent full-consensus format
// or slim format which replaces the empty root and code hash as nil byte slice.
type SlimAccount struct {
	//Nonce    uint64
	//Balance  *uint256.Int
	Root []byte // Nil if root equals to types.EmptyRootHash
	//CodeHash []byte // Nil if hash equals to types.EmptyCodeHash
}

// SlimAccountRLP encodes the state account in 'slim RLP' format.
func SlimAccountRLP(account StateAccount) []byte {
	// TODO: Implement support for nonce and balance in the slim format if needed.
	// Currently, the slim format only includes the root hash.
	slim := SlimAccount{
		//Nonce:   account.Nonce,
		//Balance: account.Balance,
	}
	if account.Root != EmptyRootHash {
		slim.Root = account.Root[:]
	}
	// if !bytes.Equal(account.CodeHash, EmptyCodeHash[:]) {
	// 	slim.CodeHash = account.CodeHash
	// }
	data, err := rlp.EncodeToBytes(slim)
	if err != nil {
		panic(err)
	}
	return data
}
