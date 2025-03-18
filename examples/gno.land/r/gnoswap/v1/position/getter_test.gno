package position

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

func setupPositionGetter(t *testing.T) {
	t.Helper()

	CreatePoolWithoutFee(t)
	std.TestSetRealm(std.NewUserRealm(adminAddr))
	position := Position{
		nonce:                    u256.Zero(),
		operator:                 positionAddr,
		poolKey:                  "gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500",
		tickLower:                -10000,
		tickUpper:                10000,
		liquidity:                u256.NewUint(1000000),
		feeGrowthInside0LastX128: u256.Zero(),
		feeGrowthInside1LastX128: u256.Zero(),
		tokensOwed0:              u256.Zero(),
		tokensOwed1:              u256.Zero(),
		burned:                   false,
	}
	positionId := GetNextId()
	createNewPosition(positionId, position)
}

func TestPositionGetPosition(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	position := PositionGetPosition(positionId)
	uassert.Equal(t, "gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500", position.poolKey)
	uassert.Equal(t, int32(-10000), position.tickLower)
	uassert.Equal(t, int32(10000), position.tickUpper)
	uassert.Equal(t, "1000000", position.liquidity.ToString())
	uassert.Equal(t, "0", position.feeGrowthInside0LastX128.ToString())
	uassert.Equal(t, "0", position.feeGrowthInside1LastX128.ToString())
	uassert.Equal(t, "0", position.tokensOwed0.ToString())
	uassert.Equal(t, "0", position.tokensOwed1.ToString())
	uassert.Equal(t, false, position.burned)
}

func TestPositionGetPositionNonce(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	nonce := PositionGetPositionNonce(positionId)
	uassert.Equal(t, "0", nonce.ToString())
}

func TestPositionGetPositionOperator(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	operator := PositionGetPositionOperator(positionId)
	uassert.Equal(t, positionAddr, operator)
}

func TestPositionGetPositionPoolKey(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	poolKey := PositionGetPositionPoolKey(positionId)
	uassert.Equal(t, "gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500", poolKey)
}

func TestPositionGetPositionTickLower(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	tickLower := PositionGetPositionTickLower(positionId)
	uassert.Equal(t, int32(-10000), tickLower)
}

func TestPositionGetPositionTickUpper(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	tickUpper := PositionGetPositionTickUpper(positionId)
	uassert.Equal(t, int32(10000), tickUpper)
}

func TestPositionGetPositionLiquidity(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	liquidity := PositionGetPositionLiquidity(positionId)
	uassert.Equal(t, "1000000", liquidity.ToString())
}

func TestPositionGetPositionFeeGrowthInside0LastX128(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	feeGrowth0 := PositionGetPositionFeeGrowthInside0LastX128(positionId)
	uassert.Equal(t, "0", feeGrowth0.ToString())
}

func TestPositionGetPositionFeeGrowthInside1LastX128(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	feeGrowth1 := PositionGetPositionFeeGrowthInside1LastX128(positionId)
	uassert.Equal(t, "0", feeGrowth1.ToString())
}

func TestPositionGetPositionTokensOwed0(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	tokenOwed0 := PositionGetPositionTokensOwed0(positionId)
	uassert.Equal(t, "0", tokenOwed0.ToString())
}

func TestPositionGetPositionTokensOwed1(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	tokenOwed1 := PositionGetPositionTokensOwed1(positionId)
	uassert.Equal(t, "0", tokenOwed1.ToString())
}

func TestPositionGetPositionIsBurned(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	isBurn := PositionGetPositionIsBurned(positionId)
	uassert.Equal(t, false, isBurn)
}

func TestPositionIsInRange(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	inRange := PositionIsInRange(positionId)
	uassert.Equal(t, true, inRange)
}

func TestPositionGetPositionNonceStr(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	nonceStr := PositionGetPositionNonceStr(positionId)
	uassert.Equal(t, "0", nonceStr)
}

func TestPositionGetPositionOperatorStr(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	operator := PositionGetPositionOperatorStr(positionId)
	uassert.Equal(t, string(positionAddr), operator)
}

func TestPositionGetPositionLiquidityStr(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	liquidity := PositionGetPositionLiquidityStr(positionId)
	uassert.Equal(t, "1000000", liquidity)
}

func TestPositionGetPositionFeeGrowthInside0LastX128Str(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	feeGrowth0 := PositionGetPositionFeeGrowthInside0LastX128Str(positionId)
	uassert.Equal(t, "0", feeGrowth0)
}

func TestPositionGetPositionFeeGrowthInside1LastX128Str(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	feeGrowth1 := PositionGetPositionFeeGrowthInside1LastX128Str(positionId)
	uassert.Equal(t, "0", feeGrowth1)
}

func TestPositionGetPositionTokensOwed0Str(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	tokenOwed0 := PositionGetPositionTokensOwed0Str(positionId)
	uassert.Equal(t, "0", tokenOwed0)
}

func TestPositionGetPositionTokensOwed1Str(t *testing.T) {
	positionId := GetNextId()
	setupPositionGetter(t)
	tokenOwed1 := PositionGetPositionTokensOwed1Str(positionId)
	uassert.Equal(t, "0", tokenOwed1)
}
