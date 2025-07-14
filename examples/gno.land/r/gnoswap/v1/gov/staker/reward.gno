package staker

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"
	"gno.land/r/gnoswap/v1/halt"
	pf "gno.land/r/gnoswap/v1/protocol_fee"
)

var (
	currentGNSBalance         uint64            = 0                       // GNS balance of gov/staker minted by emission
	currentProtocolFeeBalance map[string]uint64 = make(map[string]uint64) // protocol fee balance of gov/staker (tokenPath -> balance)

	q96         = u256.MustFromDecimal(consts.Q96)
	q128        = uint(128)
	rewardState = NewRewardState() // reward state of gov/staker

	// For Launchpad deposit
	amountByProjectWallet = avl.NewTree() // recipient wallet => amount
)

// collectAndGetGNSBalance gets the acuumulated GNS balance from emission to gov/staker
// Need to clear the distributed GNS to gov/staker in emission to get the latest accumulated GNS balance
func collectAndGetGNSBalance() uint64 {
	gotGnsForEmission := en.GetDistributedToGovStaker()
	en.ClearDistributedToGovStaker(cross)

	currentGNSBalance += uint64(gotGnsForEmission)

	return currentGNSBalance
}

// updateAndGetProtocolFeeBalance gets the accumulated protocol fee balance from protocol_fee to gov/staker
// Need to clear the accumulated protocol fee to gov/staker in protocol fee to get the latest accumulated protocol fee balance
func updateAndGetProtocolFeeBalance() map[string]uint64 {
	pf.DistributeProtocolFee(cross)
	gotAccuProtocolFee := pf.GetAccuTransferToGovStaker()

	gotAccuProtocolFee.Iterate("", "", func(key string, value any) bool {
		func(cur realm) {
			amount := value.(uint64)
			currentProtocolFeeBalance[key] += amount
		}(cross)
		return false
	})

	pf.ClearAccuTransferToGovStaker(cross)

	return currentProtocolFeeBalance
}

type StakerRewardInfo struct {
	StartHeight              uint64                // height when staker started staking
	RewardDebt               *u256.Uint            // price debt per xGNS stake, Q128
	ProtocolFeeDebt          map[string]*u256.Uint // protocol fee debt per xGNS stake, Q128
	StakedAmount             uint64                // amount of xGNS staked
	PendingGnsReward         uint64                // amount of GNS reward pending
	PendingProtocolFeeReward map[string]uint64     // protocol fee amount pending per token
	GnsRewardClaimed         uint64                // amount of GNS reward claimed so far
	ProtocolFeeClaimed       map[string]uint64     // protocol fee amount claimed per token
	LastClaimedHeight        uint64                // last height when the reward was claimed
}

func (self *StakerRewardInfo) PriceDebtUint64() uint64 {
	return u256.Zero().Rsh(self.RewardDebt, q128).Uint64()
}

type RewardState struct {
	// CurrentBalance is sum of all the previous balances, including the reward distribution.
	totalXGNSStake                 uint64                // total xGNS staked in the system
	currentGNSBalance              uint64                // current GNS balance of gov/staker, used to calculate GNSRewardAccumulation
	currentProtocolFeeBalance      map[string]uint64     // current protocol fee balance per token path
	accumulatedGNSRewardPerStake   *u256.Uint            // accumulated claimable GNS per xGNS stake, Q128 format
	accumulatedProtocolFeePerStake map[string]*u256.Uint // accumulated claimable protocol fee per xGNS stake, Q128 format
	stakerInfos                    *avl.Tree             // address -> StakerRewardInfo
}

func NewRewardState() *RewardState {
	return &RewardState{
		totalXGNSStake:                 0,
		currentGNSBalance:              0,
		currentProtocolFeeBalance:      make(map[string]uint64),
		accumulatedGNSRewardPerStake:   u256.Zero(),
		accumulatedProtocolFeePerStake: make(map[string]*u256.Uint),
		stakerInfos:                    avl.NewTree(),
	}
}

func (self *RewardState) Info(staker std.Address) StakerRewardInfo {
	infoI, exists := self.stakerInfos.Get(staker.String())
	if !exists {
		return StakerRewardInfo{
			StartHeight:              uint64(std.ChainHeight()),
			RewardDebt:               u256.Zero(),
			ProtocolFeeDebt:          make(map[string]*u256.Uint),
			StakedAmount:             0,
			PendingGnsReward:         0,
			PendingProtocolFeeReward: make(map[string]uint64),
			GnsRewardClaimed:         0,
			ProtocolFeeClaimed:       make(map[string]uint64),
			LastClaimedHeight:        0,
		}
	}
	return infoI.(StakerRewardInfo)
}

func (self *RewardState) CalculateReward(staker std.Address) (uint64, map[string]uint64) {
	info := self.Info(staker)
	stakerPrice := u256.Zero().Sub(self.accumulatedGNSRewardPerStake, info.RewardDebt)
	reward := stakerPrice.Mul(stakerPrice, u256.NewUint(info.StakedAmount))
	reward = reward.Rsh(reward, q128)
	calcReward := reward.Uint64() + info.PendingGnsReward

	protocolFeeRewards := make(map[string]uint64)
	for tokenPath, protocolFeePriceAccumulation := range self.accumulatedProtocolFeePerStake {
		protocolFeePriceDebt, ok := info.ProtocolFeeDebt[tokenPath]
		if !ok {
			protocolFeePriceDebt = u256.Zero()
		}
		protocolFeePrice := u256.Zero().Sub(protocolFeePriceAccumulation, protocolFeePriceDebt)
		protocolFeeReward := u256.Zero().Mul(protocolFeePrice, u256.NewUint(info.StakedAmount))
		protocolFeeReward = protocolFeeReward.Rsh(protocolFeeReward, q128)
		protocolFeeReward64 := protocolFeeReward.Uint64()
		pendingReward, exist := info.PendingProtocolFeeReward[tokenPath]
		if exist {
			protocolFeeReward64 += pendingReward
		}
		if protocolFeeReward64 > 0 {
			protocolFeeRewards[tokenPath] = protocolFeeReward64
		}
	}
	return calcReward, protocolFeeRewards
}

func (self *RewardState) PriceAccumulationUint64() uint64 {
	return u256.Zero().Rsh(self.accumulatedGNSRewardPerStake, q128).Uint64()
}

// amount MUST be less than or equal to the amount of xGNS staked
// This function does not check it
func (self *RewardState) deductReward(staker std.Address) (uint64, map[string]uint64) {
	info := self.Info(staker)
	if info.LastClaimedHeight == uint64(std.ChainHeight()) {
		return 0, make(map[string]uint64)
	}

	collectableReward := uint64(0)
	actualAccumulatedRewardPerStake := u256.Zero().Sub(self.accumulatedGNSRewardPerStake, info.RewardDebt)
	if actualAccumulatedRewardPerStake.Gt(u256.Zero()) {
		rewardQ128 := u256.Zero().Mul(actualAccumulatedRewardPerStake, u256.NewUint(info.StakedAmount))
		reward := rewardQ128.Rsh(rewardQ128, q128)
		totalReward := reward.Uint64()
		if info.GnsRewardClaimed > 0 {
			totalReward -= info.GnsRewardClaimed
		}
		if info.PendingGnsReward > 0 {
			totalReward += info.PendingGnsReward
			info.PendingGnsReward = 0
		}
		collectableReward = totalReward
		info.GnsRewardClaimed += collectableReward
		self.currentGNSBalance -= collectableReward
	}

	protocolFeeRewards := make(map[string]uint64)
	for tokenPath, protocolFeePerStakeAccumulation := range self.accumulatedProtocolFeePerStake {
		protocolFeePriceDebt, ok := info.ProtocolFeeDebt[tokenPath]
		if !ok {
			protocolFeePriceDebt = u256.Zero()
		}
		protocolFeePerStakeDelta := u256.Zero().Sub(protocolFeePerStakeAccumulation, protocolFeePriceDebt)
		protocolFeeRewardQ128 := u256.Zero().Mul(protocolFeePerStakeDelta, u256.NewUint(info.StakedAmount))
		protocolFeeReward := u256.Zero().Rsh(protocolFeeRewardQ128, q128)
		collectableProtocolFeeReward := protocolFeeReward.Uint64()
		pendingReward, exist := info.PendingProtocolFeeReward[tokenPath]
		if exist {
			collectableProtocolFeeReward += pendingReward
			info.PendingProtocolFeeReward[tokenPath] = 0
		}
		alreadyClaimed, exist := info.ProtocolFeeClaimed[tokenPath]
		if exist {
			collectableProtocolFeeReward -= alreadyClaimed
		}
		if !(collectableProtocolFeeReward > 0) {
			continue
		}

		protocolFeeRewards[tokenPath] = uint64(collectableProtocolFeeReward)
		info.ProtocolFeeClaimed[tokenPath] += protocolFeeRewards[tokenPath]
		self.currentProtocolFeeBalance[tokenPath] -= protocolFeeRewards[tokenPath]
	}

	info.LastClaimedHeight = uint64(std.ChainHeight())
	self.stakerInfos.Set(staker.String(), info)

	return collectableReward, protocolFeeRewards
}

// AccumulatedProtocolFeeRewardPerStakeOf returns the accumulated protocol fee reward per stake of the token path
func (self *RewardState) AccumulatedProtocolFeeRewardPerStakeOf(tokenPath string) *u256.Uint {
	accumulated, ok := self.accumulatedProtocolFeePerStake[tokenPath]
	if !ok {
		return u256.Zero()
	}
	return accumulated
}

// This function MUST be called as a part of addStake or removeStake
// currentGNSBalance / StakeChange / IsRemoveStake will be updated in those functions
func (self *RewardState) finalize(currentBalance uint64, currentProtocolFeeBalances map[string]uint64) {
	if self.totalXGNSStake == uint64(0) {
		// no staker
		return
	}

	// emission reward
	delta := currentBalance - self.currentGNSBalance
	if delta > 0 {
		deltaU256 := u256.NewUint(delta)
		deltaU256Q128 := deltaU256.Lsh(deltaU256, q128)
		emissionRewardPerOneStakedQ128 := u256.Zero().Div(deltaU256Q128, u256.NewUint(self.totalXGNSStake))
		self.accumulatedGNSRewardPerStake = u256.Zero().Add(self.accumulatedGNSRewardPerStake, emissionRewardPerOneStakedQ128)
		self.currentGNSBalance = currentBalance
	}

	// protocol fee reward
	for tokenPath, protocolFeeBalance := range currentProtocolFeeBalances {
		protocolFeeDelta := protocolFeeBalance - self.currentProtocolFeeBalance[tokenPath]
		if protocolFeeDelta == 0 {
			continue
		}
		protocolFeeDeltaU256 := u256.NewUint(protocolFeeDelta)
		protocolFeeDeltaU256Q128 := protocolFeeDeltaU256.Lsh(protocolFeeDeltaU256, q128)
		protocolFeeRewardPerOneStakedQ128 := u256.Zero().Div(protocolFeeDeltaU256Q128, u256.NewUint(self.totalXGNSStake))
		self.accumulatedProtocolFeePerStake[tokenPath] = u256.Zero().Add(self.AccumulatedProtocolFeeRewardPerStakeOf(tokenPath), protocolFeeRewardPerOneStakedQ128)
		self.currentProtocolFeeBalance[tokenPath] = protocolFeeBalance
	}
}

// updateReward updates the reward of the staker
func (rs *RewardState) updateReward(staker std.Address) {
	info := rs.Info(staker)

	if info.LastClaimedHeight == uint64(std.ChainHeight()) {
		return
	}

	actualAccumulatedRewardPerStake := u256.Zero().Sub(rs.accumulatedGNSRewardPerStake, info.RewardDebt)
	if actualAccumulatedRewardPerStake.Gt(u256.Zero()) {
		collectableRewardQ128 := u256.Zero().Mul(actualAccumulatedRewardPerStake, u256.NewUint(info.StakedAmount))
		collectableReward := collectableRewardQ128.Rsh(collectableRewardQ128, q128)
		totalReward := collectableReward.Uint64()
		if info.GnsRewardClaimed > 0 {
			totalReward -= info.GnsRewardClaimed
		}
		info.PendingGnsReward += totalReward
	}

	for tokenPath, protocolFeePerStakeAccumulation := range rs.accumulatedProtocolFeePerStake {
		protocolFeePriceDebt, ok := info.ProtocolFeeDebt[tokenPath]
		if !ok {
			protocolFeePriceDebt = u256.Zero()
		}
		protocolFeePerStakeDelta := u256.Zero().Sub(protocolFeePerStakeAccumulation, protocolFeePriceDebt)
		protocolFeeRewardQ128 := u256.Zero().Mul(protocolFeePerStakeDelta, u256.NewUint(info.StakedAmount))
		protocolFeeReward := u256.Zero().Rsh(protocolFeeRewardQ128, q128)
		collectableProtocolFeeReward := protocolFeeReward.Uint64()
		alreadyClaimed, exist := info.ProtocolFeeClaimed[tokenPath]
		if exist {
			if collectableProtocolFeeReward < alreadyClaimed {
				collectableProtocolFeeReward = 0
			} else {
				collectableProtocolFeeReward -= alreadyClaimed
			}
		}
		info.PendingProtocolFeeReward[tokenPath] = collectableProtocolFeeReward
	}

	rs.stakerInfos.Set(staker.String(), info)
}

// addStake adds the staker's stake to the reward state
func (self *RewardState) addStake(currentHeight uint64, staker std.Address, amount uint64, currentBalance uint64, currentProtocolFeeBalances map[string]uint64) {
	self.finalize(currentBalance, currentProtocolFeeBalances)

	if !staker.IsValid() {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("invalid staker address: %s", staker.String()),
		))
	}

	self.totalXGNSStake += amount
	if self.stakerInfos.Has(staker.String()) {
		self.updateReward(staker)
		info := self.Info(staker)
		info.RewardDebt = self.accumulatedGNSRewardPerStake.Clone()
		for tokenPath, protocolFeePriceAccumulation := range self.accumulatedProtocolFeePerStake {
			info.ProtocolFeeDebt[tokenPath] = protocolFeePriceAccumulation.Clone()
		}
		info.StakedAmount += amount
		self.stakerInfos.Set(staker.String(), info)
		return
	}

	info := StakerRewardInfo{
		StartHeight:              currentHeight,
		RewardDebt:               self.accumulatedGNSRewardPerStake.Clone(),
		ProtocolFeeDebt:          make(map[string]*u256.Uint),
		StakedAmount:             amount,
		PendingGnsReward:         0,
		PendingProtocolFeeReward: make(map[string]uint64),
		GnsRewardClaimed:         0,
		ProtocolFeeClaimed:       make(map[string]uint64),
		LastClaimedHeight:        0,
	}

	for tokenPath, protocolFeePriceAccumulation := range self.accumulatedProtocolFeePerStake {
		info.ProtocolFeeDebt[tokenPath] = protocolFeePriceAccumulation.Clone()
	}

	self.stakerInfos.Set(staker.String(), info)
}

func (self *RewardState) claim(staker std.Address, currentBalance uint64, currentProtocolFeeBalances map[string]uint64) (uint64, map[string]uint64) {
	if !self.stakerInfos.Has(staker.String()) {
		return 0, make(map[string]uint64)
	}

	stakerRewardInfo := self.Info(staker)
	currentHeight := uint64(std.ChainHeight())
	if stakerRewardInfo.LastClaimedHeight == currentHeight {
		return 0, make(map[string]uint64)
	}

	self.finalize(currentBalance, currentProtocolFeeBalances)

	reward, protocolFeeRewards := self.deductReward(staker)

	return reward, protocolFeeRewards
}

func (self *RewardState) removeStake(
	staker std.Address,
	amount uint64,
	currentBalance uint64,
	currentProtocolFeeBalances map[string]uint64,
) (uint64, map[string]uint64) {
	self.finalize(currentBalance, currentProtocolFeeBalances)

	reward, protocolFeeRewards := self.deductReward(staker)

	stakerStr := staker.String()
	stakerInfoValue, exists := self.stakerInfos.Get(stakerStr)
	if !exists {
		panic("staker info not found")
	}
	stakerInfo := stakerInfoValue.(StakerRewardInfo)

	// remove completely only if the total stake amount is equal to the undelegate amount
	if stakerInfo.StakedAmount == amount {
		self.stakerInfos.Remove(stakerStr)
	} else {
		// reduce the partial amount
		stakerInfo.StakedAmount -= amount
		self.stakerInfos.Set(stakerStr, stakerInfo)
	}

	self.totalXGNSStake -= amount

	return reward, protocolFeeRewards
}

// GetRewardByProjectWallet returns the amount of reward for the project wallet
func GetRewardByProjectWallet(addr std.Address) (uint64, map[string]uint64) {
	rewardState.finalize(collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())
	emissionReward, protocolFeeRewards := rewardState.CalculateReward(addr)
	return emissionReward, protocolFeeRewards
}

// SetAmountByProjectWallet sets the amount of reward for the project wallet
func SetAmountByProjectWallet(cur realm, addr std.Address, amount uint64, add bool) {
	caller := std.PreviousRealm().Address()
	if err := common.LaunchpadOnly(caller); err != nil {
		panic(err)
	}
	if err := halt.IsHalted(); err != nil {
		panic(err)
	}

	currentAmount := getAmountByProjectWallet(addr)
	if add {
		amountByProjectWallet.Set(addr.String(), currentAmount+amount)
		rewardState.addStake(uint64(std.ChainHeight()), addr, amount, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())
	} else {
		amountByProjectWallet.Set(addr.String(), currentAmount-amount)
		rewardState.removeStake(addr, amount, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())
	}
}

// isLaunchpadProjectWallet returns true if the address is a project wallet
func isLaunchpadProjectWallet(addr std.Address) bool {
	if amountByProjectWallet.Has(addr.String()) {
		return true
	}
	return false
}

// getAmountByProjectWallet returns the amount of xGNS staked by the project wallet
func getAmountByProjectWallet(addr std.Address) uint64 {
	value, exists := amountByProjectWallet.Get(addr.String())
	if !exists {
		return 0
	}
	return value.(uint64)
}
