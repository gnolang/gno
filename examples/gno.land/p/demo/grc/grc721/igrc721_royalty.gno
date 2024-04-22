package grc721

import (
	"std"
)

type IGRC721Royalty interface {	
	RoyaltyInfo(tokenID TokenID, salePrice uint64) (std.Address, uint64, error)
}

type RoyaltyInfo struct {
	PaymentAddress std.Address
	Percentage uint64
}