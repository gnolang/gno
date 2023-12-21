package grc721

import "std"

type IGRC721 interface {
	BalanceOf(owner std.Address) (uint64, error)
	OwnerOf(tid TokenID) (std.Address, error)
	SetTokenURI(tid TokenID, tURI TokenURI) (bool, error)
	SafeTransferFrom(from, to std.Address, tid TokenID) error
	TransferFrom(from, to std.Address, tid TokenID) error
	Approve(approved std.Address, tid TokenID) error
	SetApprovalForAll(operator std.Address, approved bool) error
	GetApproved(tid TokenID) (std.Address, error)
	IsApprovedForAll(owner, operator std.Address) bool
}

type (
	TokenID  string
	TokenURI string
)

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
