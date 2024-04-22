package grc721

import (
	"std"
)

// Base grc2981 interface
type IGRC2981 interface {	
	RoyaltyInfo(tokenID TokenID, salePrice uint64) (std.Address, uint64, error)
}

type RoyaltyInfo struct {
	PaymentAddress std.Address
	Percentage uint64
}