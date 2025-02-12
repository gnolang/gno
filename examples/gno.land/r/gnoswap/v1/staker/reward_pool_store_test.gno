package staker

import "testing"

func TestRewardPool(t *testing.T) {
	pool := NewRewardPool()

	// Test default values
	if pool.GetTier() != 0 {
		t.Errorf("Expected default tier to be 0, got %d", pool.GetTier())
	}
	if pool.GetRewardAmount() != 0 {
		t.Errorf("Expected default rewardAmount to be 0, got %d", pool.GetRewardAmount())
	}
	if pool.GetDistributedAmount() != 0 {
		t.Errorf("Expected default distributedAmount to be 0, got %d", pool.GetDistributedAmount())
	}
	if pool.GetLeftAmount() != 0 {
		t.Errorf("Expected default leftAmount to be 0, got %d", pool.GetLeftAmount())
	}

	// Test setters
	pool.SetTier(1)
	pool.SetRewardAmount(1000)
	pool.SetDistributedAmount(500)
	pool.SetLeftAmount(500)

	// Test getters
	if pool.GetTier() != 1 {
		t.Errorf("Expected tier to be 1, got %d", pool.GetTier())
	}
	if pool.GetRewardAmount() != 1000 {
		t.Errorf("Expected rewardAmount to be 1000, got %d", pool.GetRewardAmount())
	}
	if pool.GetDistributedAmount() != 500 {
		t.Errorf("Expected distributedAmount to be 500, got %d", pool.GetDistributedAmount())
	}
	if pool.GetLeftAmount() != 500 {
		t.Errorf("Expected leftAmount to be 500, got %d", pool.GetLeftAmount())
	}
}

func TestRewardPoolMap(t *testing.T) {
	rewardMap := NewRewardPoolMap()

	// Test default values
	if len(rewardMap.GetRewardPools()) != 0 {
		t.Errorf("Expected empty rewardPools map, got %d entries", len(rewardMap.GetRewardPools()))
	}

	// Test adding a new pool
	rewardMap.SetRewardPool("pool1", &RewardPool{tier: 1, rewardAmount: 1000})
	rewardPool := rewardMap.GetRewardPoolByPoolPath("pool1")
	if rewardPool.GetTier() != 1 {
		t.Errorf("Expected tier of pool1 to be 1, got %d", rewardPool.GetTier())
	}
	if rewardPool.GetRewardAmount() != 1000 {
		t.Errorf("Expected rewardAmount of pool1 to be 1000, got %d", rewardPool.GetRewardAmount())
	}

	// Test updating pool tier
	rewardMap.SetPoolTier("pool1", 2)
	if rewardMap.GetPoolTier("pool1") != 2 {
		t.Errorf("Expected updated tier of pool1 to be 2, got %d", rewardMap.GetPoolTier("pool1"))
	}

	// Test setting and getting reward amount for tiers
	rewardMap.SetRewardAmountForTier(TIER1_INDEX, 5000)
	if rewardMap.GetRewardAmountForTier(TIER1_INDEX) != 5000 {
		t.Errorf("Expected rewardAmountForTier[TIER1_INDEX] to be 5000, got %d", rewardMap.GetRewardAmountForTier(TIER1_INDEX))
	}

	// Test setting and getting left amount for tiers
	rewardMap.SetLeftAmountForTier(TIER2_INDEX, 3000)
	if rewardMap.GetLeftAmountForTier(TIER2_INDEX) != 3000 {
		t.Errorf("Expected leftAmountForTier[TIER2_INDEX] to be 3000, got %d", rewardMap.GetLeftAmountForTier(TIER2_INDEX))
	}
}
