package grc721

import (
	"std"

	"gno.land/p/demo/avl"
)


type royaltyNFT struct {
	*metadataNFT

	tokenRoyaltyInfo avl.Tree
}


var _ IGRC721Royalty = (*royaltyNFT)(nil)


func NewNFTWithRoyalty(name string, symbol string) *royaltyNFT {

	nft := NewNFTWithMetadata(name, symbol)

	return &royaltyNFT{
		metadataNFT:   nft,
		tokenRoyaltyInfo: avl.Tree{},
	}
}

// SetTokenRoyalty sets the royalty information for a specific token ID
func (r *royaltyNFT) SetTokenRoyalty(tid TokenID, royaltyInfo RoyaltyInfo) (bool, error) {
	if err := isValidAddress(royaltyInfo.PaymentAddress); err != nil {
		return false, ErrInvalidRoyaltyPaymentAddress 
	}

	// RoyaltyAmount will exceed the sale price
	if royaltyInfo.Percentage > 100 {
		return false, ErrInvalidRoyaltyPercentage
	}

	// Check for the right owner
	owner, err := r.metadataNFT.OwnerOf(tid)
	if err != nil {
		return false, err
	}
	caller := std.PrevRealm().Addr()
	if caller != owner {
		return false, ErrCallerIsNotOwner
	}
	r.tokenRoyaltyInfo.Set(string(tid), royaltyInfo)
	return true, nil
}


// RoyaltyInfo returns the royalty information for the given token ID and sale price
func (r *royaltyNFT) RoyaltyInfo(tid TokenID, salePrice uint64) (std.Address, uint64, error) {
	val, found := r.tokenRoyaltyInfo.Get(string(tid))
	if !found {
		return "", 0, ErrInvalidTokenId
	}

	royaltyInfo := val.(RoyaltyInfo)

	royaltyAmount := (salePrice * royaltyInfo.Percentage) / 100

	return royaltyInfo.PaymentAddress, royaltyAmount, nil
}

