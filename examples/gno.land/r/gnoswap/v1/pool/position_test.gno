package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
)

func TestPositionGetKey(t *testing.T) {
	invalidAddr := std.Address("invalidAddr")
	validAddr := testutils.TestAddress("validAddr")

	tests := []struct {
		owner       std.Address
		tickLower   int32
		tickUpper   int32
		shouldPanic bool
		panicMsg    string
		expectedKey string
	}{
		{invalidAddr, 100, 200, true, `[GNOSWAP-POOL-023] invalid address || (invalidAddr)`, ""},                                                    // invalid address
		{validAddr, 200, 100, true, `[GNOSWAP-POOL-024] tickLower is greater than or equal to tickUpper || tickLower(200), tickUpper(100)`, ""},     // tickLower > tickUpper
		{validAddr, -100, -200, true, `[GNOSWAP-POOL-024] tickLower is greater than or equal to tickUpper || tickLower(-100), tickUpper(-200)`, ""}, // tickLower > tickUpper
		{validAddr, 100, 100, true, "[GNOSWAP-POOL-024] tickLower is greater than or equal to tickUpper || tickLower(100), tickUpper(100)", ""},     // tickLower == tickUpper
		{validAddr, 100, 200, false, "", "ZzF3ZXNrYzZ0eWc5anhndWpsdGEwNDdoNmx0YTA0N2g2bGRqbHVkdV9fMTAwX18yMDA="},                                    // tickLower < tickUpper
	}

	for _, tc := range tests {
		defer func() {
			if r := recover(); r != nil {
				uassert.Equal(t, tc.panicMsg, r.(string))
			}
		}()
		if tc.shouldPanic {
			uassert.PanicsWithMessage(t, tc.panicMsg, func() { getPositionKey(tc.owner, tc.tickLower, tc.tickUpper) })
		} else {
			key := getPositionKey(tc.owner, tc.tickLower, tc.tickUpper)
			uassert.Equal(t, tc.expectedKey, key)
		}
	}
}

func TestPositionUpdateWithKey(t *testing.T) {
	var dummyPool *Pool
	var positionKey string

	t.Run("set up initial data for this test function", func(t *testing.T) {
		poolParams := newPoolParams(
			"token0",
			"token1",
			100,
			common.TickMathGetSqrtRatioAtTick(0).ToString(),
		)
		dummyPool = newPool(poolParams)

		positionKey = getPositionKey(
			testutils.TestAddress("dummyAddr"),
			100,
			200,
		)
	})

	tests := []struct {
		liquidity         *i256.Int
		amount0           *u256.Uint
		amount1           *u256.Uint
		shouldPanic       bool
		panicMsg          string
		expectedLiquidity string
	}{
		{i256.MustFromDecimal("0"), u256.Zero(), u256.Zero(), true, `[GNOSWAP-POOL-010] zero liquidity || both liquidityDelta and current position's liquidity are zero`, ""},
		{i256.MustFromDecimal("100000"), u256.Zero(), u256.Zero(), false, "", "100000"},
	}

	for _, tc := range tests {
		if tc.shouldPanic {
			uassert.PanicsWithMessage(t, tc.panicMsg, func() { dummyPool.positionUpdateWithKey(positionKey, tc.liquidity, tc.amount0, tc.amount1) })
		} else {
			newPos := dummyPool.positionUpdateWithKey(positionKey, tc.liquidity, tc.amount0, tc.amount1)
			uassert.Equal(t, newPos.liquidity.ToString(), tc.expectedLiquidity)
		}
	}
}

func TestPositionUpdate(t *testing.T) {
	tests := []struct {
		initialLiquidity             *u256.Uint
		liquidityDelta               *i256.Int
		feeGrowthInside0X128         *u256.Uint
		feeGrowthInside1X128         *u256.Uint
		shouldPanic                  bool
		panicMsg                     string
		expectedLiquidity            string
		expectedFeeGrowthInside0X128 string
		expectedFeeGrowthInside1X128 string
		expectedToken0Owed           string
		expectedToken1Owed           string
	}{
		{
			initialLiquidity:     u256.Zero(),
			liquidityDelta:       i256.MustFromDecimal("0"),
			feeGrowthInside0X128: u256.Zero(),
			feeGrowthInside1X128: u256.Zero(),
			shouldPanic:          true,
			panicMsg:             `[GNOSWAP-POOL-010] zero liquidity || both liquidityDelta and current position's liquidity are zero`,
		},
		{
			initialLiquidity:             u256.Zero(),
			liquidityDelta:               i256.MustFromDecimal("100000"),
			feeGrowthInside0X128:         u256.Zero(),
			feeGrowthInside1X128:         u256.Zero(),
			expectedLiquidity:            "100000",
			expectedFeeGrowthInside0X128: "0",
			expectedFeeGrowthInside1X128: "0",
			expectedToken0Owed:           "0",
			expectedToken1Owed:           "0",
		},
		{
			initialLiquidity:             u256.Zero(),
			liquidityDelta:               i256.MustFromDecimal("100000"),
			feeGrowthInside0X128:         u256.MustFromDecimal("100000000"),
			feeGrowthInside1X128:         u256.MustFromDecimal("100000000"),
			expectedLiquidity:            "100000",
			expectedFeeGrowthInside0X128: "100000000",
			expectedFeeGrowthInside1X128: "100000000",
			expectedToken0Owed:           "0",
			expectedToken1Owed:           "0",
		},
		{
			initialLiquidity:             u256.NewUint(100000),
			liquidityDelta:               i256.MustFromDecimal("100000"),
			feeGrowthInside0X128:         u256.MustFromDecimal("100000000"),
			feeGrowthInside1X128:         u256.MustFromDecimal("100000000"),
			expectedLiquidity:            "200000",
			expectedFeeGrowthInside0X128: "100000000",
			expectedFeeGrowthInside1X128: "100000000",
			expectedToken0Owed:           "0",
			expectedToken1Owed:           "0",
		},
		{
			initialLiquidity:             u256.MustFromDecimal("340282366920938463463374607431768211456"), // Q128 value
			liquidityDelta:               i256.Zero(),
			feeGrowthInside0X128:         u256.MustFromDecimal("100000000"),
			feeGrowthInside1X128:         u256.MustFromDecimal("200000000"),
			expectedLiquidity:            "340282366920938463463374607431768211456",
			expectedFeeGrowthInside0X128: "100000000",
			expectedFeeGrowthInside1X128: "200000000",
			expectedToken0Owed:           "100000000",
			expectedToken1Owed:           "200000000",
		},
	}

	for _, tc := range tests {
		position := PositionInfo{
			liquidity: tc.initialLiquidity,
		}

		if tc.shouldPanic {
			uassert.PanicsWithMessage(t, tc.panicMsg, func() { positionUpdate(position, tc.liquidityDelta, tc.feeGrowthInside0X128, tc.feeGrowthInside1X128) })
		} else {
			newPos := positionUpdate(position, tc.liquidityDelta, tc.feeGrowthInside0X128, tc.feeGrowthInside1X128)
			uassert.Equal(t, newPos.liquidity.ToString(), tc.expectedLiquidity)
			uassert.Equal(t, newPos.feeGrowthInside0LastX128.ToString(), tc.expectedFeeGrowthInside0X128)
			uassert.Equal(t, newPos.feeGrowthInside1LastX128.ToString(), tc.expectedFeeGrowthInside1X128)
			uassert.Equal(t, newPos.tokensOwed0.ToString(), tc.expectedToken0Owed)
			uassert.Equal(t, newPos.tokensOwed1.ToString(), tc.expectedToken1Owed)
		}
	}
}
