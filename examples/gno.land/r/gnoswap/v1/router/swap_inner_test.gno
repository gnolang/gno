package router

import (
	"std"
	"testing"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"

	"gno.land/p/demo/uassert"
)

func TestCalculateSqrtPriceLimitForSwap(t *testing.T) {
	tests := []struct {
		name              string
		zeroForOne        bool
		fee               uint32
		sqrtPriceLimitX96 *u256.Uint
		expected          *u256.Uint
	}{
		{
			name:              "already set sqrtPriceLimit",
			zeroForOne:        true,
			fee:               500,
			sqrtPriceLimitX96: u256.NewUint(1000),
			expected:          u256.NewUint(1000),
		},
		{
			name:              "when zeroForOne is true, calculate min tick",
			zeroForOne:        true,
			fee:               500,
			sqrtPriceLimitX96: u256.Zero(),
			expected: common.TickMathGetSqrtRatioAtTick(getMinTick(500)).Add(
				common.TickMathGetSqrtRatioAtTick(getMinTick(500)),
				u256.One(),
			),
		},
		{
			name:              "when zeroForOne is false, calculate max tick",
			zeroForOne:        false,
			fee:               500,
			sqrtPriceLimitX96: u256.Zero(),
			expected: common.TickMathGetSqrtRatioAtTick(getMaxTick(500)).Sub(
				common.TickMathGetSqrtRatioAtTick(getMaxTick(500)),
				u256.One(),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSqrtPriceLimitForSwap(
				tt.zeroForOne,
				tt.fee,
				tt.sqrtPriceLimitX96,
			)
			uassert.Equal(t, result.ToString(), tt.expected.ToString())
		})
	}
}

func TestSwapInner(t *testing.T) {
	user1Realm := std.NewUserRealm(user1Addr)

	tests := []struct {
		name              string
		setupFn           func(t *testing.T)
		amountSpecified   *i256.Int
		recipient         std.Address
		sqrtPriceLimitX96 *u256.Uint
		data              SwapCallbackData
		expectedRecv      string
		expectedOut       string
		expectError       bool
	}{
		{
			name: "normal swap - exact input",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)

				std.TestSetRealm(user1Realm)
				bar.Approve(consts.ROUTER_ADDR, maxApprove)
				baz.Approve(consts.ROUTER_ADDR, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
				TokenFaucet(t, bazPath, user1Addr)
			},
			amountSpecified:   i256.MustFromDecimal("100"),
			recipient:         alice,
			sqrtPriceLimitX96: u256.NewUint(4295128740),
			data: SwapCallbackData{
				tokenIn:  barPath,
				tokenOut: bazPath,
				fee:      3000,
				payer:    consts.ROUTER_ADDR,
			},
			expectedRecv: "100",
			expectedOut:  "98",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		std.TestSetRealm(user1Realm)
		bar.Approve(consts.ROUTER_ADDR, maxApprove)
		baz.Approve(consts.ROUTER_ADDR, maxApprove)
		bar.Approve(consts.POOL_ADDR, maxApprove)
		baz.Approve(consts.POOL_ADDR, maxApprove)
		TokenFaucet(t, barPath, consts.ROUTER_ADDR)

		if tt.setupFn != nil {
			tt.setupFn(t)
		}

		poolRecv, poolOut := swapInner(
			tt.amountSpecified,
			tt.recipient,
			tt.sqrtPriceLimitX96,
			tt.data,
		)

		uassert.Equal(t, poolRecv.ToString(), tt.expectedRecv)
		uassert.Equal(t, poolOut.ToString(), tt.expectedOut)
	}
}
