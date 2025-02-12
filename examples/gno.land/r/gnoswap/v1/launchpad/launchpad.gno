package launchpad

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"
	"gno.land/r/gnoswap/v1/gns"
)

var (
	projects = make(map[string]Project) // projectId -> project

	// project tier should distribute project token if deposit ever happened
	// therefore we need to keep track of project tiers without deposit
	projectTiersWithoutDeposit = make(map[string]bool) // tierId -> true

	q128 = u256.Zero()
)

func init() {
	q128 = u256.MustFromDecimal(consts.Q128)
}

// ProjectInput defines the input parameters required to create a project.
//
// Fields:
// - Name (string): The name of the project.
// - TokenPath (string): The token path for the deposit token.
// - Recipient (std.Address): The recipient address for the project's rewards.
// - DepositAmount (uint64): The total deposit amount for the project.
// - ConditionsToken (string): A string containing token paths separated by `*PAD*`.
// - ConditionsAmount (string): A string containing corresponding amounts separated by `*PAD*`.
// - Tier30Ratio (uint64): The percentage of the deposit allocated to the 30-day tier.
// - Tier90Ratio (uint64): The percentage of the deposit allocated to the 90-day tier.
// - Tier180Ratio (uint64): The percentage of the deposit allocated to the 180-day tier.
// - StartTime (uint64): The start time of the project in Unix timestamp.
type ProjectInput struct {
	Name             string
	TokenPath        string
	Recipient        std.Address
	DepositAmount    uint64
	ConditionsToken  string // separated by PAD_SEP (`*PAD*`)
	ConditionsAmount string // separated by PAD_SEP (`*PAD*`)
	Tier30Ratio      uint64
	Tier90Ratio      uint64
	Tier180Ratio     uint64
	StartTime        uint64
}

// NewProjectInput creates and returns a new `ProjectInput` object.
//
// Parameters:
// - name (string): The name of the project.
// - tokenPath (string): The token path for the deposit token.
// - recipient (std.Address): The recipient address for the project's rewards.
// - depositAmount (uint64): The total deposit amount for the project.
// - conditionsToken (string): Token paths separated by `*PAD*`.
// - conditionsAmount (string): Corresponding amounts separated by `*PAD*`.
// - tier30Ratio (uint64): Allocation percentage for the 30-day tier.
// - tier90Ratio (uint64): Allocation percentage for the 90-day tier.
// - tier180Ratio (uint64): Allocation percentage for the 180-day tier.
// - startTime (uint64): The project's start time in Unix timestamp.
//
// Returns:
// - ProjectInput: The initialized `ProjectInput` object.
func NewProjectInput(
	name string,
	tokenPath string,
	recipient std.Address,
	depositAmount uint64,
	conditionsToken string,
	conditionsAmount string,
	tier30Ratio uint64,
	tier90Ratio uint64,
	tier180Ratio uint64,
	startTime uint64,
) ProjectInput {
	return ProjectInput{
		Name:             name,
		TokenPath:        tokenPath,
		Recipient:        recipient,
		DepositAmount:    depositAmount,
		ConditionsToken:  conditionsToken,
		ConditionsAmount: conditionsAmount,
		Tier30Ratio:      tier30Ratio,
		Tier90Ratio:      tier90Ratio,
		Tier180Ratio:     tier180Ratio,
		StartTime:        startTime,
	}
}

// ValidateName checks if the project name is valid.
//
// Returns:
// - error: If the name is empty, returns an error.
func (in ProjectInput) ValidateName() error {
	if in.Name == "" {
		return errors.New(addDetailToError(
			errInvalidInput, "project name cannot be empty"))
	}
	if len(in.Name) > 100 {
		return errors.New(addDetailToError(errInvalidInput, "project name is too long"))
	}
	return nil
}

// ValidateTokenPath validates the token path.
//
// Ensures that the token path is not empty and is registered.
//
// Returns:
// - error: If the token path is invalid or unregistered, returns an error.
func (in ProjectInput) ValidateTokenPath() error {
	if in.TokenPath == "" {
		return errors.New(addDetailToError(
			errInvalidInput, "tokenPath cannot be empty"))
	}
	if err := common.IsRegistered(in.TokenPath); err != nil {
		return errors.New(addDetailToError(
			errInvalidInput, ufmt.Sprintf("tokenPath(%s) not registered", in.TokenPath)))
	}
	return nil
}

// ValidateRecipient checks if the recipient address is valid.
//
// Returns:
// - error: If the recipient address is invalid, returns an error.
func (in ProjectInput) ValidateRecipient() error {
	if !in.Recipient.IsValid() {
		return errors.New(addDetailToError(
			errInvalidAddress, ufmt.Sprintf("recipient address(%s)", in.Recipient.String())))
	}
	return nil
}

// ValidateDepositAmount ensures that the deposit amount is greater than zero.
//
// Returns:
// - error: If the deposit amount is zero, returns an error.
func (in ProjectInput) ValidateDepositAmount() error {
	if in.DepositAmount == 0 {
		return errors.New(addDetailToError(
			errInvalidInput, "deposit amount cannot be 0"))
	}
	return nil
}

// ValidateRatio checks if the sum of the tier ratios equals 100.
//
// Returns:
// - error: If the sum of `Tier30Ratio`, `Tier90Ratio`, and `Tier180Ratio` is not 100, returns an error.
func (in ProjectInput) ValidateRatio() error {
	sum := in.Tier30Ratio + in.Tier90Ratio + in.Tier180Ratio
	if sum != 100 {
		return errors.New(addDetailToError(
			errInvalidInput, ufmt.Sprintf("invalid ratio, sum of all tiers(30:%d, 90:%d, 180:%d) should be 100",
				in.Tier30Ratio, in.Tier90Ratio, in.Tier180Ratio)))
	}
	return nil
}

// ValidateStartTime checks if the start time is in the future.
//
// Parameters:
// - now (uint64): The current time in Unix timestamp.
//
// Returns:
// - error: If the start time is in the past or equal to `now`, returns an error.
func (in ProjectInput) ValidateStartTime(now uint64) error {
	if in.StartTime <= now {
		return errors.New(addDetailToError(
			errInvalidInput, ufmt.Sprintf("start time(%d) must be greater than now(%d)", in.StartTime, now)))
	}
	return nil
}

// ParseConditions parses the conditions from `ConditionsToken` and `ConditionsAmount`.
//
// Splits the token and amount strings using `*PAD*`, validates each token path, and converts
// the amounts to integers. Returns a map of token paths to their corresponding conditions.
//
// Returns:
// - map[string]Condition: A map of token paths to `Condition` objects.
// - error: If the tokens and amounts are mismatched, invalid, or unregistered, returns an error.
func (in ProjectInput) ParseConditions() (map[string]Condition, error) {
	if in.ConditionsToken == "" || in.ConditionsAmount == "" {
		return nil, errors.New(addDetailToError(
			errInvalidInput, "conditionsToken or conditionsAmount cannot be empty"))
	}

	tokens := strings.Split(in.ConditionsToken, PAD_SEP)
	amounts := strings.Split(in.ConditionsAmount, PAD_SEP)
	if len(tokens) != len(amounts) {
		return nil, errors.New(addDetailToError(
			errInvalidLength, ufmt.Sprintf("invalid conditions(numTokens(%d) != numAmounts(%d))", len(tokens), len(amounts))))
	}

	condition := make(map[string]Condition)
	for i, token := range tokens {
		if token == "" {
			return nil, errors.New(addDetailToError(
				errInvalidInput, ufmt.Sprintf("invalid token(%s)", token)))
		}
		if err := common.IsRegistered(token); err != nil {
			return nil, errors.New(addDetailToError(
				errInvalidInput, ufmt.Sprintf("token(%s) not registered", token)))
		}
		if _, exists := condition[token]; exists {
			return nil, errors.New(addDetailToError(
				errInvalidInput, ufmt.Sprintf("duplicated condition token(%s)", token)))
		}

		amtStr := amounts[i]
		amtVal, err := strconv.ParseUint(amtStr, 10, 64)
		if err != nil {
			return nil, errors.New(addDetailToError(
				errConvertFail, err.Error()))
		}

		condition[token] = Condition{
			tokenPath: token,
			minAmount: amtVal,
		}
	}
	return condition, nil
}

// ValidateAll validates all the fields of the `ProjectInput`.
//
// Calls individual validation methods for the name, token path, recipient, deposit amount,
// tier ratios, and start time. Also parses the conditions.
//
// Parameters:
// - now (uint64): The current time in Unix timestamp.
//
// Returns:
// - map[string]Condition: The parsed conditions map.
// - error: If any validation fails, returns an error.
func (in ProjectInput) ValidateAll(now uint64) (map[string]Condition, error) {
	if err := in.ValidateName(); err != nil {
		return nil, err
	}
	if err := in.ValidateTokenPath(); err != nil {
		return nil, err
	}
	if err := in.ValidateRecipient(); err != nil {
		return nil, err
	}
	if err := in.ValidateDepositAmount(); err != nil {
		return nil, err
	}
	if err := in.ValidateRatio(); err != nil {
		return nil, err
	}
	if err := in.ValidateStartTime(now); err != nil {
		return nil, err
	}

	condMap, err := in.ParseConditions()
	if err != nil {
		return nil, err
	}
	return condMap, nil
}

// ConvertStartTimeToHeight converts the project's start time to a block height.
//
// Calculates the block height at which the project should start based on the average block time
// and the difference between the start time and the current time.
//
// Parameters:
// - now (uint64): The current time in Unix timestamp.
//
// Returns:
// - uint64: The calculated block height for the project's start.
// - error: If the average block time is unavailable or the start time is in the past, returns an error.
func (in ProjectInput) ConvertStartTimeToHeight(now uint64) (uint64, error) {
	avgBlockTimeMs := uint64(gns.GetAvgBlockTimeInMs())
	if avgBlockTimeMs <= 0 {
		panic(addDetailToError(
			errInvalidInput, "average block time not available"))
	}

	diffSec := int64(in.StartTime) - int64(now)
	if diffSec <= 0 {
		return 0, errors.New(addDetailToError(
			errInvalidInput, ufmt.Sprintf("invalid start time(%d), cannot start project in past(now:%d)",
				in.StartTime, now)))
	}

	diffBlockHeight := uint64(diffSec * 1000 / int64(avgBlockTimeMs))
	currentHeight := uint64(std.GetHeight())
	return currentHeight + diffBlockHeight, nil
}

// ProjectCalculationResult middle result of project params calculation
type ProjectCalculationResult struct {
	Tier30Amount  uint64
	Tier90Amount  uint64
	Tier180Amount uint64
	StartHeight   uint64
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
// - startTime (uint64): The project's start time in Unix timestamp (seconds).
//
// Returns:
// - string: The unique ID of the created project.
func CreateProject(
	name string,
	tokenPath string,
	recipient std.Address,
	depositAmount uint64,
	conditionsToken string,
	conditionsAmount string,
	tier30Ratio uint64,
	tier90Ratio uint64,
	tier180Ratio uint64,
	startTime uint64,
) string {
	assertOnlyNotHalted()
	assertOnlyAdmin()

	input := NewProjectInput(name, tokenPath, recipient, depositAmount, conditionsToken, conditionsAmount, tier30Ratio, tier90Ratio, tier180Ratio, startTime)
	now := uint64(time.Now().Unix())
	conditions, err := input.ValidateAll(now)
	if err != nil {
		panic(err.Error())
	}

	en.MintAndDistributeGns()

	projectId := generateProjectId(input.TokenPath)
	if _, exists := projects[projectId]; exists {
		panic(addDetailToError(
			errDuplicateProject,
			ufmt.Sprintf("project(%s) already exists", projectId)))
	}

	calcResult, err := calculateProjectParams(input, now)
	if err != nil {
		panic(addDetailToError(errInvalidInput, err.Error()))
	}

	startHeight := calcResult.StartHeight
	tier30 := createTier(projectId, 30, calcResult.Tier30Amount, startHeight, input.StartTime, convertTimeToHeight(TIMESTAMP_3DAYS))
	tier90 := createTier(projectId, 90, calcResult.Tier90Amount, startHeight, input.StartTime, convertTimeToHeight(TIMESTAMP_7DAYS))
	tier180 := createTier(projectId, 180, calcResult.Tier180Amount, startHeight, input.StartTime, convertTimeToHeight(TIMESTAMP_14DAYS))
	tierMap := map[uint64]Tier{
		TIER30:  tier30,
		TIER90:  tier90,
		TIER180: tier180,
	}

	tierRatioMap := map[uint64]uint64{
		TIER30:  input.Tier30Ratio,
		TIER90:  input.Tier90Ratio,
		TIER180: input.Tier180Ratio,
	}

	projectCreated := TimeInfo{
		height: uint64(std.GetHeight()),
		time:   now,
	}

	projectStarted := TimeInfo{
		height: startHeight,
		time:   input.StartTime,
	}

	projectEnded := TimeInfo{
		height: tier180.ended.height,
		time:   tier180.ended.time,
	}

	projectStats := NewProjectStats(0, 0, 0, 0, 0)

	refundInfo := NewRefundInfo(0, 0, 0)

	project := NewProject(
		projectId,
		name,
		tokenPath,
		depositAmount,
		recipient,
		conditions,
		tierMap,
		tierRatioMap,
		projectCreated,
		projectStarted,
		projectEnded,
		*projectStats,
		refundInfo,
	)

	projects[projectId] = *project

	// Transfer tokens
	tokenTeller := common.GetTokenTeller(tokenPath)
	tokenTeller.TransferFrom(
		getPrevAddr(),
		std.Address(GetOrigPkgAddr()),
		depositAmount,
	)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"CreateProject",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"name", name,
		"tokenPath", tokenPath,
		"recipient", recipient.String(),
		"depositAmount", formatUint(depositAmount),
		"conditionsToken", conditionsToken,
		"conditionsAmount", conditionsAmount,
		"tier30Ratio", formatUint(tier30Ratio),
		"tier90Ratio", formatUint(tier90Ratio),
		"tier180Ratio", formatUint(tier180Ratio),
		"startTime", formatUint(startTime),
		"startHeight", formatUint(startHeight),
		"projectId", projectId,
		"tier30Amount", formatUint(calcResult.Tier30Amount),
		"tier30EndHeight", formatUint(tier30.ended.height),
		"tier90Amount", formatUint(calcResult.Tier90Amount),
		"tier90EndHeight", formatUint(tier90.ended.height),
		"tier180Amount", formatUint(calcResult.Tier180Amount),
		"tier180EndHeight", formatUint(tier180.ended.height),
	)

	return projectId
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
func TransferLeftFromProjectByAdmin(projectId string, recipient std.Address) uint64 {
	if err := common.SatisfyCond(isUserCall()); err != nil {
		panic(addDetailToError(errNotUserCaller, err.Error()))
	}
	assertOnlyAdmin()
	assertOnlyNotHalted()

	project, err := getProject(projectId)
	if err != nil {
		panic(err.Error())
	}

	currentHeight := uint64(std.GetHeight())
	if err := validateTransferLeft(project, recipient, currentHeight); err != nil {
		panic(addDetailToError(
			errInvalidCondition, err.Error()))
	}

	en.MintAndDistributeGns()

	accumTotalReward := uint64(0)
	accumLeftReward := uint64(0)
	accumCollectedReward := uint64(0)
	tierMap := project.Tiers()
	for _, tier := range tierMap {
		if !tier.isEnded(currentHeight) {
			panic(addDetailToError(
				errActiveProject, ufmt.Sprintf("tier(%d) is not ended", tier.ID())))
		}
		if tier.ActualParticipant() > 0 {
			panic(addDetailToError(
				errTierHasParticipants, ufmt.Sprintf("tier(%d) has (%d) participants", tier.ID(), tier.ActualParticipant())))
		}
		leftReward := tier.calculateLeftReward()
		accumLeftReward += leftReward
		accumCollectedReward += tier.UserCollectedAmount()
		accumTotalReward += tier.TierAmount()
	}

	if accumTotalReward != accumCollectedReward+accumLeftReward {
		panic(addDetailToError(
			errInvalidRewardState, ufmt.Sprintf("accumTotalReward(%d) != accumCollectedReward(%d)+accumLeftReward(%d)", accumTotalReward, accumCollectedReward, accumLeftReward)))
	}

	projectLeftReward := project.calculateLeftReward()
	project.refund = RefundInfo{
		amount: projectLeftReward,
		height: currentHeight,
		time:   uint64(time.Now().Unix()),
	}

	projects[projectId] = project

	if projectLeftReward > 0 {
		tokenTeller := common.GetTokenTeller(project.tokenPath)
		tokenTeller.Transfer(recipient, projectLeftReward)
	}

	std.Emit(
		"TransferLeftFromProjectByAdmin",
		"projectId", projectId,
		"recipient", recipient.String(),
		"tokenPath", project.tokenPath,
		"leftReward", formatUint(projectLeftReward),
		"tier30Full", formatUint(project.tiers[30].tierAmount),
		"tier30Left", formatUint(project.tiers[30].tierAmount-project.tiers[30].calculatedAmount),
		"tier90Full", formatUint(project.tiers[90].tierAmount),
		"tier90Left", formatUint(project.tiers[90].tierAmount-project.tiers[90].calculatedAmount),
		"tier180Full", formatUint(project.tiers[180].tierAmount),
		"tier180Left", formatUint(project.tiers[180].tierAmount-project.tiers[180].calculatedAmount),
	)

	return projectLeftReward
}

// calculateProjectParams calculates project-related parameters based on the input and current time.
//
// This function computes the amounts for three tiers (30, 90, and 180), as well as the block height
// when the project starts based on the provided start time.
//
// Parameters:
// - input (ProjectInput): The input parameters for the project.
//   - DepositAmount (uint64): The total deposit amount.
//   - Tier30Ratio (uint64): The percentage of the deposit allocated to the 30-day tier.
//   - Tier90Ratio (uint64): The percentage of the deposit allocated to the 90-day tier.
//   - Tier180Ratio (uint64): The percentage of the deposit allocated to the 180-day tier.
//   - StartTime (uint64): The project start time in Unix timestamp (seconds).
//
// - now (uint64): The current time in Unix timestamp (seconds).
//
// Returns:
// - *ProjectCalculationResult: A pointer to the calculated results, including tier amounts and start height.
// - error: Returns an error if the start time is earlier than the current time or if the average block time is unavailable.
func calculateProjectParams(input ProjectInput, now uint64) (*ProjectCalculationResult, error) {
	if input.StartTime <= now {
		return nil, errors.New(addDetailToError(
			errInvalidInput, ufmt.Sprintf("start time(%d) must be greater than now(%d)", input.StartTime, now)))
	}

	tier30Amount := input.DepositAmount * input.Tier30Ratio / 100
	tier90Amount := input.DepositAmount * input.Tier90Ratio / 100
	tier180Amount := input.DepositAmount * input.Tier180Ratio / 100

	// calculate when the block starts
	timeUntilStart := input.StartTime - now
	blockDurationToStart := convertTimeToHeight(timeUntilStart)
	currentHeight := uint64(std.GetHeight())
	startHeight := currentHeight + blockDurationToStart

	return &ProjectCalculationResult{
		Tier30Amount:  tier30Amount,
		Tier90Amount:  tier90Amount,
		Tier180Amount: tier180Amount,
		StartHeight:   startHeight,
	}, nil
}

// generateProjectId generates a unique project ID based on the given token path and the current block height.
//
// The generated ID combines the `tokenPath` and the current block height in the following format:
// "{tokenPath}:{height}"
//
// Parameters:
// - tokenPath (string): The path of the token associated with the project.
//
// Returns:
// - string: A unique project ID in the format "tokenPath:height".
func generateProjectId(tokenPath string) string {
	// gno.land/r/gnoswap/gns:{height}
	// gno.land/r/gnoswap/gns:30
	return ufmt.Sprintf("%s:%d", tokenPath, std.GetHeight())
}

// generateTierId generates a unique tier ID based on the given project ID and the tier duration.
//
// The generated ID combines the `projectId` and the `duration` in the following format:
// "{projectId}:{duration}"
//
// Parameters:
// - projectId (string): The unique ID of the project associated with the tier.
// - duration (uint64): The duration of the tier (e.g., 30, 90, 180 days).
//
// Returns:
// - string: A unique tier ID in the format "projectId:duration".
func generateTierId(projectId string, duration uint64) string {
	// gno.land/r/gnoswap/gns:{height}:{duration}
	// gno.land/r/gnoswap/gns:30:30(90,180)
	return ufmt.Sprintf("%s:%d", projectId, duration)
}

// createTier creates a new tier with the given parameters
func createTier(
	projectId string,
	duration uint64,
	amount uint64,
	startHeight uint64,
	startTime uint64,
	collectWaitDuration uint64) Tier {

	tierId := generateTierId(projectId, duration)

	tierStart := TimeInfo{
		height: startHeight,
		time:   startTime,
	}

	durationSecond := duration * consts.SECONDS_PER_DAY
	endTime := startTime + durationSecond
	durationHeight := convertTimeToHeight(durationSecond)
	endHeight := startHeight + durationHeight
	tierEnded := TimeInfo{
		height: endHeight,
		time:   endTime,
	}

	rewardPerBlockX128 := calcRewardPerBlockX128(amount, durationHeight)

	// reward
	reward := NewReward(u256.Zero(), startHeight, endHeight)

	tier := Tier{
		id:                     tierId,
		collectWaitDuration:    collectWaitDuration,
		tierAmount:             amount,
		tierAmountPerBlockX128: rewardPerBlockX128,
		started:                tierStart,
		ended:                  tierEnded,
		totalDepositAmount:     0,
		actualDepositAmount:    0,
		totalParticipant:       0,
		actualParticipant:      0,
		userCollectedAmount:    0,
		calculatedAmount:       0,
		reward:                 *reward,
	}

	rewardState := NewRewardState(tier.tierAmountPerBlockX128, startHeight, endHeight)
	rewardStates.Set(projectId, strconv.FormatUint(duration, 10), rewardState)

	projectTiersWithoutDeposit[tierId] = true
	return tier
}

// validateTransferLeft validates the transfer of remaining tokens
func validateTransferLeft(project Project, recipient std.Address, height uint64) error {
	if !recipient.IsValid() {
		return errors.New(ufmt.Sprintf("invalid recipient address(%s)", recipient.String()))
	}

	if height < project.Ended().height {
		return errors.New(ufmt.Sprintf("project not ended yet(current:%d, endHeight: %d)",
			height, project.ended.height))
	}

	if project.Refund().height != 0 {
		return errors.New(ufmt.Sprintf("project already refunded(height:%d)",
			project.Refund().height))
	}

	return nil
}

// calculateLeftReward calculates the remaining reward amount for each tier
func calculateLeftReward(project Project) uint64 {
	tier30 := project.tiers[30]
	tier90 := project.tiers[90]
	tier180 := project.tiers[180]

	left30 := tier30.tierAmount - tier30.calculatedAmount
	left90 := tier90.tierAmount - tier90.calculatedAmount
	left180 := tier180.tierAmount - tier180.calculatedAmount
	return left30 + left90 + left180
}

// getProject returns the project with the given project ID
func getProject(projectId string) (Project, error) {
	project, exists := projects[projectId]
	if !exists {
		return Project{}, errors.New(addDetailToError(
			errDataNotFound, ufmt.Sprintf("projectId(%s) not found", projectId)))
	}
	return project, nil
}
