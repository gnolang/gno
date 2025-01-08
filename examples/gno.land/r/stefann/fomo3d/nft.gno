package fomo3d

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
)

// Fomo3DNFT manages the NFTs for Fomo3D winners
type Fomo3DNFT struct {
	grc721.IGRC721         // implements the GRC721 interface
	grc721.IGRC721Metadata // implements the metadata interface

	name         string
	symbol       string
	roundCounter int
	owners       avl.Tree // tokenId -> OwnerAddress
	balances     avl.Tree // OwnerAddress -> TokenCount
	tokenURIs    avl.Tree // TokenId -> URIs
	approvals    avl.Tree // TokenId -> ApprovedAddress
	operators    avl.Tree // "OwnerAddress:OperatorAddress" -> bool
}

// Global instance
var gNFT = &Fomo3DNFT{
	name:         "Fomo3D Winner",
	symbol:       "FOMO",
	roundCounter: 0,
	owners:       avl.Tree{},
	balances:     avl.Tree{},
	tokenURIs:    avl.Tree{},
	approvals:    avl.Tree{},
	operators:    avl.Tree{},
}

// GetNFT returns the global NFT instance
func GetNFT() *Fomo3DNFT {
	return gNFT
}

// Name returns the name of the NFT collection
func (f *Fomo3DNFT) Name() string {
	return f.name
}

// Symbol returns the symbol of the NFT collection
func (f *Fomo3DNFT) Symbol() string {
	return f.symbol
}

// BalanceOf returns the number of NFTs owned by an address
func (f *Fomo3DNFT) BalanceOf(owner std.Address) (uint64, error) {
	if owner == "" {
		panic("zero address")
	}

	balance, found := f.balances.Get(owner.String())
	if !found {
		return 0, nil
	}
	return balance.(uint64), nil
}

// OwnerOf returns the owner of a specific token
func (f *Fomo3DNFT) OwnerOf(tid grc721.TokenID) (std.Address, error) {
	owner, found := f.owners.Get(string(tid))
	if !found {
		panic("token not found")
	}
	return owner.(std.Address), nil
}

// TokenURI returns the URI for a specific token
func (f *Fomo3DNFT) TokenURI(tid grc721.TokenID) (string, error) {
	uri, found := f.tokenURIs.Get(string(tid))
	if !found {
		panic("token not found")
	}
	return uri.(string), nil
}

// SetTokenURI sets the URI for a specific token
func (f *Fomo3DNFT) SetTokenURI(tid grc721.TokenID, uri grc721.TokenURI) (bool, error) {
	if _, found := f.owners.Get(string(tid)); !found {
		panic("token not found")
	}
	f.tokenURIs.Set(string(tid), string(uri))
	return true, nil
}

// SafeTransferFrom safely transfers a token
func (f *Fomo3DNFT) SafeTransferFrom(from, to std.Address, tid grc721.TokenID) error {
	return f.TransferFrom(from, to, tid)
}

// TransferFrom transfers a token
func (f *Fomo3DNFT) TransferFrom(from, to std.Address, tid grc721.TokenID) error {
	if to == "" {
		panic("zero address")
	}

	owner, err := f.OwnerOf(tid)
	if err != nil {
		panic(err)
	}

	if owner != from {
		panic("not owner")
	}

	caller := std.GetOrigCaller()
	if caller != from && !f.IsApprovedForAll(from, caller) {
		approved, _ := f.GetApproved(tid)
		if approved != caller {
			panic("not approved")
		}
	}

	// Update balances
	fromBalance, _ := f.BalanceOf(from)
	f.balances.Set(from.String(), fromBalance-1)

	toBalance, _ := f.BalanceOf(to)
	f.balances.Set(to.String(), toBalance+1)

	// Update owner
	f.owners.Set(string(tid), to)

	// Clear approval
	f.approvals.Remove(string(tid))

	return nil
}

// Approve approves an address to transfer a token
func (f *Fomo3DNFT) Approve(approved std.Address, tid grc721.TokenID) error {
	owner, err := f.OwnerOf(tid)
	if err != nil {
		panic(err)
	}

	caller := std.GetOrigCaller()
	if caller != owner && !f.IsApprovedForAll(owner, caller) {
		panic("not approved")
	}

	f.approvals.Set(string(tid), approved)
	return nil
}

// SetApprovalForAll approves an operator for all tokens
func (f *Fomo3DNFT) SetApprovalForAll(operator std.Address, approved bool) error {
	caller := std.GetOrigCaller()
	if caller == operator {
		panic("cannot approve self")
	}

	key := caller.String() + ":" + operator.String()
	f.operators.Set(key, approved)
	return nil
}

// GetApproved returns the approved address for a token
func (f *Fomo3DNFT) GetApproved(tid grc721.TokenID) (std.Address, error) {
	if _, err := f.OwnerOf(tid); err != nil {
		panic(err)
	}

	approved, found := f.approvals.Get(string(tid))
	if !found {
		return "", nil
	}
	return approved.(std.Address), nil
}

// IsApprovedForAll checks if an operator is approved for all tokens
func (f *Fomo3DNFT) IsApprovedForAll(owner, operator std.Address) bool {
	key := owner.String() + ":" + operator.String()
	approved, found := f.operators.Get(key)
	if !found {
		return false
	}
	return approved.(bool)
}

// MintRoundWinnerNFT mints a new NFT for the round winner
func (f *Fomo3DNFT) MintRoundWinnerNFT(winner std.Address) error {
	if winner == "" {
		panic("zero address")
	}

	f.roundCounter++
	tokenID := grc721.TokenID(strconv.Itoa(f.roundCounter))

	// Update balances
	balance, _ := f.BalanceOf(winner)
	f.balances.Set(winner.String(), balance+1)

	// Set owner
	f.owners.Set(string(tokenID), winner)

	// Set token URI
	uri := "Fomo3D Round " + strconv.Itoa(f.roundCounter) + " Winner"
	f.tokenURIs.Set(string(tokenID), uri)

	return nil
}

// GetWinnerNFTs returns all NFTs owned by an address
func (f *Fomo3DNFT) GetWinnerNFTs(owner std.Address) (uint64, error) {
	return f.BalanceOf(owner)
}
