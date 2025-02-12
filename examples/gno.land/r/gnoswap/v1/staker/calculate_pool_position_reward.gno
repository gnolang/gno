package staker

import (
	"gno.land/p/gnoswap/consts"

	u256 "gno.land/p/gnoswap/uint256"
)

// Q96
var _q96 = u256.MustFromDecimal(consts.Q96)

func isAbleToCalculateEmissionReward(prev int64, current int64) bool {
	if prev >= current {
		return false
	}
	return true
}

// Reward is a struct for storing reward for a position.
// Internal reward is the GNS reward, external reward is the reward for other incentives.
// Penalties are the amount that is deducted from the reward due to the position's warmup.
type Reward struct {
	Internal        uint64
	InternalPenalty uint64
	External        map[string]uint64 // Incentive ID -> TokenAmount
	ExternalPenalty map[string]uint64 // Incentive ID -> TokenAmount
}

// calculate position reward by each warmup
func calcPositionRewardByWarmups(currentHeight int64, positionId uint64) []Reward {
	rewards := calculatePositionReward(CalcPositionRewardParam{
		CurrentHeight: currentHeight,
		Deposits:      deposits,
		Pools:         pools,
		PoolTier:      poolTier,
		PositionId:    positionId,
	})

	return rewards
}

// calculate total position rewards and penalties
func calcPositionReward(currentHeight int64, positionId uint64) Reward {
	rewards := calculatePositionReward(CalcPositionRewardParam{
		CurrentHeight: currentHeight,
		Deposits:      deposits,
		Pools:         pools,
		PoolTier:      poolTier,
		PositionId:    positionId,
	})

	internal := uint64(0)
	for _, reward := range rewards {
		internal += reward.Internal
	}

	internalPenalty := uint64(0)
	for _, reward := range rewards {
		internalPenalty += reward.InternalPenalty
	}

	externalReward := make(map[string]uint64)
	for _, reward := range rewards {
		if reward.External != nil {
			for incentive, reward := range reward.External {
				externalReward[incentive] += reward
			}
		}
	}

	externalPenalty := make(map[string]uint64)
	for _, reward := range rewards {
		if reward.ExternalPenalty != nil {
			for incentive, penalty := range reward.ExternalPenalty {
				externalPenalty[incentive] += penalty
			}
		}
	}

	return Reward{
		Internal:        internal,
		InternalPenalty: internalPenalty,
		External:        externalReward,
		ExternalPenalty: externalPenalty,
	}
}

// CalcPositionRewardParam is a struct for calculating position reward
type CalcPositionRewardParam struct {
	// Environmental variables
	CurrentHeight int64
	Deposits      *Deposits
	Pools         *Pools
	PoolTier      *PoolTier

	// Position variables
	PositionId uint64
}

func calculatePositionReward(param CalcPositionRewardParam) []Reward {
	// cache per-pool rewards in the internal incentive(tiers)
	param.PoolTier.cacheReward(param.CurrentHeight, param.Pools)

	deposit := param.Deposits.Get(param.PositionId)
	poolPath := deposit.targetPoolPath

	pool, ok := param.Pools.Get(poolPath)
	if !ok {
		pool = NewPool(poolPath, param.CurrentHeight)
		param.Pools.set(poolPath, pool)
	}

	lastCollectHeight := deposit.lastCollectHeight

	// Initializes reward/penalty arrays for rewards and penalties for each warmup
	internalRewards := make([]uint64, len(deposit.warmups))
	internalPenalties := make([]uint64, len(deposit.warmups))
	externalRewards := make([]map[string]uint64, len(deposit.warmups))
	externalPenalties := make([]map[string]uint64, len(deposit.warmups))

	if param.PoolTier.CurrentTier(poolPath) != 0 {
		// Internal incentivized pool.
		// Calculate reward for each warmup
		internalRewards, internalPenalties = pool.RewardStateOf(deposit).calculateInternalReward(lastCollectHeight, param.CurrentHeight)
	}

	// All active incentives
	allIncentives := pool.incentives.GetAllInHeights(lastCollectHeight, param.CurrentHeight)

	for i := range externalRewards {
		externalRewards[i] = make(map[string]uint64)
		externalPenalties[i] = make(map[string]uint64)
	}

	for incentiveId, incentive := range allIncentives {
		// External incentivized pool.
		// Calculate reward for each warmup
		externalReward, externalPenalty := pool.RewardStateOf(deposit).calculateExternalReward(int64(lastCollectHeight), int64(param.CurrentHeight), incentive)

		for i := range externalReward {
			externalRewards[i][incentiveId] = externalReward[i]
			externalPenalties[i][incentiveId] = externalPenalty[i]
		}
	}

	rewards := make([]Reward, len(internalRewards))
	for i := range internalRewards {
		rewards[i] = Reward{
			Internal:        internalRewards[i],
			InternalPenalty: internalPenalties[i],
			External:        externalRewards[i],
			ExternalPenalty: externalPenalties[i],
		}
	}

	return rewards
}

// calculates internal unclaimable reward for the pool
func processUnClaimableReward(poolPath string, endHeight int64) uint64 {
	pool, ok := pools.Get(poolPath)
	if !ok {
		return 0
	}
	return pool.processUnclaimableReward(poolTier, endHeight)
}
