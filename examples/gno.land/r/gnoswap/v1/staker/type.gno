package staker

import (
	"math"
	"std"

	u256 "gno.land/p/gnoswap/uint256"
)

// ExternalIncentive is a struct for storing external incentive information.
type ExternalIncentive struct {
	incentiveId      string      // incentive id
	startTimestamp   int64       // start time for external reward
	endTimestamp     int64       // end time for external reward
	createdHeight    int64       // block height when the incentive was created
	depositGnsAmount int64       // deposited gns amount
	targetPoolPath   string      // external reward target pool path
	rewardToken      string      // external reward token path
	rewardAmount     int64       // total reward amount
	rewardLeft       int64       // remaining reward amount
	startHeight      int64       // start height for external reward
	endHeight        int64       // end height for external reward
	rewardPerBlock   int64       // reward per block
	refundee         std.Address // refundee address

	unclaimableRefunded bool // whether unclaimable reward is refunded
}

// safeMulInt64 performs safe multiplication of int64 values, panicking on overflow
func safeMulInt64(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	if a > 0 && b > 0 {
		if a > math.MaxInt64/b {
			panic("int64 multiplication overflow")
		}
	} else if a < 0 && b < 0 {
		if a < math.MaxInt64/b {
			panic("int64 multiplication overflow")
		}
	} else if a > 0 && b < 0 {
		if b < math.MinInt64/a {
			panic("int64 multiplication underflow")
		}
	} else { // a < 0 && b > 0
		if a < math.MinInt64/b {
			panic("int64 multiplication underflow")
		}
	}
	return a * b
}

// safeAddInt64 performs safe addition of int64 values, panicking on overflow
func safeAddInt64(a, b int64) int64 {
	if a > 0 && b > math.MaxInt64-a {
		panic("int64 addition overflow")
	}
	if a < 0 && b < math.MinInt64-a {
		panic("int64 addition underflow")
	}
	return a + b
}

// safeSubInt64 performs safe subtraction of int64 values, panicking on underflow
func safeSubInt64(a, b int64) int64 {
	if b > 0 && a < math.MinInt64+b {
		panic("int64 subtraction underflow")
	}
	if b < 0 && a > math.MaxInt64+b {
		panic("int64 subtraction overflow")
	}
	return a - b
}

func (e ExternalIncentive) StartTimestamp() int64 {
	return e.startTimestamp
}

func (e ExternalIncentive) EndTimestamp() int64 {
	return e.endTimestamp
}

func (e ExternalIncentive) RewardToken() string {
	return e.rewardToken
}

func (e ExternalIncentive) RewardAmount() int64 {
	return e.rewardAmount
}

func (self *ExternalIncentive) RewardSpent(currentHeight uint64) int64 {
	if currentHeight < uint64(self.startHeight) {
		return 0
	}

	if currentHeight > uint64(self.endHeight) {
		return int64(self.rewardAmount)
	}

	blocks := currentHeight - uint64(self.startHeight)
	rewardSpent := safeMulInt64(int64(blocks), self.rewardPerBlock)
	return rewardSpent
}

func (self *ExternalIncentive) RewardLeft(currentHeight uint64) int64 {
	if currentHeight <= uint64(self.startHeight) {
		return int64(self.rewardAmount)
	}

	if currentHeight > uint64(self.endHeight) {
		return 0
	}

	if currentHeight == uint64(self.endHeight) {
		return self.rewardPerBlock
	}

	blocks := uint64(self.endHeight) - currentHeight
	rewardLeft := safeMulInt64(int64(blocks), self.rewardPerBlock)
	return rewardLeft
}

func (self *ExternalIncentive) Clone() *ExternalIncentive {
	return &ExternalIncentive{
		incentiveId:         self.incentiveId,
		startTimestamp:      self.startTimestamp,
		endTimestamp:        self.endTimestamp,
		createdHeight:       self.createdHeight,
		depositGnsAmount:    self.depositGnsAmount,
		targetPoolPath:      self.targetPoolPath,
		rewardToken:         self.rewardToken,
		rewardAmount:        self.rewardAmount,
		rewardLeft:          self.rewardLeft,
		startHeight:         self.startHeight,
		endHeight:           self.endHeight,
		rewardPerBlock:      self.rewardPerBlock,
		refundee:            self.refundee,
		unclaimableRefunded: self.unclaimableRefunded,
	}
}

// NewExternalIncentive creates a new external incentive
func NewExternalIncentive(
	incentiveId string,
	targetPoolPath string,
	rewardToken string,
	rewardAmount int64,
	startTimestamp int64, // timestamp is in unix time(seconds)
	endTimestamp int64,
	refundee std.Address,
	createdHeight int64,
	depositGnsAmount int64,
	currentTime int64, // current time in unix time(seconds)
	msPerBlock int64, // msPerBlock is in milliseconds
) *ExternalIncentive {
	incentiveDuration := endTimestamp - startTimestamp
	incentiveBlock := incentiveDuration * 1000 / msPerBlock
	rewardPerBlock := rewardAmount / int64(incentiveBlock)

	blocksLeftUntilStartHeight := (startTimestamp - currentTime) * 1000 / msPerBlock
	blocksLeftUntilEndHeight := (endTimestamp - currentTime) * 1000 / msPerBlock

	startHeight := std.ChainHeight() + blocksLeftUntilStartHeight
	endHeight := std.ChainHeight() + blocksLeftUntilEndHeight

	return &ExternalIncentive{
		incentiveId:         incentiveId,
		targetPoolPath:      targetPoolPath,
		rewardToken:         rewardToken,
		rewardAmount:        rewardAmount,
		startTimestamp:      startTimestamp,
		endTimestamp:        endTimestamp,
		startHeight:         startHeight,
		endHeight:           endHeight,
		rewardPerBlock:      rewardPerBlock,
		refundee:            refundee,
		createdHeight:       createdHeight,
		depositGnsAmount:    depositGnsAmount,
		unclaimableRefunded: false,
	}
}

type Deposit struct {
	owner             std.Address // owner address
	stakeTimestamp    int64       // staked time
	stakeHeight       int64       // staked block height
	targetPoolPath    string      // staked position's pool path
	tickLower         int32       // tick lower
	tickUpper         int32       // tick upper
	liquidity         *u256.Uint  // liquidity
	lastCollectHeight int64       // last collect block height
	warmups           []Warmup    // warmup information
}
