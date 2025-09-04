package grc1155

import "chain"

type IGRC1155 interface {
	SafeTransferFrom(from, to chain.Address, tid TokenID, amount int64) error
	SafeBatchTransferFrom(from, to chain.Address, batch []TokenID, amounts []int64) error
	BalanceOf(owner chain.Address, tid TokenID) (int64, error)
	BalanceOfBatch(owners []chain.Address, batch []TokenID) ([]int64, error)
	SetApprovalForAll(operator chain.Address, approved bool) error
	IsApprovedForAll(owner, operator chain.Address) bool
}

type TokenID string

type TransferSingleEvent struct {
	Operator chain.Address
	From     chain.Address
	To       chain.Address
	TokenID  TokenID
	Amount   int64
}

type TransferBatchEvent struct {
	Operator chain.Address
	From     chain.Address
	To       chain.Address
	Batch    []TokenID
	Amounts  []int64
}

type ApprovalForAllEvent struct {
	Owner    chain.Address
	Operator chain.Address
	Approved bool
}

type UpdateURIEvent struct {
	URI string
}

type MultiTokenGetter func() IGRC1155
