package staker

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/json"

	"gno.land/p/gnoswap/consts"
)

// JsonResponse is an interface that all JSON response types must implement.
type JsonResponse interface {
	JSON() *json.Node
}

type RewardToken struct {
	PoolPath         string   `json:"poolPath"`
	RewardsTokenList []string `json:"rewardsTokenList"`
}

func newRewardToken(poolPath string, tokens []string) RewardToken {
	return RewardToken{
		PoolPath:         poolPath,
		RewardsTokenList: tokens,
	}
}

func (r RewardToken) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"poolPath": json.StringNode("poolPath", r.PoolPath),
		"tokens":   json.ArrayNode("tokens", makeRewardTokensArray(r.RewardsTokenList)),
	})
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

func newApiExternalIncentive(ictv *ExternalIncentive) ApiExternalIncentive {
	return ApiExternalIncentive{
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
	}
}

func (r ApiExternalIncentive) JSON() *json.Node {
	active := false
	if time.Now().Unix() >= r.StartTimestamp && time.Now().Unix() <= r.EndTimestamp {
		active = true
	}

	return json.ObjectNode("", map[string]*json.Node{
		"incentiveId":      json.StringNode("incentiveId", r.IncentiveId),
		"poolPath":         json.StringNode("poolPath", r.PoolPath),
		"rewardToken":      json.StringNode("rewardToken", r.RewardToken),
		"rewardAmount":     json.StringNode("rewardAmount", strconv.FormatUint(r.RewardAmount, 10)),
		"rewardLeft":       json.StringNode("rewardLeft", strconv.FormatUint(r.RewardLeft, 10)),
		"startTimestamp":   json.NumberNode("startTimestamp", float64(r.StartTimestamp)),
		"endTimestamp":     json.NumberNode("endTimestamp", float64(r.EndTimestamp)),
		"active":           json.BoolNode("active", active),
		"refundee":         json.StringNode("refundee", r.Refundee),
		"createdHeight":    json.NumberNode("createdHeight", float64(r.CreatedHeight)),
		"depositGnsAmount": json.NumberNode("depositGnsAmount", float64(r.DepositGnsAmount)),
	})
}

type ApiInternalIncentive struct {
	PoolPath       string `json:"poolPath"`
	Tier           uint64 `json:"tier"`
	StartTimestamp int64  `json:"startTimestamp"`
	RewardPerBlock string `json:"rewardPerBlock"`
}

func newApiInternalIncentive(poolPath string, tier uint64) ApiInternalIncentive {
	perBlock := calculateInternalRewardPerBlockByPoolPath(poolPath)
	return ApiInternalIncentive{
		PoolPath:       poolPath,
		Tier:           tier,
		RewardPerBlock: perBlock,
	}
}

func (r ApiInternalIncentive) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"poolPath":       json.StringNode("poolPath", r.PoolPath),
		"rewardToken":    json.StringNode("rewardToken", consts.GNS_PATH),
		"tier":           json.NumberNode("tier", float64(r.Tier)),
		"rewardPerBlock": json.StringNode("rewardPerBlock", r.RewardPerBlock),
	})
}

// LpTokenReward represents the rewards associated with a specific LP token
type LpTokenReward struct {
	LpTokenId uint64      `json:"lpTokenId"` // The ID of the LP token
	Address   string      `json:"address"`   // The address associated with the LP token
	Rewards   []ApiReward `json:"rewards"`
}

func newLpTokenReward(lpTokenId uint64, address string, rewards []ApiReward) LpTokenReward {
	return LpTokenReward{
		LpTokenId: lpTokenId,
		Address:   address,
		Rewards:   rewards,
	}
}

func (r LpTokenReward) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"lpTokenId": json.NumberNode("lpTokenId", float64(r.LpTokenId)),
		"address":   json.StringNode("address", r.Address),
		"rewards":   json.ArrayNode("rewards", makeRewardsArray(r.Rewards)),
	})
}

// Stake represents a single stake
type ApiStake struct {
	PositionId     uint64      `json:"positionId"`     // The ID of the staked LP token
	Owner          std.Address `json:"owner"`          // The address of the owner of the staked LP token
	NumberOfStakes uint64      `json:"numberOfStakes"` // The number of times this LP token has been staked
	StakeTimestamp int64       `json:"stakeTimestamp"` // The timestamp when the LP token was staked
	StakeHeight    int64       `json:"stakeHeight"`    // The block height when the LP token was staked
	TargetPoolPath string      `json:"targetPoolPath"` // The path of the target pool for the stake
}

func newApiStake(tokenId uint64, deposit *Deposit) ApiStake {
	return ApiStake{
		PositionId:     tokenId,
		Owner:          deposit.owner,
		StakeTimestamp: deposit.stakeTimestamp,
		StakeHeight:    deposit.stakeHeight,
		TargetPoolPath: deposit.targetPoolPath,
	}
}

func (s ApiStake) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"positionId":     json.NumberNode("positionId", float64(s.PositionId)),
		"owner":          json.StringNode("owner", s.Owner.String()),
		"stakeTimestamp": json.NumberNode("stakeTimestamp", float64(s.StakeTimestamp)),
		"stakeHeight":    json.NumberNode("stakeHeight", float64(s.StakeHeight)),
		"targetPoolPath": json.StringNode("targetPoolPath", s.TargetPoolPath),
	})
}

type statNode struct {
	height    int64
	timestamp int64
}

func newStatNode() statNode {
	return statNode{
		height:    std.GetHeight(),
		timestamp: time.Now().Unix(),
	}
}

func (s statNode) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(s.height)),
		"timestamp": json.NumberNode("timestamp", float64(s.timestamp)),
	})
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

func (r ApiReward) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"incentiveType":     json.StringNode("incentiveType", r.IncentiveType),
		"incentiveId":       json.StringNode("incentiveId", r.IncentiveId),
		"targetPoolPath":    json.StringNode("targetPoolPath", r.TargetPoolPath),
		"rewardTokenPath":   json.StringNode("rewardTokenPath", r.RewardTokenPath),
		"rewardTokenAmount": json.NumberNode("rewardTokenAmount", float64(r.RewardTokenAmount)),
		"stakeTimestamp":    json.NumberNode("stakeTimestamp", float64(r.StakeTimestamp)),
		"stakeHeight":       json.NumberNode("stakeHeight", float64(r.StakeHeight)),
		"incentiveStart":    json.NumberNode("incentiveStart", float64(r.IncentiveStart)),
	})
}

///////////////////// Response /////////////////////

type ApiResponse struct {
	Stat     statNode       `json:"stat"`
	Response []JsonResponse `json:"response"`
}

func (r ApiResponse) JSON() *json.Node {
	rspsNodes := make([]*json.Node, len(r.Response))
	for i, item := range r.Response {
		rspsNodes[i] = item.JSON()
	}

	return json.ObjectNode("", map[string]*json.Node{
		"stat":     r.Stat.JSON(),
		"response": json.ArrayNode("response", rspsNodes),
	})
}

func makeApiResponse(rs []JsonResponse) string {
	resp := ApiResponse{
		Stat:     newStatNode(),
		Response: rs,
	}

	b, err := json.Marshal(resp.JSON())
	if err != nil {
		panic(err.Error())
	}

	return string(b)
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
