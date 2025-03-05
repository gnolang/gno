package gnolend

import (
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/demo/uint256"
)

// Market represents a lending market for a specific token pair
type Market struct {
	TotalSupplyAssets *u256.Uint // Total assets supplied to market
	TotalSupplyShares *u256.Uint // Total shares issued for supply
	TotalBorrowAssets *u256.Uint // Total assets borrowed from market
	TotalBorrowShares *u256.Uint // Total shares issued for borrows
	LastUpdate        int64      // Last time market was updated (unix timestamp)
	Fee              int64      // Market fee as a percentage (e.g. 5 means 5%)
}

// Position represents a user's position in a market
type Position struct {
	SupplyShares *u256.Uint // Shares for supplied assets
	BorrowShares *u256.Uint // Shares for borrowed assets
	Collateral   *u256.Uint // Amount of collateral token deposited
}

// MarketParams defines parameters for market creation
type MarketParams struct {
	LoanToken       string // Token that can be borrowed
	CollateralToken string // Token used as collateral
	Oracle         string // Price oracle path
	IRM            string // Interest Rate Model path
	LLTV          int64  // Liquidation Loan-to-Value ratio as a percentage (e.g. 75 means 75%)
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

// Constants for scaling
const (
	MAX_FEE = 10 // 10% max fee
)

// IRM is the interface that all interest rate models must implement
type IRM interface {
	// BorrowRate returns the borrow rate per second as a simple integer percentage
	// For example: 5 means 5% per second
	BorrowRate(market Market, params MarketParams) int64
}
