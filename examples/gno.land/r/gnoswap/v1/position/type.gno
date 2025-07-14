package position

import (
	"std"

	u256 "gno.land/p/gnoswap/uint256"
)

// Position represents a liquidity position in a pool.
// Each position tracks the amount of liquidity, fee growth, and tokens owed to the position owner.
type Position struct {
	nonce     *u256.Uint  // nonce for permits
	operator  std.Address // address that is approved for spending this token
	poolKey   string      // poolPath of the pool which this has lp token
	tickLower int32       // the lower tick of the position, bounds are included
	tickUpper int32       // the upper tick of the position
	liquidity *u256.Uint  // liquidity of the position

	// fee growth of the aggregate position as of the last action on the individual position
	feeGrowthInside0LastX128 *u256.Uint
	feeGrowthInside1LastX128 *u256.Uint

	// how many uncollected tokens are owed to the position, as of the last computation
	tokensOwed0 *u256.Uint
	tokensOwed1 *u256.Uint

	token0Balance *u256.Uint // token0 balance of the position
	token1Balance *u256.Uint // token1 balance of the position

	burned bool // whether the position has been burned(≈ actuall we don't burn lp token even its empty, just update flag)
}

func (p Position) PoolKey() string { return p.poolKey }
func (p Position) Liquidity() *u256.Uint { return p.liquidity }
func (p Position) TickLower() int32 { return p.tickLower }
func (p Position) TickUpper() int32 { return p.tickUpper }
func (p Position) TokensOwed0() *u256.Uint { return p.tokensOwed0 }
func (p Position) TokensOwed1() *u256.Uint { return p.tokensOwed1 }
func (p Position) Token0Balance() *u256.Uint { return p.token0Balance }
func (p Position) Token1Balance() *u256.Uint { return p.token1Balance }

// isClear reports whether the position is empty
func (p Position) isClear() bool {
	return p.liquidity.IsZero() && p.tokensOwed0.IsZero() && p.tokensOwed1.IsZero()
}

type MintParams struct {
	token0         string      // token0 path for a specific pool
	token1         string      // token1 path for a specific pool
	fee            uint32      // fee for a specific pool
	tickLower      int32       // lower end of the tick range for the position
	tickUpper      int32       // upper end of the tick range for the position
	amount0Desired *u256.Uint  // desired amount of token0 to be minted
	amount1Desired *u256.Uint  // desired amount of token1 to be minted
	amount0Min     *u256.Uint  // minimum amount of token0 to be minted
	amount1Min     *u256.Uint  // minimum amount of token1 to be minted
	deadline       int64       // time by which the transaction must be included to effect the change
	mintTo         std.Address // address to mint lpToken
	caller         std.Address // address to call the function
}

// newMintParams creates `MintParams` from processed input data.
func newMintParams(input ProcessedMintInput, mintInput MintInput) MintParams {
	return MintParams{
		token0:         input.tokenPair.token0,
		token1:         input.tokenPair.token1,
		fee:            mintInput.fee,
		tickLower:      input.tickLower,
		tickUpper:      input.tickUpper,
		amount0Desired: input.amount0Desired,
		amount1Desired: input.amount1Desired,
		amount0Min:     input.amount0Min,
		amount1Min:     input.amount1Min,
		deadline:       mintInput.deadline,
		mintTo:         mintInput.mintTo,
		caller:         mintInput.caller,
	}
}

type IncreaseLiquidityParams struct {
	positionId     uint64      // positionId of the position to increase liquidity
	amount0Desired *u256.Uint  // desired amount of token0 to be minted
	amount1Desired *u256.Uint  // desired amount of token1 to be minted
	amount0Min     *u256.Uint  // minimum amount of token0 to be minted
	amount1Min     *u256.Uint  // minimum amount of token1 to be minted
	deadline       int64       // time by which the transaction must be included to effect the change
	caller         std.Address // address to call the function
}

type DecreaseLiquidityParams struct {
	positionId   uint64      // positionId of the position to decrease liquidity
	liquidity    string      // amount of liquidity to decrease
	amount0Min   *u256.Uint  // minimum amount of token0 to be minted
	amount1Min   *u256.Uint  // minimum amount of token1 to be minted
	deadline     int64       // time by which the transaction must be included to effect the change
	unwrapResult bool        // whether to unwrap the token if it's wrapped native token
	caller       std.Address // address to call the function
}

type MintInput struct {
	token0         string
	token1         string
	fee            uint32
	tickLower      int32
	tickUpper      int32
	amount0Desired string
	amount1Desired string
	amount0Min     string
	amount1Min     string
	deadline       int64
	mintTo         std.Address
	caller         std.Address
}

type TokenPair struct {
	token0         string
	token1         string
	token0IsNative bool
	token1IsNative bool
	wrappedAmount  int64
}

type ProcessedMintInput struct {
	tokenPair      TokenPair
	amount0Desired *u256.Uint
	amount1Desired *u256.Uint
	amount0Min     *u256.Uint
	amount1Min     *u256.Uint
	tickLower      int32
	tickUpper      int32
	poolPath       string
}

// FeeGrowthInside represents fee growth inside ticks
type FeeGrowthInside struct {
	feeGrowthInside0LastX128 *u256.Uint
	feeGrowthInside1LastX128 *u256.Uint
}
