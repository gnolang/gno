package staker

import (
	"math"
	"std"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
)

func ApiGetRewardTokens() string {
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

		rt := newRewardToken(pool.poolPath, thisPoolRewardTokens)
		rewardTokens = append(rewardTokens, rt)
		return false
	})

	rsps := make([]JsonResponse, len(rewardTokens))
	for i := range rewardTokens {
		rsps[i] = rewardTokens[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetRewardTokensByPoolPath(targetPoolPath string) string {
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

	rt := newRewardToken(pool.poolPath, thisPoolRewardTokens)
	rewardTokens = append(rewardTokens, rt)

	rsps := make([]JsonResponse, len(rewardTokens))
	for i := range rewardTokens {
		rsps[i] = rewardTokens[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetExternalIncentives() string {
	apiExternalIncentives := []ApiExternalIncentive{}

	pools.tree.Iterate("", "", func(key string, value interface{}) bool {
		pool := value.(*Pool)
		pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
			ictv := value.(*ExternalIncentive)
			externalIctv := newApiExternalIncentive(ictv)
			apiExternalIncentives = append(apiExternalIncentives, externalIctv)
			return false
		})
		return false
	})

	rsps := make([]JsonResponse, len(apiExternalIncentives))
	for i := range apiExternalIncentives {
		rsps[i] = apiExternalIncentives[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetExternalIncentiveById(poolPath, incentiveId string) string {
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

	externalictv := newApiExternalIncentive(incentive)
	apiExternalIncentives = append(apiExternalIncentives, externalictv)

	rsps := make([]JsonResponse, len(apiExternalIncentives))
	for i := range apiExternalIncentives {
		rsps[i] = apiExternalIncentives[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetExternalIncentivesByPoolPath(targetPoolPath string) string {
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

		externalIctv := newApiExternalIncentive(incentive)
		apiExternalIncentives = append(apiExternalIncentives, externalIctv)

		return false
	})

	rsps := make([]JsonResponse, len(apiExternalIncentives))
	for i := range apiExternalIncentives {
		rsps[i] = apiExternalIncentives[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetExternalIncentivesByRewardTokenPath(rewardTokenPath string) string {
	apiExternalIncentives := []ApiExternalIncentive{}

	pools.tree.Iterate("", "", func(key string, value interface{}) bool {
		pool := value.(*Pool)
		pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
			incentive := value.(*ExternalIncentive)
			if incentive.rewardToken != rewardTokenPath {
				return false
			}

			externalIctv := newApiExternalIncentive(incentive)
			apiExternalIncentives = append(apiExternalIncentives, externalIctv)

			return false
		})
		return false
	})

	rsps := make([]JsonResponse, len(apiExternalIncentives))
	for i := range apiExternalIncentives {
		rsps[i] = apiExternalIncentives[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetInternalIncentives() string {
	apiInternalIncentives := []ApiInternalIncentive{}

	poolTier.membership.Iterate("", "", func(key string, value interface{}) bool {
		poolPath := key
		internalTier := value.(uint64)
		internalIctv := newApiInternalIncentive(poolPath, internalTier)
		apiInternalIncentives = append(apiInternalIncentives, internalIctv)
		return false
	})

	rsps := make([]JsonResponse, len(apiInternalIncentives))
	for i := range apiInternalIncentives {
		rsps[i] = apiInternalIncentives[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetInternalIncentivesByPoolPath(targetPoolPath string) string {
	apiInternalIncentives := []ApiInternalIncentive{}

	tier := poolTier.CurrentTier(targetPoolPath)
	if tier == 0 {
		return ""
	}

	internalIctv := newApiInternalIncentive(targetPoolPath, tier)
	apiInternalIncentives = append(apiInternalIncentives, internalIctv)

	rsps := make([]JsonResponse, len(apiInternalIncentives))
	for i := range apiInternalIncentives {
		rsps[i] = apiInternalIncentives[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetInternalIncentivesByTiers(targetTier uint64) string {
	apiInternalIncentives := []ApiInternalIncentive{}

	poolTier.membership.Iterate("", "", func(key string, value interface{}) bool {
		poolPath := key
		internalTier := value.(uint64)
		if internalTier != targetTier {
			return false
		}

		internalIctv := newApiInternalIncentive(poolPath, internalTier)
		apiInternalIncentives = append(apiInternalIncentives, internalIctv)

		return false
	})

	rsps := make([]JsonResponse, len(apiInternalIncentives))
	for i := range apiInternalIncentives {
		rsps[i] = apiInternalIncentives[i]
	}

	return makeApiResponse(rsps)
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

func ApiGetRewardsByLpTokenId(targetLpTokenId uint64) string {
	deposit := deposits.Get(targetLpTokenId)

	reward := calcPositionReward(std.ChainHeight(), targetLpTokenId)

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

	rsps := make([]JsonResponse, len(rewards))
	for i := range rewards {
		rsps[i] = rewards[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetRewardsByAddress(targetAddress string) string {
	lpTokenRewards := []LpTokenReward{}

	stakers.IterateAll(std.Address(targetAddress), func(positionId uint64, deposit *Deposit) bool {
		rewards := []ApiReward{}

		reward := calcPositionReward(std.ChainHeight(), positionId)

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

		lpTokenReward := newLpTokenReward(positionId, deposit.owner.String(), rewards)
		lpTokenRewards = append(lpTokenRewards, lpTokenReward)

		return false
	})

	rsps := make([]JsonResponse, len(lpTokenRewards))
	for i := range lpTokenRewards {
		rsps[i] = lpTokenRewards[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetStakes() string {
	stakes := []ApiStake{}
	deposits.Iterate(0, math.MaxUint64, func(positionId uint64, deposit *Deposit) bool {
		stk := newApiStake(positionId, deposit)
		stakes = append(stakes, stk)
		return false
	})

	rsps := make([]JsonResponse, len(stakes))
	for i := range stakes {
		rsps[i] = stakes[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetStakesByLpTokenId(targetLpTokenId uint64) string {
	stakes := []ApiStake{}

	deposit := deposits.Get(targetLpTokenId)
	stk := newApiStake(targetLpTokenId, deposit)
	stakes = append(stakes, stk)

	rsps := make([]JsonResponse, len(stakes))
	for i := range stakes {
		rsps[i] = stakes[i]
	}

	return makeApiResponse(rsps)
}

func ApiGetStakesByAddress(targetAddress string) string {
	stakes := []ApiStake{}

	stakers.IterateAll(std.Address(targetAddress), func(positionId uint64, deposit *Deposit) bool {
		stk := newApiStake(positionId, deposit)
		stakes = append(stakes, stk)
		return false
	})

	rsps := make([]JsonResponse, len(stakes))
	for i := range stakes {
		rsps[i] = stakes[i]
	}

	return makeApiResponse(rsps)
}

// for off chain to check if lpTokenId is staked via RPC
func IsStaked(positionId uint64) bool {
	return deposits.Has(positionId)
}
