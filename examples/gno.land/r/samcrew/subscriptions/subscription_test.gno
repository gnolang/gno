package subscriptions

import (
	"chain"
	"testing"
	"time"

	"gno.land/p/nt/uassert"
)

func TestDueAmount(t *testing.T) {
	defer resetState()

	price := mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 1000)), nil)

	tests := []struct {
		name           string
		vault          chain.Coins
		renewalPeriod  time.Duration
		hoursAgo       int
		initialPaid    bool
		expectedDue    int64
		expectedActive bool
		expectPanic    bool
		panicMsg       string
	}{
		{
			name:           "initial payment not yet made",
			vault:          mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			renewalPeriod:  24 * time.Hour,
			hoursAgo:       0,
			initialPaid:    false,
			expectedDue:    1000,
			expectedActive: true,
		},
		{
			name:           "initial payment made, no time elapsed",
			vault:          mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			renewalPeriod:  24 * time.Hour,
			hoursAgo:       0,
			initialPaid:    true,
			expectedDue:    0,
			expectedActive: true,
		},
		{
			name:           "one period elapsed after initial payment",
			vault:          mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			renewalPeriod:  24 * time.Hour,
			hoursAgo:       25,
			initialPaid:    true,
			expectedDue:    1000,
			expectedActive: true,
		},
		{
			name:           "multiple periods elapsed",
			vault:          mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			renewalPeriod:  24 * time.Hour,
			hoursAgo:       73,
			initialPaid:    true,
			expectedDue:    3000,
			expectedActive: true,
		},
		{
			name:           "multiple periods elapsed with initial not paid",
			vault:          mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			renewalPeriod:  24 * time.Hour,
			hoursAgo:       49,
			initialPaid:    false,
			expectedDue:    3000,
			expectedActive: true,
		},
		{
			name:           "insufficient funds for all periods",
			vault:          mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 2000)), nil),
			renewalPeriod:  24 * time.Hour,
			hoursAgo:       73,
			initialPaid:    true,
			expectedDue:    2000,
			expectedActive: false,
		},
		{
			name:          "zero renewal period panics",
			vault:         mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			renewalPeriod: 0,
			hoursAgo:      0,
			initialPaid:   true,
			expectPanic:   true,
			panicMsg:      "renewal period must be positive",
		},
		{
			name:          "negative renewal period panics",
			vault:         mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			renewalPeriod: -1 * time.Hour,
			hoursAgo:      0,
			initialPaid:   true,
			expectPanic:   true,
			panicMsg:      "renewal period must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &subscription{
				vault:         tt.vault,
				startedAt:     time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
				lastChargedAt: time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
				initialPaid:   tt.initialPaid,
			}

			if tt.expectPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					sub.dueAmount(false, tt.renewalPeriod, price)
				})
				return
			}

			due, active := sub.dueAmount(false, tt.renewalPeriod, price)
			uassert.Equal(t, tt.expectedActive, active)
			uassert.Equal(t, tt.expectedDue, due.AmountOf("/native/ugnot"))
		})
	}

	t.Run("charge mode updates state", func(t *testing.T) {
		renewalPeriod := 24 * time.Hour
		sub := &subscription{
			vault:         mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			startedAt:     time.Now().Add(-49 * time.Hour),
			lastChargedAt: time.Now().Add(-49 * time.Hour),
			initialPaid:   false,
		}

		originalLastCharged := sub.lastChargedAt
		due, active := sub.dueAmount(true, renewalPeriod, price)

		uassert.True(t, active, "subscription should be active")
		uassert.Equal(t, int64(3000), due.AmountOf("/native/ugnot"))
		uassert.Equal(t, int64(2000), sub.vault.AmountOf("/native/ugnot"))
		uassert.True(t, sub.initialPaid, "initial should be marked as paid")
		uassert.True(t, sub.lastChargedAt.After(originalLastCharged), "lastChargedAt should be updated")
	})
}

func TestIsActive(t *testing.T) {
	defer resetState()

	renewalPeriod := 24 * time.Hour
	price := mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 1000)), nil)

	tests := []struct {
		name        string
		vault       chain.Coins
		hoursAgo    int
		initialPaid bool
		expected    bool
	}{
		{
			name:        "active with enough funds",
			vault:       mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			hoursAgo:    0,
			initialPaid: true,
			expected:    true,
		},
		{
			name:        "inactive with insufficient funds",
			vault:       mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 500)), nil),
			hoursAgo:    25,
			initialPaid: true,
			expected:    false,
		},
		{
			name:        "active with initial payment pending",
			vault:       mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 1000)), nil),
			hoursAgo:    0,
			initialPaid: false,
			expected:    true,
		},
		{
			name:        "inactive when cant afford initial payment",
			vault:       mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 500)), nil),
			hoursAgo:    0,
			initialPaid: false,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &subscription{
				vault:         tt.vault,
				startedAt:     time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
				lastChargedAt: time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
				initialPaid:   tt.initialPaid,
			}

			uassert.Equal(t, tt.expected, sub.isActive(renewalPeriod, price))
		})
	}
}

func TestSubscriptionBalance(t *testing.T) {
	defer resetState()

	renewalPeriod := 24 * time.Hour
	price := mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 1000)), nil)

	tests := []struct {
		name            string
		vault           chain.Coins
		hoursAgo        int
		initialPaid     bool
		expectedBalance int64
	}{
		{
			name:            "balance with no due amount",
			vault:           mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			hoursAgo:        0,
			initialPaid:     true,
			expectedBalance: 5000,
		},
		{
			name:            "balance after periods elapsed",
			vault:           mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 5000)), nil),
			hoursAgo:        49,
			initialPaid:     true,
			expectedBalance: 3000,
		},
		{
			name:            "balance with initial payment pending",
			vault:           mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 3000)), nil),
			hoursAgo:        0,
			initialPaid:     false,
			expectedBalance: 2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &subscription{
				vault:         tt.vault,
				startedAt:     time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
				lastChargedAt: time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
				initialPaid:   tt.initialPaid,
			}

			balance := sub.balance(renewalPeriod, price)
			uassert.Equal(t, tt.expectedBalance, balance.AmountOf("/native/ugnot"))
		})
	}
}

func TestDueAmountWithMultipleCoins(t *testing.T) {
	defer resetState()

	renewalPeriod := 24 * time.Hour
	price := chain.NewCoins(
		chain.NewCoin("/native/ugnot", 1000),
		chain.NewCoin("/grc20/r/demo/foo20", 500),
	)

	tests := []struct {
		name              string
		vault             chain.Coins
		hoursAgo          int
		expectedActive    bool
		expectedNativeDue int64
		expectedGRC20Due  int64
	}{
		{
			name: "due amount with mixed coins",
			vault: chain.NewCoins(
				chain.NewCoin("/native/ugnot", 5000),
				chain.NewCoin("/grc20/r/demo/foo20", 2500),
			),
			hoursAgo:          49,
			expectedActive:    true,
			expectedNativeDue: 2000,
			expectedGRC20Due:  1000,
		},
		{
			name: "insufficient GRC20 makes subscription inactive",
			vault: chain.NewCoins(
				chain.NewCoin("/native/ugnot", 5000),
				chain.NewCoin("/grc20/r/demo/foo20", 1000),
			),
			hoursAgo:          73,
			expectedActive:    false,
			expectedNativeDue: 2000,
			expectedGRC20Due:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &subscription{
				vault:         tt.vault,
				startedAt:     time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
				lastChargedAt: time.Now().Add(-time.Duration(tt.hoursAgo) * time.Hour),
				initialPaid:   true,
			}

			due, active := sub.dueAmount(false, renewalPeriod, price)
			uassert.Equal(t, tt.expectedActive, active)
			uassert.Equal(t, tt.expectedNativeDue, due.AmountOf("/native/ugnot"))
			uassert.Equal(t, tt.expectedGRC20Due, due.AmountOf("/grc20/r/demo/foo20"))
		})
	}
}

func TestDueAmountWithSkipHeights(t *testing.T) {
	defer resetState()

	renewalPeriod := 1 * time.Hour
	price := mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 100)), nil)

	sub := &subscription{
		vault:         mustPrefixCoins(chain.NewCoins(chain.NewCoin("ugnot", 250)), nil),
		startedAt:     time.Now(),
		lastChargedAt: time.Now(),
		initialPaid:   true,
	}

	due1, active := sub.dueAmount(false, renewalPeriod, price)
	uassert.Equal(t, int64(0), due1.AmountOf("/native/ugnot"))
	uassert.True(t, active, "subscription should be active initially")

	testing.SkipHeights(720)

	due2, active := sub.dueAmount(false, renewalPeriod, price)
	uassert.True(t, active, "subscription should still be active after 1 hour")
	uassert.Equal(t, int64(100), due2.AmountOf("/native/ugnot"))

	testing.SkipHeights(720)

	due3, active := sub.dueAmount(false, renewalPeriod, price)
	uassert.True(t, active, "subscription should still be active after 2 hours")
	uassert.Equal(t, int64(200), due3.AmountOf("/native/ugnot"))

	testing.SkipHeights(720)

	due4, active := sub.dueAmount(false, renewalPeriod, price)
	uassert.False(t, active, "subscription should be inactive after 3 hours (insufficient funds)")
	uassert.Equal(t, int64(200), due4.AmountOf("/native/ugnot"))
}
