package position

import (
	"errors"
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
		verify      func(t *testing.T, to std.Address) int64
		from        std.Address
		to          std.Address
		expected    string
		shouldPanic bool
	}{
		{
			name: "transferUGNOT is success by zero amount",
			action: func(t *testing.T, from, to std.Address) {
				transferUGNOT(from, to, 0)
			},
			verify: func(t *testing.T, to std.Address) int64 {
				return ugnotBalanceOf(t, to)
			},
			from:        alice,
			to:          bob,
			expected:    "0",
			shouldPanic: false,
		},
		{
			name: "transferUGNOT is success by valid transfer",
			action: func(t *testing.T, from, to std.Address) {
				ugnotFaucet(t, from, 100)
				testing.SetRealm(std.NewUserRealm(from))
				transferUGNOT(from, to, 100)
			},
			verify: func(t *testing.T, to std.Address) int64 {
				return ugnotBalanceOf(t, to)
			},
			from:        positionAddr,
			to:          bob,
			expected:    "100",
			shouldPanic: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.shouldPanic {
				tc.action(t, tc.from, tc.to)
				if tc.verify != nil {
					balance := tc.verify(t, tc.to)
					uassert.Equal(t, tc.expected, strconv.FormatInt(balance, 10))
				}
			} else {
				uassert.PanicsWithMessage(t, tc.expected, func() {
					tc.action(t, tc.from, tc.to)
				})
			}
		})
	}
}

func TestWrapInPosition(t *testing.T) {
	tests := []struct {
		name           string
		action         func(t *testing.T, from, to std.Address) error
		verify         func(t *testing.T, to std.Address) int64
		from           std.Address
		to             std.Address
		expected       string
		shouldPanic    bool
		expectedErrMsg string
	}{
		{
			name: "wrap is failed by amount less than minimum",
			action: func(t *testing.T, from, to std.Address) error {
				return wrap(999, to)
			},
			verify:         nil,
			from:           alice,
			to:             bob,
			expectedErrMsg: "amount(999) < minimum(1000)",
			shouldPanic:    true,
		},
		{
			name: "wrap is failed by zero amount",
			action: func(t *testing.T, from, to std.Address) error {
				return wrap(0, to)
			},
			verify:         nil,
			from:           alice,
			to:             bob,
			expectedErrMsg: "amount(0) < minimum(1000)",
			shouldPanic:    true,
		},
		{
			name: "wrap is success by valid amount",
			action: func(t *testing.T, from, to std.Address) error {
				ugnotFaucet(t, from, 1000)
				testing.SetRealm(std.NewUserRealm(from))
				return wrap(1000, to)
			},
			verify: func(t *testing.T, to std.Address) int64 {
				return TokenBalance(t, wugnotPath, to)
			},
			from:        positionAddr,
			to:          bob,
			expected:    "1000",
			shouldPanic: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.shouldPanic {
				err := tc.action(t, tc.from, tc.to)
				if err == nil && tc.verify != nil {
					balance := tc.verify(t, tc.to)
					uassert.Equal(t, tc.expected, strconv.FormatInt(balance, 10))
				}
			} else {
				uassert.ErrorIs(t, tc.action(t, tc.from, tc.to), errors.New(tc.expectedErrMsg))
			}
		})
	}
}

func TestUnWrap(t *testing.T) {
	tests := []struct {
		name        string
		action      func(t *testing.T, from, to std.Address) error
		verify      func(t *testing.T, to std.Address) int64
		from        std.Address
		to          std.Address
		expected    string
		shouldPanic bool
	}{
		{
			name: "unwrap is failed by zero amount",
			action: func(t *testing.T, from, to std.Address) error {
				return unwrap(0, to)
			},
			verify:      nil,
			from:        alice,
			to:          bob,
			expected:    "amount(0) is zero or negative",
			shouldPanic: true,
		},
		{
			name: "unwrap is success by valid amount",
			action: func(t *testing.T, from, to std.Address) error {
				ugnotFaucet(t, from, 1000)

				testing.SetRealm(std.NewUserRealm(from))
				err := wrap(1000, to)
				uassert.Nil(t, err, "wrap should not return error")

				TokenApprove(t, wugnotPath, to, from, 1000)
				return unwrap(1000, to)
			},
			verify: func(t *testing.T, to std.Address) int64 {
				return ugnotBalanceOf(t, to)
			},
			from:        positionAddr,
			to:          bob,
			expected:    "1000",
			shouldPanic: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.shouldPanic {
				err := tc.action(t, tc.from, tc.to)
				if err == nil && tc.verify != nil {
					balance := tc.verify(t, tc.to)
					uassert.Equal(t, tc.expected, strconv.FormatInt(balance, 10))
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
			name:     "isNative is success by native token",
			token:    "gnot",
			expected: true,
		},
		{
			name:     "isNative is failed by non-native token",
			token:    "usdt",
			expected: false,
		},
		{
			name:     "isNative is failed by empty token",
			token:    "",
			expected: false,
		},
		{
			name:     "isNative is failed by similar but different token",
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
			name:      "isWrappedToken is success by wrapped WUGNOT",
			tokenPath: consts.WRAPPED_WUGNOT,
			expected:  true,
		},
		{
			name:      "isWrappedToken is failed by non-wrapped WUGNOT",
			tokenPath: "gno.land/r/demo/ugnot",
			expected:  false,
		},
		{
			name:      "isWrappedToken is failed by empty tokenPath",
			tokenPath: "",
			expected:  false,
		},
		{
			name:      "isWrappedToken is failed by similar but different tokenPath",
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
		sentAmount    int64
		expectPanic   bool
		expectedWrap  int64
		expectedError string
	}{
		{
			name:          "safeWrapNativeToken is failed by zero UGNOT",
			amountDesired: "50",
			userAddress:   alice,
			sentAmount:    0,
			expectPanic:   true,
			expectedError: "[GNOSWAP-POSITION-008] can not wrap less than minimum amount || amount(50) < minimum(1000)",
		},
		{
			name:          "safeWrapNativeToken is failed by insufficient UGNOT",
			amountDesired: "150",
			userAddress:   alice,
			sentAmount:    100,
			expectPanic:   true,
			expectedError: "[GNOSWAP-POSITION-008] can not wrap less than minimum amount || amount(150) < minimum(1000)",
		},
		{
			name:          "safeWrapNativeToken is failed by invalid desired amount",
			amountDesired: "invalid",
			userAddress:   alice,
			sentAmount:    200,
			expectPanic:   true,
			expectedError: "[GNOSWAP-POSITION-003] wrap, unwrap failed || strconv.ParseInt: parsing \"invalid\": invalid syntax",
		},
		{
			name:          "safeWrapNativeToken is success by exact amount",
			amountDesired: "1050",
			userAddress:   alice,
			sentAmount:    1050,
			expectPanic:   false,
			expectedWrap:  1050,
		},
		{
			name:          "safeWrapNativeToken is success by excess refund",
			amountDesired: "1000",
			userAddress:   alice,
			sentAmount:    1500,
			expectPanic:   false,
			expectedWrap:  1000,
		},
		{
			name:          "safeWrapNativeToken is success by boundary test",
			amountDesired: "1000",
			userAddress:   alice,
			sentAmount:    1000,
			expectPanic:   false,
			expectedWrap:  1000,
		},
		{
			name:          "safeWrapNativeToken is success by zero desired amount", // unable to catch this panic
			amountDesired: "0",
			userAddress:   alice,
			sentAmount:    100,
			expectPanic:   false,
			expectedWrap:  0,
		},
		{
			name:          "safeWrapNativeToken is failed by wrap error",
			amountDesired: "100",
			userAddress:   alice,
			sentAmount:    100,
			expectPanic:   true, // Simulate wrap error internally
			expectedError: "[GNOSWAP-POSITION-008] can not wrap less than minimum amount || amount(100) < minimum(1000)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// we ignore the error here because we need to test
			// the invalid amount case
			amount, _ := strconv.ParseInt(tt.amountDesired, 10, 64)
			if amount > 0 {
				ugnotFaucet(t, positionAddr, amount)
			}
			testing.SetRealm(std.NewUserRealm(positionAddr))

			transferUGNOT(positionAddr, positionAddr, amount)

			// Verify wrapped amount
			if !tt.expectPanic {
				wrappedAmount := safeWrapNativeToken(tt.amountDesired, tt.userAddress)
				uassert.Equal(t, tt.expectedWrap, wrappedAmount)
			} else {
				uassert.PanicsWithMessage(t, tt.expectedError, func() {
					safeWrapNativeToken(tt.amountDesired, tt.userAddress)
				})
			}
		})
	}
}
