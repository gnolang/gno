package staker

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	pl "gno.land/r/gnoswap/v1/pool"
)

// EncodeInt takes an int32 and returns a zero-padded decimal string
// with up to 10 digits for the absolute value.
// If the number is negative, the '-' sign comes first, followed by zeros, then digits.
func EncodeInt(num int32) string {
	// Convert the absolute value to a decimal string.
	absValue := int64(num)
	isNegative := false
	if num < 0 {
		isNegative = true
		absValue = -absValue // Safely negate into int64 to avoid overflow.
	}

	s := strconv.FormatInt(absValue, 10)

	// Zero-pad to a total of 10 digits for the absolute value.
	// (The '-' sign will be added later if needed.)
	zerosNeeded := 10 - len(s)
	if zerosNeeded < 0 {
		zerosNeeded = 0
	}

	padded := strings.Repeat("0", zerosNeeded) + s

	// If the original number was negative, prepend '-'.
	if isNegative {
		return "-" + padded
	}
	return padded
}

// Tick mapping for each pool
type Ticks struct {
	tree *avl.Tree // int32 tickId -> tick
}

func NewTicks() Ticks {
	return Ticks{
		tree: avl.NewTree(),
	}
}

func (self *Ticks) Get(tickId int32) *Tick {
	v, ok := self.tree.Get(EncodeInt(tickId))
	if !ok {
		tick := &Tick{
			id:                   tickId,
			stakedLiquidityGross: u256.Zero(),
			stakedLiquidityDelta: i256.Zero(),
			outsideAccumulation:  NewUintTree(),
		}
		self.tree.Set(EncodeInt(tickId), tick)
		return tick
	}
	return v.(*Tick)
}

func (self *Ticks) set(tickId int32, tick *Tick) {
	if tick.stakedLiquidityGross.IsZero() {
		self.tree.Remove(EncodeInt(tickId))
		return
	}
	self.tree.Set(EncodeInt(tickId), tick)
}

func (self *Ticks) Has(tickId int32) bool {
	return self.tree.Has(EncodeInt(tickId))
}

// Tick represents the state of a specific tick in a pool.
//
// Fields:
// - id (int32): The ID of the tick.
// - stakedLiquidityGross (*u256.Uint): Total gross staked liquidity at this tick.
// - stakedLiquidityDelta (*i256.Int): Net change in staked liquidity at this tick.
// - outsideAccumulation (*UintTree): RewardRatioAccumulation outside the tick.
type Tick struct {
	id int32

	// conceptually equal with Pool.liquidityGross but only for the staked positions
	stakedLiquidityGross *u256.Uint

	// conceptually equal with Pool.liquidityNet but only for the staked positions
	stakedLiquidityDelta *i256.Int

	// currentOutsideAccumulation is the accumulation of the blockNumber / TotalStake outside the tick.
	// It is calculated by subtracting the current tick's currentOutsideAccumulation from the global reward ratio accumulation.
	outsideAccumulation *UintTree // blockNumber -> *u256.Uint
}

// CurrentOutsideAccumulation returns the latest outside accumulation for the tick
func (self *Tick) CurrentOutsideAccumulation(blockNumber int64) *u256.Uint {
	var acc *u256.Uint
	self.outsideAccumulation.ReverseIterate(0, blockNumber, func(key int64, value interface{}) bool {
		acc = value.(*u256.Uint)
		return true
	})
	if acc == nil {
		acc = u256.Zero()
	}
	return acc
}

// modifyDepositLower updates the tick's liquidity info by treating the deposit as a lower tick
func (self *Tick) modifyDepositLower(currentHeight int64, currentTick int32, liquidity *i256.Int) {
	// update staker side tick info
	self.stakedLiquidityGross = liquidityMathAddDelta(self.stakedLiquidityGross, liquidity)
	if self.stakedLiquidityGross.Lt(u256.Zero()) {
		panic("stakedLiquidityGross is negative")
	}
	self.stakedLiquidityDelta = i256.Zero().Add(self.stakedLiquidityDelta, liquidity)
}

// modifyDepositUpper updates the tick's liquidity info by treating the deposit as an upper tick
func (self *Tick) modifyDepositUpper(currentHeight int64, currentTick int32, liquidity *i256.Int) {
	self.stakedLiquidityGross = liquidityMathAddDelta(self.stakedLiquidityGross, liquidity)
	if self.stakedLiquidityGross.Lt(u256.Zero()) {
		panic("stakedLiquidityGross is negative")
	}
	self.stakedLiquidityDelta = i256.Zero().Sub(self.stakedLiquidityDelta, liquidity)
}

// updateCurrentOutsideAccumulation updates the tick's outside accumulation
// It "flips" the accumulation's inside/outside by subtracting the current outside accumulation from the global accumulation
func (self *Tick) updateCurrentOutsideAccumulation(blockNumber int64, acc *u256.Uint) {
	currentOutsideAccumulation := self.CurrentOutsideAccumulation(blockNumber)
	newOutsideAccumulation := u256.Zero().Sub(acc, currentOutsideAccumulation)
	self.outsideAccumulation.set(blockNumber, newOutsideAccumulation)
}

// TickCrossHook is a hook that is called when a tick is crossed.
// Modifies the tick's outside accumulation and updates the tick's liquidity info
func TickCrossHook(pools *Pools, height func() int64) func(poolPath string, tickId int32, zeroForOne bool) {
	return func(poolPath string, tickId int32, zeroForOne bool) {
		pool, ok := pools.Get(poolPath)
		if !ok {
			return
		}

		tick := pool.ticks.Get(tickId)

		blockNumber := height()

		var nextTick int32
		if zeroForOne {
			nextTick = tickId - 1
		} else {
			nextTick = tickId
		}

		liquidityInRangeDelta := tick.stakedLiquidityDelta
		if liquidityInRangeDelta.Sign() == 0 {
			return
		}

		if zeroForOne {
			liquidityInRangeDelta = i256.Zero().Neg(liquidityInRangeDelta)
		}
		newAcc := pool.modifyDeposit(liquidityInRangeDelta, blockNumber, nextTick)
		tick.updateCurrentOutsideAccumulation(blockNumber, newAcc)
	}
}

func init() {
	// Sets tick cross hook for pool contract
	pl.SetTickCrossHook(TickCrossHook(pools, std.ChainHeight))
}
