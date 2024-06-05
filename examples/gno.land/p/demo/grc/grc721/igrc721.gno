package grc721

import (
	"std"
)

type IGRC721 interface {
	// BalanceOf returns the number of tokens in `owner`'s account
	BalanceOf(owner std.Address) uint64

	// OwnerOf returns the owner of the `tokenId` token
	// tokenId must exist
	OwnerOf(tokenId string) std.Address

	// TransferFrom transfers `tokenId` token from `from` to `to`
	TransferFrom(from, to std.Address, tokenId string)

	// Approve gives permission to `to` to transfer `tokenId` token to another account
	// The approval is cleared when the token is transferred
	Approve(to std.Address, tokenId string)

	// SetApprovalForAll approves or removes `operator` as an operator for the caller
	// Operators can call TransferFrom for any token owned by the caller
	SetApprovalForAll(operator std.Address, approved bool)

	// GetApproved returns the account approved for `tokenId` token
	GetApproved(tokenId string) std.Address

	// IsApprovedForAll returns if the `operator` is allowed to manage all the assets of `owner`
	IsApprovedForAll(owner, operator std.Address) bool

	// TokenURI gets the tokenURI for matching `tokenId`
	TokenURI(tokenId string) string

	// SetTokenURI sets `tokenURI` as the tokenURI of `tokenId`
	SetTokenURI(tokenId string, tokenURI string) string
}
