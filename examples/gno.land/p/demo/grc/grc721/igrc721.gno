package grc721

import "chain"

type IGRC721 interface {
	BalanceOf(owner chain.Address) (int64, error)
	OwnerOf(tid TokenID) (chain.Address, error)
	SetTokenURI(tid TokenID, tURI TokenURI) (bool, error)
	SafeTransferFrom(from, to chain.Address, tid TokenID) error
	TransferFrom(from, to chain.Address, tid TokenID) error
	Approve(approved chain.Address, tid TokenID) error
	SetApprovalForAll(operator chain.Address, approved bool) error
	GetApproved(tid TokenID) (chain.Address, error)
	IsApprovedForAll(owner, operator chain.Address) bool
}

type (
	TokenID  string
	TokenURI string
)

const (
	MintEvent           = "Mint"
	BurnEvent           = "Burn"
	TransferEvent       = "Transfer"
	ApprovalEvent       = "Approval"
	ApprovalForAllEvent = "ApprovalForAll"
)

type NFTGetter func() IGRC721
