package staker

import (
	"testing"
)

func TestHandleUnstakingFee(t *testing.T) {
	tests := []struct {
		name        string
		tokenPath   string
		amount      uint64
		internal    bool
		tokenId     uint64
		poolPath    string
		expectedFee uint64
		expectedNet uint64
	}{
		{
			name:        "No fee configured",
			tokenPath:   gnsPath,
			amount:      10000,
			internal:    true,
			tokenId:     1,
			poolPath:    "pool1",
			expectedFee: 0,
			expectedNet: 10000,
		},
		{
			name:        "Standard fee",
			tokenPath:   gnsPath,
			amount:      10000,
			internal:    false,
			tokenId:     1,
			poolPath:    "pool1",
			expectedFee: 100,
			expectedNet: 9900,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unstakingFee = tc.expectedFee // Set the fee globally for the test

			netAmount := handleUnStakingFee(tc.tokenPath, tc.amount, tc.internal, tc.tokenId, tc.poolPath)

			if netAmount != tc.expectedNet {
				t.Errorf("Expected netAmount %d, got %d", tc.expectedNet, netAmount)
			}
		})
	}
}

func TestSetUnstakingFee(t *testing.T) {
	tests := []struct {
		name        string
		fee         uint64
		shouldPanic bool
	}{
		{name: "Valid fee", fee: 500, shouldPanic: false},
		{name: "Excessive fee", fee: 10001, shouldPanic: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("Unexpected panic for fee %d: %v", tc.fee, r)
				}
			}()

			setUnStakingFee(tc.fee)

			if !tc.shouldPanic {
				if unstakingFee != tc.fee {
					t.Errorf("Expected fee %d, got %d", tc.fee, unstakingFee)
				}
			}
		})
	}
}
