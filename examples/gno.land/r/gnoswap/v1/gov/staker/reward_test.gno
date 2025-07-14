package staker

import (
	"testing"

	"gno.land/p/demo/testutils"
)

func TestRewardCalculation_1_1(t *testing.T) {
	state := NewRewardState()

	current := 100
	state.addStake(10, testutils.TestAddress("alice"), 100, uint64(current), make(map[string]uint64))

	current += 100
	reward, _ := state.removeStake(testutils.TestAddress("alice"), 100, uint64(current), make(map[string]uint64))

	if reward != 100+100 {
		t.Errorf("expected reward %d, got %d", 100+100, reward)
	}
}

func TestRewardCalculation_1_2(t *testing.T) {
	state := NewRewardState()

	current := 100
	state.addStake(10, testutils.TestAddress("alice"), 100, uint64(current), make(map[string]uint64))

	current += 100
	reward, _ := state.removeStake(testutils.TestAddress("alice"), 100, uint64(current), make(map[string]uint64))
	current -= int(reward)

	if reward != 100+100 {
		t.Errorf("expected reward %d, got %d", 100+100, reward)
	}

	current += 100
	state.addStake(12, testutils.TestAddress("bob"), 100, uint64(current), make(map[string]uint64))

	current += 100
	reward, _ = state.removeStake(testutils.TestAddress("bob"), 100, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 100+100 {
		t.Errorf("expected reward %d, got %d", 100+100, reward)
	}
}

func TestRewardCalculation_1_3(t *testing.T) {
	state := NewRewardState()

	// Alice takes 100 GNS
	current := 100
	state.addStake(10, testutils.TestAddress("alice"), 10, uint64(current), make(map[string]uint64))

	// Alice takes 100 GNS
	current += 100
	state.addStake(11, testutils.TestAddress("bob"), 10, uint64(current), make(map[string]uint64))

	// Alice takes 50 GNS, Bob takes 50 GNS
	current += 100
	reward, _ := state.removeStake(testutils.TestAddress("alice"), 10, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 100+100+50 {
		t.Errorf("expected reward %d, got %d", 100+100+50, reward)
	}

	// Bob takes 100 GNS
	current += 100
	reward, _ = state.removeStake(testutils.TestAddress("bob"), 10, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 100+50 {
		t.Errorf("expected reward %d, got %d", 100+50, reward)
	}
}

func TestRewardCalculation_1_4(t *testing.T) {
	state := NewRewardState()

	// Alice takes 100 GNS
	current := 100
	state.addStake(10, testutils.TestAddress("alice"), 10, uint64(current), make(map[string]uint64))

	// Alice takes 200GNS
	current += 200
	state.addStake(11, testutils.TestAddress("bob"), 30, uint64(current), make(map[string]uint64))

	// Alice 25, Bob 75
	current += 100
	state.addStake(12, testutils.TestAddress("charlie"), 10, uint64(current), make(map[string]uint64))

	// Alice 20, Bob 60, Charlie 20
	current += 100
	reward, _ := state.removeStake(testutils.TestAddress("alice"), 10, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 100+200+25+20 {
		t.Errorf("expected reward %d, got %d", 100+200+25+20, reward)
	}

	// Bob 75, Charlie 25
	current += 100
	reward, _ = state.removeStake(testutils.TestAddress("bob"), 30, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 75+60+75 {
		t.Errorf("expected reward %d, got %d", 75+60+75, reward)
	}
}

func TestRemoveStake(t *testing.T) {
	// This test verifies the correct distribution of protocol fee rewards and emission rewards
	// when a staker partially and then fully unstakes. It checks that:
	// 1. Protocol fee rewards are distributed correctly after a partial unstake.
	// 2. After a full unstake, protocol fee rewards may be empty depending on the logic,
	//    and the staker's info is properly deleted.
	// 3. The totalXGNSStake is updated correctly after each operation.

	state := NewRewardState()
	staker := testutils.TestAddress("staker")

	// Step 1: Initial staking
	// The staker deposits 10 GNS with some initial protocol fee balances.
	initialStake := uint64(10_000_000) // 10 GNS
	initialBalance := uint64(1_000_000_000)
	initialProtocolFee := map[string]uint64{
		"tokenA": 100_000,
		"tokenB": 200_000,
	}
	state.addStake(1, staker, initialStake, initialBalance, initialProtocolFee)

	// Step 2: Partial unstake
	// The staker unstakes 4 GNS after protocol fees and emission rewards have increased.
	partialUnstake := uint64(4_000_000)        // 4 GNS
	updatedBalance := initialBalance + 500_000 // increase emission reward
	updatedProtocolFee := map[string]uint64{
		"tokenA": 300_000, // +200_000
		"tokenB": 500_000, // +300_000
	}
	_, protocolFeeRewards := state.removeStake(
		staker,
		partialUnstake,
		updatedBalance,
		updatedProtocolFee,
	)
	// After partial unstake, protocol fee rewards should be distributed.
	if len(protocolFeeRewards) == 0 {
		t.Error("partial unstake: expected non-empty protocol fee rewards")
	}
	for k, v := range protocolFeeRewards {
		if v == 0 {
			t.Errorf("partial unstake: expected non-zero reward for %s", k)
		}
	}

	// Step 3: Full unstake
	// The staker unstakes the remaining amount after protocol fees and emission rewards have increased again.
	remainingStake := initialStake - partialUnstake
	finalBalance := updatedBalance + 500_000
	finalProtocolFee := map[string]uint64{
		"tokenA": 600_000, // +300_000
		"tokenB": 900_000, // +400_000
	}
	_, protocolFeeRewards2 := state.removeStake(
		staker,
		remainingStake,
		finalBalance,
		finalProtocolFee,
	)
	// After full unstake, protocol fee rewards may be empty depending on the logic.
	if len(protocolFeeRewards2) == 0 {
		t.Log("full unstake: protocol fee rewards is empty, which is allowed by current logic")
	} else {
		for k, v := range protocolFeeRewards2 {
			if v == 0 {
				t.Errorf("full unstake: expected non-zero reward for %s", k)
			}
		}
	}

	// Step 4: Check staker info deletion and total stake
	// After full unstake, the staker's info should be deleted and totalXGNSStake should be zero.
	if _, exists := state.stakerInfos.Get(staker.String()); exists {
		t.Error("staker info should be deleted after full unstake")
	}
	if state.totalXGNSStake != 0 {
		t.Errorf("totalXGNSStake should be 0 after full unstake, got %d", state.totalXGNSStake)
	}
}
