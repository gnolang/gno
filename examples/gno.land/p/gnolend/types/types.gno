package types

import (
	"std"
	"time"

	"gno.land/p/demo/grc/grc20"
)

// LendingPool represents a single asset lending pool
type LendingPool struct {
	// Pool configuration
	AssetAddress std.Address    // Address of the underlying GRC20 token
	GTokenAddress std.Address   // Address of the interest-bearing gToken
	
	// Pool state
	TotalSupply uint64         // Total assets supplied to the pool
	TotalBorrow uint64         // Total assets borrowed from the pool
	LiquidityIndex uint64      // Current liquidity index for interest accrual
	BorrowIndex uint64         // Current borrow index for interest accrual
	LastUpdateBlock int64      // Last block when indices were updated
	
	// Risk parameters
	LTV uint64                 // Loan-to-value ratio (in basis points)
	LiquidationThreshold uint64 // Liquidation threshold (in basis points)
	LiquidationBonus uint64    // Liquidation bonus (in basis points)
	
	// Interest rates
	SupplyRate uint64          // Current supply interest rate (in basis points)
	BorrowRate uint64          // Current borrow interest rate (in basis points)
}

// UserPosition tracks a user's lending/borrowing position
type UserPosition struct {
	Owner std.Address          // Address of the position owner
	
	// Supplied assets
	SuppliedAssets map[std.Address]Supply // Map of asset address to supply data
	
	// Borrowed assets
	BorrowedAssets map[std.Address]Borrow // Map of asset address to borrow data
	
	// Position state
	HealthFactor uint64        // Current health factor (in basis points)
	LastUpdateBlock int64      // Last block when position was updated
}

// Supply represents a supplied asset position
type Supply struct {
	Amount uint64              // Amount of asset supplied
	IsCollateral bool          // Whether this supply is used as collateral
	LiquidityIndex uint64      // Liquidity index at time of supply
}

// Borrow represents a borrowed asset position
type Borrow struct {
	Amount uint64              // Amount of asset borrowed
	BorrowIndex uint64         // Borrow index at time of borrow
}

// PriceData represents price information from Gnoswap
type PriceData struct {
	Price uint64               // Current price in GNOT
	LastUpdateBlock int64      // Last block when price was updated
	Decimals uint8            // Number of decimals for the asset
}

// InterestRateParams stores the parameters for interest rate calculation
type InterestRateParams struct {
	BaseRate uint64           // Base interest rate (in basis points)
	Slope1 uint64            // Interest rate slope below optimal utilization
	Slope2 uint64            // Interest rate slope above optimal utilization
	OptimalUtilization uint64 // Optimal utilization rate (in basis points)
}

// Events
type SupplyEvent struct {
	User std.Address
	Asset std.Address
	Amount uint64
	Timestamp time.Time
}

type BorrowEvent struct {
	User std.Address
	Asset std.Address
	Amount uint64
	Timestamp time.Time
}

type RepayEvent struct {
	User std.Address
	Asset std.Address
	Amount uint64
	Timestamp time.Time
}

type WithdrawEvent struct {
	User std.Address
	Asset std.Address
	Amount uint64
	Timestamp time.Time
}

type LiquidationEvent struct {
	Liquidator std.Address
	User std.Address
	CollateralAsset std.Address
	DebtAsset std.Address
	CollateralAmount uint64
	DebtAmount uint64
	Timestamp time.Time
} 