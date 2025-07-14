package launchpad

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	u256 "gno.land/p/gnoswap/uint256"
)

// RewardManager manages the distribution of rewards for a project tier.
//
// This struct contains the necessary data and methods to calculate and track
// rewards for deposits associated with a project tier.
//
// Fields:
// - rewards (avl.Tree): A map of deposit IDs to their associated reward states.
// - distributeAmountPerBlockX128 (u256.Uint): The amount of tokens to be distributed per block, represented as a Q128 fixed-point number.
// - accumulatedRewardPerDepositX128 (u256.Uint): The accumulated reward per GNS stake, represented as a Q128 fixed-point number.
// - totalDistributeAmount (int64): The total amount of tokens to be distributed.
// - totalClaimedAmount (int64): The total amount of tokens claimed.
// - distributeStartHeight (int64): The start height of the reward calculation.
// - distributeEndHeight (int64): The end height of the reward calculation.
// - accumulatedDistributeAmount (int64): The accumulated amount of tokens distributed.
// - accumulatedHeight (int64): The last height when reward was calculated.
// - rewardClaimableDuration (int64): The duration of reward claimable.
type RewardManager struct {
	rewards *avl.Tree // depositId -> RewardState

	distributeAmountPerBlockX128    *u256.Uint // distribute amount per block, Q128
	accumulatedRewardPerDepositX128 *u256.Uint // accumulated reward per GNS stake, Q128

	totalDistributeAmount       int64 // total distributed amount
	totalClaimedAmount          int64 // total claimed amount
	distributeStartHeight       int64 // start height of reward calculation
	distributeEndHeight         int64 // end height of reward calculation
	accumulatedDistributeAmount int64 // accumulated distribute amount
	accumulatedHeight           int64 // last height when reward was calculated
	rewardClaimableDuration     int64 // duration of reward claimable
}

func (r *RewardManager) IsInitialized() bool {
	return r.rewards.Size() > 0
}

func (r *RewardManager) DistributeAmountPerBlockX128() *u256.Uint {
	return r.distributeAmountPerBlockX128
}

func (r *RewardManager) AccumulatedHeight() int64 {
	return r.accumulatedHeight
}

func (r *RewardManager) DistributeEndHeight() int64 {
	return r.distributeEndHeight
}

func (r *RewardManager) AccumulatedRewardPerDepositX128() *u256.Uint {
	return r.accumulatedRewardPerDepositX128
}

func (r *RewardManager) AccumulatedReward() int64 {
	return u256.Zero().Rsh(r.accumulatedRewardPerDepositX128, 128).Int64()
}

func (r *RewardManager) getDepositRewardState(depositId string) (*RewardState, error) {
	rewardStateI, exists := r.rewards.Get(depositId)
	if !exists {
		return nil, makeErrorWithDetails(errNotExistDeposit, ufmt.Sprintf("(%s)", depositId))
	}

	return rewardStateI.(*RewardState), nil
}

func (r *RewardManager) calculateRewardPerDepositX128(rewardPerBlockX128 *u256.Uint, totalStaked int64, currentHeight int64) (*u256.Uint, error) {
	accumulatedHeight := r.accumulatedHeight
	if r.distributeStartHeight > accumulatedHeight {
		accumulatedHeight = r.distributeStartHeight
	}

	// not started yet
	if currentHeight < accumulatedHeight {
		return u256.Zero(), nil
	}

	// past distribute end height
	if accumulatedHeight > r.distributeEndHeight {
		return u256.Zero(), nil
	}

	// past distribute end height, set to distribute end height
	if currentHeight > r.distributeEndHeight {
		currentHeight = r.distributeEndHeight
	}

	if rewardPerBlockX128.IsZero() {
		return nil, makeErrorWithDetails(
			errNoLeftReward,
			ufmt.Sprintf("rewardPerBlock(%d)", rewardPerBlockX128),
		)
	}

	// no left reward
	if totalStaked == 0 {
		return u256.Zero(), nil
	}

	// blockDuration * rewardPerBlock / totalStaked
	blockDuration := currentHeight - accumulatedHeight
	totalRewardX128 := u256.Zero().Mul(u256.NewUintFromInt64(blockDuration), rewardPerBlockX128)
	rewardPerDepositX128 := u256.Zero().Div(totalRewardX128, u256.NewUintFromInt64(totalStaked))

	return rewardPerDepositX128, nil
}

func (r *RewardManager) addRewardStateByDeposit(deposit *Deposit, averageBlickTimeMs int64) *RewardState {
	claimableBlockTimeInfo := newBlockTimeInfoByTimeAndAverageBlockTimeMs(
		deposit.createdBlockTimeInfo.BlockTime(),
		deposit.createdBlockTimeInfo.BlockHeight(),
		deposit.createdBlockTimeInfo.BlockTime(),
		int64(averageBlickTimeMs),
	)

	rewardState := NewRewardState(
		r.AccumulatedRewardPerDepositX128(),
		deposit.DepositAmount(),
		deposit.DepositHeight(),
		r.distributeEndHeight,
		claimableBlockTimeInfo,
	)

	// if the first deposit, set the distribute start height
	if !r.IsInitialized() {
		rewardState.setDistributeStartHeight(r.distributeStartHeight)
		rewardState.setDistributeEndHeight(r.distributeEndHeight)
		rewardState.setAccumulatedHeight(r.distributeStartHeight)
		rewardState.setPriceDebtX128(u256.Zero())
	}

	return r.addRewardState(deposit, rewardState)
}

func (r *RewardManager) addRewardState(deposit *Deposit, rewardState *RewardState) *RewardState {
	r.rewards.Set(deposit.ID(), rewardState)

	return rewardState
}

func (r *RewardManager) addRewardPerDepositX128(rewardPerDepositX128 *u256.Uint, currentHeight int64) {
	if rewardPerDepositX128.IsZero() {
		return
	}

	if r.accumulatedHeight > currentHeight || r.distributeStartHeight > currentHeight {
		return
	}

	if currentHeight > r.distributeEndHeight {
		currentHeight = r.distributeEndHeight
	}

	r.accumulatedRewardPerDepositX128 = u256.Zero().Add(r.accumulatedRewardPerDepositX128, rewardPerDepositX128)
	r.accumulatedHeight = currentHeight
}

// updateRewardPerDepositX128 updates the reward per deposit state.
// This function calculates and updates the accumulated reward per deposit
// based on the current total deposit amount and height.
//
// Parameters:
// - totalDepositAmount (int64): Current total deposit amount
// - height (int64): Current blockchain height
//
// Returns:
// - error: If the update fails
func (r *RewardManager) updateRewardPerDepositX128(totalDepositAmount int64, height int64) error {
	// Input validation
	if height <= 0 {
		return makeErrorWithDetails(errInvalidHeight, "height must be positive")
	}

	// Calculate and update rewards
	rewardPerDepositX128, err := r.calculateRewardPerDepositX128(
		r.distributeAmountPerBlockX128,
		totalDepositAmount,
		height,
	)
	if err != nil {
		return err
	}

	r.addRewardPerDepositX128(rewardPerDepositX128, height)

	return nil
}

func (r *RewardManager) updateDistributeAmountPerBlockX128(totalDistributeAmount int64, distributeStartHeight int64, distributeEndHeight int64) {
	blockDuration := distributeEndHeight - distributeStartHeight
	if blockDuration <= 0 {
		return
	}

	totalDistributeAmountX128 := u256.Zero().Lsh(
		u256.NewUintFromInt64(totalDistributeAmount),
		128,
	)
	amountPerBlockX128 := u256.Zero().Div(
		totalDistributeAmountX128,
		u256.NewUintFromInt64(blockDuration),
	)

	r.distributeAmountPerBlockX128 = amountPerBlockX128
}

// collectReward processes the reward collection for a specific deposit.
// This function ensures that the reward collection is valid and updates
// the claimed amount accordingly.
//
// Parameters:
// - depositId (string): The ID of the deposit
// - currentHeight (int64): Current blockchain height
//
// Returns:
// - int64: The amount of reward collected
// - error: If the collection fails
func (r *RewardManager) collectReward(depositId string, currentHeight int64) (int64, error) {
	if currentHeight < r.accumulatedHeight {
		return 0, makeErrorWithDetails(
			errInvalidRewardState,
			ufmt.Sprintf("currentHeight %d is less than AccumulatedHeight %d", currentHeight, r.accumulatedHeight),
		)
	}

	rewardState, err := r.getDepositRewardState(depositId)
	if err != nil {
		return 0, err
	}

	if !rewardState.IsClaimable(currentHeight) {
		return 0, makeErrorWithDetails(
			errInvalidRewardState,
			ufmt.Sprintf("currentHeight %d is less than claimableHeight %d", currentHeight, rewardState.ClaimableHeight()),
		)
	}

	if currentHeight < rewardState.DistributeStartHeight() {
		return 0, makeErrorWithDetails(
			errInvalidRewardState,
			ufmt.Sprintf("currentHeight %d is less than DistributeStartHeight %d", currentHeight, rewardState.DistributeStartHeight()),
		)
	}

	claimableReward := rewardState.calculateClaimableReward(r.accumulatedRewardPerDepositX128)
	if claimableReward == 0 {
		return 0, nil
	}

	rewardState.setClaimedAmount(rewardState.ClaimedAmount() + claimableReward)
	r.rewards.Set(depositId, rewardState)
	r.totalClaimedAmount += claimableReward

	return claimableReward, nil
}

// NewRewardManager returns a pointer to a new RewardManager with the given values.
func NewRewardManager(
	totalDistributeAmount int64,
	distributeStartHeight int64,
	distributeEndHeight int64,
	currentHeight int64,
) *RewardManager {
	manager := &RewardManager{
		totalDistributeAmount:           totalDistributeAmount,
		distributeStartHeight:           distributeStartHeight,
		distributeEndHeight:             distributeEndHeight,
		totalClaimedAmount:              0,
		accumulatedDistributeAmount:     0,
		accumulatedHeight:               0,
		accumulatedRewardPerDepositX128: u256.Zero(),
		distributeAmountPerBlockX128:    u256.Zero(),
		rewardClaimableDuration:         0,
		rewards:                         avl.NewTree(),
	}

	manager.updateDistributeAmountPerBlockX128(totalDistributeAmount, distributeStartHeight, distributeEndHeight)
	manager.updateRewardPerDepositX128(0, currentHeight)

	return manager
}
