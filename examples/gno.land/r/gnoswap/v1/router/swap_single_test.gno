package router

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	i256 "gno.land/p/gnoswap/int256"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
)

func TestSingleSwap(t *testing.T) {
	user1Realm := std.NewUserRealm(user1Addr)
	routerAddr, _ := access.GetAddress(access.ROLE_ROUTER)

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

				testing.SetRealm(user1Realm)
				bar.Approve(cross, routerAddr, maxApprove)
				baz.Approve(cross, routerAddr, maxApprove)
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

				testing.SetRealm(user1Realm)
				bar.Approve(cross, routerAddr, maxApprove)
				baz.Approve(cross, routerAddr, maxApprove)
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

			amountIn, amountOut := singleSwap(&tt.params)

			uassert.Equal(t, amountIn.ToString(), tt.expectedIn)
			uassert.Equal(t, amountOut.ToString(), tt.expectedOut)
		})
	}
}

func TestSingleSwapInputValidation(t *testing.T) {
	newInt256 := func(val string) *i256.Int {
		return i256.MustFromDecimal(val)
	}

	invalidTokenPath := "gno.land/r/invalid/token"
	tests := []struct {
		name         string
		setupFn      func(t *testing.T)
		params       SingleSwapParams
		expectError  bool
		errType      string
		errorMessage string
	}{
		{
			name: "Fail: Same token swap attempt",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				testing.SetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(cross, routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        barPath,
				fee:             3000,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errType:      "panic",
			errorMessage: "[GNOSWAP-ROUTER-014] cannot swap same token",
		},
		{
			name: "Fail: Unregistered input token",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				testing.SetRealm(std.NewUserRealm(user1Addr))
			},
			params: SingleSwapParams{
				tokenIn:         invalidTokenPath,
				tokenOut:        bazPath,
				fee:             3000,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errType:      "panic",
			errorMessage: "[GNOSWAP-COMMON-004] token is not registered || token(gno.land/r/invalid/token)",
		},
		{
			name: "Fail: Unregistered output token",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				testing.SetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(cross, routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        invalidTokenPath,
				fee:             3000,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errType:      "panic",
			errorMessage: "[GNOSWAP-COMMON-004] token is not registered || token(gno.land/r/invalid/token)",
		},
		{
			name: "Fail: Invalid fee tier 1",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				testing.SetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(cross, routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             0,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errType:      "abort",
			errorMessage: "[GNOSWAP-ROUTER-006] invalid pool fee tier || unknown fee(0)",
		},
		{
			name: "Fail: Invalid fee tier 2",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				testing.SetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(cross, routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             1000000,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errType:      "abort",
			errorMessage: "[GNOSWAP-ROUTER-006] invalid pool fee tier || unknown fee(1000000)",
		},
		{
			name: "Fail: Zero amount specified",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				testing.SetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(cross, routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             3000,
				amountSpecified: newInt256("0"),
			},
			expectError:  true,
			errType:      "abort",
			errorMessage: "[GNOSWAP-POOL-015] invalid swap amount || amountSpecified == 0",
		},
		{
			name: "Success: Valid parameters",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				testing.SetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(cross, routerAddr, maxApprove)
				TokenFaucet(t, barPath, user1Addr)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             3000,
				amountSpecified: newInt256("100"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			if tt.expectError {
				switch tt.errType {
				case "panic":
					uassert.PanicsWithMessage(t, tt.errorMessage, func() {
						singleSwap(&tt.params)
					})
				case "abort":
					uassert.AbortsWithMessage(t, tt.errorMessage, func() {
						singleSwap(&tt.params)
					})
				default:
					t.Errorf("unexpected error type: %s", tt.errType)
				}
			} else {
				amountIn, amountOut := singleSwap(&tt.params)

				if amountIn == nil || amountOut == nil {
					t.Error("Expected non-nil amounts for successful swap")
				}
				if amountIn.IsZero() {
					t.Error("Expected non-zero input amount")
				}
				if amountOut.IsZero() {
					t.Error("Expected non-zero output amount")
				}
			}
		})
	}
}
