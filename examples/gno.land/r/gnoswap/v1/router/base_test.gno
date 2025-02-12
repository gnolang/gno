package router

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/gnoswap/consts"
	i256 "gno.land/p/gnoswap/int256"

	"gno.land/p/demo/uassert"
)

var errDummy = errors.New("dummy error")

type mockOperation struct {
	ValidateErr error
	ProcessErr  error
	Result      *SwapResult
}

func (m *mockOperation) Validate() error {
	return m.ValidateErr
}

func (m *mockOperation) Process() (*SwapResult, error) {
	return m.Result, m.ProcessErr
}

func TestExecuteSwapOperation(t *testing.T) {
	tests := []struct {
		name        string
		operation   RouterOperation
		expectError bool
	}{
		{
			name: "success case",
			operation: &mockOperation{
				ValidateErr: nil,
				ProcessErr:  nil,
				Result:      &SwapResult{},
			},
			expectError: false,
		},
		{
			name: "validate error",
			operation: &mockOperation{
				ValidateErr: errDummy,
				ProcessErr:  nil,
				Result:      &SwapResult{},
			},
			expectError: true,
		},
		{
			name: "process error",
			operation: &mockOperation{
				ValidateErr: nil,
				ProcessErr:  errDummy,
				Result:      nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executeSwapOperation(tt.operation)
			if tt.expectError && err == nil {
				t.Errorf("expected an error but got nil (test case: %s)", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v (test case: %s)", err, tt.name)
			}
			if !tt.expectError && result == nil {
				t.Errorf("expected non-nil result but got nil (test case: %s)", tt.name)
			}
		})
	}
}

func TestHandleNativeTokenWrapping(t *testing.T) {
	tests := []struct {
		name            string
		inputToken      string
		outputToken     string
		swapType        SwapType
		specifiedAmount *i256.Int
		sentAmount      int64
		expectError     bool
	}{
		{
			name:            "Pass: non-GNOT token swap",
			inputToken:      "token1",
			outputToken:     "token2",
			swapType:        ExactIn,
			specifiedAmount: i256.NewInt(100),
			sentAmount:      0,
			expectError:     false,
		},
		{
			name:            "Pass: GNOT -> WGNOT exact amount",
			inputToken:      consts.GNOT,
			outputToken:     "token2",
			swapType:        ExactIn,
			specifiedAmount: i256.NewInt(1000),
			sentAmount:      1000,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := &baseSwapOperation{}

			testCoins := std.Coins{{"ugnot", tt.sentAmount}}
			std.TestIssueCoins(std.CurrentRealm().Addr(), testCoins)
			std.TestSetOrigSend(testCoins, nil)

			err := op.handleNativeTokenWrapping(
				tt.inputToken,
				tt.outputToken,
				tt.swapType,
				tt.specifiedAmount,
			)
			std.TestSetOrigSend(testCoins, nil)

			if tt.expectError && err == nil {
				t.Errorf("expected an error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRouteQuote(t *testing.T) {
	op := &baseSwapOperation{
		amountSpecified: i256.NewInt(1000),
	}

	tests := []struct {
		name        string
		quote       string
		index       int
		expectError bool
		expected    *i256.Int
	}{
		{
			name:        "Pass: valid quote - 100%",
			quote:       "100",
			index:       0,
			expectError: false,
			expected:    i256.NewInt(1000), // 1000 * 100 / 100 = 1000
		},
		{
			name:        "Pass: valid quote - 50%",
			quote:       "50",
			index:       0,
			expectError: false,
			expected:    i256.NewInt(500), // 1000 * 50 / 100 = 500
		},
		{
			name:        "Fail: invalid quote - string",
			quote:       "invalid",
			index:       0,
			expectError: true,
			expected:    nil,
		},
		{
			name:        "Fail: invalid quote - empty string",
			quote:       "",
			index:       0,
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := op.validateRouteQuote(tt.quote, tt.index)
			if tt.expectError {
				uassert.Error(t, err)
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.Cmp(tt.expected) != 0 {
					t.Errorf("expected %v but got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestProcessRoute(t *testing.T) {
	std.TestSetRealm(adminRealm)
	op := &baseSwapOperation{}

	t.Run("Single hop route", func(t *testing.T) {
		CreatePoolWithoutFee(t)
		MakeThirdMintPositionWithoutFee(t)
		route := "gno.land/r/onbloc/foo:gno.land/r/onbloc/bar:500"
		toSwap := i256.NewInt(1000)
		swapType := ExactIn

		amountIn, amountOut, err := op.processRoute(route, toSwap, swapType)

		uassert.Equal(t, err, nil)
		uassert.Equal(t, amountIn.ToString(), "1000")
		uassert.Equal(t, amountOut.ToString(), "979")
	})
}
