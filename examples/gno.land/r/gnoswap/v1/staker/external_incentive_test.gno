package staker

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/gnoswap/consts"
)

func TestCreateExternalIncentiveWithAllowedTokens(t *testing.T) {
	poolPath := "test_pool"
	account1 := testutils.TestAddress("account1")
	account2 := testutils.TestAddress("account2")

	// create pool
	pool := NewPool(poolPath, 100)
	pools.set(poolPath, pool)

	currentTime := time.Now().Unix()
	startTime := currentTime + 86400   // 1 day later
	endTime := startTime + (86400 * 7) // 7 days after start

	tests := []struct {
		name        string
		creator     std.Address
		rewardToken string
		amount      int64
		shouldPass  bool
	}{
		{
			name:        "create incentive with GNS token",
			creator:     account1,
			rewardToken: consts.GNS_PATH,
			amount:      1000,
			shouldPass:  true,
		},
		{
			name:        "create incentive with WUGNOT token",
			creator:     account1,
			rewardToken: consts.WUGNOT_PATH,
			amount:      2000,
			shouldPass:  true,
		},
		{
			name:        "create incentive with pool's first token",
			creator:     account2,
			rewardToken: "token0_path",
			amount:      1500,
			shouldPass:  true,
		},
		{
			name:        "create incentive with pool's second token",
			creator:     account2,
			rewardToken: "token1_path",
			amount:      2500,
			shouldPass:  true,
		},
		{
			name:        "create incentive with different token by same account",
			creator:     account1,
			rewardToken: "token0_path",
			amount:      3000,
			shouldPass:  true,
		},
		{
			name:        "create incentive with same token by different account",
			creator:     account2,
			rewardToken: consts.GNS_PATH,
			amount:      3500,
			shouldPass:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// create incentive
			incentive := &ExternalIncentive{
				startTimestamp: startTime,
				endTimestamp:   endTime,
				startHeight:    100,
				endHeight:      200,
				rewardToken:    tc.rewardToken,
				rewardAmount:   tc.amount,
				refundee:       tc.creator,
			}

			// add incentive
			pool.incentives.create(tc.creator, incentive)

			if tc.shouldPass {
				// check if incentive was actually created
				ictv, exists := pool.incentives.Get(startTime, endTime, tc.creator, tc.rewardToken)
				if !exists {
					t.Error("incentive was not created")
				}

				if ictv.rewardAmount != tc.amount {
					t.Errorf("expected reward amount %d, got %d", tc.amount, ictv.rewardAmount)
				}
			}
		})
	}

	// clean up after test
	pools.tree.Remove(poolPath)
}
