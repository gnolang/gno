package router

import (
	"std"
	"testing"

	i256 "gno.land/p/gnoswap/int256"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
	"gno.land/r/onbloc/qux"

	"gno.land/p/demo/uassert"
)

func TestMultiSwap(t *testing.T) {
	user1Realm := std.NewUserRealm(user1Addr)

	tests := []struct {
		name             string
		setupFn          func(t *testing.T)
		params           SwapParams
		currentPoolIndex int
		numPools         int
		swapPath         string
		expectedFirstIn  string
		expectedLastOut  string
		expectError      bool
	}{
		{
			name: "single hop swap BAR -> BAZ",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)

				std.TestSetRealm(user1Realm)
				bar.Approve(consts.ROUTER_ADDR, maxApprove)
				baz.Approve(consts.ROUTER_ADDR, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			params: SwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             3000,
				recipient:       alice,
				amountSpecified: i256.MustFromDecimal("100"),
			},
			currentPoolIndex: 0,
			numPools:         1,
			swapPath:         "",
			expectedFirstIn:  "100",
			expectedLastOut:  "98",
			expectError:      false,
		},
		{
			name: "multi hop swap (BAR -> BAZ -> QUX)",
			setupFn: func(t *testing.T) {
				// BAR -> BAZ
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)

				// BAZ -> QUX
				CreateSecondPoolWithoutFee(t)
				MakeSecondMintPositionWithoutFee(t)

				std.TestSetRealm(user1Realm)
				bar.Approve(consts.ROUTER_ADDR, maxApprove)
				baz.Approve(consts.ROUTER_ADDR, maxApprove)
				qux.Approve(consts.ROUTER_ADDR, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			params: SwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             3000,
				recipient:       alice,
				amountSpecified: i256.MustFromDecimal("100"),
			},
			currentPoolIndex: 0,
			numPools:         2,
			swapPath:         "gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000*POOL*gno.land/r/onbloc/baz:gno.land/r/onbloc/qux:3000",
			expectedFirstIn:  "100",
			expectedLastOut:  "96",
			expectError:      false,
		},
		{
			name: "multi hop swap with exact output",
			setupFn: func(t *testing.T) {
				// BAR -> BAZ -> QUX
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				CreateSecondPoolWithoutFee(t)
				MakeSecondMintPositionWithoutFee(t)

				std.TestSetRealm(user1Realm)
				bar.Approve(consts.ROUTER_ADDR, maxApprove)
				baz.Approve(consts.ROUTER_ADDR, maxApprove)
				qux.Approve(consts.ROUTER_ADDR, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			params: SwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             3000,
				recipient:       alice,
				amountSpecified: i256.MustFromDecimal("-96"),
			},
			currentPoolIndex: 0,
			numPools:         2,
			swapPath:         "gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000*POOL*gno.land/r/onbloc/baz:gno.land/r/onbloc/qux:3000",
			expectedFirstIn:  "98",
			expectedLastOut:  "94",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			firstAmountIn, lastAmountOut := multiSwap(
				tt.params,
				tt.currentPoolIndex,
				tt.numPools,
				tt.swapPath,
			)

			uassert.Equal(t, firstAmountIn.ToString(), tt.expectedFirstIn)
			uassert.Equal(t, lastAmountOut.ToString(), tt.expectedLastOut)
		})
	}
}
