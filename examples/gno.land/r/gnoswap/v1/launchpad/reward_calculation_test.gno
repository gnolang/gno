package launchpad

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	u256 "gno.land/p/gnoswap/uint256"
)

func TestRewardState_BasicFlow(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128) // 1000 * 2^128
	startHeight := uint64(100)
	endHeight := uint64(1000)
	state := NewRewardState(rewardPerBlock, startHeight, endHeight)

	currentHeight := uint64(150)
	depositId := "deposit1"
	stakeAmount := uint64(500)
	state.addStake(currentHeight, depositId, stakeAmount)

	uassert.Equal(t, state.TotalStake, stakeAmount, ufmt.Sprintf("TotalStake = %d, want %d", state.TotalStake, stakeAmount))

	claimHeight := uint64(200)
	reward := state.claim(depositId, claimHeight)
	uassert.NotEqual(t, reward, uint64(0), "Expected non-zero reward")

	removeHeight := uint64(250)
	finalReward := state.removeStake(depositId, stakeAmount, removeHeight)
	uassert.Equal(t, state.TotalStake, uint64(0), ufmt.Sprintf("TotalStake after removal = %d, want 0", state.TotalStake))
}

func TestRewardState_MultipleStakers(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	state.addStake(150, "deposit1", 300)
	state.addStake(160, "deposit2", 200)

	uassert.Equal(t, state.TotalStake, uint64(500), ufmt.Sprintf("TotalStake = %d, want 500", state.TotalStake))

	// check first staker's reward
	reward1 := state.claim("deposit1", 200)
	uassert.NotEqual(t, reward1, uint64(0), "Expected non-zero reward for deposit1")

	// check 2nd staker's reward
	reward2 := state.claim("deposit2", 200)
	uassert.NotEqual(t, reward2, uint64(0), "Expected non-zero reward for deposit2")

	// reward1 must be greater than reward2
	// because reward1 staked earlier than reward2
	if reward1 <= reward2 {
		t.Error("Expected reward1 > reward2")
	}
}

func TestRewardState_EmptyBlocks(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	// call `finalize` while the block is empty
	state.finalize(200)
	uassert.Equal(t, state.TotalEmptyBlock, uint64(100), ufmt.Sprintf("TotalEmptyBlock = %d, want 100", state.TotalEmptyBlock))

	// block cound must be stopped when staker added
	state.addStake(250, "deposit1", 100)
	state.finalize(300)

	uassert.Equal(t, state.TotalEmptyBlock, uint64(150), ufmt.Sprintf("TotalEmptyBlock after staking = %d, want 150", state.TotalEmptyBlock))
}

func TestRewardState_EndHeightBehavior(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	startHeight := uint64(100)
	endHeight := uint64(200)
	state := NewRewardState(rewardPerBlock, startHeight, endHeight)

	// stake before endHeight
	state.addStake(150, "deposit1", 100)

	// claim after endHeight
	reward := state.claim("deposit1", 250)

	// must be no additional reward after endHeight
	secondReward := state.claim("deposit1", 300)
	uassert.Equal(t, secondReward, uint64(0), ufmt.Sprintf("Got reward after endHeight: %d, want 0", secondReward))
}

func TestRewardState_PartialStakeRemoval(t *testing.T) {
	t.Skip("fix this test")
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	state.addStake(150, "deposit1", 500)

	// checking reward after partial removal
	reward1 := state.removeStake("deposit1", 200, 200)
	reward2 := state.claim("deposit1", 250)

	uassert.Equal(t, state.TotalStake, uint64(300), ufmt.Sprintf("TotalStake after partial removal = %d, want 300", state.TotalStake))
}

func TestRewardState_BoundaryCases(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	// stake before startHeight
	state.addStake(50, "deposit1", 100)
	if state.LastHeight != 100 {
		t.Errorf("LastHeight = %d, want 100", state.LastHeight)
	}

	state.addStake(150, "deposit2", 1)
	state.addStake(150, "deposit3", ^uint64(0)-1)

	reward1 := state.claim("deposit2", 200)
	reward2 := state.claim("deposit2", 200)
	uassert.Equal(t, reward2, uint64(0), ufmt.Sprintf("Second claim at same height should be 0, got %d", reward2))
}

func TestRewardState_MultipleStakesFromSameUser(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	// stake from the same user at different times
	state.addStake(150, "user1_deposit1", 200)
	state.addStake(200, "user1_deposit2", 300)

	reward1 := state.claim("user1_deposit1", 250)
	reward2 := state.claim("user1_deposit2", 250)

	uassert.NotEqual(t, reward1, uint64(0), "First deposit should have rewards")
	uassert.NotEqual(t, reward2, uint64(0), "Second deposit should have rewards")
}

func TestRewardState_RewardCalculationAccuracy(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	state.addStake(150, "deposit1", 100)
	state.addStake(150, "deposit2", 100)

	reward1 := state.claim("deposit1", 200)
	reward2 := state.claim("deposit2", 200)

	// in a same condition, the same reward should be paid
	uassert.Equal(t, reward1, reward2, "Equal stakes should receive equal rewards")
}

func TestRewardState_ZeroRewardPerBlock(t *testing.T) {
	// Test with zero rewards per block
	rewardPerBlock := u256.Zero()
	state := NewRewardState(rewardPerBlock, 100, 1000)

	state.addStake(150, "deposit1", 100)
	reward := state.claim("deposit1", 200)

	uassert.Equal(t, reward, uint64(0),
		"Expected zero reward when rewardPerBlock is zero")
}

func TestRewardState_ConsecutiveStakeAndUnstake(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	// Repeatedly stake and unstake
	state.addStake(150, "deposit1", 100)
	state.removeStake("deposit1", 100, 160)

	state.addStake(170, "deposit2", 200)
	state.removeStake("deposit2", 200, 180)

	state.addStake(190, "deposit3", 300)

	uassert.Equal(t, state.TotalStake, uint64(300),
		"Final stake amount should be correct after multiple stake/unstake operations")
}

func TestRewardState_ClaimAtEndHeight(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	startHeight := uint64(100)
	endHeight := uint64(200)
	state := NewRewardState(rewardPerBlock, startHeight, endHeight)

	state.addStake(150, "deposit1", 100)

	// claim exactly at endHeight
	reward1 := state.claim("deposit1", endHeight)

	// claim after endHeight
	reward2 := state.claim("deposit1", endHeight+50)

	uassert.NotEqual(t, reward1, uint64(0),
		"Should receive rewards when claiming at endHeight")
	uassert.Equal(t, reward2, uint64(0),
		"Should not receive additional rewards after endHeight")
}

func TestRewardState_StakeChangesDuringRewardPeriod(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	// First staker
	state.addStake(150, "deposit1", 1000)

	// Second staker joins
	state.addStake(160, "deposit2", 1000)

	// First staker removes half
	reward1 := state.removeStake("deposit1", 500, 170)

	// Second staker claims
	reward2 := state.claim("deposit2", 170)

	uassert.Equal(t, reward1, uint64(15000))
	uassert.Equal(t, reward2, uint64(5000))
}

func TestRewardState_SmallIntervals(t *testing.T) {
	rewardPerBlock := u256.NewUint(1000).Lsh(u256.NewUint(1000), 128)
	state := NewRewardState(rewardPerBlock, 100, 1000)

	// Test with consecutive block heights
	state.addStake(150, "deposit1", 100)
	reward1 := state.claim("deposit1", 151)
	reward2 := state.claim("deposit1", 152)

	uassert.NotEqual(t, reward1, uint64(0),
		"Should receive rewards for single block interval")

	uassert.Equal(t, reward1, reward2)
	uassert.Equal(t, reward1, uint64(1000))
}
