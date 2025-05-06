package evm

import (
	"gno.land/p/demo/uint256"
)

type TransactionType uint

const (
	CreateContract TransactionType = iota
	PayMoney
	ExecuteContract
)

type Transaction interface {
	getType() TransactionType // returns the type ID
	clone() Transaction       // creates a deep copy and initializes all fields

	getChainID() *uint256.Uint // returns the unique chain ID
	getData() *[]byte          // return data of the transaction

	getGasPrice() uint  // return used gaz
	getCaller() *Caller // return caller

	execute(env *State) // execute the contract on the env
	String() string     // string representing the transaction
}

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs            []byte
	Storage         map[uint256.Uint]*uint256.Uint
	Accounts        []*Caller
	Contracts       SmartContracts
	Transaction     []*Transaction
	ChainID         uint256.Uint
	BlockDifficulty uint256.Uint
	GasLimit        uint256.Uint
}
