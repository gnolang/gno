package launchpad

import (
	"gno.land/p/demo/ufmt"

	u256 "gno.land/p/gnoswap/uint256"
)

// ProjectTier represents a tier within a project.
//
// This struct contains the necessary data and methods to manage and distribute
// rewards for a specific tier of a project.
//
// Fields:
// - distributeAmountPerBlockX128 (u256.Uint): The amount of tokens to be distributed per block, represented as a Q128 fixed-point number.
// - startBlockTimeInfo (BlockTimeInfo): The block time and height information for the start of the tier.
// - endBlockTimeInfo (BlockTimeInfo): The block time and height information for the end of the tier.
// - id (string): The unique identifier for the tier, formatted as "{projectID}:duration".
// - totalDistributeAmount (int64): The total amount of tokens to be distributed for the tier.
// - totalDepositAmount (int64): The total amount of tokens deposited for the tier.
// - totalWithdrawAmount (int64): The total amount of tokens withdrawn from the tier.
// - totalDepositCount (int64): The total number of deposits made to the tier.
// - totalWithdrawCount (int64): The total number of withdrawals from the tier.
// - totalCollectedAmount (int64): The total amount of tokens collected as rewards for the tier.
type ProjectTier struct {
	distributeAmountPerBlockX128 *u256.Uint     // distribute amount per block, Q128
	startBlockTimeInfo           *BlockTimeInfo // project tier start block time
	endBlockTimeInfo             *BlockTimeInfo // project tier end block time

	id                    string // '{projectId}:duration' // duartion == 30, 90, 180
	totalDistributeAmount int64
	totalDepositAmount    int64 // accumulated deposit amount
	totalWithdrawAmount   int64 // accumulated withdraw amount
	totalDepositCount     int64 // accumulated deposit count
	totalWithdrawCount    int64 // accumulated withdraw count
	totalCollectedAmount  int64 // total collected amount by user (reward)
}

func (t *ProjectTier) ID() string {
	return t.id
}

func (t *ProjectTier) TotalDistributeAmount() int64 {
	return t.totalDistributeAmount
}

func (t *ProjectTier) TotalCollectedAmount() int64 {
	return t.totalCollectedAmount
}

func (t *ProjectTier) TotalDepositAmount() int64 {
	return t.totalDepositAmount
}

func (t *ProjectTier) TotalWithdrawAmount() int64 {
	return t.totalWithdrawAmount
}

func (t *ProjectTier) TotalDepositCount() int64 {
	return t.totalDepositCount
}

func (t *ProjectTier) TotalWithdrawCount() int64 {
	return t.totalWithdrawCount
}

func (t *ProjectTier) CurrentDepositCount() int64 {
	return t.totalDepositCount - t.totalWithdrawCount
}

func (t *ProjectTier) CurrentDepositAmount() int64 {
	return t.totalDepositAmount - t.totalWithdrawAmount
}

func (t *ProjectTier) DistributeAmountPerBlockX128() *u256.Uint {
	return t.distributeAmountPerBlockX128
}

func (t *ProjectTier) isActivated(currentHeight int64) bool {
	return t.startBlockTimeInfo.BlockHeight() <= currentHeight && currentHeight < t.endBlockTimeInfo.BlockHeight()
}

func (t *ProjectTier) isEnded(currentHeight int64) bool {
	return t.endBlockTimeInfo.BlockHeight() < currentHeight
}

func (t *ProjectTier) isFirstDeposit() bool {
	return t.totalDepositCount == 0
}

func (t *ProjectTier) StartTime() int64 {
	return t.startBlockTimeInfo.BlockTime()
}

func (t *ProjectTier) StartHeight() int64 {
	return t.startBlockTimeInfo.BlockHeight()
}

func (t *ProjectTier) EndTime() int64 {
	return t.endBlockTimeInfo.BlockTime()
}

func (t *ProjectTier) EndHeight() int64 {
	return t.endBlockTimeInfo.BlockHeight()
}

func (t *ProjectTier) deposit(deposit *Deposit) {
	t.totalDepositAmount += deposit.DepositAmount()
	t.totalDepositCount++
}

func (t *ProjectTier) withdraw(deposit *Deposit) {
	t.totalWithdrawAmount += deposit.DepositAmount()
	t.totalWithdrawCount++
}

func (t *ProjectTier) updateStartBlockTimeInfo(blockTime, blockHeight int64) {
	t.startBlockTimeInfo.setBlockTime(blockTime)
	t.startBlockTimeInfo.setBlockHeight(blockHeight)
}

func (t *ProjectTier) updateEndBlockTimeInfo(blockTime, blockHeight int64) {
	t.endBlockTimeInfo.setBlockTime(blockTime)
	t.endBlockTimeInfo.setBlockHeight(blockHeight)
}

func (t *ProjectTier) calculateLeftReward() int64 {
	return t.totalDistributeAmount - t.totalCollectedAmount
}

func (t *ProjectTier) updateDistributeAmountPerBlock() {
	distributeBlockCount := t.endBlockTimeInfo.BlockHeight() - t.startBlockTimeInfo.BlockHeight()
	totalDistributeAmountX128 := new(u256.Uint).Mul(u256.NewUintFromInt64(t.totalDistributeAmount), q128.Clone())
	distributeAmountPerBlockX128 := new(u256.Uint).Div(totalDistributeAmountX128, u256.NewUintFromInt64(distributeBlockCount))

	t.distributeAmountPerBlockX128 = distributeAmountPerBlockX128
}

// NewProjectTier returns a pointer to a new ProjectTier with the given values.
func NewProjectTier(
	projectID string,
	tierDuration int64,
	totalDistributeAmount int64,
	startBlockTime,
	endBlockTime,
	averageBlockTimeMs,
	currentBlockTime,
	currentBlockHeight int64,
) *ProjectTier {
	startBlockTimeInfo := newBlockTimeInfoByTimeAndAverageBlockTimeMs(
		currentBlockTime,
		currentBlockHeight,
		startBlockTime,
		averageBlockTimeMs,
	)
	endBlockTimeInfo := newBlockTimeInfoByTimeAndAverageBlockTimeMs(
		currentBlockTime,
		currentBlockHeight,
		endBlockTime,
		averageBlockTimeMs,
	)

	tier := &ProjectTier{
		id:                           makeProjectTierID(projectID, tierDuration),
		totalDistributeAmount:        totalDistributeAmount,
		distributeAmountPerBlockX128: u256.Zero(),
		startBlockTimeInfo:           startBlockTimeInfo,
		endBlockTimeInfo:             endBlockTimeInfo,
		totalDepositAmount:           0,
		totalWithdrawAmount:          0,
		totalDepositCount:            0,
		totalWithdrawCount:           0,
		totalCollectedAmount:         0,
	}

	tier.updateDistributeAmountPerBlock()

	return tier
}

// makeProjectTierID generates a unique tier ID based on the given project ID and the tier duration.
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
func makeProjectTierID(projectID string, duration int64) string {
	return ufmt.Sprintf("%s:%d", projectID, duration)
}
