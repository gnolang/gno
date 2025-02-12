package staker

const (
	TIER1_INDEX   = 0
	TIER2_INDEX   = 1
	TIER3_INDEX   = 2
	TIER_TYPE_NUM = 3
)

// RewardPool represents a reward pool.
// A RewardPool type has the following fields:
// - tier: tier of pool
// - rewardAmount: distributable reward amount
// - distributedAmount: distributed reward amount
// - leftAmount: left reward amount
type RewardPool struct {
	tier              uint64
	rewardAmount      uint64
	distributedAmount uint64
	leftAmount        uint64
}

// NewRewardPool creates a new instance of RewardPool with default values.
//
// Returns:
// - *RewardPool: A pointer to the initialized RewardPool instance.
func NewRewardPool() *RewardPool {
	return &RewardPool{
		tier:              0,
		rewardAmount:      0,
		distributedAmount: 0,
		leftAmount:        0,
	}
}

// SetTier : Sets the tier of the pool.
func (r *RewardPool) SetTier(tier uint64) {
	r.tier = tier
}

// SetRewardAmount : Sets the total reward amount for the pool.
func (r *RewardPool) SetRewardAmount(rewardAmount uint64) {
	r.rewardAmount = rewardAmount
}

// SetDistributedAmount : Sets the amount of rewards distributed.
func (r *RewardPool) SetDistributedAmount(distributedAmount uint64) {
	r.distributedAmount = distributedAmount
}

// SetLeftAmount : Sets the remaining rewards in the pool.
func (r *RewardPool) SetLeftAmount(leftAmount uint64) {
	r.leftAmount = leftAmount
}

// GetTier : Returns the tier of the pool.
func (r *RewardPool) GetTier() uint64 {
	return r.tier
}

// GetRewardAmount : Returns the total reward amount.
func (r *RewardPool) GetRewardAmount() uint64 {
	return r.rewardAmount
}

// GetDistributedAmount : Returns the distributed reward amount.
func (r *RewardPool) GetDistributedAmount() uint64 {
	return r.distributedAmount
}

// GetLeftAmount : Returns the remaining reward amount.
func (r *RewardPool) GetLeftAmount() uint64 {
	return r.leftAmount
}

// RewardPoolMap manages multiple reward pools and tier-specific reward data.
//
// Fields:
// - rewardPools (map[string]*RewardPool): Maps pool paths to RewardPool instances.
// - rewardAmountForTier ([TIER_TYPE_NUM]uint64): Total reward amount for each tier.
// - rewardAmountForTierEachPool ([TIER_TYPE_NUM]uint64): Reward amount for each pool in each tier.
// - leftAmountForTier ([TIER_TYPE_NUM]uint64): Remaining reward amount for each tier.
type RewardPoolMap struct {
	rewardPools                 map[string]*RewardPool // poolPath -> RewardPool
	rewardAmountForTier         [TIER_TYPE_NUM]uint64  // total reward amount for each tier
	rewardAmountForTierEachPool [TIER_TYPE_NUM]uint64  // reward amount for each pool in each tier
	leftAmountForTier           [TIER_TYPE_NUM]uint64  // left reward amount for each tier
}

// NewRewardPoolMap creates a new instance of RewardPoolMap with default values.
//
// Returns:
// - *RewardPoolMap: A pointer to the initialized RewardPoolMap instance.
func NewRewardPoolMap() *RewardPoolMap {
	return &RewardPoolMap{
		rewardPools:                 make(map[string]*RewardPool),
		rewardAmountForTier:         [TIER_TYPE_NUM]uint64{},
		rewardAmountForTierEachPool: [TIER_TYPE_NUM]uint64{},
		leftAmountForTier:           [TIER_TYPE_NUM]uint64{},
	}
}

// SetRewardPoolMap : Replaces the current reward pool map.
func (r *RewardPoolMap) SetRewardPoolMap(rewardPools map[string]*RewardPool) {
	r.rewardPools = rewardPools
}

// SetRewardPool : Adds or updates a reward pool for a specific pool path.
func (r *RewardPoolMap) SetRewardPool(poolPath string, rewardPool *RewardPool) {
	r.rewardPools[poolPath] = rewardPool
}

// SetPoolTier : Sets the tier of a specific pool.
func (r *RewardPoolMap) SetPoolTier(poolPath string, tier uint64) {
	if _, exist := r.rewardPools[poolPath]; !exist {
		r.rewardPools[poolPath] = NewRewardPool()
	}
	r.rewardPools[poolPath].SetTier(tier)
}

// SetPoolRewardAmount : Sets the reward amount for a specific pool.
func (r *RewardPoolMap) SetPoolRewardAmount(poolPath string, rewardAmount uint64) {
	if _, exist := r.rewardPools[poolPath]; !exist {
		r.rewardPools[poolPath] = NewRewardPool()
	}
	r.rewardPools[poolPath].SetRewardAmount(rewardAmount)
}

// SetRewardAmountForTier : Sets the total reward amount for a tier.
func (r *RewardPoolMap) SetRewardAmountForTier(tierIndex int, amount uint64) {
	r.rewardAmountForTier[tierIndex] = amount
}

// SetRewardAmountForTierEachPool : Sets the reward amount per pool for a tier.
func (r *RewardPoolMap) SetRewardAmountForTierEachPool(tierIndex int, amount uint64) {
	r.rewardAmountForTierEachPool[tierIndex] = amount
}

// SetLeftAmountForTier : Sets the remaining reward amount for a tier.
func (r *RewardPoolMap) SetLeftAmountForTier(tierIndex int, amount uint64) {
	r.leftAmountForTier[tierIndex] = amount
}

// GetRewardPools : Returns the entire reward pool map.
func (r *RewardPoolMap) GetRewardPools() map[string]*RewardPool {
	return r.rewardPools
}

// GetRewardPoolByPoolPath : Returns the RewardPool for a specific pool path.
func (r *RewardPoolMap) GetRewardPoolByPoolPath(poolPath string) *RewardPool {
	if _, exist := r.rewardPools[poolPath]; !exist {
		r.rewardPools[poolPath] = NewRewardPool()
	}
	return r.rewardPools[poolPath]
}

// GetPoolTier : Returns the tier of a specific pool.
func (r *RewardPoolMap) GetPoolTier(poolPath string) uint64 {
	if _, exist := r.rewardPools[poolPath]; !exist {
		r.rewardPools[poolPath] = NewRewardPool()
	}
	return r.rewardPools[poolPath].tier
}

// GetPoolRewardAmount : Returns the reward amount for a specific pool.
func (r *RewardPoolMap) GetPoolRewardAmount(poolPath string) uint64 {
	if _, exist := r.rewardPools[poolPath]; !exist {
		r.rewardPools[poolPath] = NewRewardPool()
	}
	return r.rewardPools[poolPath].rewardAmount
}

// GetRewardAmountForTier : Returns the total reward amount for a tier.
func (r *RewardPoolMap) GetRewardAmountForTier(tierIndex int) uint64 {
	return r.rewardAmountForTier[tierIndex]
}

// GetRewardAmountForTierEachPool : Returns the reward amount per pool for a tier.
func (r *RewardPoolMap) GetRewardAmountForTierEachPool(tierIndex int) uint64 {
	return r.rewardAmountForTierEachPool[tierIndex]
}

// GetLeftAmountForTier : Returns the remaining reward amount for a tier.
func (r *RewardPoolMap) GetLeftAmountForTier(tierIndex int) uint64 {
	return r.leftAmountForTier[tierIndex]
}
