package grc721

import (
	"math/overflow"

	"gno.land/p/nt/avl/v0"
)

// royaltyNFT represents a non-fungible token (NFT) with royalty functionality.
type royaltyNFT struct {
	*metadataNFT                   // Embedding metadataNFT for NFT functionality
	tokenRoyaltyInfo     *avl.Tree // AVL tree to store royalty information for each token
	maxRoyaltyPercentage int64     // maxRoyaltyPercentage represents the maximum royalty percentage that can be charged every sale
}

// Ensure that royaltyNFT implements the IGRC2981 interface.
var _ IGRC2981 = (*royaltyNFT)(nil)

// NewNFTWithRoyalty creates a new royalty NFT with the specified name, symbol, and royalty calculator.
func NewNFTWithRoyalty(_ int, rlm realm, name, symbol string) *royaltyNFT {
	return &royaltyNFT{
		metadataNFT:          NewNFTWithMetadata(0, rlm, name, symbol),
		tokenRoyaltyInfo:     avl.NewTree(),
		maxRoyaltyPercentage: 100,
	}
}

// SetTokenRoyalty sets the royalty information for a specific token ID.
// caller must equal the token's owner; the owning realm's wrapper
// derives caller from rlm.Previous().Address() under IsCurrent.
func (r *royaltyNFT) SetTokenRoyalty(caller address, tid TokenID, royaltyInfo RoyaltyInfo) error {
	// Validate the payment address
	if err := isValidAddress(royaltyInfo.PaymentAddress); err != nil {
		return ErrInvalidRoyaltyPaymentAddress
	}

	// Check if royalty percentage exceeds maxRoyaltyPercentage
	if royaltyInfo.Percentage > r.maxRoyaltyPercentage {
		return ErrInvalidRoyaltyPercentage
	}

	// Check if the caller is the owner of the token
	owner, err := r.metadataNFT.OwnerOf(tid)
	if err != nil {
		return err
	}
	if caller != owner {
		return ErrCallerIsNotOwner
	}

	// Set royalty information for the token
	r.tokenRoyaltyInfo.Set(tid.String(), royaltyInfo)

	return nil
}

// RoyaltyInfo returns the royalty information for the given token ID and sale price.
func (r *royaltyNFT) RoyaltyInfo(tid TokenID, salePrice int64) (address, int64, error) {
	// Retrieve royalty information for the token
	val, found := r.tokenRoyaltyInfo.Get(tid.String())
	if !found {
		return "", 0, ErrInvalidTokenId
	}

	royaltyInfo := val.(RoyaltyInfo)

	// Calculate royalty amount
	royaltyAmount, _ := r.calculateRoyaltyAmount(salePrice, royaltyInfo.Percentage)

	return royaltyInfo.PaymentAddress, royaltyAmount, nil
}

func (r *royaltyNFT) calculateRoyaltyAmount(salePrice, percentage int64) (int64, error) {
	royaltyAmount := overflow.Mul64p(salePrice, percentage) / 100
	return royaltyAmount, nil
}
