package router

import (
	"std"
	"testing"

	i256 "gno.land/p/gnoswap/int256"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"

	"gno.land/p/demo/uassert"
)

func TestSingleSwap(t *testing.T) {
	user1Realm := std.NewUserRealm(user1Addr)

	tests := []struct {
		name        string
		setupFn     func(t *testing.T)
		params      SingleSwapParams
		expectedIn  string
		expectedOut string
		expectError bool
	}{
		{
			name: "exact input swap BAR -> BAZ",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)

				std.TestSetRealm(user1Realm)
				bar.Approve(consts.ROUTER_ADDR, maxApprove)
				baz.Approve(consts.ROUTER_ADDR, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             3000,
				amountSpecified: i256.MustFromDecimal("100"),
			},
			expectedIn:  "100",
			expectedOut: "98",
			expectError: false,
		},
		{
			name: "exact output swap BAR -> BAZ",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)

				std.TestSetRealm(user1Realm)
				bar.Approve(consts.ROUTER_ADDR, maxApprove)
				baz.Approve(consts.ROUTER_ADDR, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             3000,
				amountSpecified: i256.MustFromDecimal("-98"),
			},
			expectedIn:  "100",
			expectedOut: "98",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			amountIn, amountOut := singleSwap(tt.params)

			uassert.Equal(t, amountIn.ToString(), tt.expectedIn)
			uassert.Equal(t, amountOut.ToString(), tt.expectedOut)
		})
	}
}
