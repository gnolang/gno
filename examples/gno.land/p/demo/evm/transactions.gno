package evm

import (
	"encoding/binary"
	"encoding/hex"
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

func (s *Transactions) PushExecution(gasCost uint, caller *Caller, result []byte, address uint) {
	createContract := ExecuteContractTransaction{
		gasPrice:        gasCost,
		caller:          caller,
		result:          result,
		contractAddress: address,
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
	return t.caller.EVMAddress.Hex() + " created a contract, costing " + strconv.FormatUint(gasPrice, 10) + " gas"
}

type ExecuteContractTransaction struct {
	gasPrice        uint
	caller          *Caller
	result          []byte
	contractAddress uint
}

func (t *ExecuteContractTransaction) getType() TransactionType {
	return ExecuteContract
}

func (t *ExecuteContractTransaction) clone() Transaction {
	return &ExecuteContractTransaction{
		gasPrice:        t.gasPrice,
		caller:          t.caller,
		result:          t.result,
		contractAddress: t.contractAddress,
	}
}

func (t *ExecuteContractTransaction) getGasPrice() uint {
	return t.gasPrice
}

func (t *ExecuteContractTransaction) getCaller() *Caller {
	return t.caller
}

func trimLeadingZeros(data []byte) []byte {
	for k, v := range data {
		if v != 0 {
			return data[k:]
		}
	}
	return []byte{0} // all of them are 0
}

func (t *ExecuteContractTransaction) String() string {
	x := t.caller.EVMAddress.Hex() + " executed " + strconv.FormatUint(uint64(t.contractAddress), 10)
	x += " costing " + strconv.FormatUint(gasPrice, 10)
	x += ", resulting with " + hex.EncodeToString(trimLeadingZeros(t.result))
	x += " (decimal: " + strconv.FormatUint(binary.BigEndian.Uint64(t.result[len(t.result)-8:len(t.result)]), 10) + ")"
	return x
}
