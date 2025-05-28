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
	getCaller() *Account      // returns caller

	String() string // string representing the transaction
}

func (s *Transactions) PushCreation(gasCost uint, caller *Account, newContractAddress uint64) {
	createContract := CreateContractTransaction{
		gasPrice:        gasCost,
		caller:          caller,
		contractAddress: newContractAddress,
	}
	*s = append(*s, &createContract)
}

func (s *Transactions) PushExecution(gasCost uint, caller *Account, result []byte, address uint64) {
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
	caller          *Account
	contractAddress uint64
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

func (t *CreateContractTransaction) getCaller() *Account {
	return t.caller
}

func (t *CreateContractTransaction) String() string {
	return "0x" + strconv.FormatInt(int64(t.caller.EvmAddress), 16) + " created an account for a " + t.caller.AccountType.String() + " (to change), costing " + strconv.FormatUint(gasPrice, 10) + " gas"
}

type ExecuteContractTransaction struct {
	gasPrice        uint
	caller          *Account
	result          []byte
	contractAddress uint64
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

func (t *ExecuteContractTransaction) getCaller() *Account {
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
	x := "0x" + strconv.FormatInt(int64(t.caller.EvmAddress), 16) + " executed " + "0x" + strconv.FormatInt(int64(t.contractAddress), 16)
	x += " costing " + strconv.FormatUint(gasPrice, 10)
	x += ", resulting with 0x" + hex.EncodeToString(trimLeadingZeros(t.result))
	x += " (decimal: " + strconv.FormatUint(binary.BigEndian.Uint64(t.result[len(t.result)-8:len(t.result)]), 10) + ")"
	return x
}
