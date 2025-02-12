package position

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/r/gnoswap/v1/common"
	pl "gno.land/r/gnoswap/v1/pool"
)

func InitialSetup(t *testing.T) {
	std.TestSetRealm(adminRealm)

	pl.SetPoolCreationFeeByAdmin(0)
	CreatePool(t, gnsPath, barPath, fee3000, common.TickMathGetSqrtRatioAtTick(0).ToString(), admin)
	TokenFaucet(t, gnsPath, alice)
	TokenFaucet(t, barPath, alice)
}

func TestAddLiquidity(t *testing.T) {
	poolKey := computePoolPath(gnsPath, barPath, fee3000)

	tests := []struct {
		name            string
		params          AddLiquidityParams
		expectPanic     bool
		expectedAmount0 string
		expectedAmount1 string
	}{
		{
			name: "Successful Liquidity Addition",
			params: AddLiquidityParams{
				poolKey:        poolKey,
				tickLower:      -600,
				tickUpper:      600,
				amount0Desired: u256.MustFromDecimal("1000000"),
				amount1Desired: u256.MustFromDecimal("2000000"),
				amount0Min:     u256.MustFromDecimal("400000"),
				amount1Min:     u256.MustFromDecimal("800000"),
				caller:         alice,
			},
			expectPanic:     false,
			expectedAmount0: "1000000",
			expectedAmount1: "1000000",
		},
		{
			name: "Slippage Panic",
			params: AddLiquidityParams{
				poolKey:        poolKey,
				tickLower:      -600,
				tickUpper:      600,
				amount0Desired: u256.MustFromDecimal("1000000"),
				amount1Desired: u256.MustFromDecimal("2000000"),
				amount0Min:     u256.MustFromDecimal("1100000"),
				amount1Min:     u256.MustFromDecimal("2200000"),
				caller:         alice,
			},
			expectPanic: true,
		},
		{
			name: "Zero Liquidity",
			params: AddLiquidityParams{
				poolKey:        poolKey,
				tickLower:      -100,
				tickUpper:      100,
				amount0Desired: u256.Zero(),
				amount1Desired: u256.Zero(),
				amount0Min:     u256.Zero(),
				amount1Min:     u256.Zero(),
				caller:         alice,
			},
			expectPanic:     true,
			expectedAmount0: "0",
			expectedAmount1: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("unexpected panic: %v", r)
					}
				} else {
					if tt.expectPanic {
						t.Errorf("expected panic but did not occur")
					}
				}
			}()

			InitialSetup(t)
			std.TestSetRealm(std.NewUserRealm(alice))
			TokenApprove(t, gnsPath, alice, consts.POOL_ADDR, tt.params.amount0Desired.Uint64())
			TokenApprove(t, barPath, alice, consts.POOL_ADDR, tt.params.amount1Desired.Uint64())

			_, amount0, amount1 := addLiquidity(tt.params)

			if !tt.expectPanic {
				uassert.Equal(t, tt.expectedAmount0, amount0.ToString())
				uassert.Equal(t, tt.expectedAmount1, amount1.ToString())
			}
		})
	}
}
