package evm

import (
	"gno.land/p/demo/uint256"
)

type TransactionType uint

const (
	CreateContract TransactionType = iota
	PayMoney
	ExecuteContract
	SucerPaul
)

type Transaction interface {
	getType() TransactionType // returns the type ID
	clone() Transaction       // creates a deep copy and initializes all fields

	chainID() *uint256.Uint
	data() []byte

	gasPrice() uint
	// gasTip() uint
	caller() *Caller

	String() string
}

type SmartContract struct {
	Code    []byte
	Creator *Caller
}

// Env is a global environnement to evaluate your EVM state
type Env struct {
	Logs        []byte
	Storage     map[uint256.Uint]*uint256.Uint
	Accounts    []*Caller
	Contracts   []*SmartContract
	Transaction []*Transaction
}
