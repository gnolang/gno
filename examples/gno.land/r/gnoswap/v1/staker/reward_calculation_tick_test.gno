package staker

import (
	"strconv"
	"testing"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/demo/uassert"
)

func TestEncodeInt(t *testing.T) {
	tests := []struct {
		input    int32
		expected string
	}{
		{123, "0000000123"},
		{-123, "-0000000123"},
		{0, "0000000000"},
		{2147483647, "2147483647"},   // int32 max
		{-2147483648, "-2147483648"}, // int32 min
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(int(tt.input)), func(t *testing.T) {
			uassert.Equal(t, EncodeInt(tt.input), tt.expected)
		})
	}
}

func TestTicks(t *testing.T) {
	ticks := NewTicks()

	tick := ticks.Get(100)
	if tick == nil || tick.id != 100 {
		t.Errorf("Get(100) returned %v; want Tick with ID 100", tick)
	}

	tick.stakedLiquidityGross = u256.MustFromDecimal("1")
	ticks.Set(100, tick)
	uassert.True(t, ticks.Has(100))

	tick.stakedLiquidityGross = u256.Zero()
	ticks.Set(100, tick)
	uassert.False(t, ticks.Has(100))
}

func TestTicksBasic(t *testing.T) {
	ticks := NewTicks()

	tick100 := ticks.Get(100)
	uassert.True(t, ticks.Has(100))
	uassert.Equal(t, tick100.id, int32(100))

	tick100Again := ticks.Get(100)
	uassert.Equal(t, int32(tick100Again.id), int32(tick100.id))
	uassert.True(t, tick100Again.stakedLiquidityGross.IsZero())
	uassert.True(t, tick100Again.stakedLiquidityDelta.IsZero())

	ticks.Set(100, tick100)
	uassert.False(t, ticks.Has(100))
}

func TestModifyDepositLower(t *testing.T) {
	ticks := NewTicks()
	tick := ticks.Get(100)

	// initial value must be zero
	uassert.True(t, tick.stakedLiquidityGross.IsZero())
	uassert.True(t, tick.stakedLiquidityDelta.IsZero())

	// deposit +10
	liquidityDelta := i256.NewInt(10) // +10
	tick.modifyDepositLower(50, 95, liquidityDelta)

	// stakedLiquidityGross += +10 => 10
	// stakedLiquidityDelta += +10 => 10
	if tick.stakedLiquidityGross.ToString() != "10" || tick.stakedLiquidityDelta.ToString() != "10" {
		t.Errorf("After deposit +10, stakedLiquidityGross=%v, stakedLiquidityDelta=%v; want 10,10",
			tick.stakedLiquidityGross, tick.stakedLiquidityDelta)
	}

	// deposit another +5
	tick.modifyDepositLower(60, 95, i256.NewInt(5))
	// gross=15, delta=15
	if tick.stakedLiquidityGross.ToString() != "15" || tick.stakedLiquidityDelta.ToString() != "15" {
		t.Errorf("After deposit +5, stakedLiquidityGross=%v, stakedLiquidityDelta=%v; want 15,15",
			tick.stakedLiquidityGross, tick.stakedLiquidityDelta)
	}
}

func TestModifyDepositUpper(t *testing.T) {
	ticks := NewTicks()
	tick := ticks.Get(200)

	// deposit +10 => modifyDepositUpper
	tick.modifyDepositUpper(70, 195, i256.NewInt(10))

	// stakedLiquidityGross=10, stakedLiquidityDelta = -10
	// 	upper => delta = stakedLiquidityDelta - liquidity
	if tick.stakedLiquidityGross.ToString() != "10" || tick.stakedLiquidityDelta.ToString() != "-10" {
		t.Errorf("After deposit +10(upper), stakedLiquidityGross=%v, stakedLiquidityDelta=%v; want 10,-10",
			tick.stakedLiquidityGross, tick.stakedLiquidityDelta)
	}
}

func compareRangeSlices(t *testing.T, a, b [][2]uint64) bool {
	t.Helper()
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareInt64Slices(t *testing.T, a, b []int64) bool {
	t.Helper()
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
