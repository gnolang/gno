package grc721

import (
	"std"
)

type IGRC721Royalty interface {	
	IGRC721Metadata
	
	SetTokenRoyalty(tokenID TokenID, royaltyInfo RoyaltyInfo) (bool, error) 
	RoyaltyInfo(tokenID TokenID, salePrice uint64) (std.Address, uint64, error)
}

// royaltyCalculator is an interface for calculating royalty amounts.
type RoyaltyCalculator interface {
	RoyaltyAmount(salePrice, percentage uint64) (uint64, error)
}

type RoyaltyInfo struct {
	PaymentAddress std.Address
	Percentage uint64
}