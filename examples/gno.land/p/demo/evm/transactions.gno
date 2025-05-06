package evm

import (
	"strconv"

	"gno.land/p/demo/uint256"
)

type CreateContractTransaction struct {
	transactionType TransactionType
	chainID         uint256.Uint
	data            []byte
	gasPrice        uint
	caller          *Caller
	contractAddress uint
}

func (t *CreateContractTransaction) getType() TransactionType {
	return t.transactionType
}

func (t *CreateContractTransaction) clone() CreateContractTransaction {
	tmpData := make([]byte, len(t.data))
	copy(tmpData, t.data)
	return CreateContractTransaction{
		transactionType: t.transactionType,
		chainID:         t.chainID,
		data:            tmpData,
	}
}

func (t *CreateContractTransaction) getChainID() *uint256.Uint {
	return &t.chainID
}

func (t *CreateContractTransaction) getData() *[]byte {
	return &t.data
}

func (t *CreateContractTransaction) getGasPrice() uint {
	return t.gasPrice
}

func (t *CreateContractTransaction) getCaller() *Caller {
	return t.caller
}

func (t *CreateContractTransaction) execute(env *Env) {
	env.Contracts.Push(&SmartContract{
		Code:    t.data,
		Creator: env.Accounts[0],
	})
	return
}

func (t *CreateContractTransaction) String() string {
	return t.caller.GnoAddress.String() + " created " + strconv.FormatUint(uint64(t.contractAddress), 10) + " costing " + strconv.FormatUint(gasPrice, 10)
}
