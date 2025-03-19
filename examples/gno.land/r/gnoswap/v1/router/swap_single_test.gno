package router

import (
	"std"
	"strings"
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

				std.TestSetRealm(user1Realm)
				bar.Approve(routerAddr, maxApprove)
				baz.Approve(routerAddr, maxApprove)
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
				bar.Approve(routerAddr, maxApprove)
				baz.Approve(routerAddr, maxApprove)
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
		errorMessage string
	}{
		{
			name: "Fail: Same token swap attempt",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				std.TestSetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        barPath,
				fee:             3000,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errorMessage: "cannot swap same token",
		},
		{
			name: "Fail: Unregistered input token",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				std.TestSetRealm(std.NewUserRealm(user1Addr))
			},
			params: SingleSwapParams{
				tokenIn:         invalidTokenPath,
				tokenOut:        bazPath,
				fee:             3000,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errorMessage: "token is not registered",
		},
		{
			name: "Fail: Unregistered output token",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				std.TestSetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        invalidTokenPath,
				fee:             3000,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errorMessage: "token is not registered || token(gno.land/r/invalid/token)",
		},
		{
			name: "Fail: Invalid fee tier 1",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				std.TestSetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             0,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errorMessage: "unknown fee(0)",
		},
		{
			name: "Fail: Invalid fee tier 2",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				std.TestSetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             1000000,
				amountSpecified: newInt256("100"),
			},
			expectError:  true,
			errorMessage: "unknown fee(1000000)",
		},
		{
			name: "Fail: Zero amount specified",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				std.TestSetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(routerAddr, maxApprove)
			},
			params: SingleSwapParams{
				tokenIn:         barPath,
				tokenOut:        bazPath,
				fee:             3000,
				amountSpecified: newInt256("0"),
			},
			expectError:  true,
			errorMessage: "amountSpecified == 0",
		},
		{
			name: "Success: Valid parameters",
			setupFn: func(t *testing.T) {
				CreatePoolWithoutFee(t)
				MakeMintPositionWithoutFee(t)
				std.TestSetRealm(std.NewUserRealm(user1Addr))
				bar.Approve(routerAddr, maxApprove)
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
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("Expected error but got none")
						return
					}
					errorStr, ok := r.(string)
					if !ok {
						t.Errorf("Expected error string but got %T: %v", r, r)
						return
					}
					if !strings.Contains(errorStr, tt.errorMessage) {
						t.Errorf("Expected error containing %q but got %q", tt.errorMessage, errorStr)
					}
				}()
			}

			amountIn, amountOut := singleSwap(tt.params)

			if !tt.expectError {
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
