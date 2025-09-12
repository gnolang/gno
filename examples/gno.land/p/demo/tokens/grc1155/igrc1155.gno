package grc1155

import "std"

type IGRC1155 interface {
	SafeTransferFrom(from, to std.Address, tid TokenID, amount int64) error
	SafeBatchTransferFrom(from, to std.Address, batch []TokenID, amounts []int64) error
	BalanceOf(owner std.Address, tid TokenID) (int64, error)
	BalanceOfBatch(owners []std.Address, batch []TokenID) ([]int64, error)
	SetApprovalForAll(operator std.Address, approved bool) error
	IsApprovedForAll(owner, operator std.Address) bool
}

type TokenID string

type TransferSingleEvent struct {
	Operator std.Address
	From     std.Address
	To       std.Address
	TokenID  TokenID
	Amount   int64
}

type TransferBatchEvent struct {
	Operator std.Address
	From     std.Address
	To       std.Address
	Batch    []TokenID
	Amounts  []int64
}

type ApprovalForAllEvent struct {
	Owner    std.Address
	Operator std.Address
	Approved bool
}

type UpdateURIEvent struct {
	URI string
}

type MultiTokenGetter func() IGRC1155
