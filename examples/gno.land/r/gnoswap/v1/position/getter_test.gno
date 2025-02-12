package position

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"
)

func setupPositionGetter(t *testing.T) {
	t.Helper()

	CreatePoolWithoutFee(t)
	std.TestSetRealm(std.NewUserRealm(admin))
	position := Position{
		nonce:                    u256.Zero(),
		operator:                 consts.POSITION_ADDR,
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
	tokenId := getNextId()
	setPosition(tokenId, position)
}

func TestPositionGetPosition(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	position := PositionGetPosition(tokenId)
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
	setupPositionGetter(t)
	tokenId := getNextId()
	nonce := PositionGetPositionNonce(tokenId)
	uassert.Equal(t, "0", nonce.ToString())
}

func TestPositionGetPositionOperator(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	operator := PositionGetPositionOperator(tokenId)
	uassert.Equal(t, consts.POSITION_ADDR, operator)
}

func TestPositionGetPositionPoolKey(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	poolKey := PositionGetPositionPoolKey(tokenId)
	uassert.Equal(t, "gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500", poolKey)
}

func TestPositionGetPositionTickLower(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	tickLower := PositionGetPositionTickLower(tokenId)
	uassert.Equal(t, int32(-10000), tickLower)
}

func TestPositionGetPositionTickUpper(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	tickUpper := PositionGetPositionTickUpper(tokenId)
	uassert.Equal(t, int32(10000), tickUpper)
}

func TestPositionGetPositionLiquidity(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	liquidity := PositionGetPositionLiquidity(tokenId)
	uassert.Equal(t, "1000000", liquidity.ToString())
}

func TestPositionGetPositionFeeGrowthInside0LastX128(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	feeGrowth0 := PositionGetPositionFeeGrowthInside0LastX128(tokenId)
	uassert.Equal(t, "0", feeGrowth0.ToString())
}

func TestPositionGetPositionFeeGrowthInside1LastX128(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	feeGrowth1 := PositionGetPositionFeeGrowthInside1LastX128(tokenId)
	uassert.Equal(t, "0", feeGrowth1.ToString())
}

func TestPositionGetPositionTokensOwed0(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	tokenOwed0 := PositionGetPositionTokensOwed0(tokenId)
	uassert.Equal(t, "0", tokenOwed0.ToString())
}

func TestPositionGetPositionTokensOwed1(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	tokenOwed1 := PositionGetPositionTokensOwed1(tokenId)
	uassert.Equal(t, "0", tokenOwed1.ToString())
}

func TestPositionGetPositionIsBurned(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	isBurn := PositionGetPositionIsBurned(tokenId)
	uassert.Equal(t, false, isBurn)
}

func TestPositionIsInRange(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	inRange := PositionIsInRange(tokenId)
	uassert.Equal(t, true, inRange)
}

func TestPositionGetPositionNonceStr(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	nonceStr := PositionGetPositionNonceStr(tokenId)
	uassert.Equal(t, "0", nonceStr)
}

func TestPositionGetPositionOperatorStr(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	operator := PositionGetPositionOperatorStr(tokenId)
	uassert.Equal(t, string(consts.POSITION_ADDR), operator)
}

func TestPositionGetPositionLiquidityStr(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	liquidity := PositionGetPositionLiquidityStr(tokenId)
	uassert.Equal(t, "1000000", liquidity)
}

func TestPositionGetPositionFeeGrowthInside0LastX128Str(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	feeGrowth0 := PositionGetPositionFeeGrowthInside0LastX128Str(tokenId)
	uassert.Equal(t, "0", feeGrowth0)
}

func TestPositionGetPositionFeeGrowthInside1LastX128Str(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	feeGrowth1 := PositionGetPositionFeeGrowthInside1LastX128Str(tokenId)
	uassert.Equal(t, "0", feeGrowth1)
}

func TestPositionGetPositionTokensOwed0Str(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	tokenOwed0 := PositionGetPositionTokensOwed0Str(tokenId)
	uassert.Equal(t, "0", tokenOwed0)
}

func TestPositionGetPositionTokensOwed1Str(t *testing.T) {
	setupPositionGetter(t)
	tokenId := getNextId()
	tokenOwed1 := PositionGetPositionTokensOwed1Str(tokenId)
	uassert.Equal(t, "0", tokenOwed1)
}
