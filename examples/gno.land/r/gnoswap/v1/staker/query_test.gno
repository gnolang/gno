package staker

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestQueryPoolData(t *testing.T) {
	testAddr := testutils.TestAddress("test_address")
	poolPath := "test_pool"

	pool := NewPool(poolPath, std.ChainHeight())
	pools.set(poolPath, pool)

	// Add test incentive to pool
	startTime := time.Now().Unix()
	endTime := startTime + 3600 // 1 hour
	rewardToken := "TEST_TOKEN"
	rewardAmount := uint64(1000)

	incentive := &ExternalIncentive{
		startTimestamp: startTime,
		endTimestamp:   endTime,
		startHeight:    std.ChainHeight(),
		endHeight:      std.ChainHeight() + 100,
		rewardToken:    rewardToken,
		rewardAmount:   rewardAmount,
		refundee:       testAddr,
	}

	pool.incentives.create(testAddr, incentive)

	data, err := QueryPoolData(poolPath)
	if err != nil {
		t.Errorf("QueryPoolData failed: %v", err)
	}

	uassert.Equal(t, data.PoolPath, poolPath)
	uassert.Equal(t, len(data.ActiveIncentives), 1)

	_, err = QueryPoolData("non_existent_pool")
	uassert.Error(t, err)

	// reset
	pools.tree.Remove(poolPath)
}

func TestQueryDepositData(t *testing.T) {
	testAddr := testutils.TestAddress("test_address")
	poolPath := "test_pool"
	lpTokenId := uint64(1)

	deposit := &Deposit{
		owner:          testAddr,
		targetPoolPath: poolPath,
		stakeTimestamp: time.Now().Unix(),
		liquidity:      u256.NewUint(1000),
		warmups:        make([]Warmup, 2),
	}

	deposits.set(lpTokenId, deposit)

	// Test query
	data, err := QueryDepositData(lpTokenId)
	if err != nil {
		t.Errorf("QueryDepositData failed: %v", err)
	}

	uassert.Equal(t, data.TokenID, lpTokenId)
	uassert.Equal(t, data.Owner, testAddr)
	uassert.Equal(t, data.TargetPoolPath, poolPath)
	uassert.Equal(t, data.StakeTimestamp, deposit.stakeTimestamp)
	uassert.Equal(t, data.WarmupCount, 2)

	// reset
	deposits.remove(lpTokenId)
}
