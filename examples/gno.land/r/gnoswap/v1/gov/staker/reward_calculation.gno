package staker

import (
	"std"

	en "gno.land/r/gnoswap/v1/emission"
	pf "gno.land/r/gnoswap/v1/protocol_fee"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/p/demo/avl"
	ufmt "gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"
)

var (
	currentGNSBalance         uint64            = 0                       // GNS balance of gov/staker minted by emission
	currentProtocolFeeBalance map[string]uint64 = make(map[string]uint64) // protocol fee balance of gov/staker (tokenPath -> balance)
)

// getCurrentBalance gets the acuumulated GNS balance from emission to gov/staker
// Need to clear the distributed GNS to gov/staker in emission to get the latest accumulated GNS balance
func getCurrentBalance() uint64 {
	gotGnsForEmission := en.GetDistributedToGovStaker()
	en.ClearDistributedToGovStaker()

	currentGNSBalance += gotGnsForEmission

	return currentGNSBalance
}

// getCurrentProtocolFeeBalance gets the accumulated protocol fee balance from protocol_fee to gov/staker
// Need to clear the accumulated protocol fee to gov/staker in protocol fee to get the latest accumulated protocol fee balance
func getCurrentProtocolFeeBalance() map[string]uint64 {
	pf.DistributeProtocolFee()
	gotAccuProtocolFee := pf.GetAccuTransferToGovStaker()

	gotAccuProtocolFee.Iterate("", "", func(key string, value interface{}) bool {
		amount := value.(uint64)
		currentProtocolFeeBalance[key] += amount
		return false
	})

	pf.ClearAccuTransferToGovStaker()

	return currentProtocolFeeBalance
}

type StakerRewardInfo struct {
	StartHeight          uint64                // height when staker started staking
	PriceDebt            *u256.Uint            // price debt per xGNS stake, Q128
	ProtocolFeePriceDebt map[string]*u256.Uint // protocol fee debt per xGNS stake, Q128
	Amount               uint64                // amount of xGNS staked
	Claimed              uint64                // amount of GNS reward claimed so far
	ProtocolFeeClaimed   map[string]uint64     // protocol fee amount claimed per token
}

func (self *StakerRewardInfo) Debug() string {
	return ufmt.Sprintf("{ StartHeight: %d, PriceDebt: %d, Amount: %d, Claimed: %d }", self.StartHeight, self.PriceDebtUint64(), self.Amount, self.Claimed)
}

func (self *StakerRewardInfo) PriceDebtUint64() uint64 {
	return u256.Zero().Rsh(self.PriceDebt, 128).Uint64()
}

type RewardState struct {
	// CurrentBalance is sum of all the previous balances, including the reward distribution.
	CurrentBalance               uint64                // current balance of gov/staker, used to calculate RewardAccumulation
	CurrentProtocolFeeBalance    map[string]uint64     // current protocol fee balance per token
	PriceAccumulation            *u256.Uint            // claimable GNS per xGNS stake, Q128
	ProtocolFeePriceAccumulation map[string]*u256.Uint // protocol fee debt per xGNS stake, Q128
	TotalStake                   uint64                // total xGNS staked
	LastHeight                   uint64                // last height when the reward was calculated,

	info *avl.Tree // address -> StakerRewardInfo
}

func NewRewardState() *RewardState {
	return &RewardState{
		info:                         avl.NewTree(),
		CurrentBalance:               0,
		CurrentProtocolFeeBalance:    make(map[string]uint64),
		PriceAccumulation:            u256.Zero(),
		ProtocolFeePriceAccumulation: make(map[string]*u256.Uint),
		TotalStake:                   0,
	}
}

var rewardState = NewRewardState()

func (self *RewardState) Debug() string {
	return ufmt.Sprintf("{ CurrentBalance: %d, PriceAccumulation: %d, TotalStake: %d, info: len(%d) }", self.CurrentBalance, self.PriceAccumulationUint64(), self.TotalStake, self.info.Size())
}

func (self *RewardState) Info(staker std.Address) StakerRewardInfo {
	infoI, exists := self.info.Get(staker.String())
	if !exists {
		return StakerRewardInfo{
			StartHeight:          uint64(std.GetHeight()),
			PriceDebt:            u256.Zero(),
			ProtocolFeePriceDebt: make(map[string]*u256.Uint),
			Amount:               0,
			Claimed:              0,
			ProtocolFeeClaimed:   make(map[string]uint64),
		}
	}
	return infoI.(StakerRewardInfo)
}

func (self *RewardState) CalculateReward(staker std.Address) (uint64, map[string]uint64) {
	info := self.Info(staker)
	stakerPrice := u256.Zero().Sub(self.PriceAccumulation, info.PriceDebt)
	reward := stakerPrice.Mul(stakerPrice, u256.NewUint(info.Amount))
	reward = reward.Rsh(reward, 128)

	protocolFeeRewards := make(map[string]uint64)
	for tokenPath, protocolFeePriceAccumulation := range self.ProtocolFeePriceAccumulation {
		protocolFeePriceDebt, ok := info.ProtocolFeePriceDebt[tokenPath]
		if !ok {
			protocolFeePriceDebt = u256.Zero()
		}
		protocolFeePrice := u256.Zero().Sub(protocolFeePriceAccumulation, protocolFeePriceDebt)
		protocolFeeReward := protocolFeePrice.Mul(protocolFeePrice, u256.NewUint(info.Amount))
		protocolFeeReward = protocolFeeReward.Rsh(protocolFeeReward, 128)
		protocolFeeReward64 := protocolFeeReward.Uint64()
		if protocolFeeReward64 > 0 {
			protocolFeeRewards[tokenPath] = protocolFeeReward64
		}
	}
	return reward.Uint64(), protocolFeeRewards
}

func (self *RewardState) PriceAccumulationUint64() uint64 {
	return u256.Zero().Rsh(self.PriceAccumulation, 128).Uint64()
}

// amount MUST be less than or equal to the amount of xGNS staked
// This function does not check it
func (self *RewardState) deductReward(staker std.Address, currentBalance uint64) (uint64, map[string]uint64) {
	info := self.Info(staker)
	stakerPrice := u256.Zero().Sub(self.PriceAccumulation, info.PriceDebt)
	reward := u256.Zero().Mul(stakerPrice, u256.NewUint(info.Amount))
	reward = reward.Rsh(reward, 128)
	reward64 := reward.Uint64()
	reward64 -= info.Claimed
	info.Claimed += reward64

	protocolFeeRewards := make(map[string]uint64)
	for tokenPath, protocolFeePriceAccumulation := range self.ProtocolFeePriceAccumulation {
		protocolFeePriceDebt, ok := info.ProtocolFeePriceDebt[tokenPath]
		if !ok {
			protocolFeePriceDebt = u256.Zero()
		}
		protocolFeePrice := u256.Zero().Sub(protocolFeePriceAccumulation, protocolFeePriceDebt)
		protocolFeeReward := protocolFeePrice.Mul(protocolFeePrice, u256.NewUint(info.Amount))
		protocolFeeReward = protocolFeeReward.Rsh(protocolFeeReward, 128)
		protocolFeeReward64 := protocolFeeReward.Uint64()
		if protocolFeeReward64 == 0 {
			continue
		}
		protocolFeeReward64 -= info.ProtocolFeeClaimed[tokenPath]
		protocolFeeRewards[tokenPath] = protocolFeeReward64
		info.ProtocolFeeClaimed[tokenPath] += protocolFeeReward64
	}

	self.info.Set(staker.String(), info)

	self.CurrentBalance = currentBalance - reward64
	for tokenPath, amount := range protocolFeeRewards {
		self.CurrentProtocolFeeBalance[tokenPath] -= amount
	}

	return reward64, protocolFeeRewards
}

// This function MUST be called as a part of AddStake or RemoveStake
// CurrentBalance / StakeChange / IsRemoveStake will be updated in those functions
func (self *RewardState) finalize(currentBalance uint64, currentProtocolFeeBalances map[string]uint64) {
	if self.TotalStake == uint64(0) {
		// no staker
		return
	}

	delta := currentBalance - self.CurrentBalance
	price := u256.NewUint(delta)
	price = price.Lsh(price, 128)
	price = price.Div(price, u256.NewUint(self.TotalStake))
	self.PriceAccumulation.Add(self.PriceAccumulation, price)
	self.CurrentBalance = currentBalance

	for tokenPath, currentProtocolFeeBalance := range currentProtocolFeeBalances {
		protocolFeeDelta := currentProtocolFeeBalance - self.CurrentProtocolFeeBalance[tokenPath]
		if protocolFeeDelta == 0 {
			continue
		}
		protocolFeePrice := u256.NewUint(protocolFeeDelta)
		protocolFeePrice = protocolFeePrice.Lsh(protocolFeePrice, 128)
		protocolFeePrice = protocolFeePrice.Div(protocolFeePrice, u256.NewUint(self.TotalStake))
		protocolFeePriceAccumulation, ok := self.ProtocolFeePriceAccumulation[tokenPath]
		if !ok {
			protocolFeePriceAccumulation = u256.Zero()
		}
		protocolFeePriceAccumulation.Add(protocolFeePriceAccumulation, protocolFeePrice)
		self.ProtocolFeePriceAccumulation[tokenPath] = protocolFeePriceAccumulation
		self.CurrentProtocolFeeBalance[tokenPath] = currentProtocolFeeBalance
	}
}

func (self *RewardState) AddStake(currentHeight uint64, staker std.Address, amount uint64, currentBalance uint64, currentProtocolFeeBalances map[string]uint64) {
	self.finalize(currentBalance, currentProtocolFeeBalances)

	self.TotalStake += amount

	if self.info.Has(staker.String()) {
		info := self.Info(staker)
		info.PriceDebt.Add(info.PriceDebt, u256.NewUint(info.Amount))
		info.PriceDebt.Add(info.PriceDebt, u256.Zero().Mul(self.PriceAccumulation, u256.NewUint(amount)))
		info.PriceDebt.Div(info.PriceDebt, u256.NewUint(self.TotalStake))
		for tokenPath, amount := range currentProtocolFeeBalances {
			protocolFeePriceDebt, ok := info.ProtocolFeePriceDebt[tokenPath]
			if !ok {
				info.ProtocolFeePriceDebt[tokenPath] = self.ProtocolFeePriceAccumulation[tokenPath].Clone()
				continue
			}
			protocolFeePriceDebt.Add(protocolFeePriceDebt, u256.NewUint(amount))
			protocolFeePriceDebt.Add(protocolFeePriceDebt, u256.Zero().Mul(self.ProtocolFeePriceAccumulation[tokenPath], u256.NewUint(amount)))
			protocolFeePriceDebt.Div(protocolFeePriceDebt, u256.NewUint(self.TotalStake))
			info.ProtocolFeePriceDebt[tokenPath] = protocolFeePriceDebt
		}
		info.Amount += amount
		self.info.Set(staker.String(), info)
		return
	}

	info := StakerRewardInfo{
		StartHeight:          currentHeight,
		PriceDebt:            self.PriceAccumulation.Clone(),
		Amount:               amount,
		Claimed:              0,
		ProtocolFeeClaimed:   make(map[string]uint64),
		ProtocolFeePriceDebt: make(map[string]*u256.Uint),
	}

	for tokenPath, protocolFeePriceAccumulation := range self.ProtocolFeePriceAccumulation {
		info.ProtocolFeePriceDebt[tokenPath] = protocolFeePriceAccumulation.Clone()
	}

	self.info.Set(staker.String(), info)
}

func (self *RewardState) Claim(staker std.Address, currentBalance uint64, currentProtocolFeeBalances map[string]uint64) (uint64, map[string]uint64) {
	if !self.info.Has(staker.String()) {
		return 0, make(map[string]uint64)
	}

	if self.LastHeight == uint64(std.GetHeight()) {
		return 0, make(map[string]uint64)
	}

	self.finalize(currentBalance, currentProtocolFeeBalances)

	reward, protocolFeeRewards := self.deductReward(staker, currentBalance)

	self.LastHeight = uint64(std.GetHeight())
	return reward, protocolFeeRewards
}

func (self *RewardState) RemoveStake(staker std.Address, amount uint64, currentBalance uint64, currentProtocolFeeBalances map[string]uint64) (uint64, map[string]uint64) {
	self.finalize(currentBalance, currentProtocolFeeBalances)

	reward, protocolFeeRewards := self.deductReward(staker, currentBalance)

	self.info.Remove(staker.String())

	self.TotalStake -= amount

	return reward, protocolFeeRewards
}

var (
	q96 = u256.MustFromDecimal(consts.Q96)
)

// === LAUNCHPAD DEPOSIT
var (
	amountByProjectWallet = avl.NewTree() // recipient wallet => amount
	rewardByProjectWallet = avl.NewTree() // recipient wallet => reward
)

func GetRewardByProjectWallet(addr std.Address) uint64 {
	value, exists := rewardByProjectWallet.Get(addr.String())
	if !exists {
		return 0
	}
	return value.(uint64)
}

func getAmountByProjectWallet(addr std.Address) uint64 {
	value, exists := amountByProjectWallet.Get(addr.String())
	if !exists {
		return 0
	}
	return value.(uint64)
}

func SetAmountByProjectWallet(addr std.Address, amount uint64, add bool) {
	caller := std.PrevRealm().Addr()
	if err := common.LaunchpadOnly(caller); err != nil {
		panic(err)
	}

	common.IsHalted()
	en.MintAndDistributeGns()

	currentAmount := getAmountByProjectWallet(addr)
	if add {
		amountByProjectWallet.Set(addr.String(), currentAmount+amount)
		rewardState.AddStake(uint64(std.GetHeight()), caller, amount, getCurrentBalance(), getCurrentProtocolFeeBalance())
	} else {
		amountByProjectWallet.Set(addr.String(), currentAmount-amount)
		rewardState.RemoveStake(caller, amount, getCurrentBalance(), getCurrentProtocolFeeBalance())
	}
}

// calculateXGnsRatio calculates the ratio of user's xGNS amount to total xGNS supply
func calculateXGnsRatio(amount uint64, xGnsX96 *u256.Uint) *u256.Uint {
	xGnsAmountX96 := new(u256.Uint).Mul(u256.NewUint(amount), q96)
	xGnsAmountX96 = new(u256.Uint).Mul(xGnsAmountX96, u256.NewUint(1_000_000_000))

	ratio := new(u256.Uint).Div(xGnsAmountX96, xGnsX96)
	ratio = ratio.Mul(ratio, q96)
	return ratio.Div(ratio, u256.NewUint(1_000_000_000))
}
