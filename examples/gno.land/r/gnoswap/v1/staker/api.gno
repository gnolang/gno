package staker

import (
	"math"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	en "gno.land/r/gnoswap/v1/emission"
)

type RewardToken struct {
	PoolPath         string   `json:"poolPath"`
	RewardsTokenList []string `json:"rewardsTokenList"`
}

type ApiExternalIncentive struct {
	IncentiveId      string `json:"incentiveId"`
	PoolPath         string `json:"poolPath"`
	RewardToken      string `json:"rewardToken"`
	RewardAmount     uint64 `json:"rewardAmount"`
	RewardLeft       uint64 `json:"rewardLeft"`
	StartTimestamp   int64  `json:"startTimestamp"`
	EndTimestamp     int64  `json:"endTimestamp"`
	Active           bool   `json:"active"`
	Refundee         string `json:"refundee"`
	CreatedHeight    int64  `json:"createdHeight"`
	DepositGnsAmount uint64 `json:"depositGnsAmount"`
}

type ApiInternalIncentive struct {
	PoolPath       string `json:"poolPath"`
	Tier           uint64 `json:"tier"`
	StartTimestamp int64  `json:"startTimestamp"`
	RewardPerBlock string `json:"rewardPerBlock"`
}

func ApiGetRewardTokens() string {
	en.MintAndDistributeGns()

	rewardTokens := []RewardToken{}

	pools.IterateAll(func(key string, pool *Pool) bool {
		thisPoolRewardTokens := []string{}

		// HANDLE INTERNAL
		if poolTier.IsInternallyIncentivizedPool(pool.poolPath) {
			thisPoolRewardTokens = append(thisPoolRewardTokens, consts.GNS_PATH)
		}

		// HANDLE EXTERNAL
		if pool.IsExternallyIncentivizedPool() {
			pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
				ictv := value.(*ExternalIncentive)
				if ictv.RewardToken() == "" {
					return false
				}
				thisPoolRewardTokens = append(thisPoolRewardTokens, ictv.RewardToken())
				return false
			})
		}

		if len(thisPoolRewardTokens) == 0 {
			return false
		}

		rewardTokens = append(rewardTokens, RewardToken{
			PoolPath:         pool.poolPath,
			RewardsTokenList: thisPoolRewardTokens,
		})
		return false
	})

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, rewardToken := range rewardTokens {
		_rewardTokenNode := json.ObjectNode("", map[string]*json.Node{
			"poolPath": json.StringNode("poolPath", rewardToken.PoolPath),
			"tokens":   json.ArrayNode("tokens", makeRewardTokensArray(rewardToken.RewardsTokenList)),
		})
		responses.AppendArray(_rewardTokenNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetRewardTokensByPoolPath(targetPoolPath string) string {
	en.MintAndDistributeGns()

	rewardTokens := []RewardToken{}

	pool, ok := pools.Get(targetPoolPath)
	if !ok {
		return ""
	}

	thisPoolRewardTokens := []string{}

	// HANDLE INTERNAL
	if poolTier.IsInternallyIncentivizedPool(pool.poolPath) {
		thisPoolRewardTokens = append(thisPoolRewardTokens, consts.GNS_PATH)
	}

	// HANDLE EXTERNAL
	if pool.IsExternallyIncentivizedPool() {
		pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
			ictv := value.(*ExternalIncentive)
			if ictv.RewardToken() == "" {
				return false
			}
			thisPoolRewardTokens = append(thisPoolRewardTokens, ictv.RewardToken())
			return false
		})
	}

	rewardTokens = append(rewardTokens, RewardToken{
		PoolPath:         pool.poolPath,
		RewardsTokenList: thisPoolRewardTokens,
	})

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, rewardToken := range rewardTokens {
		_rewardTokenNode := json.ObjectNode("", map[string]*json.Node{
			"poolPath": json.StringNode("poolPath", rewardToken.PoolPath),
			"tokens":   json.ArrayNode("tokens", makeRewardTokensArray(rewardToken.RewardsTokenList)),
		})
		responses.AppendArray(_rewardTokenNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetExternalIncentives() string {
	en.MintAndDistributeGns()

	apiExternalIncentives := []ApiExternalIncentive{}

	pools.tree.Iterate("", "", func(key string, value interface{}) bool {
		pool := value.(*Pool)
		pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
			ictv := value.(*ExternalIncentive)
			apiExternalIncentives = append(apiExternalIncentives, ApiExternalIncentive{
				IncentiveId:      ictv.incentiveId,
				PoolPath:         ictv.targetPoolPath,
				RewardToken:      ictv.rewardToken,
				RewardAmount:     ictv.rewardAmount,
				RewardLeft:       ictv.rewardLeft,
				StartTimestamp:   ictv.startTimestamp,
				EndTimestamp:     ictv.endTimestamp,
				Refundee:         ictv.refundee.String(),
				CreatedHeight:    ictv.createdHeight,
				DepositGnsAmount: ictv.depositGnsAmount,
			})
			return false
		})
		return false
	})
	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, incentive := range apiExternalIncentives {
		active := false
		if time.Now().Unix() >= incentive.StartTimestamp && time.Now().Unix() <= incentive.EndTimestamp {
			active = true
		}

		_incentiveNode := json.ObjectNode("", map[string]*json.Node{
			"incentiveId":      json.StringNode("incentiveId", incentive.IncentiveId),
			"poolPath":         json.StringNode("poolPath", incentive.PoolPath),
			"rewardToken":      json.StringNode("rewardToken", incentive.RewardToken),
			"rewardAmount":     json.StringNode("rewardAmount", strconv.FormatUint(incentive.RewardAmount, 10)),
			"rewardLeft":       json.StringNode("rewardLeft", strconv.FormatUint(incentive.RewardLeft, 10)),
			"startTimestamp":   json.NumberNode("startTimestamp", float64(incentive.StartTimestamp)),
			"endTimestamp":     json.NumberNode("endTimestamp", float64(incentive.EndTimestamp)),
			"active":           json.BoolNode("active", active),
			"refundee":         json.StringNode("refundee", incentive.Refundee),
			"createdHeight":    json.NumberNode("createdHeight", float64(incentive.CreatedHeight)),
			"depositGnsAmount": json.NumberNode("depositGnsAmount", float64(incentive.DepositGnsAmount)),
		})
		responses.AppendArray(_incentiveNode)
	}

	// RETURN
	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetExternalIncentiveById(poolPath, incentiveId string) string {
	en.MintAndDistributeGns()

	apiExternalIncentives := []ApiExternalIncentive{}

	pool, ok := pools.Get(poolPath)
	if !ok {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("pool(%s) not found", poolPath),
		))
	}

	incentive, exist := pool.incentives.GetByIncentiveId(incentiveId)
	if !exist {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("incentive(%s) not found", incentiveId),
		))
	}

	apiExternalIncentives = append(apiExternalIncentives, ApiExternalIncentive{
		IncentiveId:      incentiveId,
		PoolPath:         poolPath,
		RewardToken:      incentive.rewardToken,
		RewardAmount:     incentive.rewardAmount,
		RewardLeft:       incentive.rewardLeft,
		StartTimestamp:   incentive.startTimestamp,
		EndTimestamp:     incentive.endTimestamp,
		Refundee:         incentive.refundee.String(),
		CreatedHeight:    incentive.createdHeight,
		DepositGnsAmount: incentive.depositGnsAmount,
	})

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, incentive := range apiExternalIncentives {
		active := false
		if time.Now().Unix() >= incentive.StartTimestamp && time.Now().Unix() <= incentive.EndTimestamp {
			active = true
		}

		_incentiveNode := json.ObjectNode("", map[string]*json.Node{
			"incentiveId":      json.StringNode("incentiveId", incentive.IncentiveId),
			"poolPath":         json.StringNode("poolPath", incentive.PoolPath),
			"rewardToken":      json.StringNode("rewardToken", incentive.RewardToken),
			"rewardAmount":     json.StringNode("rewardAmount", strconv.FormatUint(incentive.RewardAmount, 10)),
			"rewardLeft":       json.StringNode("rewardLeft", strconv.FormatUint(incentive.RewardLeft, 10)),
			"startTimestamp":   json.NumberNode("startTimestamp", float64(incentive.StartTimestamp)),
			"endTimestamp":     json.NumberNode("endTimestamp", float64(incentive.EndTimestamp)),
			"active":           json.BoolNode("active", active),
			"refundee":         json.StringNode("refundee", incentive.Refundee),
			"createdHeight":    json.NumberNode("createdHeight", float64(incentive.CreatedHeight)),
			"depositGnsAmount": json.NumberNode("depositGnsAmount", float64(incentive.DepositGnsAmount)),
		})
		responses.AppendArray(_incentiveNode)
	}

	// RETURN
	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetExternalIncentivesByPoolPath(targetPoolPath string) string {
	en.MintAndDistributeGns()

	apiExternalIncentives := []ApiExternalIncentive{}

	pool, ok := pools.Get(targetPoolPath)
	if !ok {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("pool(%s) not found", targetPoolPath),
		))
	}

	pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
		incentive := value.(*ExternalIncentive)
		if incentive.targetPoolPath != targetPoolPath {
			return false
		}

		apiExternalIncentives = append(apiExternalIncentives, ApiExternalIncentive{
			IncentiveId:      incentive.incentiveId,
			PoolPath:         incentive.targetPoolPath,
			RewardToken:      incentive.rewardToken,
			RewardAmount:     incentive.rewardAmount,
			RewardLeft:       incentive.rewardLeft,
			StartTimestamp:   incentive.startTimestamp,
			EndTimestamp:     incentive.endTimestamp,
			Refundee:         incentive.refundee.String(),
			CreatedHeight:    incentive.createdHeight,
			DepositGnsAmount: incentive.depositGnsAmount,
		})
		return false
	})

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, incentive := range apiExternalIncentives {
		active := false
		if time.Now().Unix() >= incentive.StartTimestamp && time.Now().Unix() <= incentive.EndTimestamp {
			active = true
		}

		_incentiveNode := json.ObjectNode("", map[string]*json.Node{
			"incentiveId":      json.StringNode("incentiveId", incentive.IncentiveId),
			"poolPath":         json.StringNode("poolPath", incentive.PoolPath),
			"rewardToken":      json.StringNode("rewardToken", incentive.RewardToken),
			"rewardAmount":     json.StringNode("rewardAmount", strconv.FormatUint(incentive.RewardAmount, 10)),
			"rewardLeft":       json.StringNode("rewardLeft", strconv.FormatUint(incentive.RewardLeft, 10)),
			"startTimestamp":   json.NumberNode("startTimestamp", float64(incentive.StartTimestamp)),
			"endTimestamp":     json.NumberNode("endTimestamp", float64(incentive.EndTimestamp)),
			"active":           json.BoolNode("active", active),
			"refundee":         json.StringNode("refundee", incentive.Refundee),
			"createdHeight":    json.NumberNode("createdHeight", float64(incentive.CreatedHeight)),
			"depositGnsAmount": json.NumberNode("depositGnsAmount", float64(incentive.DepositGnsAmount)),
		})
		responses.AppendArray(_incentiveNode)
	}

	// RETURN
	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetExternalIncentivesByRewardTokenPath(rewardTokenPath string) string {
	en.MintAndDistributeGns()

	apiExternalIncentives := []ApiExternalIncentive{}

	pools.tree.Iterate("", "", func(key string, value interface{}) bool {
		pool := value.(*Pool)
		pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
			incentive := value.(*ExternalIncentive)
			if incentive.rewardToken != rewardTokenPath {
				return false
			}

			apiExternalIncentives = append(apiExternalIncentives, ApiExternalIncentive{
				IncentiveId:      incentive.incentiveId,
				PoolPath:         pool.poolPath,
				RewardToken:      incentive.rewardToken,
				RewardAmount:     incentive.rewardAmount,
				RewardLeft:       incentive.rewardLeft,
				StartTimestamp:   incentive.startTimestamp,
				EndTimestamp:     incentive.endTimestamp,
				Refundee:         incentive.refundee.String(),
				CreatedHeight:    incentive.createdHeight,
				DepositGnsAmount: incentive.depositGnsAmount,
			})
			return false
		})
		return false
	})

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, incentive := range apiExternalIncentives {
		active := false
		if time.Now().Unix() >= incentive.StartTimestamp && time.Now().Unix() <= incentive.EndTimestamp {
			active = true
		}

		_incentiveNode := json.ObjectNode("", map[string]*json.Node{
			"incentiveId":      json.StringNode("incentiveId", incentive.IncentiveId),
			"poolPath":         json.StringNode("poolPath", incentive.PoolPath),
			"rewardToken":      json.StringNode("rewardToken", incentive.RewardToken),
			"rewardAmount":     json.StringNode("rewardAmount", strconv.FormatUint(incentive.RewardAmount, 10)),
			"rewardLeft":       json.StringNode("rewardLeft", strconv.FormatUint(incentive.RewardLeft, 10)),
			"startTimestamp":   json.NumberNode("startTimestamp", float64(incentive.StartTimestamp)),
			"endTimestamp":     json.NumberNode("endTimestamp", float64(incentive.EndTimestamp)),
			"active":           json.BoolNode("active", active),
			"refundee":         json.StringNode("refundee", incentive.Refundee),
			"createdHeight":    json.NumberNode("createdHeight", float64(incentive.CreatedHeight)),
			"depositGnsAmount": json.NumberNode("depositGnsAmount", float64(incentive.DepositGnsAmount)),
		})
		responses.AppendArray(_incentiveNode)
	}

	// RETURN
	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetInternalIncentives() string {
	en.MintAndDistributeGns()

	apiInternalIncentives := []ApiInternalIncentive{}

	poolTier.membership.Iterate("", "", func(key string, value interface{}) bool {
		poolPath := key
		internalTier := value.(uint64)
		apiInternalIncentives = append(apiInternalIncentives, ApiInternalIncentive{
			PoolPath:       poolPath,
			Tier:           internalTier,
			RewardPerBlock: calculateInternalRewardPerBlockByPoolPath(poolPath),
		})
		return false
	})

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, incentive := range apiInternalIncentives {
		_incentiveNode := json.ObjectNode("", map[string]*json.Node{
			"poolPath":       json.StringNode("poolPath", incentive.PoolPath),
			"rewardToken":    json.StringNode("rewardToken", consts.GNS_PATH),
			"tier":           json.NumberNode("tier", float64(incentive.Tier)),
			"rewardPerBlock": json.StringNode("rewardPerBlock", incentive.RewardPerBlock),
		})
		responses.AppendArray(_incentiveNode)
	}

	// RETURN
	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetInternalIncentivesByPoolPath(targetPoolPath string) string {
	en.MintAndDistributeGns()

	apiInternalIncentives := []ApiInternalIncentive{}

	tier := poolTier.CurrentTier(targetPoolPath)
	if tier == 0 {
		return ""
	}

	apiInternalIncentives = append(apiInternalIncentives, ApiInternalIncentive{
		PoolPath:       targetPoolPath,
		Tier:           tier,
		RewardPerBlock: calculateInternalRewardPerBlockByPoolPath(targetPoolPath),
	})

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, incentive := range apiInternalIncentives {
		_incentiveNode := json.ObjectNode("", map[string]*json.Node{
			"poolPath":       json.StringNode("poolPath", incentive.PoolPath),
			"rewardToken":    json.StringNode("rewardToken", consts.GNS_PATH),
			"tier":           json.NumberNode("tier", float64(incentive.Tier)),
			"rewardPerBlock": json.StringNode("rewardPerBlock", incentive.RewardPerBlock),
		})
		responses.AppendArray(_incentiveNode)
	}

	// RETURN
	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetInternalIncentivesByTiers(targetTier uint64) string {
	en.MintAndDistributeGns()

	apiInternalIncentives := []ApiInternalIncentive{}

	poolTier.membership.Iterate("", "", func(key string, value interface{}) bool {
		poolPath := key
		internalTier := value.(uint64)
		if internalTier != targetTier {
			return false
		}

		apiInternalIncentives = append(apiInternalIncentives, ApiInternalIncentive{
			PoolPath:       poolPath,
			Tier:           internalTier,
			RewardPerBlock: calculateInternalRewardPerBlockByPoolPath(poolPath),
		})

		return false
	})

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, incentive := range apiInternalIncentives {
		_incentiveNode := json.ObjectNode("", map[string]*json.Node{
			"poolPath":       json.StringNode("poolPath", incentive.PoolPath),
			"rewardToken":    json.StringNode("rewardToken", consts.GNS_PATH),
			"tier":           json.NumberNode("tier", float64(incentive.Tier)),
			"rewardPerBlock": json.StringNode("rewardPerBlock", incentive.RewardPerBlock),
		})
		responses.AppendArray(_incentiveNode)
	}

	// RETURN
	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func makeRewardTokensArray(rewardsTokenList []string) []*json.Node {
	rewardsTokenArray := make([]*json.Node, len(rewardsTokenList))
	for i, rewardToken := range rewardsTokenList {
		rewardsTokenArray[i] = json.StringNode("", rewardToken)
	}
	return rewardsTokenArray
}

func calculateInternalRewardPerBlockByPoolPath(poolPath string) string {
	reward := poolTier.CurrentRewardPerPool(poolPath)
	return ufmt.Sprintf("%d", reward)
}

// LpTokenReward represents the rewards associated with a specific LP token
type LpTokenReward struct {
	LpTokenId uint64      `json:"lpTokenId"` // The ID of the LP token
	Address   string      `json:"address"`   // The address associated with the LP token
	Rewards   []ApiReward `json:"rewards"`
}

// Reward represents a single reward for a staked LP token
type ApiReward struct {
	IncentiveType     string `json:"incentiveType"`     // The type of incentive (INTERNAL or EXTERNAL)
	IncentiveId       string `json:"incentiveId"`       // The unique identifier of the incentive
	TargetPoolPath    string `json:"targetPoolPath"`    // The path of the target pool for the reward
	RewardTokenPath   string `json:"rewardTokenPath"`   // The pathe of the reward token
	RewardTokenAmount uint64 `json:"rewardTokenAmount"` // The amount of the reward token
	StakeTimestamp    int64  `json:"stakeTimestamp"`    // The timestamp when the LP token was staked
	StakeHeight       int64  `json:"stakeHeight"`       // The block height when the LP token was staked
	IncentiveStart    int64  `json:"incentiveStart"`    // The timestamp when the incentive started
}

// Stake represents a single stake
type ApiStake struct {
	TokenId        uint64      `json:"tokenId"`        // The ID of the staked LP token
	Owner          std.Address `json:"owner"`          // The address of the owner of the staked LP token
	NumberOfStakes uint64      `json:"numberOfStakes"` // The number of times this LP token has been staked
	StakeTimestamp int64       `json:"stakeTimestamp"` // The timestamp when the LP token was staked
	StakeHeight    int64       `json:"stakeHeight"`    // The block height when the LP token was staked
	TargetPoolPath string      `json:"targetPoolPath"` // The path of the target pool for the stake
}

// ResponseQueryBase contains basic information about a query response.
type ResponseQueryBase struct {
	Height    int64 `json:"height"`    // The block height at the time of the query
	Timestamp int64 `json:"timestamp"` // The timestamp at the time of the query
}

// ResponseApiGetRewards represents the API response for getting rewards.
type ResponseApiGetRewards struct {
	Stat     ResponseQueryBase `json:"stat"`     // Basic query information
	Response []LpTokenReward   `json:"response"` // A slice of LpTokenReward structs
}

// ResponseApiGetRewardByLpTokenId represents the API response for getting rewards for a specific LP token.
type ResponseApiGetRewardByLpTokenId struct {
	Stat     ResponseQueryBase `json:"stat"`     // Basic query information
	Response LpTokenReward     `json:"response"` // The LpTokenReward for the specified LP token
}

// ResponseApiGetStakes represents the API response for getting stakes.
type ResponseApiGetStakes struct {
	Stat     ResponseQueryBase `json:"stat"`     // Basic query information
	Response []ApiStake        `json:"response"` // A slice of Stake structs
}

func ApiGetRewardsByLpTokenId(targetLpTokenId uint64) string {
	en.MintAndDistributeGns()

	deposit := deposits.Get(targetLpTokenId)

	reward := calcPositionReward(std.GetHeight(), targetLpTokenId)

	rewards := []ApiReward{}

	if reward.Internal > 0 {
		rewards = append(rewards, ApiReward{
			IncentiveType:     "INTERNAL",
			IncentiveId:       "",
			TargetPoolPath:    deposit.targetPoolPath,
			RewardTokenPath:   consts.GNS_PATH,
			RewardTokenAmount: reward.Internal,
			StakeTimestamp:    deposit.stakeTimestamp,
			StakeHeight:       deposit.stakeHeight,
			IncentiveStart:    deposit.stakeTimestamp,
		})
	}

	for incentiveId, externalReward := range reward.External {
		rewards = append(rewards, ApiReward{
			IncentiveType:     "EXTERNAL",
			IncentiveId:       incentiveId,
			TargetPoolPath:    deposit.targetPoolPath,
			RewardTokenPath:   consts.GNS_PATH,
			RewardTokenAmount: externalReward,
			StakeTimestamp:    deposit.stakeTimestamp,
			StakeHeight:       deposit.stakeHeight,
			IncentiveStart:    deposit.stakeTimestamp,
		})
	}

	qb := ResponseQueryBase{
		Height:    std.GetHeight(),
		Timestamp: time.Now().Unix(),
	}

	r := ResponseApiGetRewards{
		Stat: qb,
		Response: []LpTokenReward{
			{
				LpTokenId: targetLpTokenId,
				Address:   deposit.owner.String(),
				Rewards:   rewards,
			},
		},
	}

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, reward := range r.Response {
		_rewardNode := json.ObjectNode("", map[string]*json.Node{
			"lpTokenId": json.NumberNode("lpTokenId", float64(reward.LpTokenId)),
			"address":   json.StringNode("address", reward.Address),
			"rewards":   json.ArrayNode("rewards", makeRewardsArray(reward.Rewards)),
		})
		responses.AppendArray(_rewardNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetRewardsByAddress(targetAddress string) string {
	en.MintAndDistributeGns()

	lpTokenRewards := []LpTokenReward{}

	stakers.IterateAll(std.Address(targetAddress), func(tokenId uint64, deposit *Deposit) bool {
		rewards := []ApiReward{}

		reward := calcPositionReward(std.GetHeight(), tokenId)

		// get internal gns reward
		if reward.Internal > 0 {
			rewards = append(rewards, ApiReward{
				IncentiveType:     "INTERNAL",
				IncentiveId:       "",
				TargetPoolPath:    deposit.targetPoolPath,
				RewardTokenPath:   consts.GNS_PATH,
				RewardTokenAmount: reward.Internal,
				StakeTimestamp:    deposit.stakeTimestamp,
				StakeHeight:       deposit.stakeHeight,
				IncentiveStart:    deposit.stakeTimestamp,
			})
		}

		// find all external reward list for poolPath which lpTokenId is staked
		for incentiveId, externalReward := range reward.External {
			rewards = append(rewards, ApiReward{
				IncentiveType:     "EXTERNAL",
				IncentiveId:       incentiveId,
				TargetPoolPath:    deposit.targetPoolPath,
				RewardTokenPath:   consts.GNS_PATH,
				RewardTokenAmount: externalReward,
				StakeTimestamp:    deposit.stakeTimestamp,
				StakeHeight:       deposit.stakeHeight,
				IncentiveStart:    deposit.stakeTimestamp,
			})
		}
		lpTokenReward := LpTokenReward{
			LpTokenId: tokenId,
			Address:   deposit.owner.String(),
			Rewards:   rewards,
		}
		lpTokenRewards = append(lpTokenRewards, lpTokenReward)

		return false
	})

	qb := ResponseQueryBase{
		Height:    std.GetHeight(),
		Timestamp: time.Now().Unix(),
	}

	r := ResponseApiGetRewards{
		Stat:     qb,
		Response: lpTokenRewards,
	}

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, reward := range r.Response {
		_rewardNode := json.ObjectNode("", map[string]*json.Node{
			"lpTokenId": json.NumberNode("lpTokenId", float64(reward.LpTokenId)),
			"address":   json.StringNode("address", reward.Address),
			"rewards":   json.ArrayNode("rewards", makeRewardsArray(reward.Rewards)),
		})
		responses.AppendArray(_rewardNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetStakes() string {
	en.MintAndDistributeGns()

	stakes := []ApiStake{}
	deposits.Iterate(0, math.MaxUint64, func(tokenId uint64, deposit *Deposit) bool {
		stakes = append(stakes, ApiStake{
			TokenId:        tokenId,
			Owner:          deposit.owner,
			StakeTimestamp: deposit.stakeTimestamp,
			StakeHeight:    deposit.stakeHeight,
			TargetPoolPath: deposit.targetPoolPath,
		})
		return false
	})

	qb := ResponseQueryBase{
		Height:    std.GetHeight(),
		Timestamp: time.Now().Unix(),
	}

	r := ResponseApiGetStakes{
		Stat:     qb,
		Response: stakes,
	}

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, stake := range r.Response {
		_stakeNode := json.ObjectNode("", map[string]*json.Node{
			"tokenId":        json.NumberNode("tokenId", float64(stake.TokenId)),
			"owner":          json.StringNode("owner", stake.Owner.String()),
			"stakeTimestamp": json.NumberNode("stakeTimestamp", float64(stake.StakeTimestamp)),
			"stakeHeight":    json.NumberNode("stakeHeight", float64(stake.StakeHeight)),
			"targetPoolPath": json.StringNode("targetPoolPath", stake.TargetPoolPath),
		})
		responses.AppendArray(_stakeNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetStakesByLpTokenId(targetLpTokenId uint64) string {
	en.MintAndDistributeGns()

	stakes := []ApiStake{}

	deposit := deposits.Get(targetLpTokenId)
	stakes = append(stakes, ApiStake{
		TokenId:        targetLpTokenId,
		Owner:          deposit.owner,
		StakeTimestamp: deposit.stakeTimestamp,
		StakeHeight:    deposit.stakeHeight,
		TargetPoolPath: deposit.targetPoolPath,
	})

	qb := ResponseQueryBase{
		Height:    std.GetHeight(),
		Timestamp: time.Now().Unix(),
	}

	r := ResponseApiGetStakes{
		Stat:     qb,
		Response: stakes,
	}

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, stake := range r.Response {
		_stakeNode := json.ObjectNode("", map[string]*json.Node{
			"tokenId":        json.NumberNode("tokenId", float64(stake.TokenId)),
			"owner":          json.StringNode("owner", stake.Owner.String()),
			"numberOfStakes": json.NumberNode("numberOfStakes", float64(stake.NumberOfStakes)),
			"stakeTimestamp": json.NumberNode("stakeTimestamp", float64(stake.StakeTimestamp)),
			"stakeHeight":    json.NumberNode("stakeHeight", float64(stake.StakeHeight)),
			"targetPoolPath": json.StringNode("targetPoolPath", stake.TargetPoolPath),
		})
		responses.AppendArray(_stakeNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

func ApiGetStakesByAddress(targetAddress string) string {
	en.MintAndDistributeGns()

	stakes := []ApiStake{}

	stakers.IterateAll(std.Address(targetAddress), func(tokenId uint64, deposit *Deposit) bool {
		stakes = append(stakes, ApiStake{
			TokenId:        tokenId,
			Owner:          deposit.owner,
			StakeTimestamp: deposit.stakeTimestamp,
			StakeHeight:    deposit.stakeHeight,
			TargetPoolPath: deposit.targetPoolPath,
		})

		return false
	})

	qb := ResponseQueryBase{
		Height:    std.GetHeight(),
		Timestamp: time.Now().Unix(),
	}

	r := ResponseApiGetStakes{
		Stat:     qb,
		Response: stakes,
	}

	// STAT NODE
	_stat := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	// RESPONSE (ARRAY) NODE
	responses := json.ArrayNode("", []*json.Node{})
	for _, stake := range r.Response {
		_stakeNode := json.ObjectNode("", map[string]*json.Node{
			"tokenId":        json.NumberNode("tokenId", float64(stake.TokenId)),
			"owner":          json.StringNode("owner", stake.Owner.String()),
			"stakeTimestamp": json.NumberNode("stakeTimestamp", float64(stake.StakeTimestamp)),
			"stakeHeight":    json.NumberNode("stakeHeight", float64(stake.StakeHeight)),
			"targetPoolPath": json.StringNode("targetPoolPath", stake.TargetPoolPath),
		})
		responses.AppendArray(_stakeNode)
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     _stat,
		"response": responses,
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

// for off chain to check if lpTokenId is staked via RPC
func IsStaked(tokenId uint64) bool {
	return deposits.Has(tokenId)
}

func makeRewardsArray(rewards []ApiReward) []*json.Node {
	rewardsArray := make([]*json.Node, len(rewards))

	for i, reward := range rewards {
		rewardsArray[i] = json.ObjectNode("", map[string]*json.Node{
			"incentiveType":     json.StringNode("incentiveType", reward.IncentiveType),
			"incentiveId":       json.StringNode("incentiveId", reward.IncentiveId),
			"targetPoolPath":    json.StringNode("targetPoolPath", reward.TargetPoolPath),
			"rewardTokenPath":   json.StringNode("rewardTokenPath", reward.RewardTokenPath),
			"rewardTokenAmount": json.NumberNode("rewardTokenAmount", float64(reward.RewardTokenAmount)),
			"stakeTimestamp":    json.NumberNode("stakeTimestamp", float64(reward.StakeTimestamp)),
			"stakeHeight":       json.NumberNode("stakeHeight", float64(reward.StakeHeight)),
			"incentiveStart":    json.NumberNode("incentiveStart", float64(reward.IncentiveStart)),
		})
	}
	return rewardsArray
}
