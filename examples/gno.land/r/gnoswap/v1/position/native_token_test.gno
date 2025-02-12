package position

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
)

func TestTransferUGNOT(t *testing.T) {
	tests := []struct {
		name        string
		action      func(t *testing.T, from, to std.Address)
		verify      func(t *testing.T, to std.Address) uint64
		from        std.Address
		to          std.Address
		expected    string
		shouldPanic bool
	}{
		{
			name: "Success - Zero amount",
			action: func(t *testing.T, from, to std.Address) {
				transferUGNOT(from, to, 0)
			},
			verify: func(t *testing.T, to std.Address) uint64 {
				return ugnotBalanceOf(t, to)
			},
			from:        alice,
			to:          bob,
			expected:    "0",
			shouldPanic: false,
		},
		{
			name: "Success - Valid transfer",
			action: func(t *testing.T, from, to std.Address) {
				ugnotFaucet(t, from, 100)
				std.TestSetRealm(std.NewUserRealm(from))
				transferUGNOT(from, to, 100)
			},
			verify: func(t *testing.T, to std.Address) uint64 {
				return ugnotBalanceOf(t, to)
			},
			from:        consts.POSITION_ADDR,
			to:          bob,
			expected:    "100",
			shouldPanic: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					if tc.shouldPanic {
						t.Errorf(">>> %s: expected panic but got none", tc.name)
						return
					}
				} else {
					switch r.(type) {
					case string:
						if r.(string) != tc.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
						}
					case error:
						if r.(error).Error() != tc.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r.(error).Error(), tc.expected)
						}
					default:
						t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
					}
				}
			}()

			if !tc.shouldPanic {
				tc.action(t, tc.from, tc.to)
				if tc.verify != nil {
					balance := tc.verify(t, tc.to)
					uassert.Equal(t, tc.expected, strconv.FormatUint(balance, 10))
				}
			} else {
				tc.action(t, tc.from, tc.to)
			}
		})
	}
}

func TestWrapInPosition(t *testing.T) {
	tests := []struct {
		name        string
		action      func(t *testing.T, from, to std.Address) error
		verify      func(t *testing.T, to std.Address) uint64
		from        std.Address
		to          std.Address
		expected    string
		shouldPanic bool
	}{
		{
			name: "Failure - Amount less than minimum",
			action: func(t *testing.T, from, to std.Address) error {
				return wrap(999, to)
			},
			verify:      nil,
			from:        alice,
			to:          bob,
			expected:    "amount(999) < minimum(1000)",
			shouldPanic: true,
		},
		{
			name: "Failure - Zero amount",
			action: func(t *testing.T, from, to std.Address) error {
				return wrap(0, to)
			},
			verify:      nil,
			from:        alice,
			to:          bob,
			expected:    "amount(0) < minimum(1000)",
			shouldPanic: true,
		},
		{
			name: "Success - Valid amount",
			action: func(t *testing.T, from, to std.Address) error {
				ugnotFaucet(t, from, 1000)
				std.TestSetRealm(std.NewUserRealm(from))
				return wrap(1000, to)
			},
			verify: func(t *testing.T, to std.Address) uint64 {
				return TokenBalance(t, wugnotPath, to)
			},
			from:        consts.POSITION_ADDR,
			to:          bob,
			expected:    "1000",
			shouldPanic: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					switch r.(type) {
					case string:
						if r.(string) != tc.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
						}
					case error:
						if r.(error).Error() != tc.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r.(error).Error(), tc.expected)
						}
					default:
						t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
					}
				}
			}()

			if !tc.shouldPanic {
				err := tc.action(t, tc.from, tc.to)
				if err == nil && tc.verify != nil {
					balance := tc.verify(t, tc.to)
					uassert.Equal(t, tc.expected, strconv.FormatUint(balance, 10))
				}
			} else {
				err := tc.action(t, tc.from, tc.to)
				if err != nil {
					uassert.Equal(t, tc.expected, err.Error())
				} else {
					t.Errorf(">>> %s: expected panic but got none", tc.name)
				}
			}
		})
	}
}

func TestUnWrap(t *testing.T) {
	tests := []struct {
		name        string
		action      func(t *testing.T, from, to std.Address) error
		verify      func(t *testing.T, to std.Address) uint64
		from        std.Address
		to          std.Address
		expected    string
		shouldPanic bool
	}{
		{
			name: "Failure - Zero amount",
			action: func(t *testing.T, from, to std.Address) error {
				return unwrap(0, to)
			},
			verify:      nil,
			from:        alice,
			to:          bob,
			expected:    "amount(0) is zero",
			shouldPanic: true,
		},
		{
			name: "Success - Valid amount",
			action: func(t *testing.T, from, to std.Address) error {
				ugnotFaucet(t, from, 1000)
				std.TestSetRealm(std.NewUserRealm(from))
				wrap(1000, to)
				std.TestSetRealm(std.NewUserRealm(to))
				TokenApprove(t, wugnotPath, to, from, 1000)
				return unwrap(1000, to)
			},
			verify: func(t *testing.T, to std.Address) uint64 {
				return TokenBalance(t, wugnotPath, to)
			},
			from:        consts.POSITION_ADDR,
			to:          bob,
			expected:    "1000",
			shouldPanic: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					switch r.(type) {
					case string:
						if r.(string) != tc.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
						}
					case error:
						if r.(error).Error() != tc.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r.(error).Error(), tc.expected)
						}
					default:
						t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
					}
				}
			}()

			if !tc.shouldPanic {
				err := tc.action(t, tc.from, tc.to)
				if err == nil && tc.verify != nil {
					balance := tc.verify(t, tc.to)
					uassert.Equal(t, tc.expected, strconv.FormatUint(balance, 10))
				}
			} else {
				err := tc.action(t, tc.from, tc.to)
				if err != nil {
					uassert.Equal(t, tc.expected, err.Error())
				} else {
					t.Errorf(">>> %s: expected panic but got none", tc.name)
				}
			}
		})
	}
}

func TestIsNative(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "Native Token - GNOT",
			token:    "gnot",
			expected: true,
		},
		{
			name:     "Non-Native Token",
			token:    "usdt",
			expected: false,
		},
		{
			name:     "Empty Token",
			token:    "",
			expected: false,
		},
		{
			name:     "Similar but Different Token",
			token:    "GNOT",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNative(tt.token)
			uassert.Equal(t, tt.expected, result, "Unexpected result for token: "+tt.token)
		})
	}
}

func TestIsWrappedToken(t *testing.T) {
	tests := []struct {
		name      string
		tokenPath string
		expected  bool
	}{
		{
			name:      "Success - Token is Wrapped WUGNOT",
			tokenPath: consts.WRAPPED_WUGNOT,
			expected:  true,
		},
		{
			name:      "Fail - Token is not Wrapped WUGNOT",
			tokenPath: "gno.land/r/demo/ugnot",
			expected:  false,
		},
		{
			name:      "Fail - Empty tokenPath",
			tokenPath: "",
			expected:  false,
		},
		{
			name:      "Fail - Similar but Different Token Path",
			tokenPath: "gno.land/r/demo/Wugnot",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWrappedToken(tt.tokenPath)
			if result != tt.expected {
				t.Errorf(
					"expected %s but got %s",
					strconv.FormatBool(tt.expected),
					strconv.FormatBool(result),
				)
			}
		})
	}
}

func TestSafeWrapNativeToken(t *testing.T) {
	tests := []struct {
		name          string
		amountDesired string
		userAddress   std.Address
		sentAmount    uint64
		expectPanic   bool
		expectedWrap  uint64
	}{
		{
			name:          "Panic - Zero UGNOT",
			amountDesired: "50",
			userAddress:   alice,
			sentAmount:    0,
			expectPanic:   true,
		},
		{
			name:          "Panic - Insufficient UGNOT",
			amountDesired: "150",
			userAddress:   alice,
			sentAmount:    100,
			expectPanic:   true,
		},
		{
			name:          "Panic - Invalid Desired Amount",
			amountDesired: "invalid",
			userAddress:   alice,
			sentAmount:    200,
			expectPanic:   true,
		},
		{
			name:          "Successful wrap - Exact Amount",
			amountDesired: "1050",
			userAddress:   alice,
			sentAmount:    1050,
			expectPanic:   false,
			expectedWrap:  1050,
		},
		{
			name:          "Excess Refund",
			amountDesired: "1000",
			userAddress:   alice,
			sentAmount:    1500,
			expectPanic:   false,
			expectedWrap:  1000,
		},
		{
			name:          "Boundary Test - Exact Match",
			amountDesired: "1000",
			userAddress:   alice,
			sentAmount:    1000,
			expectPanic:   false,
			expectedWrap:  1000,
		},
		// {
		// 	name:          "Zero Desired Amount", // unable to catch this panic
		// 	amountDesired: "0",
		// 	userAddress:   alice,
		// 	sentAmount:    100,
		// 	expectPanic:   true,
		// },
		{
			name:          "Wrap Error Test",
			amountDesired: "100",
			userAddress:   alice,
			sentAmount:    100,
			expectPanic:   true, // Simulate wrap error internally
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

			amount, err := strconv.ParseUint(tt.amountDesired, 10, 64)
			if err != nil {
				panic("can not parse amount")
			}
			ugnotFaucet(t, consts.POSITION_ADDR, amount)
			std.TestSetRealm(std.NewUserRealm(consts.POSITION_ADDR))

			transferUGNOT(consts.POSITION_ADDR, consts.POSITION_ADDR, amount)

			// Perform wrapping
			wrappedAmount := safeWrapNativeToken(tt.amountDesired, tt.userAddress)

			// Verify wrapped amount
			if !tt.expectPanic {
				uassert.Equal(t, tt.expectedWrap, wrappedAmount)
			}
		})
	}
}
