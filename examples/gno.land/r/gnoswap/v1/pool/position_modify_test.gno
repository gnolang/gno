package pool

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/r/gnoswap/v1/common"
)

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
				fee500,
				sqrtPrice.ToString(),
			)
			pool := newPool(poolParams)

			params := ModifyPositionParams{
				owner:          consts.POSITION_ADDR,
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
			fee500,
			sqrtPrice,
		)
		pool := newPool(poolParams)
		params := ModifyPositionParams{
			owner:          consts.POSITION_ADDR,
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
			fee500,
			sqrtPrice,
		)
		pool := newPool(poolParams)
		params := ModifyPositionParams{
			owner:          consts.POSITION_ADDR,
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
			expected:    "[GNOSWAP-POOL-028] tickLower is invalid || tickLower(-887273) < MIN_TICK(-887272)",
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
			shouldPanic: false,
			expected:    "",
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
			expected:    "[GNOSWAP-POOL-029] tickUpper is invalid || tickUpper(887273) > MAX_TICK(887272)",
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
