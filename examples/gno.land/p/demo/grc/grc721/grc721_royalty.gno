package grc721

import (
	"std"

	"gno.land/p/demo/avl"
)

// royaltyNFT represents a non-fungible token (NFT) with royalty functionality.
type royaltyNFT struct {
	*metadataNFT              // Embedding metadataNFT for NFT functionality
	tokenRoyaltyInfo avl.Tree // AVL tree to store royalty information for each token
}

// Ensure that royaltyNFT implements the IGRC2981 interface.
var _ IGRC2981 = (*royaltyNFT)(nil)

// NewNFTWithRoyalty creates a new royalty NFT with the specified name, symbol, and royalty calculator.
func NewNFTWithRoyalty(name string, symbol string) *royaltyNFT {
	// Create a new NFT with metadata
	nft := NewNFTWithMetadata(name, symbol)

	return &royaltyNFT{
		metadataNFT:      nft,
		tokenRoyaltyInfo: avl.Tree{},
	}
}

// SetTokenRoyalty sets the royalty information for a specific token ID.
func (r *royaltyNFT) SetTokenRoyalty(tid TokenID, royaltyInfo RoyaltyInfo) error {
	// Validate the payment address
	if err := isValidAddress(royaltyInfo.PaymentAddress); err != nil {
		return ErrInvalidRoyaltyPaymentAddress
	}

	// Check if royalty percentage exceeds 100%
	if royaltyInfo.Percentage > 100 {
		return ErrInvalidRoyaltyPercentage
	}

	// Check if the caller is the owner of the token
	owner, err := r.metadataNFT.OwnerOf(tid)
	if err != nil {
		return err
	}
	caller := std.PrevRealm().Addr()
	if caller != owner {
		return ErrCallerIsNotOwner
	}

	// Set royalty information for the token
	r.tokenRoyaltyInfo.Set(string(tid), royaltyInfo)
	return nil
}

// RoyaltyInfo returns the royalty information for the given token ID and sale price.
func (r *royaltyNFT) RoyaltyInfo(tid TokenID, salePrice uint64) (std.Address, uint64, error) {
	// Retrieve royalty information for the token
	val, found := r.tokenRoyaltyInfo.Get(string(tid))
	if !found {
		return "", 0, ErrInvalidTokenId
	}

	royaltyInfo := val.(RoyaltyInfo)

	// Calculate royalty amount using the provided calculator
	royaltyAmount, err := r.calculateRoyaltyAmount(salePrice, royaltyInfo.Percentage)
	if err != nil {
		return "", 0, ErrCannotCalculateRoyaltyAmount
	}

	return royaltyInfo.PaymentAddress, royaltyAmount, nil
}

func (r *royaltyNFT) calculateRoyaltyAmount(salePrice, percentage uint64) (uint64, error) {
	royaltyAmount := (salePrice * percentage) / 100
	return royaltyAmount, nil
}
