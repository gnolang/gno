package staker

import (
	"testing"

	"gno.land/p/demo/testutils"
)

func TestRewardCalculation_1_1(t *testing.T) {
	state := NewRewardState()

	current := 100
	state.AddStake(10, testutils.TestAddress("alice"), 100, uint64(current), make(map[string]uint64))

	current += 100
	reward, _ := state.RemoveStake(testutils.TestAddress("alice"), 100, uint64(current), make(map[string]uint64))

	if reward != 100+100 {
		t.Errorf("expected reward %d, got %d", 100+100, reward)
	}
}

func TestRewardCalculation_1_2(t *testing.T) {
	state := NewRewardState()

	current := 100
	state.AddStake(10, testutils.TestAddress("alice"), 100, uint64(current), make(map[string]uint64))

	current += 100
	reward, _ := state.RemoveStake(testutils.TestAddress("alice"), 100, uint64(current), make(map[string]uint64))
	current -= int(reward)

	if reward != 100+100 {
		t.Errorf("expected reward %d, got %d", 100+100, reward)
	}

	current += 100
	state.AddStake(12, testutils.TestAddress("bob"), 100, uint64(current), make(map[string]uint64))

	current += 100
	reward, _ = state.RemoveStake(testutils.TestAddress("bob"), 100, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 100+100 {
		t.Errorf("expected reward %d, got %d", 100+100, reward)
	}
}

func TestRewardCalculation_1_3(t *testing.T) {
	state := NewRewardState()

	// Alice takes 100 GNS
	current := 100
	state.AddStake(10, testutils.TestAddress("alice"), 10, uint64(current), make(map[string]uint64))

	// Alice takes 100 GNS
	current += 100
	state.AddStake(11, testutils.TestAddress("bob"), 10, uint64(current), make(map[string]uint64))

	// Alice takes 50 GNS, Bob takes 50 GNS
	current += 100
	reward, _ := state.RemoveStake(testutils.TestAddress("alice"), 10, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 100+100+50 {
		t.Errorf("expected reward %d, got %d", 100+100+50, reward)
	}

	// Bob takes 100 GNS
	current += 100
	reward, _ = state.RemoveStake(testutils.TestAddress("bob"), 10, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 100+50 {
		t.Errorf("expected reward %d, got %d", 100+50, reward)
	}
}


func TestRewardCalculation_1_4(t *testing.T) {
	state := NewRewardState()

	// Alice takes 100 GNS
	current := 100
	state.AddStake(10, testutils.TestAddress("alice"), 10, uint64(current), make(map[string]uint64))

	// Alice takes 200GNS
	current += 200
	state.AddStake(11, testutils.TestAddress("bob"), 30, uint64(current), make(map[string]uint64))

	// Alice 25, Bob 75
	current += 100
	state.AddStake(12, testutils.TestAddress("charlie"), 10, uint64(current), make(map[string]uint64))

	// Alice 20, Bob 60, Charlie 20
	current += 100
	reward, _ := state.RemoveStake(testutils.TestAddress("alice"), 10, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 100+200+25+20 {
		t.Errorf("expected reward %d, got %d", 100+200+25+20, reward)
	}

	// Bob 75, Charlie 25
	current += 100
	reward, _ = state.RemoveStake(testutils.TestAddress("bob"), 30, uint64(current), make(map[string]uint64))
	current -= int(reward)
	if reward != 75+60+75 {
		t.Errorf("expected reward %d, got %d", 75+60+75, reward)
	}
}
