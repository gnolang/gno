package staker

import (
	"math"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	gns "gno.land/r/gnoswap/v1/gns"
)

type Warmup struct {
	Index            int
	BlockDuration    int64
	NextWarmupHeight int64 // set to 0 for template
	WarmupRatio      uint64
}

var warmupTemplate []Warmup = DefaultWarmupTemplate()

func DefaultWarmupTemplate() []Warmup {
	msInDay := int64(86400000)
	blocksInDay := msInDay / int64(gns.GetAvgBlockTimeInMs())
	blocksIn5Days := int64(5 * blocksInDay)
	blocksIn10Days := int64(10 * blocksInDay)
	blocksIn30Days := int64(30 * blocksInDay)

	// NextWarmupHeights are set to 0 for template.
	// It will be set by InstantiateWarmup()
	return []Warmup{
		{
			Index:         0,
			BlockDuration: blocksIn5Days,
			//NextWarmupHeight: currentHeight + blocksIn5Days,
			WarmupRatio: 30,
		},
		{
			Index:         1,
			BlockDuration: blocksIn10Days,
			//NextWarmupHeight: currentHeight + blocksIn10Days,
			WarmupRatio: 50,
		},
		{
			Index:         2,
			BlockDuration: blocksIn30Days,
			//NextWarmupHeight: currentHeight + blocksIn30Days,
			WarmupRatio: 70,
		},
		{
			Index:         3,
			BlockDuration: math.MaxInt64,
			//NextWarmupHeight: math.MaxInt64,
			WarmupRatio: 100,
		},
	}
}

// expected to be called by governance
func modifyWarmup(index int, blockDuration int64) {
	if index >= len(warmupTemplate) {
		panic(ufmt.Sprintf("index(%d) is out of range", index))
	}

	warmupTemplate[index].BlockDuration = blockDuration
}

func instantiateWarmup(currentHeight int64) []Warmup {
	warmups := make([]Warmup, 0)
	for _, warmup := range warmupTemplate {
		nextWarmupHeight := currentHeight + warmup.BlockDuration
		if nextWarmupHeight < 0 {
			nextWarmupHeight = math.MaxInt64
		}

		warmups = append(warmups, Warmup{
			Index:            warmup.Index,
			BlockDuration:    warmup.BlockDuration,
			NextWarmupHeight: nextWarmupHeight,
			WarmupRatio:      warmup.WarmupRatio,
		})
		currentHeight += warmup.BlockDuration
	}
	return warmups
}

func (warmup *Warmup) apply(poolReward uint64, positionLiquidity, stakedLiquidity *u256.Uint) (uint64, uint64) {
	poolRewardUint := u256.NewUint(poolReward)
	perPositionReward := u256.Zero().Mul(poolRewardUint, positionLiquidity)
	perPositionReward = u256.Zero().Div(perPositionReward, stakedLiquidity)
	rewardRatio := u256.NewUint(warmup.WarmupRatio)
	penaltyRatio := u256.NewUint(100 - warmup.WarmupRatio)
	totalReward := u256.Zero().Mul(perPositionReward, rewardRatio)
	totalReward = u256.Zero().Div(totalReward, u256.NewUint(100))
	totalPenalty := u256.Zero().Mul(perPositionReward, penaltyRatio)
	totalPenalty = u256.Zero().Div(totalPenalty, u256.NewUint(100))
	return totalReward.Uint64(), totalPenalty.Uint64()
}

func (self *Deposit) FindWarmup(currentHeight int64) int {
	for i, warmup := range self.warmups {
		if currentHeight < warmup.NextWarmupHeight {
			return i
		}
	}
	return len(self.warmups) - 1
}

func (self *Deposit) GetWarmup(index int) Warmup {
	return self.warmups[index]
}
