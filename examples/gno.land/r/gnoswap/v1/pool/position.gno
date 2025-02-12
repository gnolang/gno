package pool

import (
	"encoding/base64"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

var (
	Q128 = u256.MustFromDecimal(consts.Q128)
)

// getPositionKey generates a unique, encoded key for a liquidity position.
//
// This function creates a unique key for identifying a liquidity position in a pool. The key is based
// on the position's owner address, lower tick, and upper tick values. The generated key is then encoded
// as a base64 string to ensure compatibility and uniqueness.
//
// Parameters:
//   - owner: std.Address, the address of the position's owner.
//   - tickLower: int32, the lower tick boundary for the position.
//   - tickUpper: int32, the upper tick boundary for the position.
//
// Returns:
//   - string: A base64-encoded string representing the unique position key.
//
// Workflow:
//  1. Validates that the `owner` address is valid using `assertOnlyValidAddress`.
//  2. Ensures `tickLower` is less than `tickUpper` using `assertTickLowerLessThanUpper`.
//  3. Constructs the position key as a formatted string:
//     "<owner>__<tickLower>__<tickUpper>"
//  4. Encodes the generated position key into a base64 string for safety and uniqueness.
//  5. Returns the encoded position key.
//
// Example:
//
//	owner := std.Address("0x123456789")
//	positionKey := getPositionKey(owner, 100, 200)
//	fmt.Println("Position Key:", positionKey)
//	// Output: base64-encoded string representing "0x123456789__100__200"
//
// Notes:
//   - The base64 encoding ensures that the position key can be safely used as an identifier
//     across different systems or data stores.
//   - The function will panic if:
//   - The `owner` address is invalid.
//   - `tickLower` is greater than or equal to `tickUpper`.
func getPositionKey(
	owner std.Address,
	tickLower int32,
	tickUpper int32,
) string {
	assertOnlyValidAddress(owner)
	assertTickLowerLessThanUpper(tickLower, tickUpper)

	positionKey := ufmt.Sprintf("%s__%d__%d", owner.String(), tickLower, tickUpper)
	encodedPositionKey := base64.StdEncoding.EncodeToString([]byte(positionKey))
	return encodedPositionKey
}

// positionUpdate calculates and returns an updated PositionInfo.
func positionUpdate(
	position PositionInfo,
	liquidityDelta *i256.Int,
	feeGrowthInside0X128 *u256.Uint,
	feeGrowthInside1X128 *u256.Uint,
) PositionInfo {
	position.valueOrZero()

	var liquidityNext *u256.Uint
	if liquidityDelta.IsZero() {
		if position.liquidity.IsZero() {
			panic(addDetailToError(
				errZeroLiquidity,
				"both liquidityDelta and current position's liquidity are zero",
			))
		}

		liquidityNext = position.liquidity
	} else {
		liquidityNext = liquidityMathAddDelta(position.liquidity, liquidityDelta)
	}

	tokensOwed0 := u256.Zero()
	diff0 := new(u256.Uint).Sub(feeGrowthInside0X128, position.feeGrowthInside0LastX128)
	tokensOwed0 = u256.MulDiv(diff0, position.liquidity, Q128)

	tokensOwed1 := u256.Zero()
	diff1 := new(u256.Uint).Sub(feeGrowthInside1X128, position.feeGrowthInside1LastX128)
	tokensOwed1 = u256.MulDiv(diff1, position.liquidity, Q128)

	if !(liquidityDelta.IsZero()) {
		position.liquidity = liquidityNext
	}

	position.feeGrowthInside0LastX128 = feeGrowthInside0X128
	position.feeGrowthInside1LastX128 = feeGrowthInside1X128
	if tokensOwed0.Gt(u256.Zero()) || tokensOwed1.Gt(u256.Zero()) {
		position.tokensOwed0 = position.tokensOwed0.Add(position.tokensOwed0, tokensOwed0)
		position.tokensOwed1 = position.tokensOwed1.Add(position.tokensOwed1, tokensOwed1)
	}

	return position
}

// positionUpdateWithKey updates a position in the pool and returns the updated position.
func (p *Pool) positionUpdateWithKey(
	positionKey string,
	liquidityDelta *i256.Int,
	feeGrowthInside0X128 *u256.Uint,
	feeGrowthInside1X128 *u256.Uint,
) PositionInfo {
	// if pointer is nil, set to zero for calculation
	liquidityDelta = liquidityDelta.NilToZero()
	feeGrowthInside0X128 = feeGrowthInside0X128.NilToZero()
	feeGrowthInside1X128 = feeGrowthInside1X128.NilToZero()

	positionToUpdate, _ := p.GetPosition(positionKey)
	positionAfterUpdate := positionUpdate(positionToUpdate, liquidityDelta, feeGrowthInside0X128, feeGrowthInside1X128)

	p.setPosition(positionKey, positionAfterUpdate)

	return positionAfterUpdate
}

// PositionLiquidity returns the liquidity of a position.
func (p *Pool) PositionLiquidity(key string) *u256.Uint {
	return p.mustGetPosition(key).liquidity
}

// PositionFeeGrowthInside0LastX128 returns the fee growth of token0 inside a position.
func (p *Pool) PositionFeeGrowthInside0LastX128(key string) *u256.Uint {
	return p.mustGetPosition(key).feeGrowthInside0LastX128
}

// PositionFeeGrowthInside1LastX128 returns the fee growth of token1 inside a position.
func (p *Pool) PositionFeeGrowthInside1LastX128(key string) *u256.Uint {
	return p.mustGetPosition(key).feeGrowthInside1LastX128
}

// PositionTokensOwed0 returns the amount of token0 owed by a position.
func (p *Pool) PositionTokensOwed0(key string) *u256.Uint {
	return p.mustGetPosition(key).tokensOwed0
}

// PositionTokensOwed1 returns the amount of token1 owed by a position.
func (p *Pool) PositionTokensOwed1(key string) *u256.Uint {
	return p.mustGetPosition(key).tokensOwed1
}

// GetPosition returns the position info for a given key.
func (p *Pool) GetPosition(key string) (PositionInfo, bool) {
	iPositionInfo, exist := p.positions.Get(key)
	if !exist {
		newPosition := PositionInfo{}
		newPosition.valueOrZero()
		return newPosition, false
	}

	return iPositionInfo.(PositionInfo), true
}

// setPosition sets the position info for a given key.
func (p *Pool) setPosition(posKey string, positionInfo PositionInfo) {
	p.positions.Set(posKey, positionInfo)
}

// mustGetPosition returns the position info for a given key.
func (p *Pool) mustGetPosition(positionKey string) PositionInfo {
	positionInfo, exist := p.GetPosition(positionKey)
	if !exist {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("positionKey(%s) does not exist", positionKey),
		))
	}
	return positionInfo
}
