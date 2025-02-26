package staker

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
)

func TestGetDelegatedCumulative(t *testing.T) {
	delegationSnapShotHistory = avl.NewTree()

	addr1 := testutils.TestAddress("test1")
	now := uint64(time.Now().Unix())

	tests := []struct {
		name         string
		setupHistory []DelegationSnapShotHistory
		delegator    std.Address
		endTimestamp uint64
		expectAmount uint64
		expectPanic  bool
	}{
		{
			name:         "no history returns zero",
			delegator:    addr1,
			endTimestamp: now,
			expectAmount: 0,
		},
		{
			name: "single history before timestamp",
			setupHistory: []DelegationSnapShotHistory{
				{
					to:           addr1,
					amount:       100,
					updatedBlock: 1,
					updatedAt:    now - 100,
				},
			},
			delegator:    addr1,
			endTimestamp: now,
			expectAmount: 100,
		},
		{
			name: "multiple histories returns latest before timestamp",
			setupHistory: []DelegationSnapShotHistory{
				{
					to:           addr1,
					amount:       100,
					updatedBlock: 1,
					updatedAt:    now - 200,
				},
				{
					to:           addr1,
					amount:       150,
					updatedBlock: 2,
					updatedAt:    now - 100,
				},
				{
					to:           addr1,
					amount:       200,
					updatedBlock: 3,
					updatedAt:    now + 100, // Future update
				},
			},
			delegator:    addr1,
			endTimestamp: now,
			expectAmount: 150,
		},
		{
			name:         "future timestamp panics",
			delegator:    addr1,
			endTimestamp: now + 1000,
			expectPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delegationSnapShotHistory = avl.NewTree()

			if len(tt.setupHistory) > 0 {
				delegationSnapShotHistory.Set(tt.delegator.String(), tt.setupHistory)
			}

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic but got none")
					}
				}()
			}

			result := GetDelegatedCumulative(tt.delegator, tt.endTimestamp)

			if !tt.expectPanic && result != tt.expectAmount {
				t.Errorf("expected amount %d but got %d", tt.expectAmount, result)
			}
		})
	}
}
