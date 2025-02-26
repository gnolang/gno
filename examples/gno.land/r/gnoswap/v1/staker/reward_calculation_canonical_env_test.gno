package staker

// "Canonical" implementation of reward calculation.
// Used for testing and reference.

import (
	"errors"
	"math"
	"std"
	"testing"

	ufmt "gno.land/p/demo/ufmt"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	en "gno.land/r/gnoswap/v1/emission"
)

type canonicalPool struct {
	poolPath  string
	tier      uint64
	tick      int32
	incentive []*ExternalIncentive

	tickCrossHook func(poolPath string, tickId int32, zeroForOne bool)
}

func (self *canonicalPool) InternalReward(emission uint64, ratio TierRatio, count uint64) uint64 {
	switch self.tier {
	case 0:
		return 0
	case 1:
		return emission * ratio.Tier1 / count / 100
	case 2:
		return emission * ratio.Tier2 / count / 100
	case 3:
		return emission * ratio.Tier3 / count / 100
	default:
		panic("invalid tier")
	}
}

func (self *canonicalPool) ExternalReward(currentHeight int64) map[string]uint64 {
	reward := make(map[string]uint64)

	for _, incentive := range self.incentive {
		if incentive.startHeight > int64(currentHeight) || incentive.endHeight < int64(currentHeight) {
			continue
		}

		reward[incentive.rewardToken] += incentive.rewardPerBlock
	}

	return reward
}

type canonicalRewardState struct {
	t *testing.T

	global *emulatedGlobalState

	Pool          map[string]*canonicalPool
	tickCrossHook func(poolPath string, tickId int32, zeroForOne bool)

	Reward []map[uint64]Reward // blockNumber -> depositId -> reward

	emissionUpdates *UintTree

	MsPerBlock       int64
	PerBlockEmission uint64

	CurrentTimestamp int64

	// emulated reward claimed by unstake
	emulatedClaimedReward map[uint64]uint64
}

func NewCanonicalRewardState(t *testing.T, pools *Pools, deposits *Deposits, tickCrossHook func(pools *Pools, height func() int64) func(poolPath string, tickId int32, zeroForOne bool)) *canonicalRewardState {
	std.TestSkipHeights(1)

	result := &canonicalRewardState{
		t: t,
		global: &emulatedGlobalState{
			poolTier: poolTier,
			pools:    pools,
			deposits: deposits,
		},
		Pool:             make(map[string]*canonicalPool),
		Reward:           make([]map[uint64]Reward, 124),
		emissionUpdates:  NewUintTree(),
		MsPerBlock:       1000,
		PerBlockEmission: 1000000000,
		CurrentTimestamp: 0,

		emulatedClaimedReward: make(map[uint64]uint64),
	}
	result.tickCrossHook = tickCrossHook(pools, func() int64 {
		return int64(result.CurrentHeight())
	})

	// Technically, the block number being provided to the NewPoolTier should be 124, but because of duplicate creation of the default pool, we put 123 here to just make rewardCache work for the testing.
	result.global.poolTier = NewPoolTier(pools, 123, test_gnousdc, func() uint64 { return result.PerBlockEmission }, func(start, end int64) ([]int64, []uint64) {
		heights := make([]int64, 0)
		emissions := make([]uint64, 0)
		result.emissionUpdates.Iterate(start, end, func(key int64, value interface{}) bool {
			heights = append(heights, key)
			emissions = append(emissions, value.(uint64))
			return false
		})
		return heights, emissions
	})

	result.NextBlock() // must skip height 0

	result.SetEmissionUpdate(1000000000)

	if std.ChainHeight() != result.CurrentHeight() {
		panic(ufmt.Sprintf("height mismatch: %d != %d", std.ChainHeight(), result.CurrentHeight()))
	}

	return result
}

type emulatedGlobalState struct {
	poolTier *PoolTier
	pools    *Pools
	deposits *Deposits
}

func (self *canonicalRewardState) isInRange(deposit *Deposit) bool {
	tick := self.Pool[deposit.targetPoolPath].tick
	return deposit.tickLower <= tick && tick < deposit.tickUpper
}

func (self *canonicalRewardState) SetEmissionUpdate(emission uint64) {
	self.emissionUpdates.set(self.CurrentHeight(), emission)
	self.PerBlockEmission = emission
}

func (self *canonicalRewardState) LiquidityPerPool() map[string]*u256.Uint {
	liquidity := make(map[string]*u256.Uint)
	self.global.deposits.Iterate(0, math.MaxUint64, func(positionId uint64, deposit *Deposit) bool {
		if !self.isInRange(deposit) {
			return false
		}

		poolLiquidity, ok := liquidity[deposit.targetPoolPath]
		if !ok {
			poolLiquidity = u256.Zero()
		}

		poolLiquidity = poolLiquidity.Add(poolLiquidity, deposit.liquidity)
		liquidity[deposit.targetPoolPath] = poolLiquidity
		return false
	})

	return liquidity
}

func (self *canonicalRewardState) InternalRewardPerPool(emission uint64) map[string]uint64 {
	reward := make(map[string]uint64)
	tierCount := []uint64{0, 0, 0, 0}

	for _, pool := range self.Pool {
		tierCount[pool.tier]++
	}
	ratio := TierRatioFromCounts(tierCount[1], tierCount[2], tierCount[3])

	for _, pool := range self.Pool {
		reward[pool.poolPath] = pool.InternalReward(emission, ratio, tierCount[pool.tier])
	}

	return reward
}

func (self *canonicalRewardState) ExternalRewardPerPool(currentHeight int64) map[string]map[string]uint64 {
	reward := make(map[string]map[string]uint64)

	for _, pool := range self.Pool {
		reward[pool.poolPath] = pool.ExternalReward(currentHeight)
	}

	return reward
}

func (self *canonicalRewardState) CurrentHeight() int64 {
	return int64(len(self.Reward)) // due to testing requirement
}

// Process block with canonical reward calculation
func (self *canonicalRewardState) CalculateCanonicalReward() map[uint64]Reward {
	currentHeight := self.CurrentHeight() + 1
	rewards := make(map[uint64]Reward)

	liquidityPerPool := self.LiquidityPerPool()
	internalRewardPerPool := self.InternalRewardPerPool(self.PerBlockEmission)
	externalRewardPerPool := self.ExternalRewardPerPool(int64(currentHeight))

	self.global.deposits.Iterate(0, math.MaxUint64, func(positionId uint64, deposit *Deposit) bool {
		if !self.isInRange(deposit) {
			return false
		}

		warmup := deposit.warmups[deposit.FindWarmup(int64(currentHeight))]
		internal, internalPenalty := warmup.apply(internalRewardPerPool[deposit.targetPoolPath], deposit.liquidity, liquidityPerPool[deposit.targetPoolPath])
		poolExternals := externalRewardPerPool[deposit.targetPoolPath]
		externals := make(map[string]uint64)
		externalPenalties := make(map[string]uint64)
		for key, value := range poolExternals {
			external, externalPenalty := warmup.apply(value, deposit.liquidity, liquidityPerPool[deposit.targetPoolPath])
			externals[key] = external
			externalPenalties[key] = externalPenalty
		}
		rewards[positionId] = Reward{
			Internal:        internal,
			External:        externals,
			InternalPenalty: internalPenalty,
			ExternalPenalty: externalPenalties,
		}
		return false
	})

	return rewards
}

func (self *canonicalRewardState) NextBlock() {
	self.Reward = append(self.Reward, self.CalculateCanonicalReward())
	self.CurrentTimestamp += self.MsPerBlock
	std.TestSkipHeights(1)
	if int64(self.CurrentHeight()) != std.ChainHeight() {
		panic(ufmt.Sprintf("height mismatch: %d != %d", self.CurrentHeight(), std.ChainHeight()))
	}
}

func (self *canonicalRewardState) NextBlockNoCanonical() {
	self.Reward = append(self.Reward, nil) // just placeholder
	self.CurrentTimestamp += self.MsPerBlock
	std.TestSkipHeights(1)
	if int64(self.CurrentHeight()) != std.ChainHeight() {
		panic(ufmt.Sprintf("height mismatch: %d != %d", self.CurrentHeight(), std.ChainHeight()))
	}
}

func (self *canonicalRewardState) UnclaimableExternalRewardOf(depositId uint64, incentiveId string) uint64 {
	pool, ok := self.global.pools.Get(self.global.deposits.Get(depositId).targetPoolPath)
	if !ok {
		panic("pool not found")
	}

	return pool.incentives.calculateUnclaimableReward(incentiveId)
}

func (self *canonicalRewardState) CanonicalRewardOf(depositId uint64) Reward {
	return self.Reward[self.CurrentHeight()-1][depositId]
}

func (self *canonicalRewardState) SafeCanonicalRewardOf(depositId uint64) (Reward, bool) {
	rewards := self.Reward[self.CurrentHeight()-1]
	reward, ok := rewards[depositId]
	return reward, ok
}

func (self *canonicalRewardState) CanonicalRewardOfHeight(depositId uint64, height uint64) Reward {
	return self.Reward[height][depositId]
}

func (self *canonicalRewardState) EmulateCalcPositionReward(positionId uint64) ([]uint64, []uint64, []map[string]uint64, []map[string]uint64) {
	currentHeight := self.CurrentHeight()

	// cache per-tier and per-pool rewards
	self.global.poolTier.cacheReward(currentHeight, self.global.pools)

	deposit := self.global.deposits.Get(positionId)

	poolPath := deposit.targetPoolPath

	pool, ok := self.global.pools.Get(poolPath)
	if !ok {
		pool = NewPool(poolPath, currentHeight)
		self.global.pools.set(poolPath, pool)
	}

	lastCollectHeight := deposit.lastCollectHeight

	internalRewards, internalPenalties := pool.RewardStateOf(deposit).calculateInternalReward(int64(lastCollectHeight), int64(currentHeight))

	externalRewards := make([]map[string]uint64, 4)
	externalPenalties := make([]map[string]uint64, 4)
	for i := range externalRewards {
		externalRewards[i] = make(map[string]uint64)
		externalPenalties[i] = make(map[string]uint64)
	}

	allIncentives := pool.incentives.GetAllInHeights(int64(lastCollectHeight), int64(currentHeight))

	for incentiveId, incentive := range allIncentives {
		externalReward, externalPenalty := pool.RewardStateOf(deposit).calculateExternalReward(int64(lastCollectHeight), int64(currentHeight), incentive)

		for i := range externalReward {
			externalRewards[i][incentive.incentiveId] = externalReward[i]
			externalPenalties[i][incentive.incentiveId] = externalPenalty[i]
		}
	}

	return internalRewards, internalPenalties, externalRewards, externalPenalties
}

func (self *canonicalRewardState) EmulatedRewardOf(depositId uint64) Reward {
	if !self.global.deposits.Has(depositId) {
		claimed := self.emulatedClaimedReward[depositId]
		self.emulatedClaimedReward[depositId] = 0
		return Reward{
			Internal:        claimed,
			InternalPenalty: 0,
			External:        make(map[string]uint64),
			ExternalPenalty: make(map[string]uint64),
		}
	}

	rewards, penalties, externalRewards, externalPenalties := self.EmulateCalcPositionReward(depositId)

	deposit := self.global.deposits.Get(depositId)
	deposit.lastCollectHeight = self.CurrentHeight()

	internal := uint64(0)
	for _, reward := range rewards {
		internal += reward
	}
	claimed, ok := self.emulatedClaimedReward[depositId]
	if ok {
		internal += claimed
		self.emulatedClaimedReward[depositId] = 0
	}
	internalPenalty := uint64(0)
	for _, penalty := range penalties {
		internalPenalty += penalty
	}
	external := make(map[string]uint64)
	for _, er := range externalRewards {
		for incentiveId, reward := range er {
			external[incentiveId] += reward
		}
	}
	externalPenalty := make(map[string]uint64)
	for _, ep := range externalPenalties {
		for incentiveId, penalty := range ep {
			externalPenalty[incentiveId] += penalty
		}
	}

	return Reward{
		Internal:        internal,
		InternalPenalty: internalPenalty,
		External:        external,
		ExternalPenalty: externalPenalty,
	}
}

// Emulation of staker.gno public entrypoints
func (self *canonicalRewardState) StakeToken(positionId uint64, targetPoolPath string, owner std.Address, tickLower int32, tickUpper int32, liquidity *u256.Uint) error {
	currentHeight := self.CurrentHeight()
	pool, ok := self.global.pools.Get(targetPoolPath)
	if !ok {
		panic("should not happen 1")
	}

	deposit := &Deposit{
		owner:             owner,
		stakeHeight:       int64(currentHeight),
		targetPoolPath:    targetPoolPath,
		tickLower:         tickLower,
		tickUpper:         tickUpper,
		liquidity:         liquidity,
		lastCollectHeight: currentHeight,
		warmups:           instantiateWarmup(currentHeight),
	}
	canonicalPool, ok := self.Pool[deposit.targetPoolPath]
	if !ok {
		return errors.New("pool not found")
	}
	if canonicalPool.tier == 0 && len(canonicalPool.incentive) == 0 {
		return errors.New("pool has no tier or incentive")
	}

	// update global state
	self.global.deposits.set(positionId, deposit)

	self.global.poolTier.cacheReward(currentHeight, self.global.pools)

	signedLiquidity := i256.FromUint256(deposit.liquidity)
	if self.isInRange(deposit) {
		pool.modifyDeposit(signedLiquidity, currentHeight, canonicalPool.tick)
	}
	// historical tick must be set regardless of the deposit's range
	pool.historicalTick.set(currentHeight, canonicalPool.tick)

	pool.ticks.Get(deposit.tickLower).modifyDepositLower(currentHeight, canonicalPool.tick, signedLiquidity)
	pool.ticks.Get(deposit.tickUpper).modifyDepositUpper(currentHeight, canonicalPool.tick, signedLiquidity)

	return nil
}

func (self *canonicalRewardState) UnstakeToken(positionId uint64) {
	deposit := self.global.deposits.Get(positionId)

	currentHeight := self.CurrentHeight()

	canonicalPool, ok := self.Pool[deposit.targetPoolPath]
	if !ok {
		panic("should not happen 2")
	}

	// Emulating CollectReward()
	reward := self.EmulatedRewardOf(positionId)
	self.emulatedClaimedReward[positionId] += reward.Internal

	// update global state
	// we will not gonna actually remove the deposit in sake of logic simplicity
	self.global.deposits.remove(positionId)

	pool, ok := self.global.pools.Get(deposit.targetPoolPath)
	if !ok {
		panic("should not happen 3")
	}
	signedLiquidity := i256.FromUint256(deposit.liquidity)
	signedLiquidity = signedLiquidity.Neg(signedLiquidity)
	if self.isInRange(deposit) {
		pool.modifyDeposit(signedLiquidity, currentHeight, canonicalPool.tick)
	}
	pool.ticks.Get(deposit.tickLower).modifyDepositLower(currentHeight, canonicalPool.tick, signedLiquidity)
	pool.ticks.Get(deposit.tickUpper).modifyDepositUpper(currentHeight, canonicalPool.tick, signedLiquidity)
}

func newExternalIncentiveByHeight(
	targetPoolPath string,
	rewardToken string,
	rewardAmount uint64,
	startTimestamp int64,
	endTimestamp int64,
	startHeight int64,
	endHeight int64,
	refundee std.Address,
) *ExternalIncentive {
	rewardPerBlock := rewardAmount / uint64(endHeight-startHeight)

	byTimeId := incentiveIdByTime(startTimestamp, endTimestamp, refundee, rewardToken)

	return &ExternalIncentive{
		incentiveId:      byTimeId,
		targetPoolPath:   targetPoolPath,
		rewardToken:      rewardToken,
		rewardAmount:     rewardAmount,
		startTimestamp:   startTimestamp,
		endTimestamp:     endTimestamp,
		startHeight:      startHeight,
		endHeight:        endHeight,
		rewardPerBlock:   rewardPerBlock,
		refundee:         refundee,
		createdHeight:    startHeight,
		depositGnsAmount: 0,
	}
}

func (self *canonicalRewardState) CreateExternalIncentive(targetPoolPath string, rewardToken string, rewardAmount uint64, startTimestamp, endTimestamp, startHeight, endHeight int64, refundee std.Address) string {
	incentive := newExternalIncentiveByHeight(targetPoolPath, rewardToken, rewardAmount, startTimestamp, endTimestamp, startHeight, endHeight, refundee)

	// update canonical state
	pool, ok := self.Pool[targetPoolPath]
	if !ok {
		self.Pool[targetPoolPath] = &canonicalPool{
			poolPath:      targetPoolPath,
			tier:          0,
			tick:          0,
			incentive:     make([]*ExternalIncentive, 0),
			tickCrossHook: self.tickCrossHook,
		}
	}
	pool.incentive = append(pool.incentive, incentive)

	// update global state
	self.global.pools.GetOrCreate(targetPoolPath).incentives.create(refundee, incentive)

	return incentive.incentiveId
}

func (self *canonicalRewardState) ChangePoolTier(poolPath string, tier uint64) {
	// update canonical state
	pool, ok := self.Pool[poolPath]
	if !ok {
		pool = &canonicalPool{
			poolPath:      poolPath,
			tier:          tier,
			tick:          0,
			incentive:     make([]*ExternalIncentive, 0),
			tickCrossHook: self.tickCrossHook,
		}
		self.Pool[poolPath] = pool
	}
	pool.tier = tier

	// update global state
	if !self.global.pools.Has(poolPath) {
		self.global.pools.set(poolPath, NewPool(poolPath, self.CurrentHeight()))
	}
	self.global.poolTier.changeTier(self.CurrentHeight(), self.global.pools, poolPath, tier)
}

func (self *canonicalRewardState) CreatePool(poolPath string, initialTier uint64, initialTick int32) {
	self.Pool[poolPath] = &canonicalPool{
		poolPath:      poolPath,
		tier:          initialTier,
		tick:          initialTick,
		incentive:     make([]*ExternalIncentive, 0),
		tickCrossHook: self.tickCrossHook,
	}
	self.global.pools.set(poolPath, NewPool(poolPath, self.CurrentHeight()))
	self.global.poolTier.changeTier(self.CurrentHeight(), self.global.pools, poolPath, initialTier)
}

func (self *canonicalRewardState) MoveTick(poolPath string, tick int32) {
	pool, ok := self.Pool[poolPath]
	if !ok {
		panic("should not happen 4")
	}
	globalPool, ok := self.global.pools.Get(poolPath)
	if !ok {
		panic("should not happen 5")
	}

	if pool.tick == tick {
		return
	}

	self.t.Logf("	[%d] (%d->%d) %s", self.CurrentHeight(), pool.tick, tick, pool.poolPath)

	zeroForOne := tick < pool.tick // true if moving left, false if moving right
	if zeroForOne {
		// backward
		for i := pool.tick; i > tick; i-- {
			// uninitialized tick
			if !globalPool.ticks.Has(i) {
				continue
			}

			// update global state
			pool.tickCrossHook(pool.poolPath, i, zeroForOne)
		}
	} else {
		// forward
		for i := pool.tick + 1; i <= tick; i++ {
			// uninitialized tick
			if !globalPool.ticks.Has(i) {
				continue
			}

			// update global state
			pool.tickCrossHook(pool.poolPath, i, zeroForOne)
		}
	}

	// update canonical state
	pool.tick = tick
}

// Testing helpers

func (self *canonicalRewardState) AssertCanonicalInternalRewardPerPool(poolPath string, expected uint64) {
	internalRewardPerPool := self.InternalRewardPerPool(self.PerBlockEmission)
	actual := internalRewardPerPool[poolPath]
	if actual != expected {
		panic(ufmt.Sprintf("internal reward per pool mismatch: expected %d, got %d", expected, actual))
	}
}

// returns true if actual is within 0.0001% of expected
func isInErrorRange(expected uint64, actual uint64) bool {
	maxSafeValue := uint64(math.MaxUint64 / 100001)
	var lowerBound, upperBound uint64
	if expected > maxSafeValue {
		lowerBound = expected / 1000000 * 999999
		upperBound = expected / 1000000 * 1000001
	} else {
		lowerBound = expected * 999999 / 1000000
		upperBound = expected * 1000001 / 1000000
	}
	return actual >= lowerBound && actual <= upperBound
}

func (self *canonicalRewardState) AssertEmulatedRewardOf(depositId uint64, expected uint64) {
	reward := self.EmulatedRewardOf(depositId)
	if !isInErrorRange(expected, reward.Internal) {
		self.t.Errorf("emulated reward of %d mismatch: expected %d, got %d", depositId, expected, reward.Internal)
		panic("emulated reward mismatch")
	}
}

func (self *canonicalRewardState) AssertEmulatedExternalRewardOf(depositId uint64, incentiveId string, expected uint64) {
	reward := self.EmulatedRewardOf(depositId)
	if !isInErrorRange(expected, reward.External[incentiveId]) {
		self.t.Errorf("!!!!emulated external reward of %d mismatch: expected %d, got %d", depositId, expected, reward.External[incentiveId])
	}
}

func (self *canonicalRewardState) AssertCanonicalRewardOf(depositId uint64, expected uint64) {
	reward := self.CanonicalRewardOf(depositId)
	if !isInErrorRange(expected, reward.Internal) {
		self.t.Errorf("canonical reward of %d mismatch: expected %d, got %d", depositId, expected, reward.Internal)
	}
}

func (self *canonicalRewardState) AssertEquivalence(depositId uint64) {
	reward := self.CanonicalRewardOf(depositId)
	emulatedReward := self.EmulatedRewardOf(depositId)
	if !isInErrorRange(reward.Internal, emulatedReward.Internal) {
		self.t.Errorf("canonical reward of %d mismatch: expected %d, got %d", depositId, reward.Internal, emulatedReward.Internal)
	}
}

func (self *canonicalRewardState) AssertEmulatedRewardMap(expected map[uint64]uint64) {
	for key, value := range expected {
		self.AssertEmulatedRewardOf(key, value)
	}
}
