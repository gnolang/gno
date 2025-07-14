package router

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	i256 "gno.land/p/gnoswap/int256"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
	_ "gno.land/r/onbloc/qux"
)

func TestMultiSwap(t *testing.T) {
	user1Realm := std.NewUserRealm(user1Addr)
	routerAddr, _ := access.GetAddress(access.ROLE_ROUTER)

	tests := []struct {
		name            string
		setupFn         func(t *testing.T)
		params          SwapParams
		numPools        int
		swapPath        string
		expectedFirstIn string
		expectedLastOut string
		expectError     bool
	}{
		{
			name: "single hop swap BAR -> BAZ",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)

				testing.SetRealm(user1Realm)
				bar.Approve(cross, routerAddr, maxApprove)
				baz.Approve(cross, routerAddr, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			params: SwapParams{
				recipient: alice,
				SingleSwapParams: SingleSwapParams{
					tokenIn:         barPath,
					tokenOut:        bazPath,
					fee:             3000,
					amountSpecified: i256.MustFromDecimal("100"),
				},
			},
			numPools:        1,
			swapPath:        "",
			expectedFirstIn: "100",
			expectedLastOut: "98",
			expectError:     false,
		},
		// {
		// 	name: "multi hop swap (BAR -> BAZ -> QUX)",
		// 	setupFn: func(t *testing.T) {
		// 		// BAR -> BAZ
		// 		CreatePoolWithoutFee(t)
		// 		MakeMintPositionWithoutFee(t)

		// 		// BAZ -> QUX
		// 		CreateSecondPoolWithoutFee(t)
		// 		MakeSecondMintPositionWithoutFee(t)

		// 		testing.SetRealm(user1Realm)
		// 		bar.Approve(cross, routerAddr, maxApprove)
		// 		baz.Approve(cross, routerAddr, maxApprove)
		// 		qux.Approve(cross, routerAddr, maxApprove)
		// 		TokenFaucet(t, barPath, user1Addr)
		// 	},
		// 	params: SwapParams{
		// 		recipient: alice,
		// 		SingleSwapParams: SingleSwapParams{
		// 			tokenIn:         barPath,
		// 			tokenOut:        bazPath,
		// 			fee:             3000,
		// 			amountSpecified: i256.MustFromDecimal("100"),
		// 		},
		// 	},
		// 	numPools:        2,
		// 	swapPath:        "gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000*POOL*gno.land/r/onbloc/baz:gno.land/r/onbloc/qux:3000",
		// 	expectedFirstIn: "100",
		// 	expectedLastOut: "96",
		// 	expectError:     false,
		// },
		// {
		// 	name: "multi hop swap with exact output",
		// 	setupFn: func(t *testing.T) {
		// 		// BAR -> BAZ -> QUX
		// 		CreatePoolWithoutFee(t)
		// 		MakeMintPositionWithoutFee(t)
		// 		CreateSecondPoolWithoutFee(t)
		// 		MakeSecondMintPositionWithoutFee(t)

		// 		testing.SetRealm(user1Realm)
		// 		bar.Approve(cross, routerAddr, maxApprove)
		// 		baz.Approve(cross, routerAddr, maxApprove)
		// 		qux.Approve(cross, routerAddr, maxApprove)
		// 		TokenFaucet(t, barPath, user1Addr)
		// 	},
		// 	params: SwapParams{
		// 		recipient: alice,
		// 		SingleSwapParams: SingleSwapParams{
		// 			tokenIn:         barPath,
		// 			tokenOut:        bazPath,
		// 			fee:             3000,
		// 			amountSpecified: i256.MustFromDecimal("-96"),
		// 		},
		// 	},
		// 	numPools:        2,
		// 	swapPath:        "gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000*POOL*gno.land/r/onbloc/baz:gno.land/r/onbloc/qux:3000",
		// 	expectedFirstIn: "98",
		// 	expectedLastOut: "94",
		// 	expectError:     false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initRouterTest(t)

			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			firstAmountIn, lastAmountOut := multiSwap(
				tt.params,
				tt.numPools,
				tt.swapPath,
			)

			uassert.Equal(t, firstAmountIn.ToString(), tt.expectedFirstIn)
			uassert.Equal(t, lastAmountOut.ToString(), tt.expectedLastOut)
		})
	}
}
