package grc721

import "std"

type IGRC721 interface {
	BalanceOf(owner std.Address) (count int64)
	OwnerOf(tid TokenID) std.Address
	SafeTransferFrom(from, to std.Address, tid TokenID)
	TransferFrom(from, to std.Address, tid TokenID)
	Approve(approved std.Address, tid TokenID)
	SetApprovalForAll(operator std.Address, approved bool)
	GetApproved(tid TokenID) std.Address
	IsApprovedForAll(owner, operator std.Address) bool
}

type TokenID string

type TransferEvent struct {
	From    std.Address
	To      std.Address
	TokenID TokenID
}

type ApprovalEvent struct {
	Owner    std.Address
	Approved std.Address
	TokenID  TokenID
}

type ApprovalForAllEvent struct {
	Owner    std.Address
	Operator std.Address
	Approved bool
}
