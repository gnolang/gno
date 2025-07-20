package volos

import (
	u256 "gno.land/p/gnoswap/uint256"

	pl "gno.land/r/gnoswap/v1/pool"
)

// Market represents a lending market for a specific token pair
type Market struct {
	TotalSupplyAssets *u256.Uint // Total assets supplied to market
	TotalSupplyShares *u256.Uint // Total shares issued for supply
	TotalBorrowAssets *u256.Uint // Total assets borrowed from market
	TotalBorrowShares *u256.Uint // Total shares issued for borrows
	LastUpdate        int64      // Last time market was updated (unix timestamp)
	Fee               *u256.Uint // Market fee
}

// Position represents a user's position in a market
type Position struct {
	SupplyShares *u256.Uint // Shares for supplied assets
	BorrowShares *u256.Uint // Shares for borrowed assets
	Collateral   *u256.Uint // Amount of collateral token deposited
}

// MarketParams defines parameters for market creation
type MarketParams struct {
	PoolPath     string     // Gnoswap pool path (e.g. "token0:token1:3000") - also used as oracle
	IRM          string     // Interest Rate Model path
	LLTV         *u256.Uint // Liquidation Loan-to-Value ratio as a percentage (e.g. 75 means 75%)
	IsToken0Loan bool       // Whether token0 is the loan token (if false, token1 is the loan token)
}

// ID generates a unique identifier for a market using the Gnoswap pool path
func (mp *MarketParams) ID() string {
	return mp.PoolPath
}

// GetLoanToken returns the loan token path from the pool path
func (mp *MarketParams) GetLoanToken() string {
	if mp.IsToken0Loan {
		return pl.PoolGetToken0Path(mp.PoolPath)
	}
	return pl.PoolGetToken1Path(mp.PoolPath)
}

// GetCollateralToken returns the collateral token path from the pool path
func (mp *MarketParams) GetCollateralToken() string {
	if mp.IsToken0Loan {
		return pl.PoolGetToken1Path(mp.PoolPath)
	}
	return pl.PoolGetToken0Path(mp.PoolPath)
}

// IRM is the interface that all interest rate models must implement
type IRM interface {
	// BorrowRate returns the borrow rate per second as a simple integer percentage
	// For example: 5 means 5% per second
	BorrowRate(totalSupply, totalBorrow *u256.Uint) *u256.Uint

	// Name returns a human readable name for the IRM
	Name() string
}
