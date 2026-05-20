package grc721

import (
	"chain/runtime"
	"math/overflow"

	"gno.land/p/nt/avl/v0"
)

// bpsDenominator is the EIP-2981 fee denominator (basis points).
const bpsDenominator = 10000

// royaltyNFT represents a non-fungible token (NFT) with royalty functionality.
type royaltyNFT struct {
	*metadataNFT               // Embedding metadataNFT for NFT functionality
	tokenRoyaltyInfo *avl.Tree // AVL tree to store royalty information for each token
	maxRoyaltyBps    int64     // maxRoyaltyBps is the maximum allowed royalty in basis points (10000 = 100%).
}

// Ensure that royaltyNFT implements the IGRC2981 interface.
var _ IGRC2981 = (*royaltyNFT)(nil)

// NewNFTWithRoyalty creates a new royalty NFT with the specified name, symbol, and royalty calculator.
func NewNFTWithRoyalty(name string, symbol string) *royaltyNFT {
	// Create a new NFT with metadata
	nft := NewNFTWithMetadata(name, symbol)

	return &royaltyNFT{
		metadataNFT:      nft,
		tokenRoyaltyInfo: avl.NewTree(),
		maxRoyaltyBps:    bpsDenominator,
	}
}

// SetTokenRoyalty sets the royalty information for a specific token ID.
func (r *royaltyNFT) SetTokenRoyalty(tid TokenID, royaltyInfo RoyaltyInfo) error {
	// Validate the payment address
	if err := isValidAddress(royaltyInfo.PaymentAddress); err != nil {
		return ErrInvalidRoyaltyPaymentAddress
	}

	// Check that royalty bps is within [0, maxRoyaltyBps].
	if royaltyInfo.Bps < 0 || royaltyInfo.Bps > r.maxRoyaltyBps {
		return ErrInvalidRoyaltyBps
	}

	// Check if the caller is the owner of the token
	owner, err := r.metadataNFT.OwnerOf(tid)
	if err != nil {
		return err
	}
	caller := runtime.CurrentRealm().Address()
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

	var royaltyAmount int64
	if product, ok := overflow.Mul64(salePrice, royaltyInfo.Bps); ok {
		royaltyAmount = product / bpsDenominator
	} else {
		royaltyAmount = (salePrice / bpsDenominator) * royaltyInfo.Bps
	}

	return royaltyInfo.PaymentAddress, royaltyAmount, nil
}
