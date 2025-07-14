package protocol_fee

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestProtocolFeeStateAddAccuToGovStaker(t *testing.T) {
	tests := []struct {
		name        string
		tokenPath   string
		amounts     []uint64
		expectedSum uint64
	}{
		{
			name:        "add single amount to new token",
			tokenPath:   "test_token",
			amounts:     []uint64{100},
			expectedSum: 100,
		},
		{
			name:        "add multiple amounts to same token",
			tokenPath:   "test_token",
			amounts:     []uint64{100, 50},
			expectedSum: 150,
		},
		{
			name:        "add zero amount",
			tokenPath:   "test_token",
			amounts:     []uint64{0},
			expectedSum: 0,
		},
		{
			name:        "add multiple amounts including zero",
			tokenPath:   "test_token",
			amounts:     []uint64{100, 0, 75},
			expectedSum: 175,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pfs := NewProtocolFeeState()

			// Initially should be 0
			uassert.Equal(t, pfs.GetAccuTransferToGovStakerByTokenPath(test.tokenPath), uint64(0))

			// Add amounts sequentially
			for _, amount := range test.amounts {
				pfs.addAccuToGovStaker(test.tokenPath, amount)
			}

			// Check final sum
			uassert.Equal(t, pfs.GetAccuTransferToGovStakerByTokenPath(test.tokenPath), test.expectedSum)
		})
	}
}

func TestProtocolFeeStateAddAccuToDevOps(t *testing.T) {
	tests := []struct {
		name        string
		tokenPath   string
		amounts     []uint64
		expectedSum uint64
	}{
		{
			name:        "add single amount to new token",
			tokenPath:   "test_token",
			amounts:     []uint64{200},
			expectedSum: 200,
		},
		{
			name:        "add multiple amounts to same token",
			tokenPath:   "test_token",
			amounts:     []uint64{200, 75},
			expectedSum: 275,
		},
		{
			name:        "add zero amount",
			tokenPath:   "test_token",
			amounts:     []uint64{0},
			expectedSum: 0,
		},
		{
			name:        "add large amounts",
			tokenPath:   "test_token",
			amounts:     []uint64{1000, 2000, 3000},
			expectedSum: 6000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pfs := NewProtocolFeeState()

			// Initially should be 0
			uassert.Equal(t, pfs.GetAccuTransferToDevOpsByTokenPath(test.tokenPath), uint64(0))

			// Add amounts sequentially
			for _, amount := range test.amounts {
				pfs.addAccuToDevOps(test.tokenPath, amount)
			}

			// Check final sum
			uassert.Equal(t, pfs.GetAccuTransferToDevOpsByTokenPath(test.tokenPath), test.expectedSum)
		})
	}
}

func TestProtocolFeeStateSetDevOpsPct(t *testing.T) {
	tests := []struct {
		name        string
		percentage  uint64
		shouldPanic bool
		panicMsg    string
		expectedPct uint64
	}{
		{
			name:        "set valid percentage - 0%",
			percentage:  0,
			shouldPanic: false,
			expectedPct: 0,
		},
		{
			name:        "set valid percentage - 25%",
			percentage:  2500,
			shouldPanic: false,
			expectedPct: 2500,
		},
		{
			name:        "set valid percentage - 100%",
			percentage:  10000,
			shouldPanic: false,
			expectedPct: 10000,
		},
		{
			name:        "set invalid percentage - over 100%",
			percentage:  10001,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-PROTOCOL_FEE-002] invalid percentage || pct(10001) should not be bigger than 10000",
		},
		{
			name:        "set invalid percentage - very large",
			percentage:  50000,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-PROTOCOL_FEE-002] invalid percentage || pct(50000) should not be bigger than 10000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pfs := NewProtocolFeeState()

			// Initially should be 0
			uassert.Equal(t, pfs.DevOpsPct, uint64(0))

			if test.shouldPanic {
				uassert.PanicsWithMessage(t, test.panicMsg, func() {
					pfs.setDevOpsPct(test.percentage)
				})
			} else {
				pfs.setDevOpsPct(test.percentage)
				uassert.Equal(t, pfs.DevOpsPct, test.expectedPct)
			}
		})
	}
}

func TestProtocolFeeStateClearTokenListWithAmount(t *testing.T) {
	tests := []struct {
		name        string
		initialData map[string]uint64
		expectedLen int
	}{
		{
			name:        "clear empty list",
			initialData: map[string]uint64{},
			expectedLen: 0,
		},
		{
			name: "clear list with single token",
			initialData: map[string]uint64{
				"token1": 100,
			},
			expectedLen: 0,
		},
		{
			name: "clear list with multiple tokens",
			initialData: map[string]uint64{
				"token1": 100,
				"token2": 200,
				"token3": 300,
			},
			expectedLen: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pfs := NewProtocolFeeState()

			// Setup initial data
			for token, amount := range test.initialData {
				pfs.TokenListWithAmount[token] = amount
			}

			// Verify initial state
			uassert.Equal(t, len(pfs.TokenListWithAmount), len(test.initialData))

			// Clear the list
			pfs.clearTokenListWithAmount()

			// Verify final state
			uassert.Equal(t, len(pfs.TokenListWithAmount), test.expectedLen)
		})
	}
}

func TestProtocolFeeStateUpdateDistributedToGovStakerHistory(t *testing.T) {
	tests := []struct {
		name        string
		tokenPath   string
		amounts     []uint64
		expectedSum uint64
	}{
		{
			name:        "update history with single amount",
			tokenPath:   "test_token",
			amounts:     []uint64{100},
			expectedSum: 100,
		},
		{
			name:        "update history with multiple amounts",
			tokenPath:   "test_token",
			amounts:     []uint64{100, 50},
			expectedSum: 150,
		},
		{
			name:        "update history with zero amount",
			tokenPath:   "test_token",
			amounts:     []uint64{0},
			expectedSum: 0,
		},
		{
			name:        "update history multiple times",
			tokenPath:   "test_token",
			amounts:     []uint64{25, 75, 100, 200},
			expectedSum: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pfs := NewProtocolFeeState()

			// Update history sequentially
			for _, amount := range test.amounts {
				pfs.updateDistributedToGovStakerHistory(test.tokenPath, amount)
			}

			// Check if history was updated correctly
			historyValue, exists := pfs.DistributedToGovStakerHistory.Get(test.tokenPath)
			uassert.True(t, exists)
			uassert.Equal(t, historyValue.(uint64), test.expectedSum)
		})
	}
}

func TestProtocolFeeStateUpdateDistributedToDevOpsHistory(t *testing.T) {
	tests := []struct {
		name        string
		tokenPath   string
		amounts     []uint64
		expectedSum uint64
	}{
		{
			name:        "update history with single amount",
			tokenPath:   "test_token",
			amounts:     []uint64{200},
			expectedSum: 200,
		},
		{
			name:        "update history with multiple amounts",
			tokenPath:   "test_token",
			amounts:     []uint64{200, 75},
			expectedSum: 275,
		},
		{
			name:        "update history with zero amount",
			tokenPath:   "test_token",
			amounts:     []uint64{0},
			expectedSum: 0,
		},
		{
			name:        "update history multiple times",
			tokenPath:   "test_token",
			amounts:     []uint64{50, 150, 300, 100},
			expectedSum: 600,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pfs := NewProtocolFeeState()

			// Update history sequentially
			for _, amount := range test.amounts {
				pfs.updateDistributedToDevOpsHistory(test.tokenPath, amount)
			}

			// Check if history was updated correctly
			historyValue, exists := pfs.DistributedToDevOpsHistory.Get(test.tokenPath)
			uassert.True(t, exists)
			uassert.Equal(t, historyValue.(uint64), test.expectedSum)
		})
	}
}
