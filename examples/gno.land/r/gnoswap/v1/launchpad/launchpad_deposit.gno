package launchpad

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gns"
	gov_staker "gno.land/r/gnoswap/v1/gov/staker"
	"gno.land/r/gnoswap/v1/gov/xgns"
	"gno.land/r/gnoswap/v1/referral"
)

var (
	// depositId -> Deposit
	deposits = avl.NewTree()
)

// DepositGns handles the deposit process for a specific project tier.
//
// This function validates the deposit conditions, checks the activation status of the project and tier,
// and creates a new deposit for the caller. It updates the project, tier, and reward states accordingly.
//
// Parameters:
// - targetProjectTierId (string): The ID of the target project and tier in the format "{projectId}:{tierType}".
// - amount (uint64): The amount to deposit.
// - referrer (string): The referrer address.
//
// Returns:
// - string: The ID of the newly created deposit.
//
// Panics:
// - If the system is halted.
// - If the amount is invalid (e.g., zero).
// - If the project or tier is inactive.
// - If the deposit conditions are not satisfied.
// - If an error occurs during deposit creation or reward state update.
//
// Logic:
// 1. **Validation**:
//   - Check if the system is halted or the deposit amount is valid.
//   - Retrieve the project and validate its activation status.
//   - Validate deposit conditions for the caller.
//   - Retrieve and validate the tier's activation status.
//
// 2. **Deposit Creation**:
//   - Create a deposit using the project and tier information.
//   - Update deposit indices and tier details (e.g., `StartHeight`, `TierAmountPerBlockX128`).
//
// 3. **Reward Updates**:
//   - Update the tier's reward state, including reward per deposit and participant counts.
//   - Update the project's stats (e.g., total deposit amount, participant count).
//
// 4. **Transfer and Lock Tokens**:
//   - Transfer the deposit amount to the `launchpad` contract and mint `xGNS` tokens.
//
// 5. **Emit Events**:
//   - Emit events for both the deposit and the first deposit for the tier (if applicable).
//
// ref: https://docs.gnoswap.io/contracts/launchpad/launchpad_deposit.gno#depositgns
func DepositGns(cur realm, targetProjectTierID string, depositAmount int64, referrer string) string {
	previousAddress := std.PreviousRealm().Address()
	previousPkgPath := std.PreviousRealm().PkgPath()

	assertOnlyNotHalted()
	assertOnlyUser()
	assertValidAmount(depositAmount)

	projectID, tierDuration := parseProjectTierID(targetProjectTierID)
	deposit, rewardState, isFirstDeposit, distributeAmountPerBlockX128, err := depositGns(
		projectID,
		tierDuration,
		depositAmount,
		previousAddress,
	)
	if err != nil {
		panic(err.Error())
	}

	actualReferrer, success := registerReferral(referrer, previousAddress)
	if !success {
		actualReferrer = referral.GetReferral(std.PreviousRealm().Address().String())
	}

	if isFirstDeposit {
		std.Emit(
			"FirstDepositForProjectTier",
			"prevAddr", previousAddress.String(),
			"prevRealm", previousPkgPath,
			"targetProjectTierId", targetProjectTierID,
			"amount", formatInt(depositAmount),
			"depositId", deposit.ID(),
			"claimableHeight", formatInt(rewardState.ClaimableHeight()),
			"claimableTime", formatInt(rewardState.ClaimableTime()),
			"tierAmountPerBlockX128", distributeAmountPerBlockX128,
		)
	}

	std.Emit(
		"DepositGns",
		"prevAddr", previousAddress.String(),
		"prevRealm", previousPkgPath,
		"targetProjectTierId", targetProjectTierID,
		"amount", formatInt(depositAmount),
		"depositId", deposit.ID(),
		"claimableHeight", formatInt(rewardState.ClaimableHeight()),
		"claimableTime", formatInt(rewardState.ClaimableTime()),
		"referrer", actualReferrer,
	)

	launchpadAddress := std.CurrentRealm().Address()

	// stake governance token to the project
	err = stakeGovernance(projectID, depositAmount, launchpadAddress, previousAddress)
	if err != nil {
		panic(err.Error())
	}

	return deposit.ID()
}

// depositGns deposits GNS to a project tier
//
// Parameters:
// - projectID (string): The ID of the project.
// - tierDuration (int64): The duration of the tier.
// - depositAmount (int64): The amount of the deposit.
// - callerAddress (std.Address): The address of the caller.
//
// Returns:
// - *Deposit: The deposit created.
// - *RewardState: The reward state for the deposit.
// - bool: Whether this is the first deposit for the tier.
// - string: The distribute amount per block in Q128 format.
// - error: If the deposit creation fails, returns an error.
//
// Panics:
// - If the project is not found.
// - If the project tier is not found.
// - If the reward manager is not found.
// - If the deposit creation fails.
func depositGns(
	projectID string,
	tierDuration int64,
	depositAmount int64,
	callerAddress std.Address,
) (*Deposit, *RewardState, bool, string, error) {
	project, err := getProject(projectID)
	if err != nil {
		return nil, nil, false, "", err
	}

	balanceOfFn := func(tokenPath string, caller std.Address) int64 {
		if tokenPath == consts.GOV_XGNS_PATH {
			return xgns.BalanceOf(caller)
		}

		return common.BalanceOf(tokenPath, caller)
	}

	project.CheckConditions(callerAddress, balanceOfFn)

	projectTier, err := project.getTier(tierDuration)
	if err != nil {
		return nil, nil, false, "", err
	}

	currentHeight := std.ChainHeight()

	if !projectTier.isActivated(currentHeight) {
		return nil, nil, false, "", makeErrorWithDetails(errInactiveProject, projectID)
	}

	currentTime := time.Now().Unix()
	createdBlockTimeInfo := newBlockTimeInfo(currentTime, currentHeight)
	deposit := NewDeposit(
		projectID,
		tierDuration,
		callerAddress,
		depositAmount,
		createdBlockTimeInfo,
		projectTier.endBlockTimeInfo,
	)
	deposits.Set(deposit.ID(), deposit)

	projectTier.deposit(deposit)

	rewardManager, err := getProjectTierRewardManager(projectTier.ID())
	if err != nil {
		return nil, nil, false, "", err
	}

	isFirstDeposit := !rewardManager.IsInitialized()

	rewardState := rewardManager.addRewardStateByDeposit(deposit, currentHeight)

	err = rewardManager.updateRewardPerDepositX128(projectTier.CurrentDepositAmount(), currentHeight)
	if err != nil {
		return nil, nil, false, "", err
	}

	return deposit,
		rewardState,
		isFirstDeposit,
		rewardManager.DistributeAmountPerBlockX128().ToString(),
		nil
}

// registerReferral registers a referral for a caller
//
// Parameters:
// - referrer (string): The referrer address.
// - callerAddress (std.Address): The address of the caller.
//
// Returns:
// - string: The actual referrer address.
// - bool: Whether the referral registration was successful.
//
// Panics:
// - If the referral registration fails.
func registerReferral(referrer string, callerAddress std.Address) (string, bool) {
	success := referral.TryRegister(cross, callerAddress, referrer)
	actualReferrer := referrer
	if !success {
		actualReferrer = referral.GetReferral(callerAddress.String())
	}

	return actualReferrer, success
}

// stakeGovernance stakes governance token to the project
//
// Parameters:
// - projectID (string): The ID of the project.
// - depositAmount (int64): The amount of the deposit.
// - launchpadAddress (std.Address): The address of the launchpad.
// - callerAddress (std.Address): The address of the caller.
//
// Returns:
// - error: If the governance token staking fails, returns an error.
//
// Panics:
// - If the project is not found.
func stakeGovernance(projectID string, depositAmount int64, launchpadAddress std.Address, callerAddress std.Address) error {
	project, err := getProject(projectID)
	if err != nil {
		return err
	}

	gov_staker.SetAmountByProjectWallet(cross, project.Recipient(), uint64(depositAmount), true)

	gns.TransferFrom(
		cross,
		callerAddress,
		launchpadAddress,
		depositAmount,
	)

	xgns.Mint(cross, launchpadAddress, depositAmount)

	return nil
}

func getDeposit(depositID string) (*Deposit, error) {
	depositI, ok := deposits.Get(depositID)
	if !ok {
		return nil, makeErrorWithDetails(errNotExistDeposit, ufmt.Sprintf("(%s)", depositID))
	}

	return depositI.(*Deposit), nil
}
