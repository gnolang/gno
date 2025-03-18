package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
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
			GetFeeAmountTickSpacing(100),
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

func TestModifyPosition(t *testing.T) {
	const (
		fee500              = uint32(500)
		test_liquidityDelta = int64(100000000)
	)

	tests := []struct {
		name         string
		sqrtPrice    string
		tickLower    int32
		tickUpper    int32
		expectedAmt0 string
		expectedAmt1 string
	}{
		{
			name:         "current price is lower than range",
			sqrtPrice:    common.TickMathGetSqrtRatioAtTick(-12000).ToString(),
			tickLower:    -11000,
			tickUpper:    -9000,
			expectedAmt0: "16492846",
			expectedAmt1: "0",
		},
		{
			name:         "current price is in range",
			sqrtPrice:    common.TickMathGetSqrtRatioAtTick(-10000).ToString(),
			tickLower:    -11000,
			tickUpper:    -9000,
			expectedAmt0: "8040316",
			expectedAmt1: "2958015",
		},
		{
			name:         "current price is higher than range",
			sqrtPrice:    common.TickMathGetSqrtRatioAtTick(-8000).ToString(),
			tickLower:    -11000,
			tickUpper:    -9000,
			expectedAmt0: "0",
			expectedAmt1: "6067683",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqrtPrice := u256.MustFromDecimal(tt.sqrtPrice)
			poolParams := newPoolParams(
				barPath,
				fooPath,
				FeeTier500,
				sqrtPrice.ToString(),
				GetFeeAmountTickSpacing(fee500),
			)
			pool := newPool(poolParams)

			params := ModifyPositionParams{
				owner:          positionAddr,
				tickLower:      tt.tickLower,
				tickUpper:      tt.tickUpper,
				liquidityDelta: new(i256.Int).SetInt64(test_liquidityDelta),
			}

			_, amount0, amount1 := pool.modifyPosition(params)

			uassert.Equal(t, amount0.ToString(), tt.expectedAmt0)
			uassert.Equal(t, amount1.ToString(), tt.expectedAmt1)
		})
	}
}

func TestModifyPositionEdgeCases(t *testing.T) {
	const fee500 = uint32(500)
	sqrtPrice := common.TickMathGetSqrtRatioAtTick(-10000).ToString()

	t.Run("liquidityDelta is zero", func(t *testing.T) {
		poolParams := newPoolParams(
			barPath,
			fooPath,
			FeeTier500,
			sqrtPrice,
			GetFeeAmountTickSpacing(fee500),
		)
		pool := newPool(poolParams)
		params := ModifyPositionParams{
			owner:          positionAddr,
			tickLower:      -11000,
			tickUpper:      -9000,
			liquidityDelta: i256.Zero(),
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic. not happened")
			} else {
				expectedError := "[GNOSWAP-POOL-010] zero liquidity"
				if err, ok := r.(string); !ok || err[:len(expectedError)] != expectedError {
					t.Errorf("expected error message. got: %v, want: %v", r, expectedError)
				}
			}
		}()

		pool.modifyPosition(params)
	})

	t.Run("liquidityDelta is negative", func(t *testing.T) {
		poolParams := newPoolParams(
			barPath,
			fooPath,
			FeeTier500,
			sqrtPrice,
			GetFeeAmountTickSpacing(fee500),
		)
		pool := newPool(poolParams)
		params := ModifyPositionParams{
			owner:          positionAddr,
			tickLower:      -11000,
			tickUpper:      -9000,
			liquidityDelta: i256.MustFromDecimal("100000000"),
		}
		pool.modifyPosition(params)

		// remove liquidity
		params.liquidityDelta = i256.MustFromDecimal("-100000000")
		_, amount0, amount1 := pool.modifyPosition(params)

		// remove amount should be same as added amount
		uassert.Equal(t, amount0.ToString(), "8040315")
		uassert.Equal(t, amount1.ToString(), "2958014")
	})
}

func TestAssertTickLowerLessThanUpper(t *testing.T) {
	tests := []struct {
		name        string
		tickLower   int32
		tickUpper   int32
		shouldPanic bool
		expected    string
	}{
		{
			name:        "tickLower is less than tickUpper",
			tickLower:   -100,
			tickUpper:   100,
			shouldPanic: false,
			expected:    "",
		},
		{
			name:        "tickLower equals tickUpper",
			tickLower:   50,
			tickUpper:   50,
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-024] tickLower is greater than or equal to tickUpper || tickLower(50), tickUpper(50)",
		},
		{
			name:        "tickLower greater than tickUpper",
			tickLower:   200,
			tickUpper:   100,
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-024] tickLower is greater than or equal to tickUpper || tickLower(200), tickUpper(100)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					if tt.shouldPanic {
						t.Errorf("unexpected panic: %v", r)
					}
					return
				} else {
					if tt.shouldPanic {
						uassert.Equal(t, tt.expected, r.(string))
					} else {
						t.Errorf("expected no panic, but got: %v", r)
					}
				}
			}()
			assertTickLowerLessThanUpper(tt.tickLower, tt.tickUpper)
		})
	}
}

func TestAssertValidTickLower(t *testing.T) {
	tests := []struct {
		name        string
		tickLower   int32
		tickUpper   int32
		shouldPanic bool
		expected    string
	}{
		{
			name:        "tickLower equals MIN_TICK",
			tickLower:   consts.MIN_TICK,
			tickUpper:   100,
			shouldPanic: false,
			expected:    "",
		},
		{
			name:        "tickLower greater than MIN_TICK",
			tickLower:   consts.MIN_TICK + 1,
			tickUpper:   50,
			shouldPanic: false,
			expected:    "",
		},
		{
			name:        "tickLower less than MIN_TICK (panic)",
			tickLower:   consts.MIN_TICK - 1,
			tickUpper:   100,
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-013] tickLower is invalid || tickLower(-887273) < MIN_TICK(-887272)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					if tt.shouldPanic {
						t.Errorf("unexpected panic: %v", r)
					}
					return
				} else {
					if tt.shouldPanic {
						uassert.Equal(t, tt.expected, r.(string))
					} else {
						t.Errorf("expected no panic, but got: %v", r)
					}
				}
			}()
			assertValidTickLower(tt.tickLower)
		})
	}
}

func TestAssertValidTickUpper(t *testing.T) {
	tests := []struct {
		name        string
		tickLower   int32
		tickUpper   int32
		shouldPanic bool
		expected    string
	}{
		{
			name:        "tickUpper equals MAX_TICK",
			tickLower:   consts.MIN_TICK,
			tickUpper:   consts.MAX_TICK,
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-014] tickUpper is invalid || tickUpper(887272) > MAX_TICK(887272)",
		},
		{
			name:        "tickUpper less than MAX_TICK",
			tickLower:   consts.MIN_TICK + 1,
			tickUpper:   consts.MAX_TICK - 1,
			shouldPanic: false,
			expected:    "",
		},
		{
			name:        "tickUpper greater than MAX_TICK (panic)",
			tickLower:   consts.MIN_TICK - 1,
			tickUpper:   consts.MAX_TICK + 1,
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-014] tickUpper is invalid || tickUpper(887273) > MAX_TICK(887272)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					if tt.shouldPanic {
						t.Errorf("unexpected panic: %v", r)
					}
					return
				} else {
					if tt.shouldPanic {
						uassert.Equal(t, tt.expected, r.(string))
					} else {
						t.Errorf("expected no panic, but got: %v", r)
					}
				}
			}()
			assertValidTickUpper(tt.tickUpper)
		})
	}
}

func TestUpdatePosition(t *testing.T) {
	poolParams := &createPoolParams{
		token0Path:   "token0",
		token1Path:   "token1",
		fee:          500,
		tickSpacing:  10,
		sqrtPriceX96: u256.MustFromDecimal("1000000000000000000"), // 1.0
	}
	p := newPool(poolParams)

	tests := []struct {
		name            string
		positionParams  ModifyPositionParams
		expectLiquidity *u256.Uint
	}{
		{
			name: "add new position",
			positionParams: ModifyPositionParams{
				owner:          positionAddr,
				tickLower:      -100,
				tickUpper:      100,
				liquidityDelta: i256.MustFromDecimal("1000000"),
			},
			expectLiquidity: u256.MustFromDecimal("1000000"),
		},
		{
			name: "add liquidity to existing position",
			positionParams: ModifyPositionParams{
				owner:          positionAddr,
				tickLower:      -100,
				tickUpper:      100,
				liquidityDelta: i256.MustFromDecimal("500000"),
			},
			expectLiquidity: u256.MustFromDecimal("1500000"),
		},
		{
			name: "remove liquidity from position",
			positionParams: ModifyPositionParams{
				owner:          positionAddr,
				tickLower:      -100,
				tickUpper:      100,
				liquidityDelta: i256.MustFromDecimal("-500000"),
			},
			expectLiquidity: u256.MustFromDecimal("1000000"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tick := p.Slot0Tick()
			position := p.updatePosition(tt.positionParams, tick)

			if !position.liquidity.Eq(tt.expectLiquidity) {
				t.Errorf("liquidity mismatch: expected %s, got %s",
					tt.expectLiquidity.ToString(),
					position.liquidity.ToString())
			}

			if !tt.positionParams.liquidityDelta.IsZero() {
				lowerTick := p.mustGetTick(tt.positionParams.tickLower)
				upperTick := p.mustGetTick(tt.positionParams.tickUpper)

				if !lowerTick.initialized {
					t.Error("lower tick not initialized")
				}
				if !upperTick.initialized {
					t.Error("upper tick not initialized")
				}
			}
		})
	}
}
