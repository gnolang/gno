package evm

import (
	// "encoding/binary"
	"encoding/hex"
	"strconv"
)

// TransactionType type of a transaction in the EVM
type TransactionType uint

// Transactions list of all transactions
type Transactions []Transaction

const (
	CreateContract  TransactionType = iota // account creation
	PayMoney                               // money transfer
	ExecuteContract                        // contract execution
)

type Transaction interface {
	getType() TransactionType // returns the type ID
	clone() Transaction       // creates a deep copy and initializes all fields
	getGasPrice() uint        // returns used gas
	getCaller() *Account      // returns caller

	String() string // string representing the transaction
}

func (s *Transactions) PushCallerCreation(gasCost uint, caller *Account, newContractAddress uint64) {
	createContract := CreateContractTransaction{
		gasPrice:        gasCost,
		caller:          caller,
		contractAddress: newContractAddress,
		AccountType:     CallerAccount,
	}
	*s = append(*s, &createContract)
}

func (s *Transactions) PushContractCreation(gasCost uint, caller *Account, newContractAddress uint64) {
	createContract := CreateContractTransaction{
		gasPrice:        gasCost,
		caller:          caller,
		contractAddress: newContractAddress,
		AccountType:     ContractAccount,
	}
	*s = append(*s, &createContract)
}

func (s *Transactions) PushExecution(gasCost uint, caller *Account, result []byte, address uint64, rawCode bool) {
	createContract := ExecuteContractTransaction{
		gasPrice:        gasCost,
		caller:          caller,
		result:          result,
		rawCode:         rawCode,
		contractAddress: address,
	}
	*s = append(*s, &createContract)
}

type CreateContractTransaction struct {
	gasPrice        uint
	AccountType     AccountType
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
	return "0x" + strconv.FormatUint(uint64(t.caller.EvmAddress), 16) + " created an account for a " + t.AccountType.String() + ", costing " + strconv.FormatUint(uint64(t.gasPrice), 10) + " gas"
}

type ExecuteContractTransaction struct {
	gasPrice        uint
	caller          *Account
	result          []byte
	rawCode         bool
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
	tmp := "string"
	if t.rawCode {
		tmp = "raw code"
	} else {
		tmp = "contract at 0x" + strconv.FormatInt(int64(t.contractAddress), 16)
	}
	x := "0x" + strconv.FormatInt(int64(t.caller.EvmAddress), 16) + " executed " + tmp
	x += " costing " + strconv.FormatUint(uint64(t.gasPrice), 10)
	x += ", resulting with 0x" + hex.EncodeToString(trimLeadingZeros(t.result))
	return x
}
