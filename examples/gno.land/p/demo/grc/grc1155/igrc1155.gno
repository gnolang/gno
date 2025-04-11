package grc1155

import "chain"

type IGRC1155 interface {
	SafeTransferFrom(from, to chain.Address, tid TokenID, amount uint64) error
	SafeBatchTransferFrom(from, to chain.Address, batch []TokenID, amounts []uint64) error
	BalanceOf(owner chain.Address, tid TokenID) (uint64, error)
	BalanceOfBatch(owners []chain.Address, batch []TokenID) ([]uint64, error)
	SetApprovalForAll(operator chain.Address, approved bool) error
	IsApprovedForAll(owner, operator chain.Address) bool
}

type TokenID string

type TransferSingleEvent struct {
	Operator chain.Address
	From     chain.Address
	To       chain.Address
	TokenID  TokenID
	Amount   uint64
}

type TransferBatchEvent struct {
	Operator chain.Address
	From     chain.Address
	To       chain.Address
	Batch    []TokenID
	Amounts  []uint64
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
