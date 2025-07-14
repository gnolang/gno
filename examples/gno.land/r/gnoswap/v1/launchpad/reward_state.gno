package launchpad

import (
	u256 "gno.land/p/gnoswap/uint256"
)

// RewardState represents the state of a reward for a deposit.
//
// This struct contains the necessary data and methods to manage and distribute
// rewards for a specific deposit.
//
// Fields:
// - priceDebtX128 (u256.Uint): The price debt per GNS stake, represented as a Q128 fixed-point number.
// - claimableBlockTimeInfo (BlockTimeInfo): The block time and height information for the claimable reward.
// - depositAmount (int64): The amount of GNS staked.
// - distributeStartHeight (int64): The start height of the reward calculation.
// - distributeEndHeight (int64): The end height of the reward calculation.
// - accumulatedRewardAmount (int64): The accumulated amount of reward.
// - accumulatedHeight (int64): The last height when reward was calculated.
// - claimedAmount (int64): The amount of reward claimed so far.
type RewardState struct {
	priceDebtX128          *u256.Uint // price debt per GNS stake, Q128
	claimableBlockTimeInfo *BlockTimeInfo

	depositAmount           int64 // amount of GNS staked
	distributeStartHeight   int64 // height when launchpad started staking
	distributeEndHeight     int64 // end height of reward calculation
	accumulatedRewardAmount int64 // calculated, not collected
	accumulatedHeight       int64 // last height when reward was calculated
	claimedAmount           int64 // amount of reward claimed so far
}

func (r *RewardState) PriceDebtX128() *u256.Uint {
	return r.priceDebtX128
}

func (r *RewardState) setPriceDebtX128(v *u256.Uint) {
	r.priceDebtX128 = v
}

func (r *RewardState) DepositAmount() int64 {
	return r.depositAmount
}

func (r *RewardState) setDepositAmount(v int64) {
	r.depositAmount = v
}

func (r *RewardState) AccumulatedRewardAmount() int64 {
	return r.accumulatedRewardAmount
}

func (r *RewardState) setAccumulatedRewardAmount(v int64) {
	r.accumulatedRewardAmount = v
}

func (r *RewardState) ClaimedAmount() int64 {
	return r.claimedAmount
}

func (r *RewardState) setClaimedAmount(v int64) {
	r.claimedAmount = v
}

func (r *RewardState) DistributeStartHeight() int64 {
	return r.distributeStartHeight
}

func (r *RewardState) setDistributeStartHeight(v int64) {
	r.distributeStartHeight = v
}

func (r *RewardState) DistributeEndHeight() int64 {
	return r.distributeEndHeight
}

func (r *RewardState) setDistributeEndHeight(v int64) {
	r.distributeEndHeight = v
}

func (r *RewardState) AccumulatedHeight() int64 {
	return r.accumulatedHeight
}

func (r *RewardState) setAccumulatedHeight(v int64) {
	r.accumulatedHeight = v
}

func (r *RewardState) IsClaimable(currentHeight int64) bool {
	return currentHeight >= r.claimableBlockTimeInfo.BlockHeight()
}

func (r *RewardState) ClaimableHeight() int64 {
	return r.claimableBlockTimeInfo.BlockHeight()
}

func (r *RewardState) ClaimableTime() int64 {
	return r.claimableBlockTimeInfo.BlockTime()
}

// calculateReward calculates the total reward amount based on
// the accumulated reward per deposit.
//
// Parameters:
// - accumRewardPerDepositX128 (*u256.Uint): Current accumulated reward per deposit
//
// Returns:
// - int64: The total reward amount
func (r *RewardState) calculateReward(accumRewardPerDepositX128 *u256.Uint) int64 {
	if accumRewardPerDepositX128 == nil || r.PriceDebtX128() == nil {
		return 0
	}

	actualRewardPerDepositX128 := u256.Zero().Sub(accumRewardPerDepositX128, r.PriceDebtX128())
	if actualRewardPerDepositX128.IsZero() {
		return 0
	}

	reward := u256.Zero().Mul(actualRewardPerDepositX128, u256.NewUintFromInt64(r.DepositAmount()))
	reward = u256.Zero().Rsh(reward, 128)

	return reward.Int64()
}

// calculateClaimableReward calculates the amount of reward that can be claimed
// based on the current accumulated reward per deposit.
//
// Parameters:
// - accumRewardPerDepositX128 (*u256.Uint): Current accumulated reward per deposit
//
// Returns:
// - int64: The amount of reward that can be claimed
func (r *RewardState) calculateClaimableReward(accumRewardPerDepositX128 *u256.Uint) int64 {
	if accumRewardPerDepositX128 == nil {
		return 0
	}

	// Return 0 if accumulated reward is less than price debt
	if accumRewardPerDepositX128.Lt(r.priceDebtX128) {
		return 0
	}

	reward := r.calculateReward(accumRewardPerDepositX128)
	claimedAmount := r.ClaimedAmount()

	if reward <= claimedAmount {
		return 0
	}

	return reward - claimedAmount
}

// NewRewardState returns a pointer to a new RewardState with the given values.
func NewRewardState(
	accumulatedRewardPerDepositX128 *u256.Uint,
	depositAmount,
	distributeStartHeight,
	distributeEndHeight int64,
	claimableBlockTimeInfo *BlockTimeInfo,
) *RewardState {
	return &RewardState{
		priceDebtX128:           accumulatedRewardPerDepositX128,
		depositAmount:           depositAmount,
		distributeStartHeight:   distributeStartHeight,
		distributeEndHeight:     distributeEndHeight,
		claimableBlockTimeInfo:  claimableBlockTimeInfo,
		accumulatedRewardAmount: 0,
		claimedAmount:           0,
		accumulatedHeight:       0,
	}
}
