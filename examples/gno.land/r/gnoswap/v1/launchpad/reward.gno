package launchpad

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"
	gs "gno.land/r/gnoswap/v1/gov/staker"
)

// RewardCalculationResult represents the result of reward calculation
type RewardCalculationResult struct {
	Reward       uint64
	UpdatedTiers map[string]Tier
}

// CollectProtocolFee collects protocol fee from gov/staker
// each project's recipient wallet will be rewarded
// ref: https://docs.gnoswap.io/contracts/launchpad/launchpad_reward.gno#collectprotocolfee
func CollectProtocolFee() {
	caller := getPrevAddr()
	gs.CollectRewardFromLaunchPad(caller)
}

// CollectRewardByProjectId collects reward from entire deposit of certain project by caller
// Returns collected reward amount
// ref: https://docs.gnoswap.io/contracts/launchpad/launchpad_reward.gno#collectrewardbyprojectid
func CollectRewardByProjectId(projectId string) uint64 {
	assertOnlyNotHalted()

	project, err := getProject(projectId)
	if err != nil {
		panic(err.Error())
	}
	caller := getPrevAddr()
	if _, exists := depositsByUserByProject[caller]; !exists {
		panic(addDetailToError(errInvalidCaller, ufmt.Sprintf("caller(%s) not found", caller)))
	}
	depositIds, exists := depositsByUserByProject[caller][projectId]
	if !exists {
		panic(addDetailToError(
			errNotExistDeposit, ufmt.Sprintf("depositId (%s) not found", projectId)))
	}
	en.MintAndDistributeGns()

	currentHeight := uint64(std.ChainHeight())
	totalReward := uint64(0) // toUser
	prevPkgPath := getPrevPkgPath()

	for _, depositId := range depositIds {
		deposit, ok := deposits[depositId]
		if !ok {
			continue
		}

		if !deposit.isClaimable(currentHeight) {
			continue
		}
		if err = validateRewardCollection(deposit, currentHeight); err != nil {
			panic(err.Error())
		}

		tier, err := project.Tier(convertTierTypeStrToUint64(deposit.Tier()))
		if err != nil {
			panic(err.Error())
		}
		reward := tier.Reward()
		rewardPerDeposit, err := reward.finalize(currentHeight, tier.ActualDepositAmount(), tier.TierAmountPerBlockX128())
		if err != nil {
			panic(err.Error())
		}
		reward.addRewardPerDeposit(rewardPerDeposit)
		reward.setLastHeight(currentHeight)
		rewardAmount := reward.deductReward(depositId, currentHeight)
		if rewardAmount == 0 {
			continue
		}

		tier.userCollectedAmount += rewardAmount
		project.Stats().totalCollected += rewardAmount

		totalReward += rewardAmount

		std.Emit(
			"CollectRewardByProjectId",
			"prevAddr", caller.String(),
			"prevRealm", prevPkgPath,
			"projectId", projectId,
			"depositId", depositId,
			"amount", formatUint(rewardAmount),
		)
		project.setTier(convertTierTypeStrToUint64(deposit.Tier()), tier)

		// Update deposit
		deposit.setRewardCollected(deposit.RewardCollected() + rewardAmount)
		deposit.setRewardCollectHeight(currentHeight)
		deposit.setRewardCollectTime(uint64(time.Now().Unix()))
		deposits[depositId] = deposit
	}

	projects[projectId] = project

	if totalReward > 0 {
		tokenTeller := common.GetTokenTeller(project.tokenPath)
		tokenTeller.Transfer(caller, totalReward)
	}

	return totalReward
}

// CollectRewardByDepositId collects reward from certain deposit by caller
// Returns collected reward amount
// ref: https://docs.gnoswap.io/contracts/launchpad/launchpad_reward.gno#collectrewardbydepositid
func CollectRewardByDepositId(depositId string) uint64 {
	assertOnlyNotHalted()

	deposit, exists := deposits[depositId]
	if !exists {
		panic(ufmt.Sprintf("depositId(%s) not found", depositId))
	}

	currentHeight := uint64(std.ChainHeight())
	if err := validateRewardCollection(deposit, currentHeight); err != nil {
		panic(err.Error())
	}

	project, err := getProject(deposit.ProjectID())
	if err != nil {
		panic(err.Error())
	}

	en.MintAndDistributeGns()

	tier, err := project.Tier(convertTierTypeStrToUint64(deposit.Tier()))
	if err != nil {
		panic(err.Error())
	}

	reward := tier.Reward()
	if reward.GetLastHeight() == currentHeight {
		panic(addDetailToError(errAlreadyCollected, ufmt.Sprintf("depositId(%s)", depositId)))
	}

	rewardPerDeposit, err := reward.finalize(currentHeight, tier.ActualDepositAmount(), tier.TierAmountPerBlockX128())
	if err != nil {
		panic(err.Error())
	}
	reward.addRewardPerDeposit(rewardPerDeposit)
	reward.setLastHeight(currentHeight)

	rewardAmount := reward.deductReward(depositId, currentHeight)
	if rewardAmount == 0 {
		panic(addDetailToError(errAlreadyCollected, ufmt.Sprintf("depositId(%s)", depositId)))
	}

	tier.setReward(reward)
	tier.userCollectedAmount += rewardAmount
	project.Stats().totalCollected += rewardAmount

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"CollectRewardByDepositId",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"depositId", depositId,
		"amount", formatUint(rewardAmount),
	)

	project.setTier(convertTierTypeStrToUint64(deposit.Tier()), tier)

	// Update deposit
	deposit.setRewardCollected(deposit.RewardCollected() + rewardAmount)
	deposit.setRewardCollectHeight(currentHeight)
	deposit.setRewardCollectTime(uint64(time.Now().Unix()))
	deposits[depositId] = deposit

	projects[deposit.projectId] = project

	// Transfer reward
	if rewardAmount > 0 {
		tokenTeller := common.GetTokenTeller(project.tokenPath)
		tokenTeller.Transfer(getPrevAddr(), rewardAmount)
	}

	return rewardAmount
}

// validateRewardCollection validates if reward can be collected
func validateRewardCollection(deposit Deposit, height uint64) error {
	caller := getPrevAddr()
	if err := validateRewardClaimable(deposit, height); err != nil {
		return err
	}
	if err := validateRewardOwner(deposit.id, caller); err != nil {
		return err
	}
	return nil
}

// calculateTierRewards calculates rewards for each tier
func calculateTierRewards(tier Tier, currentHeight uint64, lastCalcHeight uint64) (*u256.Uint, uint64, error) {
	if tier.tierAmountPerBlockX128 == nil {
		return u256.Zero(), 0, nil
	}

	if tier.tierAmountPerBlockX128.IsZero() {
		return u256.Zero(), 0, nil
	}

	endHeight := minU64(currentHeight, tier.ended.height)
	sinceLast := endHeight - minU64(endHeight, lastCalcHeight)

	if sinceLast == 0 {
		return u256.Zero(), 0, nil
	}

	rewardX128 := new(u256.Uint).Mul(tier.tierAmountPerBlockX128.NilToZero().Clone(), u256.NewUint(sinceLast))
	reward := new(u256.Uint).Div(rewardX128, q128).Uint64()

	return rewardX128, reward, nil
}

// calcDepositRatioX128 calculates the deposit ratio with Q128 precision
func calcDepositRatioX128(tierAmount uint64, amount uint64) *u256.Uint {
	amountX128 := new(u256.Uint).Mul(u256.NewUint(amount), q128.Clone())
	amountX128x := new(u256.Uint).Mul(amountX128, u256.NewUint(1_000_000_000))

	tierAmountX128 := new(u256.Uint).Mul(u256.NewUint(tierAmount), q128.Clone())

	depositRatioX96 := new(u256.Uint).Div(amountX128x, tierAmountX128)
	depositRatioX96 = new(u256.Uint).Mul(depositRatioX96, q128.Clone())
	depositRatioX96 = new(u256.Uint).Div(depositRatioX96, u256.NewUint(1_000_000_000))
	return depositRatioX96
}

// calcProjectTiersRewardPerBlockX96 calculates the reward per block with Q96 precision
func calcProjectTiersRewardPerBlockX128(tier Tier) *u256.Uint {
	tierAmountX128 := new(u256.Uint).Mul(u256.NewUint(tier.tierAmount), q128)
	return new(u256.Uint).Div(tierAmountX128, u256.NewUint(tier.ended.height-tier.started.height))
}

// calcRewardPerBlockX128 calculates the reward per block with Q128 precision
func calcRewardPerBlockX128(tierAmount uint64, duration uint64) *u256.Uint {
	tierAmountX128 := new(u256.Uint).Mul(u256.NewUint(tierAmount), q128)
	return new(u256.Uint).Div(tierAmountX128, u256.NewUint(duration))
}

// validateRewardOwner checks if the deposit is truly owned by `caller`.
func validateRewardOwner(depositId string, caller std.Address) error {
	dep, ok := deposits[depositId]
	if !ok {
		return errors.New(addDetailToError(errNotExistDeposit, ufmt.Sprintf("depositId(%s)", depositId)))
	}
	projMap, userHasProject := depositsByUserByProject[caller]
	if !userHasProject {
		return errors.New(addDetailToError(errInvalidCaller, ufmt.Sprintf("caller(%s) not found", caller)))
	}
	depositIds, userHasDeposits := projMap[dep.projectId]
	if !userHasDeposits {
		return errors.New(addDetailToError(errNotExistDeposit, ufmt.Sprintf("caller(%s) no deposit for project(%s)", caller, dep.projectId)))
	}
	if !containsString(depositIds, depositId) {
		return errors.New(addDetailToError(errInvalidOwner, ufmt.Sprintf("caller(%s) does not own depositId(%s)", caller, depositId)))
	}
	return nil
}

// validateRewardClaimable checks if reward can be claimed
// e.g. deposit.claimableHeight, deposit.claimableTime
func validateRewardClaimable(deposit Deposit, currentHeight uint64) error {
	if deposit.rewardCollectTime == 0 {
		if currentHeight < deposit.claimableHeight {
			return errors.New(addDetailToError(errNotYetClaimReward, ufmt.Sprintf("need block >= %d, current=%d", deposit.claimableHeight, currentHeight)))
		}
	}
	return nil
}
