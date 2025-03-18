package staker

import (
	"std"

	"gno.land/p/demo/avl"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
)

var (
	// Q128 is 2^128
	q128 = u256.MustFromDecimal(consts.Q128)
	// Q192 is 2^192
	q192 = u256.MustFromDecimal("6277101735386680763835789423207666416102355444464034512895")

	// pools is the global pool storage
	pools *Pools
)

func init() {
	pools = NewPools()
}

// Pools represents the global pool storage
type Pools struct {
	tree *avl.Tree // string poolPath -> pool
}

func NewPools() *Pools {
	return &Pools{
		tree: avl.NewTree(),
	}
}

// Get returns the pool for the given poolPath
func (self *Pools) Get(poolPath string) (*Pool, bool) {
	v, ok := self.tree.Get(poolPath)
	if !ok {
		return nil, false
	}
	return v.(*Pool), true
}

// GetOrCreate returns the pool for the given poolPath, or creates a new pool if it does not exist
func (self *Pools) GetOrCreate(poolPath string) *Pool {
	pool, ok := self.Get(poolPath)
	if !ok {
		pool = NewPool(poolPath, std.ChainHeight())
		self.set(poolPath, pool)
	}
	return pool
}

// set sets the pool for the given poolPath
func (self *Pools) set(poolPath string, pool *Pool) {
	self.tree.Set(poolPath, pool)
}

// Has returns true if the pool exists for the given poolPath
func (self *Pools) Has(poolPath string) bool {
	return self.tree.Has(poolPath)
}

func (self *Pools) IterateAll(fn func(key string, pool *Pool) bool) {
	self.tree.Iterate("", "", func(key string, value interface{}) bool {
		return fn(key, value.(*Pool))
	})
}

// Pool is a struct for storing an incentivized pool information
// Each pool stores Incentives and Ticks associated with it.
//
// Fields:
// - poolPath: The path of the pool.
//
//   - currentStakedLiquidity:
//     The current total staked liquidity of the in-range positions for the pool.
//     Updated when tick cross happens or stake/unstake happens.
//     Used to calculate the global reward ratio accumulation or
//     decide whether to enter/exit unclaimable period.
//
//   - lastUnclaimableHeight:
//     The height at which the unclaimable period started.
//     Set to 0 when the pool is not in an unclaimable period.
//
//   - unclaimableAcc:
//     The accumulated undisributed unclaimable reward.
//     Reset to 0 when processUnclaimableReward is called and sent to community pool.
//
//   - rewardCache:
//     The cached per-block reward emitted for this pool.
//     Stores new entry only when the reward is changed.
//     PoolTier.cacheReward() updates this.
//
// - incentives: The external incentives associated with the pool.
//
// - ticks: The Ticks associated with the pool.
//
//   - globalRewardRatioAccumulation:
//     Global ratio of BlockNumber / TotalStake accumulation(since the pool creation)
//     Stores new entry only when tick cross or stake/unstake happens.
//     It is used to calculate the reward for a staked position at certain height.
//
//   - historicalTick:
//     The historical tick for the pool at a given height.
//     It does not reflect the exact tick at the blockNumber,
//     but it provides correct ordering for the staked position's ticks.
//     Therefore, you should not compare it for equality, only for ordering.
//     Set when tick cross happens or a new position is created.
type Pool struct {
	poolPath string

	stakedLiquidity *UintTree // uint64 blockNumber -> *u256.Uint(Q128)

	lastUnclaimableHeight int64
	unclaimableAcc        uint64

	rewardCache *UintTree // uint64 blockNumber -> uint64 gnsReward

	incentives Incentives

	ticks Ticks // int32 tickId -> Tick tick

	globalRewardRatioAccumulation *UintTree // uint64 blockNumber -> *u256.Uint(Q128) rewardRatioAccumulation

	historicalTick *UintTree // uint64 blockNumber -> int32 tickId
}

// NewPool creates a new pool with the given poolPath and currentHeight.
func NewPool(poolPath string, currentHeight int64) *Pool {
	pool := &Pool{
		poolPath:                      poolPath,
		stakedLiquidity:               NewUintTree(),
		lastUnclaimableHeight:         currentHeight,
		unclaimableAcc:                0,
		rewardCache:                   NewUintTree(),
		incentives:                    NewIncentives(),
		ticks:                         NewTicks(),
		globalRewardRatioAccumulation: NewUintTree(),
		historicalTick:                NewUintTree(),
	}

	pool.globalRewardRatioAccumulation.set(currentHeight, u256.Zero())
	pool.rewardCache.set(currentHeight, uint64(0))
	pool.stakedLiquidity.set(currentHeight, u256.Zero())

	return pool
}

// Get the latest global reward ratio accumulation in [0, currentHeight] range.
// Returns the height and the accumulation.
func (self *Pool) CurrentGlobalRewardRatioAccumulation(currentHeight int64) (int64, *u256.Uint) {
	var height int64
	var acc *u256.Uint
	self.globalRewardRatioAccumulation.ReverseIterate(0, currentHeight, func(key int64, value interface{}) bool {
		height = key
		acc = value.(*u256.Uint)
		return true
	})
	if acc == nil {
		panic("should not happen, globalRewardRatioAccumulation must be set when pool is created")
	}
	return height, acc
}

// Get the latest tick in [0, currentHeight] range.
// Returns the tick.
func (self *Pool) CurrentTick(currentHeight int64) int32 {
	var tick int32
	self.historicalTick.ReverseIterate(0, currentHeight, func(key int64, value interface{}) bool {
		tick = value.(int32)
		return true
	})
	return tick
}

func (self *Pool) CurrentStakedLiquidity(currentHeight int64) *u256.Uint {
	var liquidity *u256.Uint
	self.stakedLiquidity.ReverseIterate(0, currentHeight, func(key int64, value interface{}) bool {
		liquidity = value.(*u256.Uint)
		return true
	})
	return liquidity
}

// IsExternallyIncentivizedPool returns true if the pool has any external incentives.
func (self *Pool) IsExternallyIncentivizedPool() bool {
	return self.incentives.byTime.Size() != 0
}

// Get the latest reward in [0, currentHeight] range.
// Returns the reward.
func (self *Pool) CurrentReward(currentHeight int64) uint64 {
	var reward uint64
	self.rewardCache.ReverseIterate(0, currentHeight, func(key int64, value interface{}) bool {
		reward = value.(uint64)
		return true
	})
	return reward
}

// cacheReward sets the current reward for the pool
// If the pool is in unclaimable period, it will end the unclaimable period,  updates the reward, and start the unclaimable period again.
func (self *Pool) cacheReward(currentHeight int64, currentTierReward uint64) {
	oldTierReward := self.CurrentReward(currentHeight)
	if oldTierReward == currentTierReward {
		return
	}

	isInUnclaimable := self.CurrentStakedLiquidity(currentHeight).IsZero()
	if isInUnclaimable {
		self.endUnclaimablePeriod(currentHeight)
	}

	self.rewardCache.set(currentHeight, currentTierReward)

	if isInUnclaimable {
		self.startUnclaimablePeriod(currentHeight)
	}
}

// cacheInternalReward caches the current emission and updates the global reward ratio accumulation.
func (self *Pool) cacheInternalReward(currentHeight int64, currentEmission uint64) {
	self.cacheReward(currentHeight, currentEmission)

	currentStakedLiquidity := self.CurrentStakedLiquidity(currentHeight)
	if currentStakedLiquidity.IsZero() {
		self.endUnclaimablePeriod(currentHeight)
		self.startUnclaimablePeriod(currentHeight)
	}

	self.updateGlobalRewardRatioAccumulation(currentHeight, currentStakedLiquidity)
}

func (self *Pool) calculateGlobalRewardRatioAccumulation(currentHeight int64, currentStakedLiquidity *u256.Uint) *u256.Uint {
	oldAccHeight, oldAcc := self.CurrentGlobalRewardRatioAccumulation(currentHeight)
	blockDiff := currentHeight - oldAccHeight
	if blockDiff == 0 {
		return oldAcc.Clone()
	}

	acc := u256.NewUint(uint64(blockDiff))
	acc = acc.Mul(acc, q128)
	acc = acc.Div(acc, currentStakedLiquidity)

	return u256.Zero().Add(oldAcc, acc)
}

// updateGlobalRewardRatioAccumulation updates the global reward ratio accumulation and returns the new accumulation.
func (self *Pool) updateGlobalRewardRatioAccumulation(currentHeight int64, currentStakedLiquidity *u256.Uint) *u256.Uint {
	newAcc := self.calculateGlobalRewardRatioAccumulation(currentHeight, currentStakedLiquidity)

	self.globalRewardRatioAccumulation.set(currentHeight, newAcc)
	return newAcc
}

// RewardState is a struct for storing the intermediate state for reward calculation.
type RewardState struct {
	pool    *Pool
	deposit *Deposit

	// accumulated rewards for each warmup
	rewards   []uint64
	penalties []uint64
}

// RewardStateOf initializes a new RewardState for the given deposit.
func (self *Pool) RewardStateOf(deposit *Deposit) *RewardState {
	result := &RewardState{
		pool:      self,
		deposit:   deposit,
		rewards:   make([]uint64, len(deposit.warmups)),
		penalties: make([]uint64, len(deposit.warmups)),
	}

	for i := range result.rewards {
		result.rewards[i] = 0
		result.penalties[i] = 0
	}

	return result
}

// calculateInternalReward calculates the internal reward for the deposit.
// It calls rewardPerWarmup for each rewardCache interval, applies warmup, and returns the rewards and penalties.
func (self *RewardState) calculateInternalReward(startHeight, endHeight int64) ([]uint64, []uint64) {
	currentReward := self.pool.CurrentReward(startHeight)
	self.pool.rewardCache.Iterate(startHeight, endHeight, func(key int64, value interface{}) bool {
		// we calculate per-position reward
		self.rewardPerWarmup(startHeight, int64(key), currentReward)
		currentReward = value.(uint64)
		startHeight = int64(key)
		return false
	})

	if startHeight < endHeight {
		self.rewardPerWarmup(startHeight, endHeight, currentReward)
	}

	self.applyWarmup()

	return self.rewards, self.penalties
}

// calculateExternalReward calculates the external reward for the deposit.
// It calls rewardPerWarmup for startHeight to endHeight(clamped to the incentive period), applies warmup and returns the rewards and penalties.
func (self *RewardState) calculateExternalReward(startHeight, endHeight int64, incentive *ExternalIncentive) ([]uint64, []uint64) {
	if startHeight < int64(self.deposit.lastCollectHeight) {
		// This must not happen, but adding some guards just in case.
		startHeight = int64(self.deposit.lastCollectHeight)
	}

	if endHeight < incentive.startHeight {
		return nil, nil // Not started yet
	}

	if startHeight < incentive.startHeight {
		startHeight = incentive.startHeight
	}

	if endHeight > incentive.endHeight {
		endHeight = incentive.endHeight
	}

	if startHeight > incentive.endHeight {
		return nil, nil // Already ended
	}

	rewardPerBlock := incentive.rewardPerBlock

	self.rewardPerWarmup(startHeight, endHeight, rewardPerBlock)

	self.applyWarmup()

	return self.rewards, self.penalties
}

// applyWarmup applies the warmup to the rewards and calculate penalties.
func (self *RewardState) applyWarmup() {
	for i, warmup := range self.deposit.warmups {
		refactorReward := self.rewards[i]
		self.rewards[i] = refactorReward * warmup.WarmupRatio / 100
		self.penalties[i] = refactorReward - self.rewards[i]
	}
}

// rewardPerWarmup calculates the reward for each warmup, adds to the RewardState's rewards array.
func (self *RewardState) rewardPerWarmup(startHeight, endHeight int64, rewardPerBlock uint64) {
	for i, warmup := range self.deposit.warmups {
		if startHeight >= warmup.NextWarmupHeight {
			// passed the warmup
			continue
		}

		if endHeight < warmup.NextWarmupHeight {
			rewardAcc := self.pool.CalculateRewardForPosition(startHeight, self.pool.CurrentTick(startHeight), endHeight, self.pool.CurrentTick(endHeight), self.deposit)

			rewardAcc = rewardAcc.Mul(rewardAcc, self.deposit.liquidity)
			rewardAcc = rewardAcc.Mul(rewardAcc, u256.NewUint(rewardPerBlock))
			rewardAcc = rewardAcc.Div(rewardAcc, q128)
			self.rewards[i] += rewardAcc.Uint64()

			// done
			break
		}

		rewardAcc := self.pool.CalculateRewardForPosition(startHeight, self.pool.CurrentTick(startHeight), warmup.NextWarmupHeight, self.pool.CurrentTick(warmup.NextWarmupHeight), self.deposit)

		rewardAcc = rewardAcc.Mul(rewardAcc, self.deposit.liquidity)
		rewardAcc = rewardAcc.Mul(rewardAcc, u256.NewUint(rewardPerBlock))
		rewardAcc = rewardAcc.Div(rewardAcc, q128)
		self.rewards[i] += rewardAcc.Uint64()

		startHeight = warmup.NextWarmupHeight
	}
}

// modifyDeposit updates the pool's staked liquidity and returns the new staked liquidity.
// updates when there is a change in the staked liquidity(tick cross, stake, unstake)
func (self *Pool) modifyDeposit(delta *i256.Int, currentHeight int64, nextTick int32) *u256.Uint {
	// update staker side pool info
	lastStakedLiquidity := self.CurrentStakedLiquidity(currentHeight)
	deltaApplied := liquidityMathAddDelta(lastStakedLiquidity, delta)
	result := self.updateGlobalRewardRatioAccumulation(currentHeight, lastStakedLiquidity)

	// historical tick does NOT actually reflect the tick at the blockNumber, but it provides correct ordering for the staked positions
	// because TickCrossHook is assured to be called for the staked-initialized ticks
	self.historicalTick.set(currentHeight, nextTick)

	switch deltaApplied.Sign() {
	case -1:
		panic("stakedLiquidity is less than 0, should not happen")
	case 0:
		if lastStakedLiquidity.Sign() == 1 {
			// StakedLiquidity moved from positive to zero, start unclaimable period
			self.startUnclaimablePeriod(currentHeight)
			self.incentives.startUnclaimablePeriod(currentHeight)
		}
	case 1:
		if lastStakedLiquidity.Sign() == 0 {
			// StakedLiquidity moved from zero to positive, end unclaimable period
			self.endUnclaimablePeriod(currentHeight)
			self.incentives.endUnclaimablePeriod(currentHeight)
		}
	}

	self.stakedLiquidity.set(currentHeight, deltaApplied)

	return result
}

// startUnclaimablePeriod starts the unclaimable period.
func (self *Pool) startUnclaimablePeriod(currentHeight int64) {
	if self.lastUnclaimableHeight == 0 {
		// We set only if it's the first time entering(0 indicates not set yet)
		self.lastUnclaimableHeight = currentHeight
	}
}

// endUnclaimablePeriod ends the unclaimable period.
// Accumulates to unclaimableAcc and resets lastUnclaimableHeight to 0.
func (self *Pool) endUnclaimablePeriod(currentHeight int64) {
	if self.lastUnclaimableHeight == 0 {
		// This should not happen, but guarding just in case
		return
	}
	unclaimableHeights := currentHeight - self.lastUnclaimableHeight
	self.unclaimableAcc += uint64(unclaimableHeights) * self.CurrentReward(self.lastUnclaimableHeight)
	self.lastUnclaimableHeight = 0
}

// processUnclaimableReward processes the unclaimable reward and returns the accumulated reward.
// It resets unclaimableAcc to 0 and updates lastUnclaimableHeight to endHeight.
func (self *Pool) processUnclaimableReward(poolTier *PoolTier, endHeight int64) uint64 {
	internalUnClaimable := self.unclaimableAcc
	self.unclaimableAcc = 0
	self.lastUnclaimableHeight = endHeight
	return internalUnClaimable
}

// Calculates reward for a position *without* considering debt or warmup
// It calculates the theoretical total reward for the position if it has been staked since the pool creation
func (self *Pool) CalculateRawRewardForPosition(currentHeight int64, currentTick int32, deposit *Deposit) *u256.Uint {
	var rewardAcc *u256.Uint

	globalAcc := self.calculateGlobalRewardRatioAccumulation(currentHeight, self.CurrentStakedLiquidity(currentHeight))
	lowerAcc := self.ticks.Get(deposit.tickLower).CurrentOutsideAccumulation(currentHeight)
	upperAcc := self.ticks.Get(deposit.tickUpper).CurrentOutsideAccumulation(currentHeight)
	if currentTick < deposit.tickLower {
		rewardAcc = u256.Zero().Sub(lowerAcc, upperAcc)
	} else if currentTick >= deposit.tickUpper {
		rewardAcc = u256.Zero().Sub(upperAcc, lowerAcc)
	} else {
		rewardAcc = u256.Zero().Sub(globalAcc, lowerAcc)
		rewardAcc = rewardAcc.Sub(rewardAcc, upperAcc)
	}

	return rewardAcc
}

// Calculate actual reward in [startHeight, endHeight) for a position by
// subtracting the startHeight's raw reward from the endHeight's raw reward
func (self *Pool) CalculateRewardForPosition(startHeight int64, startTick int32, endHeight int64, endTick int32, deposit *Deposit) *u256.Uint {
	rewardAcc := self.CalculateRawRewardForPosition(endHeight, endTick, deposit)

	debtAcc := self.CalculateRawRewardForPosition(startHeight, startTick, deposit)

	rewardAcc = rewardAcc.Sub(rewardAcc, debtAcc)

	return rewardAcc
}
