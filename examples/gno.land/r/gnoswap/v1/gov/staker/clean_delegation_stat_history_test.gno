package staker

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
)

var (
	testAddr1 = testutils.TestAddress("test1")
	testAddr  = testutils.TestAddress("test")
)

type mockEnv struct {
	height  uint64
	isAdmin bool
}

func (m *mockEnv) GetHeight() int64 {
	return int64(m.height)
}

func (m *mockEnv) IsAdmin() bool {
	return m.isAdmin
}

func TestCleanDelegationStatHistory(t *testing.T) {
	mock := &mockEnv{height: 1000, isAdmin: true}
	std.TestSetOrigCaller(testAddr1)
	delegationSnapShotHistory = avl.NewTree()

	addr := testAddr.String()
	history := []DelegationSnapShotHistory{
		{updatedBlock: 500}, // Old
		{updatedBlock: 900}, // Within threshold
		{updatedBlock: 950}, // Latest
	}
	delegationSnapShotHistory.Set(addr, history)

	tests := []struct {
		name        string
		setupHeight uint64
		lastCleaned uint64
		threshold   int64
		expectedLen int
	}{
		{
			name:        "no clean needed",
			setupHeight: 1000,
			lastCleaned: 999,
			threshold:   100,
			expectedLen: 3,
		},
		{
			name:        "clean old records",
			setupHeight: 1000,
			lastCleaned: 800,
			threshold:   100,
			expectedLen: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock.height = tc.setupHeight
			lastCleanedHeight = tc.lastCleaned
			thresholdVotingWeightBlockHeight = tc.threshold

			cleanDelegationStatHistory()

			value, exists := delegationSnapShotHistory.Get(addr)
			if !exists {
				t.Fatal("history should exist")
			}

			history := value.([]DelegationSnapShotHistory)
			if len(history) != tc.expectedLen {
				t.Errorf("expected history length %d, got %d", tc.expectedLen, len(history))
			}
		})
	}
}
