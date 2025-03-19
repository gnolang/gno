package staker

import (
	"std"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"
)

type PoolData struct {
	PoolPath         string
	Tier             uint64
	ActiveIncentives []string
	StakedLiquidity  *u256.Uint
}

type IncentiveData struct {
	IncentiveID    string
	StartTimestamp int64
	EndTimestamp   int64
	RewardToken    string
	RewardAmount   *u256.Uint
	Refundee       std.Address
	PoolPath       string
}

type DepositData struct {
	TokenID        uint64
	Owner          std.Address
	TargetPoolPath string
	StakeTimestamp int64
	Liquidity      *u256.Uint
	WarmupCount    int
}

// QueryPoolData returns combined pool data including tier, incentives and current staked liquidity
func QueryPoolData(poolPath string) (*PoolData, error) {
	pool, exist := pools.Get(poolPath)
	if !exist {
		return nil, ufmt.Errorf("pool %s not found", poolPath)
	}

	currentHeight := std.ChainHeight()
	tier := poolTier.CurrentTier(poolPath)

	ictvIds := filterActiveIncentives(pool, currentHeight)

	return &PoolData{
		PoolPath:         poolPath,
		Tier:             tier,
		ActiveIncentives: ictvIds,
		StakedLiquidity:  pool.CurrentStakedLiquidity(std.ChainHeight()),
	}, nil
}

// QueryIncentiveData returns detailed information about a specific incentive
func QueryIncentiveData(incentiveId string) (*IncentiveData, error) {
	var found bool
	var data IncentiveData

	pools.tree.Iterate("", "", func(key string, value interface{}) bool {
		pool := value.(*Pool)

		pool.incentives.byTime.Iterate("", "", func(key string, value interface{}) bool {
			if key == incentiveId {
				ictv := value.(*ExternalIncentive)
				data = IncentiveData{
					IncentiveID:    incentiveId,
					StartTimestamp: ictv.startTimestamp,
					EndTimestamp:   ictv.endTimestamp,
					RewardToken:    ictv.rewardToken,
					RewardAmount:   u256.NewUint(ictv.rewardAmount),
					Refundee:       ictv.refundee,
					PoolPath:       pool.poolPath,
				}
				found = true
				return true
			}
			return false
		})

		return found
	})

	if !found {
		return nil, ufmt.Errorf("incentiveId(%s) incentive does not exist", incentiveId)
	}

	return &data, nil
}

// QueryDepositData returns detailed information about a specific deposit
func QueryDepositData(lpTokenId uint64) (*DepositData, error) {
	deposit := deposits.Get(lpTokenId)
	if deposit == nil {
		return nil, ufmt.Errorf("positionId(%d) deposit does not exist", lpTokenId)
	}

	return &DepositData{
		TokenID:        lpTokenId,
		Owner:          deposit.owner,
		TargetPoolPath: deposit.targetPoolPath,
		StakeTimestamp: deposit.stakeTimestamp,
		Liquidity:      deposit.liquidity,
		WarmupCount:    len(deposit.warmups),
	}, nil
}

func filterActiveIncentives(pool *Pool, currentHeight int64) []string {
	ictvIds := make([]string, 0)
	pool.incentives.byHeight.Iterate("", "", func(key string, value interface{}) bool {
		ictv := value.(*ExternalIncentive)
		if ictv.startHeight <= currentHeight && currentHeight < ictv.endHeight {
			ictvId := incentiveIdByTime(
				ictv.startHeight,
				ictv.endHeight,
				ictv.refundee,
				ictv.rewardToken,
			)
			ictvIds = append(ictvIds, ictvId)
		}
		return false
	})
	return ictvIds
}
