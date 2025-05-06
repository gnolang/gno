package evm

import (
	"strconv"
)

type TransactionType uint

type Transactions []Transaction

const (
	CreateContract TransactionType = iota
	PayMoney
	ExecuteContract
)

type Transaction interface {
	getType() TransactionType // returns the type ID
	clone() Transaction       // creates a deep copy and initializes all fields
	getGasPrice() uint        // returns used gas
	getCaller() *Caller       // returns caller

	String() string // string representing the transaction
}

func (s *Transactions) PushCreation(gasCost uint, caller *Caller, newContractAdress uint) {
	createContract := CreateContractTransaction{
		gasPrice:        gasCost,
		caller:          caller,
		contractAddress: newContractAdress,
	}
	*s = append(*s, &createContract)
}

func (s *Transactions) PushExecution(gasCost uint, caller *Caller, newContractAdress uint) {
	createContract := ExecuteContractTransaction{
		gasPrice:        gasCost,
		caller:          caller,
		contractAddress: newContractAdress,
	}
	*s = append(*s, &createContract)
}

type CreateContractTransaction struct {
	gasPrice        uint
	caller          *Caller
	contractAddress uint
}

func (t *CreateContractTransaction) getType() TransactionType {
	return CreateContract
}

func (t *CreateContractTransaction) clone() Transaction {
	return &CreateContractTransaction{
		gasPrice:        t.gasPrice,
		caller:          t.caller,
		contractAddress: t.contractAddress,
	}
}

func (t *CreateContractTransaction) getGasPrice() uint {
	return t.gasPrice
}

func (t *CreateContractTransaction) getCaller() *Caller {
	return t.caller
}

func (t *CreateContractTransaction) String() string {
	return t.caller.GnoAddress.String() + " created " + strconv.FormatUint(uint64(t.contractAddress), 10) + " costing " + strconv.FormatUint(gasPrice, 10)
}

type ExecuteContractTransaction struct {
	gasPrice        uint
	caller          *Caller
	contractAddress uint
}

func (t *ExecuteContractTransaction) getType() TransactionType {
	return ExecuteContract
}

func (t *ExecuteContractTransaction) clone() Transaction {
	return &ExecuteContractTransaction{
		gasPrice:        t.gasPrice,
		caller:          t.caller,
		contractAddress: t.contractAddress,
	}
}

func (t *ExecuteContractTransaction) getGasPrice() uint {
	return t.gasPrice
}

func (t *ExecuteContractTransaction) getCaller() *Caller {
	return t.caller
}

func (t *ExecuteContractTransaction) String() string {
	return t.caller.GnoAddress.String() + " executed " + strconv.FormatUint(uint64(t.contractAddress), 10) + " costing " + strconv.FormatUint(gasPrice, 10)
}
