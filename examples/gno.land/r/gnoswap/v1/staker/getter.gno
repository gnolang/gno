package staker

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gns"
)

// GetPoolByPoolPath retrieves the pool's path using the provided pool path.
//
// Parameters:
// - poolPath (string): The path of the pool to retrieve.
//
// Returns:
// - string: The `poolPath` of the found pool.
//
// Panics:
// - If the pool corresponding to the given `poolPath` does not exist in the `pools` collection.
func GetPoolByPoolPath(poolPath string) *Pool {
	pool, ok := pools.Get(poolPath)
	if !ok {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("poolPath(%s) pool does not exist", poolPath)),
		)
	}

	return pool
}

// GetPoolIncentiveIdList returns the list of incentive IDs for a given pool
//
// Parameters:
//   - poolPath (string): The path of the pool to get incentives for
//
// Returns:
//   - A slice og incentive IDs associated with the pool
//
// Panics:
//   - If the pool incentives do not exist for the given pool path
func GetPoolIncentiveIdList(poolPath string) []string {
	pool := GetPoolByPoolPath(poolPath)

	ids := []string{}
	pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
		ids = append(ids, key)
		return true
	})

	return ids
}

// GetIncentive retrieves an external incentive by its ID from a specified pool.
//
// This function searches for an incentive within the specified pool's incentives,
// identified by the given `incentiveId`. If the pool or the incentive does not exist,
// the function panics with an appropriate error message.
//
// Parameters:
// - poolPath (string): The path of the pool containing the incentive.
// - incentiveId (string): The ID of the incentive to retrieve.
//
// Returns:
// - *ExternalIncentive: A pointer to the retrieved ExternalIncentive.
//
// Panics:
// - If the pool corresponding to the given `poolPath` does not exist.
// - If the incentive with the specified `incentiveId` does not exist.
func GetIncentive(poolPath string, incentiveId string) *ExternalIncentive {
	pool := GetPoolByPoolPath(poolPath)

	incentive, exist := pool.incentives.byTime.Get(incentiveId)
	if !exist {
		panic(ufmt.Sprintf("incentiveId(%s) incentive does not exist", incentiveId))
	}

	return incentive.(*ExternalIncentive)
}

// GetIncentiveStartTimestamp returns the start timestamp for a given incentive
//
// Parameters:
//   - incentiveId (string): The ID of the incentive
//
// Returns:
//   - int64: The start timestamp of the incentive
//
// Panics:
//   - If the incentive does not exist for the given incentiveId
func GetIncentiveStartTimestamp(poolPath string, incentiveId string) int64 {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.startTimestamp
}

// GetIncentiveEndTimestamp returns the end timestamp for a given incentive
//
// Parameters:
//   - incentiveId (string): The ID of the incentive
//
// Returns:
//   - int64: The end timestamp of the incentive
//
// Panics:
//   - If the incentive does not exist for the given incentiveId
func GetIncentiveEndTimestamp(poolPath string, incentiveId string) int64 {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.endTimestamp
}

// GetTargetPoolPathByIncentiveId returns the target pool path for a given incentive
//
// Parameters:
//   - incentiveId (string): The ID of the incentive
//
// Returns:
//   - The target pool path (string) associated with the incentive
//
// Panics:
//   - If the incentive does nor exist for the given incentive ID
func GetTargetPoolPathByIncentiveId(poolPath string, incentiveId string) string {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.targetPoolPath
}

// GetCreatedHeightOfIncentive retrieves the creation height of a specified incentive.
//
// Parameters:
// - poolPath (string): The path of the pool containing the incentive.
// - incentiveId (string): The ID of the incentive to retrieve the creation height for.
//
// Returns:
// - int64: The creation height of the specified incentive.
func GetCreatedHeightOfIncentive(poolPath string, incentiveId string) int64 {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.createdHeight
}

// GetIncentiveRewardToken returns the reward token for a given incentive
//
// Parameters:
//   - incentiveId (string): The ID of the incentive
//
// Returns:
//   - The reward token (string) associated with the incentive
//
// Panics:
//   - If the incentive does not exist for the given incentiveId
func GetIncentiveRewardToken(poolPath string, incentiveId string) string {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.rewardToken
}

// GetIncentiveRewardAmount returns the reward amount for a given incentive as a Uint256
//
// Parameters:
//   - incentiveId (string): The ID of the incentive
//
// Returns:
//   - *u256.Uint: The reward amount associated with the incentive
//
// Panics:
//   - If the incentive does not exist for the given incentiveId
func GetIncentiveRewardAmount(poolPath string, incentiveId string) *u256.Uint {
	incentive := GetIncentive(poolPath, incentiveId)

	return u256.NewUint(incentive.rewardAmount)
}

// GetIncentiveRewardAmountAsString returns the reward amount for a given incentive as a string
//
// Parameters:
//   - incentiveId (string): The ID of the incentive
//
// Returns:
//   - string: The reward amount associated with the incentive as a string
//
// Panics:
//   - If the incentive does not exist for the given incentiveId
func GetIncentiveRewardAmountAsString(poolPath string, incentiveId string) string {
	rewardAmount := GetIncentiveRewardAmount(incentiveId, poolPath)

	return rewardAmount.ToString()
}

// GetIncentiveStartHeight retrieves the start height of a specified incentive.
//
// This function looks up an incentive within the specified pool using the `poolPath` and `incentiveId`.
// It then returns the height at which the incentive starts.
//
// Parameters:
// - poolPath (string): The path of the pool containing the incentive.
// - incentiveId (string): The ID of the incentive to retrieve the start height for.
//
// Returns:
// - int64: The starting height of the specified incentive.
func GetIncentiveStartHeight(poolPath string, incentiveId string) int64 {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.startHeight
}

// GetIncentiveEndHeight retrieves the end height of a specified incentive.
//
// This function looks up an incentive within the specified pool using the `poolPath` and `incentiveId`.
// It then returns the height at which the incentive ends.
//
// Parameters:
// - poolPath (string): The path of the pool containing the incentive.
// - incentiveId (string): The ID of the incentive to retrieve the end height for.
//
// Returns:
// - int64: The ending height of the specified incentive.
func GetIncentiveEndHeight(poolPath string, incentiveId string) int64 {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.endHeight
}

// GetIncentiveRewardPerBlock retrieves the reward per block of a specified incentive.
//
// This function looks up an incentive within the specified pool using the `poolPath` and `incentiveId`.
// It then returns the reward amount distributed per block for the incentive.
//
// Parameters:
// - poolPath (string): The path of the pool containing the incentive.
// - incentiveId (string): The ID of the incentive to retrieve the reward per block for.
//
// Returns:
// - uint64: The reward amount distributed per block for the specified incentive.
func GetIncentiveRewardPerBlock(poolPath string, incentiveId string) uint64 {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.rewardPerBlock
}

// GetIncentiveRefundee returns the refundee address for a given incentive
//
// Parameters:
//   - incentiveId (string): The ID of the incentive
//
// Returns:
//   - std.Address: The refundee address of the incentive
//
// Panics:
//   - If the incentive does not exist for the given incentiveId
func GetIncentiveRefundee(poolPath string, incentiveId string) std.Address {
	incentive := GetIncentive(poolPath, incentiveId)

	return incentive.refundee
}

// GetDeposit retrieves a deposit associated with a given LP token ID.
//
// This function looks up the deposit record for the specified LP token ID in the global `deposits` collection.
// If the deposit does not exist, it panics with an appropriate error message.
//
// Parameters:
// - lpTokenId (uint64): The unique identifier of the LP token for which the deposit is retrieved.
//
// Returns:
// - *Deposit: A pointer to the `Deposit` object associated with the given LP token ID.
func GetDeposit(lpTokenId uint64) *Deposit {
	deposit := deposits.Get(lpTokenId)
	if deposit == nil {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("lpTokenId(%d) deposit does not exist", lpTokenId)),
		)
	}

	return deposit
}

// GetDepositOwner returns the owner address of a deposit for a given LP token ID
//
// Parameters:
//   - lpTokenId (uint64): The ID of the LP token
//
// Returns:
//   - std.Address: The owner address of the deposit
//
// Panics:
//   - If the deposit does not exist for the given lpTokenId
func GetDepositOwner(lpTokenId uint64) std.Address {
	deposit := GetDeposit(lpTokenId)

	return deposit.owner
}

// GetDepositStakeTimestamp returns the stake timestamp for a given LP token ID
//
// Parameters:
//   - lpTokenId (uint64): The ID of the LP token
//
// Returns:
//   - int64: The stake timestamp of the deposit
//
// Panics:
//   - If the deposit does not exist for the given lpTokenId
func GetDepositStakeTimestamp(lpTokenId uint64) int64 {
	deposit := GetDeposit(lpTokenId)

	return deposit.stakeTimestamp
}

// GetDepositStakeHeight retrieves the stake height of a deposit associated with a given LP token ID.
//
// This function looks up the deposit record for the specified LP token ID and returns the height
// at which the deposit was staked. It relies on the `GetDeposit` function to fetch the deposit.
// If the deposit does not exist, the function panics.
//
// Parameters:
// - lpTokenId (uint64): The unique identifier of the LP token for which the stake height is retrieved.
//
// Returns:
// - int64: The stake height of the deposit associated with the given LP token ID.
func GetDepositStakeHeight(lpTokenId uint64) int64 {
	deposit := GetDeposit(lpTokenId)

	return deposit.stakeHeight
}

// GetDepositTargetPoolPath returns the target pool path for a given LP token ID
//
// Parameters:
//   - lpTokenId (uint64): The ID of the LP token
//
// Returns:
//   - string: The target pool path of the deposit
//
// Panics:
//   - If the deposit does not exist for the given lpTokenId
func GetDepositTargetPoolPath(lpTokenId uint64) string {
	deposit := GetDeposit(lpTokenId)

	return deposit.targetPoolPath
}

// GetDepositTickLower retrieves the lower tick of a deposit associated with a given LP token ID.
//
// This function looks up the deposit record for the specified LP token ID and returns
// the lower tick value of the deposit. If the deposit does not exist, the function panics.
//
// Parameters:
// - lpTokenId (uint64): The unique identifier of the LP token for which the lower tick is retrieved.
//
// Returns:
// - int32: The lower tick value of the deposit.
func GetDepositTickLower(lpTokenId uint64) int32 {
	deposit := GetDeposit(lpTokenId)

	return deposit.tickLower
}

// GetDepositTickUpper retrieves the upper tick of a deposit associated with a given LP token ID.
//
// This function looks up the deposit record for the specified LP token ID and returns
// the upper tick value of the deposit. If the deposit does not exist, the function panics.
//
// Parameters:
// - lpTokenId (uint64): The unique identifier of the LP token for which the upper tick is retrieved.
//
// Returns:
// - int32: The upper tick value of the deposit.
func GetDepositTickUpper(lpTokenId uint64) int32 {
	deposit := GetDeposit(lpTokenId)

	return deposit.tickUpper
}

// GetDepositLiquidity retrieves the liquidity of a deposit associated with a given LP token ID.
//
// This function looks up the deposit record for the specified LP token ID and returns
// the liquidity value as a `*u256.Uint` object. If the deposit does not exist, the function panics.
//
// Parameters:
// - lpTokenId (uint64): The unique identifier of the LP token for which the liquidity is retrieved.
//
// Returns:
// - *u256.Uint: The liquidity value of the deposit.
func GetDepositLiquidity(lpTokenId uint64) *u256.Uint {
	deposit := GetDeposit(lpTokenId)

	return deposit.liquidity
}

// GetDepositLiquidityAsString retrieves the liquidity of a deposit associated with a given LP token ID
// and returns it as a string.
//
// This function looks up the deposit record for the specified LP token ID, retrieves its liquidity,
// and converts it to a string. If the deposit does not exist, the function panics.
//
// Parameters:
// - lpTokenId (uint64): The unique identifier of the LP token for which the liquidity is retrieved.
//
// Returns:
// - string: The liquidity value of the deposit as a string.
func GetDepositLiquidityAsString(lpTokenId uint64) string {
	liquidity := GetDepositLiquidity(lpTokenId)

	return liquidity.ToString()
}

// GetDepositLastCollectHeight retrieves the last collection height of a deposit associated with a given LP token ID.
//
// This function looks up the deposit record for the specified LP token ID and returns
// the last height at which rewards were collected. If the deposit does not exist, the function panics.
//
// Parameters:
// - lpTokenId (uint64): The unique identifier of the LP token for which the last collection height is retrieved.
//
// Returns:
// - uint64: The last collection height of the deposit.
func GetDepositLastCollectHeight(lpTokenId uint64) int64 {
	deposit := GetDeposit(lpTokenId)

	return deposit.lastCollectHeight
}

// GetDepositWarmUp retrieves the warm-up records of a deposit associated with a given LP token ID.
//
// This function looks up the deposit record for the specified LP token ID and returns
// the list of warm-up records associated with the deposit. If the deposit does not exist, the function panics.
//
// Parameters:
// - lpTokenId (uint64): The unique identifier of the LP token for which the warm-up records are retrieved.
//
// Returns:
// - []Warmup: A slice of warm-up records associated with the deposit.
func GetDepositWarmUp(lpTokenId uint64) []Warmup {
	deposit := GetDeposit(lpTokenId)

	return deposit.warmups
}

// GetPoolTier returns the tier of a given pool
//
// Parameters:
//   - poolPath (string): The path of the pool
//
// Returns:
//   - uint64: The tier of the pool
//
// Panics:
//   - If the pool tier does not exist for the given poolPath
func GetPoolTier(poolPath string) uint64 {
	return poolTier.CurrentTier(poolPath)
}

// GetPoolTierRatio retrieves the current reward ratio for the specified pool tier.
//
// This function calculates the reward ratio of a pool by determining its tier
// using `GetPoolTier` and then fetching the tier's current ratio based on the current block height.
//
// Parameters:
// - poolPath (string): The path of the pool whose tier ratio is retrieved.
//
// Returns:
// - uint64: The current reward ratio for the specified pool tier.
func GetPoolTierRatio(poolPath string) uint64 {
	tier := GetPoolTier(poolPath)
	return poolTier.tierRatio.Get(tier)
}

// GetPoolTierCount retrieves the current count of pools in the specified tier.
//
// This function checks the number of pools in the given tier based on the current block height.
// If the tier is `0`, it returns `0`.
//
// Parameters:
// - tier (uint64): The tier for which the pool count is retrieved.
//
// Returns:
// - uint64: The current count of pools in the specified tier.
func GetPoolTierCount(tier uint64) uint64 {
	if tier == 0 {
		return 0
	}
	return uint64(poolTier.CurrentCount(tier))
}

// GetPoolReward retrieves the current reward amount for the specified pool tier.
//
// This function fetches the reward for the given tier based on the current block height.
//
// Parameters:
// - tier (uint64): The tier for which the reward is retrieved.
//
// Returns:
// - uint64: The current reward amount for the specified tier.
func GetPoolReward(tier uint64) uint64 {
	return poolTier.CurrentReward(tier)
}

// GetExternalIncentive retrieves an external incentive by its ID.
//
// This function looks up an external incentive in the global `externalIncentives` tree using the specified `incentiveId`.
// If the incentive does not exist, the function panics with an appropriate error message.
//
// Parameters:
// - incentiveId (string): The ID of the incentive to retrieve.
//
// Returns:
// - *ExternalIncentive: A pointer to the retrieved `ExternalIncentive` object.
func GetExternalIncentive(incentiveId string) *ExternalIncentive {
	incentive, exist := externalIncentives.tree.Get(incentiveId)
	if !exist {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("incentiveId(%s) incentive does not exist", incentiveId)),
		)
	}

	return incentive.(*ExternalIncentive)
}

// GetExternalIncentiveByPoolPath retrieves all external incentives associated with a specific pool path.
//
// This function iterates through all external incentives in the global `externalIncentives` tree
// and returns a list of incentives whose `targetPoolPath` matches the specified `poolPath`.
//
// Parameters:
// - poolPath (string): The path of the pool for which incentives are retrieved.
//
// Returns:
// - []ExternalIncentive: A slice of `ExternalIncentive` objects associated with the specified pool path.
func GetExternalIncentiveByPoolPath(poolPath string) []ExternalIncentive {
	incentives := []ExternalIncentive{}
	externalIncentives.tree.Iterate("", "", func(key string, value interface{}) bool {
		incentive := value.(*ExternalIncentive)
		if incentive.targetPoolPath == poolPath {
			incentives = append(incentives, *incentive)
		}
		return false
	})

	return incentives
}

func GetPrintExternalInfo() string {
	externalDebug := ApiExternalDebugInfo{
		Height: std.GetHeight(),
		Time:   time.Now().Unix(),
	}

	externalPositions := []ApiExternalDebugPosition{}
	deposits.Iterate(uint64(0), uint64(deposits.Size()), func(tokenId uint64, deposit *Deposit) bool {
		externalPosition := ApiExternalDebugPosition{
			LpTokenId:       tokenId,
			StakedHeight:    deposit.stakeHeight,
			StakedTimestamp: deposit.stakeTimestamp,
		}

		externalIncentivesList := []ApiExternalDebugIncentive{}
		externalIncentives.tree.Iterate("", "", func(key string, value interface{}) bool {
			incentive := value.(*ExternalIncentive)
			if incentive.targetPoolPath == deposit.targetPoolPath {
				externalIncentive := ApiExternalDebugIncentive{
					PoolPath:          incentive.targetPoolPath,
					IncentiveId:       key,
					RewardToken:       incentive.rewardToken,
					RewardAmount:      strconv.FormatUint(incentive.rewardAmount, 10),
					RewardLeft:        strconv.FormatUint(incentive.rewardLeft, 10),
					StartTimestamp:    incentive.startTimestamp,
					EndTimestamp:      incentive.endTimestamp,
					RewardPerBlock:    strconv.FormatUint(incentive.rewardPerBlock, 10),
					Refundee:          incentive.refundee,
					tokenAmountFull:   incentive.depositGnsAmount,
					tokenAmountToGive: incentive.RewardSpent(uint64(std.GetHeight())),
					StartHeight:       incentive.startHeight,
					EndHeight:         incentive.endHeight,
				}

				externalIncentivesList = append(externalIncentivesList, externalIncentive)
			}
			return false
		})

		externalPosition.Incentive = externalIncentivesList
		externalPositions = append(externalPositions, externalPosition)
		return false
	})

	externalDebug.Position = externalPositions

	// JSON Serialization
	node := json.ObjectNode("", map[string]*json.Node{
		"height":   json.NumberNode("", float64(externalDebug.Height)),
		"time":     json.NumberNode("", float64(externalDebug.Time)),
		"position": json.ArrayNode("", makeExternalPositionsNode(externalDebug.Position)),
	})

	b, err := json.Marshal(node)
	if err != nil {
		return "JSON MARSHAL ERROR"
	}

	return string(b)
}

func makeExternalPositionsNode(positions []ApiExternalDebugPosition) []*json.Node {
	externalPositions := make([]*json.Node, 0)

	for _, externalPosition := range positions {
		incentives := make([]*json.Node, 0)
		for _, incentive := range externalPosition.Incentive {
			stakedOrExternalDuration := std.GetHeight() - max(incentive.StartHeight, externalPosition.StakedHeight)

			incentives = append(incentives, json.ObjectNode("", map[string]*json.Node{
				"poolPath":                 json.StringNode("poolPath", incentive.PoolPath),
				"rewardToken":              json.StringNode("rewardToken", incentive.RewardToken),
				"rewardAmount":             json.StringNode("rewardAmount", incentive.RewardAmount),
				"rewardLeft":               json.StringNode("rewardLeft", incentive.RewardLeft),
				"startTimestamp":           json.NumberNode("startTimestamp", float64(incentive.StartTimestamp)),
				"endTimestamp":             json.NumberNode("endTimestamp", float64(incentive.EndTimestamp)),
				"rewardPerBlock":           json.StringNode("rewardPerBlock", incentive.RewardPerBlock),
				"stakedOrExternalDuration": json.NumberNode("stakedOrExternalDuration", float64(stakedOrExternalDuration)),
				"tokenAmountFull":          json.NumberNode("tokenAmountFull", float64(incentive.tokenAmountFull)),
				"tokenAmountToGive":        json.NumberNode("tokenAmountToGive", float64(incentive.tokenAmountToGive)),
			}))
		}

		externalPositions = append(externalPositions, json.ObjectNode("", map[string]*json.Node{
			"lpTokenId":       json.NumberNode("lpTokenId", float64(externalPosition.LpTokenId)),
			"stakedHeight":    json.NumberNode("stakedHeight", float64(externalPosition.StakedHeight)),
			"stakedTimestamp": json.NumberNode("stakedTimestamp", float64(externalPosition.StakedTimestamp)),
			"incentive":       json.ArrayNode("", incentives),
		}))
	}

	return externalPositions
}

type currentExternalInfo struct {
	height             int64
	time               int64
	externalIncentives []ExternalIncentive
}

type ApiExternalDebugInfo struct {
	Height   int64                      `json:"height"`
	Time     int64                      `json:"time"`
	Position []ApiExternalDebugPosition `json:"pool"`
}

type ApiExternalDebugPosition struct {
	LpTokenId       uint64                      `json:"lpTokenId"`
	StakedHeight    int64                       `json:"stakedHeight"`
	StakedTimestamp int64                       `json:"stakedTimestamp"`
	Incentive       []ApiExternalDebugIncentive `json:"incentive"`
}

type ApiExternalDebugIncentive struct {
	PoolPath          string      `json:"poolPath"`
	IncentiveId       string      `json:"incentiveId"`
	RewardToken       string      `json:"rewardToken"`
	RewardAmount      string      `json:"rewardAmount"`
	RewardLeft        string      `json:"rewardLeft"`
	StartTimestamp    int64       `json:"startTimestamp"`
	EndTimestamp      int64       `json:"endTimestamp"`
	RewardPerBlockX96 string      `json:"rewardPerBlockX96"`
	RewardPerBlock    string      `json:"rewardPerBlock"`
	Refundee          std.Address `json:"refundee"`
	StartHeight       int64       `json:"startHeight"`
	EndHeight         int64       `json:"endHeight"`
	// FROM positionExternal -> externalRewards
	tokenAmountX96    *u256.Uint `json:"tokenAmountX96"`
	tokenAmount       uint64     `json:"tokenAmount"`
	tokenAmountFull   uint64     `json:"tokenAmountFull"`
	tokenAmountToGive uint64     `json:"tokenAmountToGive"`
	// FROM externalWarmUpAmount
	full30  uint64 `json:"full30"`
	give30  uint64 `json:"give30"`
	full50  uint64 `json:"full50"`
	give50  uint64 `json:"give50"`
	full70  uint64 `json:"full70"`
	give70  uint64 `json:"give70"`
	full100 uint64 `json:"full100"`
}

// DEBUG INTERNAL (GNS EMISSION)
type currentInfo struct {
	height           int64
	time             int64
	gnsStaker        uint64
	gnsDevOps        uint64
	gnsCommunityPool uint64
	gnsGovStaker     uint64
	gnsProtocolFee   uint64
	gnsADMIN         uint64
}

func getCurrentInfo() currentInfo {
	return currentInfo{
		height:           std.GetHeight(),
		time:             time.Now().Unix(),
		gnsStaker:        gns.BalanceOf(consts.STAKER_ADDR),
		gnsDevOps:        gns.BalanceOf(consts.DEV_OPS),
		gnsCommunityPool: gns.BalanceOf(consts.COMMUNITY_POOL_ADDR),
		gnsGovStaker:     gns.BalanceOf(consts.GOV_STAKER_ADDR),
		gnsProtocolFee:   gns.BalanceOf(consts.PROTOCOL_FEE_ADDR),
		gnsADMIN:         gns.BalanceOf(consts.ADMIN),
	}
}
