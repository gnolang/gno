package router

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestFinalizeSwap(t *testing.T) {
	gnot := consts.GNOT

	newUint256 := func(val string) *u256.Uint {
		return u256.MustFromDecimal(val)
	}

	tests := []struct {
		name                    string
		inputToken              string
		outputToken             string
		resultAmountIn          *u256.Uint
		resultAmountOut         *u256.Uint
		swapType                SwapType
		tokenAmountLimit        *u256.Uint
		userBeforeWugnotBalance uint64
		userWrappedWugnot       uint64
		amountSpecified         *u256.Uint
		expectError             bool
		errorMessage            string
	}{
		{
			name:                    "Pass: ExactIn",
			inputToken:              barPath,
			outputToken:             bazPath,
			resultAmountIn:          newUint256("100"),
			resultAmountOut:         newUint256("90"),
			swapType:                ExactIn,
			tokenAmountLimit:        newUint256("85"),
			userBeforeWugnotBalance: 0,
			userWrappedWugnot:       0,
			amountSpecified:         newUint256("100"),
			expectError:             false,
		},
		{
			name:                    "Pass: ExactOut",
			inputToken:              barPath,
			outputToken:             bazPath,
			resultAmountIn:          newUint256("110"),
			resultAmountOut:         newUint256("100"),
			swapType:                ExactOut,
			tokenAmountLimit:        newUint256("120"),
			userBeforeWugnotBalance: 0,
			userWrappedWugnot:       0,
			amountSpecified:         newUint256("100"),
			expectError:             false,
		},
		{
			name:                    "ExactOut: Slippage error",
			inputToken:              barPath,
			outputToken:             bazPath,
			resultAmountIn:          newUint256("100"),
			resultAmountOut:         newUint256("90"),
			swapType:                ExactOut,
			tokenAmountLimit:        newUint256("100"),
			userBeforeWugnotBalance: 0,
			userWrappedWugnot:       0,
			amountSpecified:         newUint256("100"),
			expectError:             true,
			errorMessage:            "[GNOSWAP-ROUTER-002] slippage check failed || Received more than requested in [EXACT_OUT] requested=100, actual=90",
		},
		{
			name:                    "GNOT: Slippage error",
			inputToken:              gnot,
			outputToken:             barPath,
			resultAmountIn:          newUint256("300"),
			resultAmountOut:         newUint256("90"),
			swapType:                ExactIn,
			tokenAmountLimit:        newUint256("85"),
			userBeforeWugnotBalance: 1000000,
			userWrappedWugnot:       20000,
			expectError:             true,
			errorMessage:            "too much wugnot spent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("Error expected but not occurred")
						return
					}
					errorStr, ok := r.(string)
					if !ok {
						t.Errorf("Unexpected error type: %v", r)
						return
					}
					if tt.errorMessage != "" && !strings.Contains(errorStr, tt.errorMessage) {
						t.Errorf("Expected error message not included. got: %v, want: %v", errorStr, tt.errorMessage)
					}
				}()
			}

			amountIn, amountOut := finalizeSwap(
				tt.inputToken,
				tt.outputToken,
				tt.resultAmountIn,
				tt.resultAmountOut,
				tt.swapType,
				tt.tokenAmountLimit,
				tt.userBeforeWugnotBalance,
				tt.userWrappedWugnot,
				tt.amountSpecified,
			)

			if !tt.expectError {
				uassert.NotEqual(t, amountIn, "")
				uassert.NotEqual(t, amountOut, "")

				outVal := i256.MustFromDecimal(amountOut)
				if !outVal.IsNeg() {
					t.Error("amountOut is not negative")
				}
			}
		})
	}
}
