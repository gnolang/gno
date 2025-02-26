package gnolend

import (
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
)

// Market represents a lending market for a specific token pair
type Market struct {
	LastUpdate        time.Time
	LoanToken         string // Token that can be borrowed
	CollateralToken   string // Token used as collateral
	TotalSupplyAssets int64  // Total assets supplied to market
	TotalSupplyShares int64  // Total shares issued for supply
	TotalBorrowAssets int64  // Total assets borrowed from market
	TotalBorrowShares int64  // Total shares issued for borrows
}

// Position represents a user's position in a market
type Position struct {
	SupplyShares int64 // Shares for supplied assets
	BorrowShares int64 // Shares for borrowed assets
	Collateral   int64 // Amount of collateral token deposited
}

// MarketParams defines parameters for market creation
type MarketParams struct {
	LoanToken       string // Token to borrow
	CollateralToken string // Token to use as collateral
}

// ID generates a unique identifier for a market by combining loan and collateral token paths
func (mp *MarketParams) ID() string {
	// Sort token paths alphabetically for consistent ID generation
	loanToken, collateralToken := mp.LoanToken, mp.CollateralToken
	if strings.Compare(collateralToken, loanToken) < 0 {
		loanToken, collateralToken = collateralToken, loanToken
	}
	return ufmt.Sprintf("%s:%s", loanToken, collateralToken)
}
