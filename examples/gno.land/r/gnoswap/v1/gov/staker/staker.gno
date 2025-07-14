package staker

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/gnoswap/v1/gov/xgns"
	"gno.land/r/gnoswap/v1/halt"
	"gno.land/r/gnoswap/v1/referral"
)

var govStaker, _ = access.GetAddress(access.ROLE_GOV_STAKER)

// lockedGNS represents a locked GNS with its amount, unlock time, and collected status
type lockedGNS struct {
	amount    uint64
	unlock    uint64
	collected bool
}

const (
	SECONDS_IN_7_DAYS = uint64(604800)
)

var (
	addrLockedGns = avl.NewTree() // address -> []lockedGNS
	lockedAmount  = uint64(0)
	minimumAmount = uint64(1_000_000)
)

// Delegate delegates GNS tokens to a specified address.
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#delegate
func Delegate(cur realm, to std.Address, amount uint64, referrer string) {
	if !to.IsValid() {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("invalid address %s to delegate", to.String()),
		))
	}

	if amount < minimumAmount {
		panic(addDetailToError(
			errLessThanMinimum,
			ufmt.Sprintf("minimum amount to delegate is %d (requested:%d)", minimumAmount, amount),
		))
	}

	if amount%minimumAmount != 0 {
		panic(addDetailToError(
			errInvalidAmount,
			ufmt.Sprintf("amount must be multiple of %d", minimumAmount),
		))
	}

	caller := std.PreviousRealm().Address()
	gnsBalanceOfCaller := gns.BalanceOf(caller)
	if uint64(gnsBalanceOfCaller) < amount {
		panic(addDetailToError(
			errNotEnoughBalance,
			ufmt.Sprintf("invalid GNS balance(%d) to delegate(%d)", gnsBalanceOfCaller, amount),
		))
	}

	// Transfer the caller's GNS to GovStaker
	gns.TransferFrom(cross, caller, std.CurrentRealm().Address(), int64(amount))

	success := referral.TryRegister(cross, caller, referrer)
	actualReferrer := referrer
	if !success {
		actualReferrer = referral.GetReferral(caller.String())
	}
	en.MintAndDistributeGns(cross)

	if running {
		cleanDelegationStatHistory()
	}
	rewardState.addStake(uint64(std.ChainHeight()), caller, amount, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())

	// actual delegate
	delegate(to, amount)

	// xGNS mint to caller
	xgns.Mint(cross, caller, int64(amount))

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Delegate",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"from", caller.String(),
		"to", to.String(),
		"amount", formatUint(amount),
		"referrer", actualReferrer,
	)
}

// Redelegate redelegates xGNS from existing delegate to another.
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#redelegate
func Redelegate(cur realm, from, to std.Address, amount uint64) {
	if !from.IsValid() {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("invalid from address %s to redelegate", to.String()),
		))
	}

	if !to.IsValid() {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("invalid to address %s to redelegate", to.String()),
		))
	}

	if amount < minimumAmount {
		panic(addDetailToError(
			errLessThanMinimum,
			ufmt.Sprintf("minimum amount to redelegate is %d (requested:%d)", minimumAmount, amount),
		))
	}

	if amount%minimumAmount != 0 {
		panic(addDetailToError(
			errInvalidAmount,
			ufmt.Sprintf("amount must be multiple of %d", minimumAmount),
		))
	}

	caller := std.PreviousRealm().Address()
	delegatedBalanceOfCaller := xgns.BalanceOf(caller)
	if uint64(delegatedBalanceOfCaller) < amount {
		panic(addDetailToError(
			errNotEnoughBalance,
			ufmt.Sprintf("not enough xGNS delegated(%d) to redelegate(%d)", delegatedBalanceOfCaller, amount),
		))
	}

	en.MintAndDistributeGns(cross)

	if running {
		cleanDelegationStatHistory()
	}

	CollectReward(cur)
	unDelegate(from, amount)
	delegate(to, amount)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Redelegate",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"from", from.String(),
		"to", to.String(),
		"amount", formatUint(amount),
	)
}

// Undelegate undelegates xGNS from the existing delegate.
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#undelegate
func Undelegate(cur realm, from std.Address, amount uint64) {
	if !from.IsValid() {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("invalid address %s to undelegate", from.String()),
		))
	}

	if amount < minimumAmount {
		panic(addDetailToError(
			errLessThanMinimum,
			ufmt.Sprintf("minimum amount to undelegate is %d (requested:%d)", minimumAmount, amount),
		))
	}

	if amount%minimumAmount != 0 {
		panic(addDetailToError(
			errInvalidAmount,
			ufmt.Sprintf("amount must be multiple of %d", minimumAmount),
		))
	}

	caller := std.PreviousRealm().Address()
	delegatedBalanceOfCaller := xgns.BalanceOf(caller)
	if uint64(delegatedBalanceOfCaller) < amount {
		panic(addDetailToError(
			errNotEnoughBalance,
			ufmt.Sprintf("not enough xGNS delegated(%d) to undelegate(%d)", delegatedBalanceOfCaller, amount),
		))
	}

	en.MintAndDistributeGns(cross)

	if running {
		cleanDelegationStatHistory()
	}

	reward, protocolFeeRewards := rewardState.removeStake(caller, amount, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())
	gns.Transfer(cross, caller, int64(reward))

	for tokenPath, rewardAmount := range protocolFeeRewards {
		if rewardAmount > 0 {
			transferProtocolFee(tokenPath, from, rewardAmount)
		}
	}

	// actual unDelegate
	unDelegate(from, amount)
	// burn equivalent amount of xGNS
	xgns.Burn(cross, caller, int64(amount))

	// lock up
	userLocked := lockedGNS{
		amount: amount,
		unlock: uint64(time.Now().Unix()) + SECONDS_IN_7_DAYS, // after 7 days, call Collect() to receive GNS
	}

	var lockedList []lockedGNS
	if value, exists := addrLockedGns.Get(caller.String()); exists {
		lockedList = value.([]lockedGNS)
	}

	lockedList = append(lockedList, userLocked)
	addrLockedGns.Set(caller.String(), lockedList)
	lockedAmount += amount

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Undelegate",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"from", from.String(),
		"amount", formatUint(amount),
	)
}

// CollectUndelegatedGns collects the amount of the undelegated GNS.
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#collectundelegatedgns
func CollectUndelegatedGns(cur realm) uint64 {
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		// Check if withdrawals are specifically halted, not just if the system is halted
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			panic(err)
		}

		if halt.IsContractHalted(halt.OpTypeGovStaker) {
			panic("gov/staker contract operations are currently disabled")
		}
	}

	en.MintAndDistributeGns(cross)

	caller := std.PreviousRealm().Address()
	value, exists := addrLockedGns.Get(caller.String())
	if !exists {
		return 0
	}

	lockedList := value.([]lockedGNS)
	if len(lockedList) == 0 {
		return 0
	}

	prevAddr, prevPkgPath := getPrev()
	collected := uint64(0)
	currentTime := uint64(time.Now().Unix())

	newLockedList := make([]lockedGNS, 0)
	for _, locked := range lockedList {
		if currentTime >= locked.unlock { // passed 7 days
			// transfer GNS to caller
			gns.Transfer(cross, caller, int64(locked.amount))
			lockedAmount -= locked.amount
			collected += locked.amount
		} else {
			newLockedList = append(newLockedList, locked)
		}
	}

	if len(newLockedList) > 0 {
		addrLockedGns.Set(caller.String(), newLockedList)
	} else {
		_, removed := addrLockedGns.Remove(caller.String())
		if !removed {
			panic("failed to remove locked GNS list")
		}
	}

	if collected > 0 {
		std.Emit(
			"CollectUndelegatedGns",
			"prevAddr", prevAddr,
			"prevRealm", prevPkgPath,
			"from", govStaker.String(),
			"to", caller.String(),
			"collectedAmount", formatUint(collected),
		)
	}

	return collected
}

// CollectReward collects the rewards from the protocol fee contract based on the holdings of xGNS.
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#collectreward
func CollectReward(cur realm) {
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		// Check if withdrawals are specifically halted, not just if the system is halted
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			panic(err)
		}

		if halt.IsContractHalted(halt.OpTypeGovStaker) {
			panic("gov/staker contract operations are currently disabled")
		}
	}

	prevAddr, prevPkgPath := getPrev()
	caller := std.PreviousRealm().Address()

	en.MintAndDistributeGns(cross)

	reward, protocolFeeRewards := rewardState.claim(caller, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())

	if reward > 0 {
		if currentGNSBalance < reward {
			panic(addDetailToError(
				errNotEnoughBalance,
				ufmt.Sprintf("not enough GNS balance(%d) to collect(%d)", currentGNSBalance, reward),
			))
		}

		currentGNSBalance -= reward
		gns.Transfer(cross, caller, int64(reward))
		std.Emit(
			"CollectEmissionReward",
			"prevAddr", prevAddr,
			"prevRealm", prevPkgPath,
			"from", govStaker.String(),
			"to", caller.String(),
			"emissionRewardAmount", formatUint(reward),
		)
	}

	for tokenPath, amount := range protocolFeeRewards {
		if amount > 0 {
			transferProtocolFee(tokenPath, caller, amount)
			std.Emit(
				"CollectProtocolFeeReward",
				"prevAddr", prevAddr,
				"prevRealm", prevPkgPath,
				"tokenPath", tokenPath,
				"from", govStaker.String(),
				"to", caller.String(),
				"collectedAmount", formatUint(amount),
			)
		}
	}
}

// CollectRewardFromLaunchPad collects the rewards from the protocol fee contract based on the holdings of xGNS in the launchpad contract.
// Only launchpad contract can call this function
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#collectrewardfromlaunchpad
func CollectRewardFromLaunchPad(cur realm, to std.Address) {
	assertCallerIsLaunchpad()

	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		// Check if withdrawals are specifically halted, not just if the system is halted
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			panic(err)
		}

		if halt.IsContractHalted(halt.OpTypeGovStaker) {
			panic("gov/staker contract operations are currently disabled")
		}
	}
	if !isLaunchpadProjectWallet(to) {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("%s is not project wallet from launchpad", to.String()),
		))
	}

	en.MintAndDistributeGns(cross)

	prevAddr, prevPkgPath := getPrev()
	emissionReward, protocolFeeRewards := rewardState.claim(to, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())
	if emissionReward > 0 {
		if currentGNSBalance < emissionReward {
			panic(addDetailToError(
				errNotEnoughBalance,
				ufmt.Sprintf("not enough GNS balance(%d) to collect(%d)", currentGNSBalance, emissionReward),
			))
		}

		currentGNSBalance -= emissionReward
		gns.Transfer(cross, to, int64(emissionReward))
		std.Emit(
			"CollectEmissionFromLaunchPad",
			"prevAddr", prevAddr,
			"prevRealm", prevPkgPath,
			"from", govStaker.String(),
			"to", to.String(),
			"emissionRewardAmount", formatUint(emissionReward),
		)
	}

	for tokenPath, amount := range protocolFeeRewards {
		if amount > 0 {
			transferProtocolFee(tokenPath, to, amount)
			std.Emit(
				"CollectProtocolFeeFromLaunchPad",
				"prevAddr", prevAddr,
				"prevRealm", prevPkgPath,
				"tokenPath", tokenPath,
				"from", govStaker.String(),
				"to", to.String(),
				"collectedAmount", formatUint(amount),
			)
		}
	}
}

func transferProtocolFee(tokenPath string, to std.Address, amount uint64) {
	common.MustRegistered(tokenPath)
	if !to.IsValid() {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("invalid address %s to transfer protocol fee", to.String()),
		))
	}
	if amount <= 0 {
		panic(addDetailToError(
			errInvalidAmount,
			ufmt.Sprintf("invalid amount %d to transfer protocol fee", amount),
		))
	}
	if currentProtocolFeeBalance[tokenPath] < amount {
		panic(addDetailToError(
			errNotEnoughBalance,
			ufmt.Sprintf("not enough %s balance(%d) to collect(%d)", tokenPath, currentProtocolFeeBalance[tokenPath], amount),
		))
	}

	currentProtocolFeeBalance[tokenPath] -= amount
	if tokenPath == consts.WUGNOT_PATH {
		wugnot.Withdraw(cross, int64(amount))
		banker := std.NewBanker(std.BankerTypeRealmSend)
		banker.SendCoins(govStaker, to, std.Coins{{"ugnot", int64(amount)}})
	} else {
		checkTransferError(common.Transfer(cross, tokenPath, to, int64(amount)))
	}
}
