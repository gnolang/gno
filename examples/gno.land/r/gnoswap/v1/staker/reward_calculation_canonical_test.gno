package staker

// Evaluate against the canonical implementation

import (
	"std"
	"strings"
	"testing"

	ufmt "gno.land/p/demo/ufmt"

	u256 "gno.land/p/gnoswap/uint256"
)

func Setup(t *testing.T) *canonicalRewardState {
	pools := NewPools()
	deposits := NewDeposits()

	return NewCanonicalRewardState(t, pools, deposits, TickCrossHook)
}

func TestCanonicalSimple(t *testing.T) {
	canonical := Setup(t)

	gnousdc := test_gnousdc
	canonical.CreatePool(gnousdc, 1, 150)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	canonical.NextBlock()

	// gnousdc takes the entire emission
	canonical.AssertCanonicalInternalRewardPerPool(gnousdc, canonical.PerBlockEmission)

	expected := canonical.PerBlockEmission * 30 / 100
	canonical.AssertEmulatedRewardOf(0, expected)
	canonical.AssertCanonicalRewardOf(0, expected)
}

// To check precision error
func TestCanonicalLargeStakedLiquidity(t *testing.T) {
	canonical := Setup(t)

	gnousdc := test_gnousdc
	canonical.CreatePool(gnousdc, 1, 150)

	canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		q128,
	)

	canonical.NextBlock()

	canonical.AssertCanonicalInternalRewardPerPool(gnousdc, canonical.PerBlockEmission)

	expected := canonical.PerBlockEmission * 30 / 100
	canonical.AssertEmulatedRewardOf(0, expected)
	canonical.AssertCanonicalRewardOf(0, expected)
}

// To check precision error
func TestCanonicalLargeStakedLiquidity_2(t *testing.T) {
	canonical := Setup(t)

	gnousdc := test_gnousdc
	canonical.CreatePool(gnousdc, 1, 150)

	u2_30 := uint64(1073741824)
	u2_33 := uint64(8589934592)

	// Estimated per-block emission for staker is estimated to not get more than 100 million, but we are stress testing.
	// If more than 100 million is emitted, the overflow could occur inside of the PoolTier distribution logic.
	canonical.SetEmissionUpdate(10000000000000)

	canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(u2_30),
	)

	canonical.StakeToken(
		1,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		100,
		200,
		u256.NewUint(u2_33),
	)

	canonical.NextBlock()

	canonical.AssertCanonicalInternalRewardPerPool(gnousdc, canonical.PerBlockEmission)

	expected := canonical.PerBlockEmission / 100 * 30
	canonical.AssertEmulatedRewardOf(0, expected/9)
	canonical.AssertCanonicalRewardOf(0, expected/9)

	canonical.AssertEmulatedRewardOf(1, expected/9*8)
	canonical.AssertCanonicalRewardOf(1, expected/9*8)
}

// Tests simple case with tick crossing
func TestCanonicalTickCross_0(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 150)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	// The position remains inrange, no change in reward.
	canonical.MoveTick(gnousdc, 120)

	canonical.NextBlock()

	// gnousdc takes the entire emission
	canonical.AssertCanonicalInternalRewardPerPool(gnousdc, canonical.PerBlockEmission)

	expected := canonical.PerBlockEmission * 30 / 100
	canonical.AssertEmulatedRewardOf(0, expected)
	canonical.AssertCanonicalRewardOf(0, expected)

	// The position is now outrange. The reward must be reduced.
	canonical.MoveTick(gnousdc, 90)

	canonical.NextBlock()

	// gnousdc still takes the entire emission, independent from position inrangeness
	canonical.AssertCanonicalInternalRewardPerPool(gnousdc, canonical.PerBlockEmission)

	canonical.AssertEmulatedRewardOf(0, 0)
	canonical.AssertCanonicalRewardOf(0, 0)

	// The position remains outrange.
	canonical.MoveTick(gnousdc, 80)

	canonical.NextBlock()

	// gnousdc still takes the entire emission, independent from position inrangeness
	canonical.AssertCanonicalInternalRewardPerPool(gnousdc, canonical.PerBlockEmission)

	canonical.AssertEmulatedRewardOf(0, 0)
	canonical.AssertCanonicalRewardOf(0, 0)

	// The tick passes through the positions range, remains outrange.
	canonical.MoveTick(gnousdc, 220)

	canonical.NextBlock()

	// gnousdc still takes the entire emission, independent from position inrangeness
	canonical.AssertCanonicalInternalRewardPerPool(gnousdc, canonical.PerBlockEmission)

	canonical.AssertEmulatedRewardOf(0, 0)
	canonical.AssertCanonicalRewardOf(0, 0)

	// The tick goes back to the range.
	canonical.MoveTick(gnousdc, 180)

	canonical.NextBlock()

	// gnousdc still takes the entire emission, independent from position inrangeness
	canonical.AssertCanonicalInternalRewardPerPool(gnousdc, canonical.PerBlockEmission)

	// The position is now inrange, so it takes the entire emission
	canonical.AssertEmulatedRewardOf(0, expected)
	canonical.AssertCanonicalRewardOf(0, expected)
}

// Tests tick crossing with lazy evaluation of position reward
func TestCanonicalTickCross_1(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 150)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	// The position remains inrange, no change in reward.
	canonical.MoveTick(gnousdc, 120)

	canonical.NextBlockNoCanonical()

	// The position is now outrange. The reward must be reduced.
	canonical.MoveTick(gnousdc, 90)

	canonical.NextBlockNoCanonical()

	// The position remains outrange.
	canonical.MoveTick(gnousdc, 80)

	canonical.NextBlockNoCanonical()

	// The tick passes through the positions range, remains outrange.
	canonical.MoveTick(gnousdc, 220)

	canonical.NextBlockNoCanonical()

	// The tick goes back to the range.
	canonical.MoveTick(gnousdc, 180)

	canonical.NextBlockNoCanonical()

	// We check that emulated reward is lazy evaluated when we finally calculate it.
	// It takes the two block's emission into account.
	expected := canonical.PerBlockEmission * 30 / 100
	canonical.AssertEmulatedRewardOf(0, expected*2)
}

// Test tick crossing with multiple positions with same tick, same liquidity
func TestCanonicalTickCross_2(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 150)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		1,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	// eligible
	canonical.MoveTick(gnousdc, 120)
	canonical.NextBlockNoCanonical()

	// not eligible
	canonical.MoveTick(gnousdc, 90)
	canonical.NextBlockNoCanonical()

	// not eligible
	canonical.MoveTick(gnousdc, 80)
	canonical.NextBlockNoCanonical()

	// not eligible
	canonical.MoveTick(gnousdc, 220)
	canonical.NextBlockNoCanonical()

	// eligible
	canonical.MoveTick(gnousdc, 180)
	canonical.NextBlockNoCanonical()

	// eligible
	canonical.MoveTick(gnousdc, 120)
	canonical.NextBlockNoCanonical()

	expected := canonical.PerBlockEmission * 30 / 100
	canonical.AssertEmulatedRewardOf(0, expected*3/2)
	canonical.AssertEmulatedRewardOf(1, expected*3/2)
}

// Test tick crossing with multiple positions with same tick, different liquidity
func TestCanonicalTickCross_3(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 150)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)
	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		1,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		100,
		200,
		u256.NewUint(3000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	// eligible
	canonical.MoveTick(gnousdc, 120)
	canonical.NextBlockNoCanonical()

	// not eligible
	canonical.MoveTick(gnousdc, 90)
	canonical.NextBlockNoCanonical()

	// not eligible
	canonical.MoveTick(gnousdc, 80)
	canonical.NextBlockNoCanonical()

	// not eligible
	canonical.MoveTick(gnousdc, 220)
	canonical.NextBlockNoCanonical()

	// eligible
	canonical.MoveTick(gnousdc, 180)
	canonical.NextBlockNoCanonical()

	// not eligible
	canonical.MoveTick(gnousdc, 80)
	canonical.NextBlockNoCanonical()

	expected := canonical.PerBlockEmission * 30 / 100
	canonical.AssertEmulatedRewardOf(0, expected*2*1/4)
	canonical.AssertEmulatedRewardOf(1, expected*2*3/4)
}

// Test tick crossing with multiple positions with different tick, same liquidity
func TestCanonicalTickCross_4(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		1,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		200,
		400,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	expected := canonical.PerBlockEmission * 30 / 100

	// 0 eligible, 1 not eligible (100:0)
	canonical.MoveTick(gnousdc, 101)
	canonical.NextBlock()
	canonical.AssertCanonicalRewardOf(0, expected)
	canonical.AssertCanonicalRewardOf(1, 0)
	// 0 eligible, 1 not eligible (100:0)
	canonical.MoveTick(gnousdc, 101)
	canonical.NextBlock()
	canonical.AssertCanonicalRewardOf(0, expected)
	canonical.AssertCanonicalRewardOf(1, 0)

	// 0 eligible, 1 eligible (50:50)
	canonical.MoveTick(gnousdc, 201)
	canonical.NextBlock()
	canonical.AssertCanonicalRewardOf(0, expected/2)
	canonical.AssertCanonicalRewardOf(1, expected/2)

	// 0 not eligible, 1 not eligible (0:0)
	canonical.MoveTick(gnousdc, 401)
	canonical.NextBlock()
	canonical.AssertCanonicalRewardOf(0, 0)
	canonical.AssertCanonicalRewardOf(1, 0)

	// 0 not eligible, 1 eligible (0:100)
	canonical.MoveTick(gnousdc, 301)
	canonical.NextBlock()
	canonical.AssertCanonicalRewardOf(0, 0)
	canonical.AssertCanonicalRewardOf(1, expected)

	// 0 total ratio: 250
	// 1 total ratio: 150

	canonical.AssertEmulatedRewardOf(0, expected*4*250/400)
	canonical.AssertEmulatedRewardOf(1, expected*4*150/400)
}

// Test tick crossing at tick boundaries, forward direction
func TestCanonicalTickCross_5(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	expected := canonical.PerBlockEmission * 30 / 100

	// not eligible
	// block 1
	canonical.MoveTick(gnousdc, 99)
	canonical.NextBlock()

	// block 2
	canonical.AssertCanonicalRewardOf(0, 0)
	canonical.AssertEmulatedRewardOf(0, 0)

	// eligible
	// entered range
	canonical.MoveTick(gnousdc, 100)
	canonical.NextBlock()

	// block 3
	canonical.AssertCanonicalRewardOf(0, expected)
	canonical.AssertEmulatedRewardOf(0, expected)

	// eligible
	canonical.MoveTick(gnousdc, 101)
	canonical.NextBlock()

	// block 4
	canonical.AssertCanonicalRewardOf(0, expected)
	canonical.AssertEmulatedRewardOf(0, expected)

	// eligible
	canonical.MoveTick(gnousdc, 299)
	canonical.NextBlock()

	canonical.AssertCanonicalRewardOf(0, expected)
	canonical.AssertEmulatedRewardOf(0, expected)

	// not eligible
	canonical.MoveTick(gnousdc, 300)
	canonical.NextBlock()

	canonical.AssertCanonicalRewardOf(0, 0)
	canonical.AssertEmulatedRewardOf(0, 0)

	// not eligible
	canonical.MoveTick(gnousdc, 301)
	canonical.NextBlock()

	canonical.AssertCanonicalRewardOf(0, 0)
	canonical.AssertEmulatedRewardOf(0, 0)
}

// Test tick crossing at tick boundaries, backward direction
func TestCanonicalTickCross_6(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	expected := canonical.PerBlockEmission * 30 / 100

	// not eligible
	canonical.MoveTick(gnousdc, 301)
	canonical.NextBlockNoCanonical()

	canonical.AssertEmulatedRewardOf(0, 0)

	// not eligible
	canonical.MoveTick(gnousdc, 300)
	canonical.NextBlockNoCanonical()

	canonical.AssertEmulatedRewardOf(0, 0)

	// eligible
	canonical.MoveTick(gnousdc, 299)
	canonical.NextBlockNoCanonical()

	canonical.AssertEmulatedRewardOf(0, expected)

	// eligible
	canonical.MoveTick(gnousdc, 101)
	canonical.NextBlockNoCanonical()

	canonical.AssertEmulatedRewardOf(0, expected)

	// eligible
	canonical.MoveTick(gnousdc, 100)
	canonical.NextBlockNoCanonical()

	canonical.AssertEmulatedRewardOf(0, expected)

	// not eligible
	canonical.MoveTick(gnousdc, 99)
	canonical.NextBlockNoCanonical()

	canonical.AssertEmulatedRewardOf(0, 0)
}

func TestCanonicalExternalReward_1(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	currentHeight := int64(canonical.CurrentHeight())
	incentiveId := canonical.CreateExternalIncentive(gnousdc, gnsPath, 100000000, currentHeight+1, currentHeight+5, currentHeight+1, currentHeight+5, std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"))

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgg"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	canonical.NextBlock()

	// Incentive has been started, but but we can collect rewards accumulated until the previous block

	canonical.NextBlock()

	// now there is a single block worth of reward

	expected := uint64(100000000/4) * 30 / 100

	canonical.AssertEmulatedExternalRewardOf(uint64(0), incentiveId, expected)
}

func TestCanonicalExternalReward_2(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	currentHeight := int64(canonical.CurrentHeight())
	incentiveId := canonical.CreateExternalIncentive(gnousdc, gnsPath, 100000000, currentHeight+1, currentHeight+5, currentHeight+1, currentHeight+5, std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"))

	canonical.NextBlock()
	canonical.NextBlock()

	// Incentive has been already started

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgg"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	canonical.NextBlock()

	// now there is a single block worth of reward

	expected := uint64(100000000/4) * 30 / 100

	canonical.AssertEmulatedExternalRewardOf(uint64(0), incentiveId, expected)
}

func TestCanonicalExternalReward_3(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	currentHeight := int64(canonical.CurrentHeight())
	incentiveId := canonical.CreateExternalIncentive(gnousdc, gnsPath, 100000000, currentHeight+1, currentHeight+5, currentHeight+1, currentHeight+5, std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"))

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgg"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}
	canonical.NextBlock()

	// eligible

	canonical.NextBlock()

	// eligible

	canonical.NextBlock()

	// not eligible

	canonical.MoveTick(gnousdc, 400)

	// not eligible

	canonical.NextBlock()

	// incentive has been ended

	expected := uint64(100000000/4*2) * 30 / 100

	canonical.AssertEmulatedExternalRewardOf(uint64(0), incentiveId, expected)

	if canonical.UnclaimableExternalRewardOf(uint64(0), incentiveId) != uint64(100000000/4*2) {
		t.Errorf("UnclaimableExternalRewardOf(uint64(0), incentiveId) = %d; want %d", canonical.UnclaimableExternalRewardOf(uint64(0), incentiveId), uint64(100000000/4*2))
	}
}

func TestCanonicalMultiplePool_1(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	gnousdc2 := GetPoolPath(wugnotPath, gnsPath, 10000)
	canonical.CreatePool(gnousdc, 1, 200)
	canonical.CreatePool(gnousdc2, 1, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		1,
		gnousdc2,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		200,
		400,
		u256.NewUint(2000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	canonical.NextBlock()

	expected := canonical.PerBlockEmission * 30 / 100

	canonical.AssertEmulatedRewardOf(0, expected/2)
	canonical.AssertEmulatedRewardOf(1, expected/2)

	canonical.MoveTick(gnousdc, 100)
	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected/2)
	canonical.AssertEmulatedRewardOf(1, expected/2)

	canonical.MoveTick(gnousdc2, 100)
	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected/2)
	canonical.AssertEmulatedRewardOf(1, 0)

	canonical.MoveTick(gnousdc, 100)
	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected/2)
	canonical.AssertEmulatedRewardOf(1, 0)

	canonical.MoveTick(gnousdc, 300)
	canonical.MoveTick(gnousdc2, 300)
	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, 0)
	canonical.AssertEmulatedRewardOf(1, expected/2)
}

func TestCanonicalMultiplePool_2(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	gnousdc2 := GetPoolPath(wugnotPath, gnsPath, 10000)
	gnousdc3 := GetPoolPath(wugnotPath, gnsPath, 30000)
	canonical.CreatePool(gnousdc, 1, 200)
	canonical.CreatePool(gnousdc2, 2, 200)
	canonical.CreatePool(gnousdc3, 3, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		1,
		gnousdc2,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		200,
		400,
		u256.NewUint(2000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		2,
		gnousdc3,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgz"),
		300,
		500,
		u256.NewUint(3000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	expected := canonical.PerBlockEmission * 30 / 100

	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected*50/100)
	canonical.AssertEmulatedRewardOf(1, expected*30/100)
	canonical.AssertEmulatedRewardOf(2, 0)
	canonical.MoveTick(gnousdc, 100)
	canonical.MoveTick(gnousdc2, 100)
	canonical.MoveTick(gnousdc3, 100)

	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected*50/100)
	canonical.AssertEmulatedRewardOf(1, 0)
	canonical.AssertEmulatedRewardOf(2, 0)
	canonical.MoveTick(gnousdc, 400)
	canonical.MoveTick(gnousdc2, 400)
	canonical.MoveTick(gnousdc3, 400)

	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, 0)
	canonical.AssertEmulatedRewardOf(1, 0)
	canonical.AssertEmulatedRewardOf(2, expected*20/100)
	canonical.ChangePoolTier(gnousdc2, 1)
	canonical.MoveTick(gnousdc, 200)
	canonical.MoveTick(gnousdc2, 300)
	canonical.MoveTick(gnousdc3, 400)

	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected*40/100)
	canonical.AssertEmulatedRewardOf(1, expected*40/100)
	canonical.AssertEmulatedRewardOf(2, expected*20/100)

}

// Large number of blocks passed
func TestCanonicalLargeBlocksPassed(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 150)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	for i := 0; i < 10000; i++ {
		canonical.NextBlockNoCanonical()
	}

	expected := canonical.PerBlockEmission * 30 / 100
	canonical.AssertEmulatedRewardOf(0, expected*10000)
}

func GetPoolPath(token0Path, token1Path string, fee uint32) string {
	if strings.Compare(token1Path, token0Path) < 0 {
		token0Path, token1Path = token1Path, token0Path
	}
	return ufmt.Sprintf("%s:%s:%d", token0Path, token1Path, fee)
}

func TestCanonicalWarmup_1(t *testing.T) {
	modifyWarmup(0, 10)
	modifyWarmup(1, 10)
	modifyWarmup(2, 10)

	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 150)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	expected := canonical.PerBlockEmission * 30 / 100

	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected*10)

	expected = canonical.PerBlockEmission * 50 / 100

	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected*10)

	expected = canonical.PerBlockEmission * 70 / 100

	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected*10)

	expected = canonical.PerBlockEmission * 100 / 100

	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.AssertEmulatedRewardOf(0, expected*10)

}

func TestCanonicalWarmup_2(t *testing.T) {
	modifyWarmup(0, 10)
	modifyWarmup(1, 10)
	modifyWarmup(2, 10)

	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 150)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		200,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	expected0 := canonical.PerBlockEmission * 30 / 100

	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()

	expected1 := canonical.PerBlockEmission * 50 / 100

	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()

	expected2 := canonical.PerBlockEmission * 70 / 100

	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()

	expected3 := canonical.PerBlockEmission * 100 / 100

	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()
	canonical.NextBlock()

	canonical.AssertEmulatedRewardOf(0, expected0*10+expected1*10+expected2*10+expected3*10)

}

// ================
// randomized

// Test tick crossing at tick boundaries, random direction
func TestCanonicalTickCross_7(t *testing.T) {
	t.Skip("gets affect by previous")
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	expected := canonical.PerBlockEmission * 30 / 100

	tcs := []struct {
		tick     int32
		expected uint64
	}{
		{99, 0},
		{100, expected},
		{101, expected},
		{299, expected},
		{300, 0},
		{301, 0},
	}

	index := 0
	for i := 0; i < 100; i++ {
		index = (index + 10007) % len(tcs)
		canonical.MoveTick(gnousdc, tcs[index].tick)
		canonical.NextBlockNoCanonical()
		canonical.AssertEmulatedRewardOf(0, tcs[index].expected)
	}
}

// Test tick crossing with multiple positions with different tick, different liquidity
// Equivalence test
func TestCanonicalTickCross_8(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		1,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		200,
		400,
		u256.NewUint(2000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		2,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgz"),
		300,
		500,
		u256.NewUint(3000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	tick := int32(0)

	for i := 0; i < 500; i++ {
		tick = (tick + 10007) % 700
		canonical.MoveTick(gnousdc, tick)
		canonical.NextBlock()
		canonical.AssertEquivalence(0)
		canonical.AssertEquivalence(1)
		canonical.AssertEquivalence(2)
	}
}

// Test tick crossing with multiple positions with different tick, different liquidity, emulated reward flushed for every 100 blocks
// Equivalence test
func TestCanonicalTickCross_9(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		1,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		200,
		400,
		u256.NewUint(2000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		2,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgz"),
		300,
		500,
		u256.NewUint(3000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	tick := int32(0)
	for i := 0; i < 20; i++ {
		canonicalRewardMap := make(map[uint64]uint64)
		for j := 0; j < 20; j++ {
			tick = (tick + 10007) % 700
			canonical.MoveTick(gnousdc, tick)
			canonical.NextBlock()
			canonicalRewardMap[0] += canonical.CanonicalRewardOf(0).Internal
			canonicalRewardMap[1] += canonical.CanonicalRewardOf(1).Internal
			canonicalRewardMap[2] += canonical.CanonicalRewardOf(2).Internal
		}
		canonical.AssertEmulatedRewardMap(canonicalRewardMap)
	}
}

// Test tick crossing with multiple positions with different tick, different liquidity, emulated reward flushed for every 100 blocks
// Equivalence test
func TestCanonicalTickCross_10(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, -200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		-300,
		-100,
		u256.NewUint(1000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		1,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgq"),
		-400,
		-200,
		u256.NewUint(2000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	err = canonical.StakeToken(
		2,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgz"),
		-500,
		-300,
		u256.NewUint(3000000000000000000),
	)

	if err != nil {
		t.Errorf("StakeToken failed: %s", err.Error())
	}

	tick := int32(0)
	for i := 0; i < 20; i++ {
		canonicalRewardMap := make(map[uint64]uint64)
		for j := 0; j < 20; j++ {
			tick = (tick + 10007) % 700
			canonical.MoveTick(gnousdc, -tick)
			canonical.NextBlock()
			canonicalRewardMap[0] += canonical.CanonicalRewardOf(0).Internal
			canonicalRewardMap[1] += canonical.CanonicalRewardOf(1).Internal
			canonicalRewardMap[2] += canonical.CanonicalRewardOf(2).Internal
		}
		canonical.AssertEmulatedRewardMap(canonicalRewardMap)
	}
}

func TestCanonicalEmissionChange_0(t *testing.T) {
	canonical := Setup(t)

	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	err := canonical.StakeToken(
		0,
		gnousdc,
		std.Address("gno1qyqszqgpqyqszqgpqyqszqgpqyqszqgp"),
		100,
		300,
		u256.NewUint(1000000000000000000),
	)

	canonical.NextBlock()

	expected0 := canonical.PerBlockEmission * 30 / 100

	t.Logf("expected0: %d", expected0)

	canonical.SetEmissionUpdate(canonical.PerBlockEmission * 10 / 100)

	canonical.NextBlock()

	expected1 := canonical.PerBlockEmission * 30 / 100

	t.Logf("expected1: %d", expected1)

	canonical.AssertEmulatedRewardOf(0, expected0+expected1)
}

type TestEventCreatePool struct {
	poolPath    string
	initialTier uint64
	initialTick int32
}

func (event *TestEventCreatePool) Apply(canonical *canonicalRewardState) {
	canonical.CreatePool(event.poolPath, event.initialTier, event.initialTick)
}

func (event *TestEventCreatePool) IsValid(canonical *canonicalRewardState) bool {
	_, ok := canonical.global.pools.Get(event.poolPath)
	return !ok
}

func (event *TestEventCreatePool) String() string {
	return ufmt.Sprintf("CreatePool(%s, %d, %d)", event.poolPath, event.initialTier, event.initialTick)
}

type TestEventChangeTier struct {
	poolPath   string
	targetTier uint64
}

func (event *TestEventChangeTier) IsValid(canonical *canonicalRewardState) bool {
	_, ok := canonical.global.pools.Get(event.poolPath)
	return ok
}

func (event *TestEventChangeTier) Apply(canonical *canonicalRewardState) {
	canonical.ChangePoolTier(event.poolPath, event.targetTier)
}

func (event *TestEventChangeTier) String() string {
	return ufmt.Sprintf("ChangeTier(%s, %d)", event.poolPath, event.targetTier)
}

type TestEventStakeToken struct {
	tokenId   uint64
	poolPath  string
	address   std.Address
	liquidity *u256.Uint
	tickLower int32
	tickUpper int32
}

func (event *TestEventStakeToken) Apply(canonical *canonicalRewardState) {
	if canonical.global.deposits.Has(event.tokenId) {
		if canonical.global.deposits.Get(event.tokenId).liquidity.IsZero() {
			canonical.StakeToken(event.tokenId, event.poolPath, event.address, event.tickLower, event.tickUpper, event.liquidity)
		} else {
			canonical.UnstakeToken(event.tokenId)
		}
	} else {
		canonical.StakeToken(event.tokenId, event.poolPath, event.address, event.tickLower, event.tickUpper, event.liquidity)
	}
}

func (event *TestEventStakeToken) IsValid(canonical *canonicalRewardState) bool {
	_, ok := canonical.global.pools.Get(event.poolPath)
	return ok
}

func (event *TestEventStakeToken) String() string {
	return ufmt.Sprintf("StakeToken(%d, %s, %s, %d, %d, %s)", event.tokenId, event.poolPath, event.address, event.tickLower, event.tickUpper, event.liquidity.ToString())
}

type TestEventMoveTick struct {
	poolPath string
	tick     int32
}

func (event *TestEventMoveTick) Apply(canonical *canonicalRewardState) {
	canonical.MoveTick(event.poolPath, event.tick)
}

func (event *TestEventMoveTick) IsValid(canonical *canonicalRewardState) bool {
	_, ok := canonical.global.pools.Get(event.poolPath)
	return ok
}

func (event *TestEventMoveTick) String() string {
	return ufmt.Sprintf("MoveTick(%s, %d)", event.poolPath, event.tick)
}

type TestEventSetEmissionUpdate struct {
	emission uint64
}

func (event *TestEventSetEmissionUpdate) Apply(canonical *canonicalRewardState) {
	canonical.SetEmissionUpdate(event.emission)
}

func (event *TestEventSetEmissionUpdate) IsValid(canonical *canonicalRewardState) bool {
	return true
}

func (event *TestEventSetEmissionUpdate) String() string {
	return ufmt.Sprintf("SetEmissionUpdate(%d)", event.emission)
}

type SimulationEvent interface {
	IsValid(canonical *canonicalRewardState) bool
	Apply(canonical *canonicalRewardState)
	String() string
}

type tokenStake struct {
	tokenId   uint64
	address   std.Address
	liquidity *u256.Uint
	tickLower int32
	tickUpper int32
}

func events(
	poolPath string,
	initialTier uint64,
	initialTick int32,
	tokenStakes []tokenStake,
	moveTicks []int32,
) []SimulationEvent {
	result := []SimulationEvent{
		&TestEventCreatePool{
			poolPath:    poolPath,
			initialTier: initialTier,
			initialTick: initialTick,
		},
	}
	for i := 0; i < len(tokenStakes); i++ {
		result = append(result, &TestEventStakeToken{
			tokenId:   tokenStakes[i].tokenId,
			poolPath:  poolPath,
			address:   tokenStakes[i].address,
			liquidity: tokenStakes[i].liquidity,
			tickLower: tokenStakes[i].tickLower,
			tickUpper: tokenStakes[i].tickUpper,
		})
	}
	for i := 0; i < len(moveTicks); i++ {
		result = append(result, &TestEventMoveTick{
			poolPath: poolPath,
			tick:     moveTicks[i],
		})
	}
	if poolPath == "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000" {
		return result
	}
	for i := 0; i < 4; i++ {
		result = append(result, &TestEventChangeTier{
			poolPath:   poolPath,
			targetTier: uint64(i),
		})
	}
	return result
}
func TestCanonicalSimulation_0(t *testing.T) {
	t.Skip("gets affect by previous")
	wugnotgns := events(
		"gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000",
		1,
		200,
		[]tokenStake{
			{0, std.Address("gno1token0"), u256.NewUint(1000000000000000000), 100, 300},
			{1, std.Address("gno1token1"), u256.NewUint(2000000000000000000), 200, 400},
			{2, std.Address("gno1token2"), u256.NewUint(3000000000000000000), 300, 500},
		},
		[]int32{50, 100, 150, 200, 250, 300, 350, 400, 450, 500, 550, 600, 650, 700, 750},
	)

	barbaz := events(
		"gno.land/r/demo/bar:gno.land/r/gnoswap/v1/baz:3000",
		0,
		200,
		[]tokenStake{
			{3, std.Address("gno1token3"), u256.NewUint(4000000000000000000), 50, 100},
			{4, std.Address("gno1token4"), u256.NewUint(5000000000000000000), 100, 200},
			{5, std.Address("gno1token5"), u256.NewUint(6000000000000000000), 300, 500},
		},
		[]int32{50, 100, 150, 200, 250, 300, 350, 400, 450, 500, 550, 600, 650, 700, 750},
	)

	bazquux := events(
		"gno.land/r/demo/baz:gno.land/r/gnoswap/v1/quux:3000",
		0,
		200,
		[]tokenStake{
			{6, std.Address("gno1token6"), u256.NewUint(7000000000000000000), 400, 600},
			{7, std.Address("gno1token7"), u256.NewUint(8000000000000000000), 600, 800},
			{8, std.Address("gno1token8"), u256.NewUint(9000000000000000000), 700, 900},
		},
		[]int32{50, 100, 150, 200, 250, 300, 350, 400, 450, 500, 550, 600, 650, 700, 750},
	)

	quuxgns := events(
		"gno.land/r/demo/quux:gno.land/r/gnoswap/v1/gns:3000",
		0,
		200,
		[]tokenStake{
			{9, std.Address("gno1token9"), u256.NewUint(1000000000000000000), 50, 100},
			{10, std.Address("gno1token10"), u256.NewUint(2000000000000000000), 100, 200},
			{11, std.Address("gno1token11"), u256.NewUint(3000000000000000000), 200, 300},
		},
		[]int32{50, 100, 150, 200, 250, 300, 350, 400, 450, 500, 550, 600, 650, 700, 750},
	)

	events := []SimulationEvent{}
	events = append(events, wugnotgns...)
	events = append(events, barbaz...)
	events = append(events, bazquux...)
	events = append(events, quuxgns...)

	canonical := Setup(t)

	// required to match with the default tier 1
	gnousdc := GetPoolPath(wugnotPath, gnsPath, 3000)
	canonical.CreatePool(gnousdc, 1, 200)

	eventId := 0

	for i := 0; i < 10; i++ {
		println("=======ITERATION : ", i, "========")

		canonicalRewardMap := make(map[uint64]uint64)
		for j := 0; j < 100; j++ {
			var event SimulationEvent
			for {
				eventId = (eventId + 17) % len(events)
				event = events[eventId]
				if event.IsValid(canonical) {
					break
				}
			}
			println("=======HEIGHT : ", std.GetHeight(), "========")
			println("	", eventId, " : ", event.String())
			event.Apply(canonical)
			canonical.NextBlock()
			for i := 0; i < 12; i++ {
				reward, ok := canonical.SafeCanonicalRewardOf(uint64(i))
				if !ok {
					continue // no reward for this deposit
				}
				canonicalRewardMap[uint64(i)] += reward.Internal
			}
		}

		println("======ASSERTING========")
		for k, v := range canonicalRewardMap {
			println("	", k, " : ", v)
		}
		canonical.AssertEmulatedRewardMap(canonicalRewardMap)
	}
}
