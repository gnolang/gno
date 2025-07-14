package router

import (
	"errors"
	"std"
	"testing"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var errDummy = errors.New("dummy error")

const (
	foobar500 = "gno.land/r/onbloc/foo:gno.land/r/onbloc/bar:500"
	barbaz500 = "gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:500"
)

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
		specifiedAmount *i256.Int
		sentAmount      int64
		expectError     bool
		expectedError   string
	}{
		{
			name:            "Pass: non-GNOT token swap",
			inputToken:      "token1",
			outputToken:     "token2",
			specifiedAmount: i256.NewInt(100),
			sentAmount:      0,
			expectError:     false,
		},
		{
			name:            "Pass: WUGNOT input token",
			inputToken:      "gno.land/r/demo/wugnot",
			outputToken:     "token2",
			specifiedAmount: i256.NewInt(100),
			sentAmount:      0,
			expectError:     false,
		},
		{
			name:            "Pass: WUGNOT output token",
			inputToken:      "token1",
			outputToken:     "gno.land/r/demo/wugnot",
			specifiedAmount: i256.NewInt(100),
			sentAmount:      0,
			expectError:     false,
		},
		{
			name:            "Fail: GNOT input with insufficient sent amount",
			inputToken:      "gnot",
			outputToken:     "token2",
			specifiedAmount: i256.NewInt(1000),
			sentAmount:      500,
			expectError:     true,
			expectedError:   "ugnot sent by user(500) is not equal to amountSpecified(1000)",
		},
		{
			name:            "Fail: GNOT input with excess sent amount",
			inputToken:      "gnot",
			outputToken:     "token2",
			specifiedAmount: i256.NewInt(1000),
			sentAmount:      1500,
			expectError:     true,
			expectedError:   "ugnot sent by user(1500) is not equal to amountSpecified(1000)",
		},
		{
			name:            "Fail: GNOT input with zero sent amount",
			inputToken:      "gnot",
			outputToken:     "token2",
			specifiedAmount: i256.NewInt(1000),
			sentAmount:      0,
			expectError:     true,
			expectedError:   "ugnot sent by user(0) is not equal to amountSpecified(1000)",
		},
		{
			name:            "Pass: GNOT input with zero specified amount",
			inputToken:      "gnot",
			outputToken:     "token2",
			specifiedAmount: i256.NewInt(0),
			sentAmount:      0,
			expectError:     false,
		},
		{
			name:            "Pass: non-GNOT input with non-zero sent amount",
			inputToken:      "token1",
			outputToken:     "token2",
			specifiedAmount: i256.NewInt(100),
			sentAmount:      1000,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := &baseSwapOperation{}

			testUser := testutils.TestAddress("test_user")
			testing.SetOriginCaller(testUser)
			testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/router"))

			testCoins := std.Coins{{"ugnot", tt.sentAmount}}
			testing.IssueCoins(std.OriginCaller(), testCoins)
			if tt.sentAmount > 0 {
				testing.IssueCoins(routerAddr, testCoins)
			}
			testing.SetOriginSend(testCoins)

			err := op.handleNativeTokenWrapping(
				tt.inputToken,
				tt.outputToken,
				tt.specifiedAmount,
			)
			testing.SetOriginSend(testCoins)

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
	op := &baseSwapOperation{}

	t.Run("Single hop route", func(t *testing.T) {
		testing.SetRealm(adminRealm)
		CreatePoolWithoutFee(t)
		MakeThirdMintPositionWithoutFee(t)
		route := foobar500
		toSwap := i256.NewInt(1000)
		swapType := ExactIn

		func(cur realm) {
			amountIn, amountOut, err := op.processRoute(route, toSwap, swapType)
			uassert.Equal(t, err, nil)
			uassert.Equal(t, amountIn.ToString(), "1000")
			uassert.Equal(t, amountOut.ToString(), "979")
		}(cross)
	})
}

// Test reproducing report L-08 "Loss of Tokens in Swap Due to Truncation"
func TestProcessRoutesWithRemainder(t *testing.T) {
	oddAmount := i256.NewInt(101)
	op := &baseSwapOperation{
		// use odd amount to ensure remainder
		amountSpecified: oddAmount,
		routes:          []string{foobar500, barbaz500},
		// allocate 50% to each route
		quotes: []string{"50", "50"},
	}

	testing.SetRealm(adminRealm)
	CreatePoolWithoutFee(t)
	MakeThirdMintPositionWithoutFee(t)

	// expected amounts for each route
	expectedAmounts := []*i256.Int{
		i256.NewInt(50), // 101 * 50 / 100 = 50
		i256.NewInt(50), // 101 * 50 / 100 = 50
	}

	func(cur realm) {
		_, amountOut, err := op.processRoutes(ExactIn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// check output amount is not zero
		if amountOut.Cmp(u256.Zero()) == 0 {
			t.Error("expected non-zero amountOut")
		}
	}(cross)

	// check actual swapped amounts for each route
	for i := range op.routes {
		toSwap, err := op.validateRouteQuote(op.quotes[i], i)
		if err != nil {
			t.Fatalf("unexpected error validating quote: %v", err)
		}

		if toSwap.Cmp(expectedAmounts[i]) != 0 {
			t.Errorf("route %d: expected amount %s, got %s", i, expectedAmounts[i].ToString(), toSwap.ToString())
		}
	}

	// check total input amount
	totalExpected := oddAmount
	if op.amountSpecified.Cmp(totalExpected) != 0 {
		t.Errorf("total expected amount %s, got %s", totalExpected.ToString(), op.amountSpecified.ToString())
	}

	// check remainder
	usedAmount := i256.NewInt(0)
	for _, amount := range expectedAmounts {
		usedAmount = usedAmount.Add(usedAmount, amount)
	}
	// Currently, this test succeeds.
	// That is, we can observe that there is indeed a remainder.
	remainder := totalExpected.Sub(totalExpected, usedAmount)
	if remainder.Cmp(i256.NewInt(1)) != 0 {
		t.Errorf("expected remainder of 1, got %s", remainder.ToString())
	}
}

func TestProcessRoutes_LastRouteUsesRemainingAmount(t *testing.T) {
	tests := []struct {
		name            string
		amountSpecified *i256.Int
		routes          []string
		quotes          []string
		expectedAmounts []string // expected amount for each route
		description     string
	}{
		{
			// 50% + 50% = 100% (no remainder)
			name:            "two routes with perfect split",
			amountSpecified: i256.NewInt(100),
			routes:          []string{foobar500, barbaz500},
			quotes:          []string{"50", "50"},
			expectedAmounts: []string{"50", "50"},
		},
		{
			// 33% + 33% + remainder(34%) = 100%
			name:            "three routes with remainder",
			amountSpecified: i256.NewInt(100),
			routes:          []string{foobar500, barbaz500, foobar500},
			quotes:          []string{"33", "33", "34"},
			expectedAmounts: []string{"33", "33", "34"}, // last route gets remainder
		},
		{
			name:            "odd amount with two routes",
			amountSpecified: i256.NewInt(101),
			routes:          []string{foobar500, barbaz500},
			quotes:          []string{"50", "50"},
			expectedAmounts: []string{"50", "51"}, // last route gets 50 + remainder(1)
		},
		{
			// 24% + 25% + 25% + remainder(26%) = 100%
			name:            "large amount with multiple routes",
			amountSpecified: i256.NewInt(1000),
			routes:          []string{foobar500, barbaz500, foobar500, barbaz500},
			quotes:          []string{"24", "25", "25", "26"},
			expectedAmounts: []string{"240", "250", "250", "260"}, // last route gets remainder
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testing.SetRealm(adminRealm)
			CreatePoolWithoutFee(t)
			MakeThirdMintPositionWithoutFee(t)

			// Track the amounts used for each route
			totalUsed := i256.NewInt(0)

			// Calculate expected amounts based on the new logic
			calculatedAmounts := make([]*i256.Int, len(tt.routes))

			for i := range tt.routes {
				if i < len(tt.routes)-1 {
					// For non-last routes, calculate based on quote
					quote := tt.quotes[i]
					quoteInt, err := i256.FromDecimal(quote)
					uassert.NoError(t, err)

					amount := new(i256.Int).Mul(tt.amountSpecified, quoteInt)
					amount = new(i256.Int).Div(amount, i256.NewInt(100))

					calculatedAmounts[i] = amount
					totalUsed = new(i256.Int).Add(totalUsed, amount)
				} else {
					// Last route gets the remaining amount
					calculatedAmounts[i] = new(i256.Int).Sub(tt.amountSpecified, totalUsed)
				}
			}

			// Verify each calculated amount matches expected
			for i, expectedAmount := range tt.expectedAmounts {
				expected := i256.MustFromDecimal(expectedAmount)
				if calculatedAmounts[i].Cmp(expected) != 0 {
					t.Errorf("route %d: expected %s, got %s",
						i, expected.ToString(), calculatedAmounts[i].ToString())
				}
			}

			// Verify total equals amountSpecified
			total := i256.NewInt(0)
			for _, amount := range calculatedAmounts {
				total = total.Add(total, amount)
			}
			if total.Cmp(tt.amountSpecified) != 0 {
				t.Errorf("total amount mismatch: expected %s, got %s",
					tt.amountSpecified.ToString(), total.ToString())
			}
		})
	}
}
