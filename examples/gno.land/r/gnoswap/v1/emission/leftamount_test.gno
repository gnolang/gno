package emission

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"

	"gno.land/r/gnoswap/v1/gns"
)

func TestLeftAmountFix(t *testing.T) {
	resetObject(t)

	testing.SetRealm(adminRealm)
	gns.Transfer(cross, emissionAddr, 10000000) // 10M GNS tokens

	tests := []struct {
		name               string
		initialLeftAmount  int64
		distributionAmount int64
		distributions      map[string]int64 // target -> bps
		expectedTotalSent  int64
		expectedLeftAmount int64
		setupFunc          func()
		verifyFunc         func(t *testing.T, actualLeft int64)
		shouldPanic        bool
		panicMsg           string
	}{
		{
			name:               "Verify leftAmount is correctly tracked when rounding occurs",
			initialLeftAmount:  0,
			distributionAmount: 10001,
			distributions: map[string]int64{
				"1": 3333, // 33.33%
				"2": 3333, // 33.33%
				"3": 3333, // 33.33%
				"4": 0,    // 0%
			},
			expectedTotalSent:  9999, // 3333 * 3
			expectedLeftAmount: 2,    // 10001 - 9999
			shouldPanic:        false,
			panicMsg:           "",
			setupFunc: func() {
				leftGNSAmount = 0
				distributionBpsPct = avl.NewTree()
			},
		},
		{
			name:               "Verify fix works with zero initial leftGNSAmount",
			initialLeftAmount:  0,
			distributionAmount: 10001,
			distributions: map[string]int64{
				"1": 3333,
				"2": 3333,
				"3": 3333,
				"4": 0,
			},
			expectedTotalSent:  9999,
			expectedLeftAmount: 2,
			shouldPanic:        false,
			panicMsg:           "",
			setupFunc: func() {
				leftGNSAmount = 0
				distributionBpsPct = avl.NewTree()
			},
		},
		{
			name:               "Integration test with different percentages",
			initialLeftAmount:  0,
			distributionAmount: 123456,
			distributions: map[string]int64{
				"1": 3333, // 33.33%
				"2": 3333, // 33.33%
				"3": 3333, // 33.33%
				"4": 1,    // 0.01% to make total 9999 bps
			},
			expectedTotalSent:  0, // Will be calculated based on actual distribution
			expectedLeftAmount: 0, // Will be calculated based on actual distribution
			shouldPanic:        false,
			panicMsg:           "",
			setupFunc: func() {
				leftGNSAmount = 0
				lastExecutedHeight = 0
				distributionBpsPct = avl.NewTree()
			},
			verifyFunc: func(t *testing.T, actualLeft int64) {
				// Custom verification for integration test
				uassert.True(t, actualLeft >= 0)
			},
		},
		{
			name:               "Test with exact division (no remainder)",
			initialLeftAmount:  0,
			distributionAmount: 10000,
			distributions: map[string]int64{
				"1": 5000, // 50%
				"2": 5000, // 50%
			},
			expectedTotalSent:  10000, // 5000 + 5000
			expectedLeftAmount: 0,     // No remainder
			setupFunc: func() {
				leftGNSAmount = 0
				distributionBpsPct = avl.NewTree()
			},
		},
		{
			name:               "Test negative leftAmount prevention",
			initialLeftAmount:  0,
			distributionAmount: 10000,
			distributions: map[string]int64{
				"1": 6000, // 60%
				"2": 5000, // 50% (total: 110% exceeds 100%)
			},
			expectedTotalSent:  0,
			expectedLeftAmount: 0,
			shouldPanic:        true,
			panicMsg:           "[GNOSWAP-EMISSION-002] invalid emission target || total distributed amount (11000) exceeds original amount (10000)",
			setupFunc: func() {
				leftGNSAmount = 0
				distributionBpsPct = avl.NewTree()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Set distribution percentages
			for target, bps := range tt.distributions {
				distributionBpsPct.Set(target, bps)
			}

			beforeLeft := GetLeftGNSAmount()

			if tt.shouldPanic {
				uassert.AbortsWithMessage(t, tt.panicMsg, func() {
					distributeToTarget(cross, tt.distributionAmount)
				})
			} else {
				// Execute distribution
				totalSent := distributeToTarget(cross, tt.distributionAmount)

				// Get leftGNSAmount after distribution
				afterLeft := GetLeftGNSAmount()
				actualLeftDifference := afterLeft - beforeLeft

				if tt.expectedTotalSent > 0 {
					uassert.Equal(t, tt.expectedTotalSent, totalSent)
				}

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, afterLeft)
				} else {
					if tt.expectedLeftAmount > 0 {
						// For cases where we expect a specific left amount difference
						leftAmount := tt.distributionAmount - totalSent
						uassert.Equal(t, tt.expectedLeftAmount, leftAmount)
						uassert.Equal(t, leftAmount, actualLeftDifference)
					}
				}

				// Additional verification that leftGNSAmount is updated correctly
				if tt.name != "Integration test with different percentages" {
					actualLeftGNSAmount := GetLeftGNSAmount()
					if tt.expectedLeftAmount > 0 {
						uassert.Equal(t, tt.expectedLeftAmount, actualLeftGNSAmount)
					}
				}
			}
		})
	}
}
