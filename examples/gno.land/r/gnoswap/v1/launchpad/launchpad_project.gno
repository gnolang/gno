package launchpad

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/emission"
	"gno.land/r/gnoswap/v1/gns"
)

var (
	// projectId -> Project
	projects *avl.Tree

	// projectTierId -> RewardManager
	projectTierRewardManagers *avl.Tree
)

func init() {
	projects = avl.NewTree()
	projectTierRewardManagers = avl.NewTree()
}

// CreateProject creates a new project with specified tiers, conditions, and token deposit.
//
// This function initializes a project by validating the input, transferring the deposit amount,
// and creating the project's tiers. The project is stored globally and an event is emitted.
//
// Parameters:
// - name (string): The name of the project.
// - tokenPath (string): The token path for the deposit token.
// - recipient (std.Address): The address to receive the project's rewards.
// - depositAmount (uint64): The total amount of tokens to be deposited for the project.
// - conditionsToken (string): The token used for reward conditions.
// - conditionsAmount (string): The amount of the conditions token required.
// - tier30Ratio (uint64): The percentage of the deposit allocated to the 30-day tier.
// - tier90Ratio (uint64): The percentage of the deposit allocated to the 90-day tier.
// - tier180Ratio (uint64): The percentage of the deposit allocated to the 180-day tier.
// - startTime (uint64): The project's start time in Unix timestamp (seconds), must be greater than now + 1 hour.
//
// Returns:
// - string: The unique ID of the created project.
func CreateProject(
	cur realm,
	name string,
	tokenPath string,
	recipient std.Address,
	depositAmount int64,
	conditionTokens string,
	conditionAmounts string,
	tier30Ratio int64,
	tier90Ratio int64,
	tier180Ratio int64,
	startTime int64,
) string {
	previousAddress := std.PreviousRealm().Address()
	previousPkgPath := std.PreviousRealm().PkgPath()
	launchpadAddr := std.CurrentRealm().Address()

	assertOnlyNotHalted()
	assertOnlyAdmin(previousAddress)

	averageBlockTimeMs := gns.GetAvgBlockTimeInMs()
	currentHeight := std.ChainHeight()
	currentTime := time.Now().Unix()

	params := &createProjectParams{
		name:                  name,
		tokenPath:             tokenPath,
		recipient:             recipient,
		depositAmount:         depositAmount,
		conditionTokens:       conditionTokens,
		conditionAmounts:      conditionAmounts,
		tier30Ratio:           tier30Ratio,
		tier90Ratio:           tier90Ratio,
		tier180Ratio:          tier180Ratio,
		startTime:             startTime,
		currentTime:           currentTime,
		currentHeight:         currentHeight,
		averageBlockTimeMs:    averageBlockTimeMs,
		minimumStartDelayTime: projectMinimumStartDelayTime,
	}

	project, err := createProject(params)
	if err != nil {
		panic(err)
	}

	tokenBalance := common.BalanceOf(tokenPath, previousAddress)
	if tokenBalance < depositAmount {
		panic(makeErrorWithDetails(errInsufficientBalance, ufmt.Sprintf("caller(%s) balance(%d) < depositAmount(%d)", previousAddress.String(), tokenBalance, depositAmount)))
	}

	err = common.TransferFrom(
		cross,
		tokenPath,
		previousAddress,
		launchpadAddr,
		depositAmount,
	)
	if err != nil {
		panic(err)
	}

	std.Emit(
		"CreateProject",
		"prevAddr", previousAddress.String(),
		"prevRealm", previousPkgPath,
		"name", name,
		"tokenPath", tokenPath,
		"recipient", recipient.String(),
		"depositAmount", formatInt(depositAmount),
		"conditionsToken", params.conditionTokens,
		"conditionsAmount", params.conditionAmounts,
		"tier30Ratio", formatInt(params.tier30Ratio),
		"tier90Ratio", formatInt(params.tier90Ratio),
		"tier180Ratio", formatInt(params.tier180Ratio),
		"startTime", formatInt(params.startTime),
		"startHeight", formatInt(project.StartHeight()),
		"projectId", project.ID(),
		"tier30Amount", formatInt(project.tiers[projectTier30].TotalDepositAmount()),
		"tier30EndHeight", formatInt(project.tiers[projectTier30].EndHeight()),
		"tier90Amount", formatInt(project.tiers[projectTier90].TotalDepositAmount()),
		"tier90EndHeight", formatInt(project.tiers[projectTier90].EndHeight()),
		"tier180Amount", formatInt(project.tiers[projectTier180].TotalDepositAmount()),
		"tier180EndHeight", formatInt(project.tiers[projectTier180].EndHeight()),
	)

	return project.ID()
}

// createProject creates a new project with the given parameters.
//
// This function validates the input parameters, creates the project structure,
// and sets up the project tiers and reward managers.
//
// Parameters:
// - params (createProjectParams): The parameters for creating the project.
//
// Returns:
// - *Project: The created project.
// - error: If the project creation fails, returns an error.
func createProject(params *createProjectParams) (*Project, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}

	createdBlockTimeInfo := newBlockTimeInfoByTimeAndAverageBlockTimeMs(
		params.currentTime,
		params.currentHeight,
		params.currentTime,
		params.averageBlockTimeMs,
	)

	// create project
	project := NewProject(
		params.name,
		params.tokenPath,
		params.depositAmount,
		params.recipient,
		createdBlockTimeInfo,
	)

	// check duplicate project
	if projects.Has(project.ID()) {
		return nil, makeErrorWithDetails(
			errDuplicateProject,
			ufmt.Sprintf("project(%s) already exists", project.ID()),
		)
	}

	projectConditions := NewProjectConditions(params.conditionTokens, params.conditionAmounts)
	for _, condition := range projectConditions {
		project.addProjectCondition(condition.tokenPath, condition)
	}

	projectTierRatios := map[int64]int64{
		projectTier30:  params.tier30Ratio,
		projectTier90:  params.tier90Ratio,
		projectTier180: params.tier180Ratio,
	}

	accumulatedTierDistributeAmount := int64(0)

	for _, duration := range projectTierDurations {
		rewardCollectableDuration := projectTierRewardCollectableDuration[duration]
		tierDurationTime := projectTierDurationTimes[duration]
		tierDistributeAmount := params.depositAmount * projectTierRatios[duration] / 100
		accumulatedTierDistributeAmount += tierDistributeAmount

		// if the last tier, distribute the remaining amount
		if duration == projectTier180 {
			remainTierDistributeAmount := params.depositAmount - accumulatedTierDistributeAmount
			tierDistributeAmount += remainTierDistributeAmount
		}

		projectTier := NewProjectTier(
			project.ID(),
			rewardCollectableDuration,
			tierDistributeAmount,
			params.startTime,
			params.startTime+tierDurationTime,
			params.averageBlockTimeMs,
			params.currentTime,
			params.currentHeight,
		)
		project.addProjectTier(duration, projectTier)

		projectTierRewardManagers.Set(projectTier.ID(), NewRewardManager(
			projectTier.TotalDistributeAmount(),
			projectTier.StartHeight(),
			projectTier.EndHeight(),
			params.currentHeight,
		))
	}

	projects.Set(project.ID(), project)

	return project, nil
}

// TransferLeftFromProjectByAdmin transfers the remaining rewards of a project to a specified recipient.
//
// This function is called by an admin to transfer any unclaimed rewards from a project to a recipient address.
// It validates the project ID, checks the recipient conditions, calculates the remaining rewards, and performs the transfer.
//
// Parameters:
// - projectId (string): The unique identifier of the project.
// - recipient (std.Address): The recipient address to transfer the remaining rewards.
//
// Returns:
// - uint64: The amount of rewards transferred to the recipient.
func TransferLeftFromProjectByAdmin(cur realm, projectID string, recipient std.Address) int64 {
	previousAddress := std.PreviousRealm().Address()
	previousPkgPath := std.PreviousRealm().PkgPath()
	currentHeight := int64(std.ChainHeight())

	assertOnlyUser()
	assertOnlyNotHalted()
	assertOnlyAdmin(previousAddress)

	projectLeftReward, err := transferLeftFromProject(projectID, recipient, currentHeight)
	if err != nil {
		panic(err)
	}

	project, err := getProject(projectID)
	if err != nil {
		panic(err)
	}

	std.Emit(
		"TransferLeftFromProjectByAdmin",
		"prevAddr", previousAddress.String(),
		"prevRealm", previousPkgPath,
		"projectId", projectID,
		"recipient", recipient.String(),
		"tokenPath", project.tokenPath,
		"leftReward", formatInt(projectLeftReward),
		"tier30Full", formatInt(project.tiers[projectTier30].TotalDepositAmount()),
		"tier30Left", formatInt(project.tiers[projectTier30].calculateLeftReward()),
		"tier90Full", formatInt(project.tiers[projectTier90].TotalDepositAmount()),
		"tier90Left", formatInt(project.tiers[projectTier90].calculateLeftReward()),
		"tier180Full", formatInt(project.tiers[projectTier180].TotalDepositAmount()),
		"tier180Left", formatInt(project.tiers[projectTier180].calculateLeftReward()),
	)

	return projectLeftReward
}

// transferLeftFromProject transfers the remaining rewards of a project to a specified recipient.
//
// This function is called by an admin to transfer any unclaimed rewards from a project to a recipient address.
// It validates the project ID, checks the recipient conditions, calculates the remaining rewards, and performs the transfer.
//
// Parameters:
// - projectId (string): The unique identifier of the project.
// - recipient (std.Address): The recipient address to transfer the remaining rewards.
// - currentHeight (int64): The current height of the chain.
//
// Returns:
// - uint64: The amount of rewards transferred to the recipient.
func transferLeftFromProject(projectID string, recipient std.Address, currentHeight int64) (int64, error) {
	project, err := getProject(projectID)
	if err != nil {
		return 0, err
	}

	if err := validateRefundProject(project, recipient, currentHeight); err != nil {
		return 0, err
	}

	emission.MintAndDistributeGns(cross)

	accumTotalDistributeAmount := int64(0)
	accumLeftReward := int64(0)
	accumCollectedReward := int64(0)

	tierMap := project.getTiers()
	for _, tier := range tierMap {
		if !tier.isEnded(currentHeight) {
			return 0, errors.New(ufmt.Sprintf("tier(%d) is not ended", tier.ID()))
		}

		if tier.CurrentDepositCount() > 0 {
			return 0, errors.New(ufmt.Sprintf("tier(%d) has (%d) participants", tier.ID(), tier.CurrentDepositCount()))
		}

		leftReward := tier.calculateLeftReward()
		accumLeftReward += leftReward
		accumCollectedReward += tier.TotalCollectedAmount()
		accumTotalDistributeAmount += tier.TotalDistributeAmount()
	}

	if accumTotalDistributeAmount != accumCollectedReward+accumLeftReward {
		return 0, errors.New(ufmt.Sprintf("accumTotalDistributeAmount(%d) != accumCollectedReward(%d)+accumLeftReward(%d)", accumTotalDistributeAmount, accumCollectedReward, accumLeftReward))
	}

	projectLeftReward := project.RemainingAmount()

	if projectLeftReward > 0 {
		if err := common.Transfer(cross, project.tokenPath, recipient, int64(projectLeftReward)); err != nil {
			return 0, makeErrorWithDetails(errTransferFailed, ufmt.Sprintf("token(%s), amount(%d)", project.tokenPath, projectLeftReward))
		}
	}

	return projectLeftReward, nil
}

// validateTransferLeft validates the transfer of remaining tokens
func validateRefundProject(project *Project, recipient std.Address, height int64) error {
	if !recipient.IsValid() {
		return errors.New(ufmt.Sprintf("invalid recipient address(%s)", recipient.String()))
	}

	return project.validateRefundRemainingAmount(height)
}

func getProject(projectID string) (*Project, error) {
	project, ok := projects.Get(projectID)
	if !ok {
		return nil, makeErrorWithDetails(errDataNotFound, ufmt.Sprintf("project(%s) not found", projectID))
	}

	return project.(*Project), nil
}

func getProjectTier(projectID string, tierDuration int64) (*ProjectTier, error) {
	project, err := getProject(projectID)
	if err != nil {
		return nil, err
	}

	tier, ok := project.tiers[tierDuration]
	if !ok {
		return nil, makeErrorWithDetails(errDataNotFound, ufmt.Sprintf("tier(%d) not found", tierDuration))
	}

	return tier, nil
}

func getProjectTierRewardManager(projectTierID string) (*RewardManager, error) {
	rewardManager, ok := projectTierRewardManagers.Get(projectTierID)
	if !ok {
		return nil, makeErrorWithDetails(errDataNotFound, ufmt.Sprintf("reward manager(%s) not found", projectTierID))
	}

	return rewardManager.(*RewardManager), nil
}

type createProjectParams struct {
	name                  string
	tokenPath             string
	recipient             std.Address
	depositAmount         int64
	conditionTokens       string
	conditionAmounts      string
	tier30Ratio           int64
	tier90Ratio           int64
	tier180Ratio          int64
	startTime             int64
	currentTime           int64
	currentHeight         int64
	averageBlockTimeMs    int64
	minimumStartDelayTime int64
}

func (p *createProjectParams) validate() error {
	if err := p.validateName(); err != nil {
		return err
	}

	if err := p.validateTokenPath(); err != nil {
		return err
	}

	if err := p.validateRecipient(); err != nil {
		return err
	}

	if err := p.validateDepositAmount(); err != nil {
		return err
	}

	if err := p.validateRatio(); err != nil {
		return err
	}

	if err := p.validateStartTime(p.currentTime, p.minimumStartDelayTime); err != nil {
		return err
	}

	if err := p.validateConditions(); err != nil {
		return err
	}

	return nil
}

// ValidateName checks if the project name is valid.
//
// Returns:
// - error: If the name is empty, returns an error.
func (p *createProjectParams) validateName() error {
	if p.name == "" {
		return makeErrorWithDetails(errInvalidInput, "project name cannot be empty")
	}

	if len(p.name) > 100 {
		return makeErrorWithDetails(errInvalidInput, "project name is too long")
	}

	return nil
}

// ValidateTokenPath validates the token path.
//
// Ensures that the token path is not empty and is registered.
//
// Returns:
// - error: If the token path is invalid or unregistered, returns an error.
func (p *createProjectParams) validateTokenPath() error {
	if p.tokenPath == "" {
		return makeErrorWithDetails(errInvalidInput, "tokenPath cannot be empty")
	}

	if err := common.IsRegistered(p.tokenPath); err != nil && !isGovernanceToken(p.tokenPath) {
		return makeErrorWithDetails(errInvalidInput, ufmt.Sprintf("tokenPath(%s) not registered", p.tokenPath))
	}

	return nil
}

// ValidateRecipient checks if the recipient address is valid.
//
// Returns:
// - error: If the recipient address is invalid, returns an error.
func (p *createProjectParams) validateRecipient() error {
	if !p.recipient.IsValid() {
		return makeErrorWithDetails(errInvalidAddress, ufmt.Sprintf("recipient address(%s)", p.recipient.String()))
	}

	return nil
}

// ValidateDepositAmount ensures that the deposit amount is greater than zero.
//
// Returns:
// - error: If the deposit amount is zero, returns an error.
func (p *createProjectParams) validateDepositAmount() error {
	if p.depositAmount == 0 {
		return makeErrorWithDetails(errInvalidInput, "deposit amount cannot be 0")
	}

	if p.depositAmount < 0 {
		return makeErrorWithDetails(errInvalidInput, "deposit amount cannot be negative")
	}

	return nil
}

// ValidateRatio checks if the sum of the tier ratios equals 100.
//
// Returns:
// - error: If the sum of `Tier30Ratio`, `Tier90Ratio`, and `Tier180Ratio` is not 100, returns an error.
func (p *createProjectParams) validateRatio() error {
	sum := p.tier30Ratio + p.tier90Ratio + p.tier180Ratio
	if sum != 100 {
		return makeErrorWithDetails(
			errInvalidInput,
			ufmt.Sprintf("invalid ratio, sum of all tiers(30:%d, 90:%d, 180:%d) should be 100", p.tier30Ratio, p.tier90Ratio, p.tier180Ratio),
		)
	}

	return nil
}

// ValidateStartTime checks if the start time is available.
//
// Parameters:
// - now (uint64): The current time in Unix timestamp.
// - minimumStartDelayTime (uint64): The minimum start delay time in Unix timestamp. (default: 1 hour)
//
// Returns:
// - error: If the start time is in the past or equal to `now`, returns an error.
func (p *createProjectParams) validateStartTime(now int64, minimumStartDelayTime int64) error {
	availableStartTime := now + minimumStartDelayTime

	if p.startTime < availableStartTime {
		return makeErrorWithDetails(errInvalidInput, ufmt.Sprintf("start time(%d) must be greater than now(%d)", p.startTime, availableStartTime))
	}

	return nil
}

func (p *createProjectParams) validateConditions() error {
	if p.conditionTokens == "" && p.conditionAmounts == "" {
		return nil
	}

	tokenPaths := strings.Split(p.conditionTokens, stringSplitterPad)
	minimumAmounts := strings.Split(p.conditionAmounts, stringSplitterPad)

	if len(tokenPaths) != len(minimumAmounts) {
		return makeErrorWithDetails(errInvalidInput, ufmt.Sprintf("conditionTokens and conditionAmounts are not matched"))
	}

	tokenPathMap := make(map[string]bool)

	for _, tokenPath := range tokenPaths {
		err := common.IsRegistered(tokenPath)
		if err != nil && !isGovernanceToken(tokenPath) {
			return makeErrorWithDetails(errInvalidInput, ufmt.Sprintf("tokenPath(%s) not registered", tokenPath))
		}

		if tokenPathMap[tokenPath] {
			return makeErrorWithDetails(errInvalidInput, ufmt.Sprintf("tokenPath(%s) is duplicated", tokenPath))
		}

		tokenPathMap[tokenPath] = true
	}

	for _, amountStr := range minimumAmounts {
		minimumAmount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return makeErrorWithDetails(errInvalidInput, ufmt.Sprintf("invalid condition amount(%s)", amountStr))
		}

		if minimumAmount <= 0 {
			return makeErrorWithDetails(errInvalidInput, ufmt.Sprintf("condition amount(%s) is not available", amountStr))
		}
	}

	return nil
}

func isGovernanceToken(tokenPath string) bool {
	return tokenPath == consts.GOV_XGNS_PATH
}
