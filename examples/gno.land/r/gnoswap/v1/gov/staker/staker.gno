package staker

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/r/demo/wugnot"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/gnoswap/v1/gov/xgns"

	en "gno.land/r/gnoswap/v1/emission"
	"gno.land/r/gnoswap/v1/referral"
)

// lockedGNS represents a locked GNS with its amount, unlock time, and collected status
type lockedGNS struct {
	amount    uint64
	unlock    uint64
	collected bool
}

const SECONDS_IN_7_DAYS = uint64(604800)

var (
	addrLockedGns = avl.NewTree() // address -> []lockedGNS
	lockedAmount  = uint64(0)
)

var minimumAmount = uint64(1_000_000)

// Delegate delegates GNS tokens to a specified address.
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#delegate
func Delegate(to std.Address, amount uint64, referrer string) {
	if running {
		cleanDelegationStatHistory()
	}

	if amount == 0 {
		panic(addDetailToError(
			errInvalidAmount,
			"delegation amount cannot be 0",
		))
	}

	CollectReward()

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

	caller := std.PrevRealm().Addr()
	referral.TryRegister(caller, referrer)

	gnsBalance := gns.BalanceOf(caller)
	if gnsBalance < amount {
		panic(addDetailToError(
			errNotEnoughBalance,
			ufmt.Sprintf("invalid GNS balance(%d) to delegate(%d)", gnsBalance, amount),
		))
	}

	rewardState.addStake(uint64(std.GetHeight()), caller, amount, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())

	// GNS // caller -> GovStaker
	gns.TransferFrom(caller, std.CurrentRealm().Addr(), amount)

	// actual delegate
	delegate(to, amount)

	// xGNS mint to caller
	xgns.Mint(caller, amount)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Delegate",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"from", caller.String(),
		"to", to.String(),
		"amount", formatUint(amount),
		"referrer", referrer,
	)
}

// Redelegate redelegates xGNS from existing delegate to another.
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#redelegate
func Redelegate(from, to std.Address, amount uint64) {
	if running {
		cleanDelegationStatHistory()
	}

	CollectReward()

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

	caller := std.PrevRealm().Addr()

	delegated := xgns.BalanceOf(caller)
	if delegated < amount {
		panic(addDetailToError(
			errNotEnoughBalance,
			ufmt.Sprintf("not enough xGNS delegated(%d) to redelegate(%d)", delegated, amount),
		))
	}

	undelegate(from, amount)
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
func Undelegate(from std.Address, amount uint64) {
	if running {
		cleanDelegationStatHistory()
	}

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

	caller := std.PrevRealm().Addr()
	delegated := xgns.BalanceOf(caller)

	if delegated < amount {
		panic(addDetailToError(
			errNotEnoughBalance,
			ufmt.Sprintf("not enough xGNS delegated(%d) to undelegate(%d)", delegated, amount),
		))
	}

	reward, protocolFeeRewards := rewardState.removeStake(caller, amount, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())

	// burn equivalent amount of xGNS
	xgns.Burn(caller, amount)

	gns.Transfer(caller, reward)

	for tokenPath, amount := range protocolFeeRewards {
		transferProtocolFee(tokenPath, from, amount)
	}

	// actual undelegate
	undelegate(from, amount)

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
func CollectUndelegatedGns() uint64 {
	common.IsHalted()
	en.MintAndDistributeGns()

	caller := std.PrevRealm().Addr()

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
			gns.Transfer(caller, locked.amount)
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
			"from", consts.GOV_STAKER_ADDR.String(),
			"to", caller.String(),
			"collectedAmount", formatUint(collected),
		)
	}

	return collected
}

// CollectReward collects the rewards from the protocol fee contract based on the holdings of xGNS.
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#collectreward
func CollectReward() {
	common.IsHalted()
	en.MintAndDistributeGns()

	prevAddr, prevPkgPath := getPrev()
	caller := std.PrevRealm().Addr()

	reward, protocolFeeRewards := rewardState.claim(caller, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())

	// XXX (@notJoon): There could be cases where the reward pool is empty, In such case,
	// it seems appropriate to return 0 and continue processing.
	//
	// This isn't necessarily an abnormal situation, particularly
	// since it could be because rewards haven't occurred yet or
	// have already been fully collected.
	//
	// still, this is a tangled with the policy issue, so should be discussed.

	if reward > 0 {
		currentGNSBalance -= reward
		gns.Transfer(caller, reward)
		std.Emit(
			"CollectEmissionReward",
			"prevAddr", prevAddr,
			"prevRealm", prevPkgPath,
			"from", consts.GOV_STAKER_ADDR.String(),
			"to", caller.String(),
			"emissionRewardAmount", formatUint(reward),
		)
	}

	for tokenPath, amount := range protocolFeeRewards {
		if tokenPath == consts.WUGNOT_PATH {
			if amount > 0 {
				if currentProtocolFeeBalance[tokenPath] < amount {
					panic(addDetailToError(
						errNotEnoughBalance,
						ufmt.Sprintf("not enough %s balance(%d) to collect(%d)", tokenPath, currentProtocolFeeBalance[tokenPath], amount),
					))
				}
				currentProtocolFeeBalance[tokenPath] -= amount
				wugnot.Withdraw(amount)
				banker := std.GetBanker(std.BankerTypeRealmSend)
				banker.SendCoins(consts.GOV_STAKER_ADDR, caller, std.Coins{{"ugnot", int64(amount)}})

			}
		} else {
			transferProtocolFee(tokenPath, caller, amount)
		}

		std.Emit(
			"CollectProtocolFeeReward",
			"prevAddr", prevAddr,
			"prevRealm", prevPkgPath,
			"tokenPath", tokenPath,
			"from", consts.GOV_STAKER_ADDR.String(),
			"to", caller.String(),
			"collectedAmount", formatUint(amount),
		)
	}
}

// CollectRewardFromLaunchPad collects the rewards from the protocol fee contract based on the holdings of xGNS in the launchpad contract.
// Only launchpad contract can call this function
// ref: https://docs.gnoswap.io/contracts/governance/staker.gno#collectrewardfromlaunchpad
func CollectRewardFromLaunchPad(to std.Address) {
	assertCallerIsLaunchpad()

	common.IsHalted()
	en.MintAndDistributeGns()

	prevAddr, prevPkgPath := getPrev()

	emissionReward, protocolFeeRewards := rewardState.claim(to, collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())

	if emissionReward > 0 {
		gns.Transfer(to, emissionReward)
		std.Emit(
			"CollectEmissionFromLaunchPad",
			"prevAddr", prevAddr,
			"prevRealm", prevPkgPath,
			"from", consts.GOV_STAKER_ADDR.String(),
			"to", to.String(),
			"emissionRewardAmount", formatUint(emissionReward),
		)
	}

	for tokenPath, amount := range protocolFeeRewards {
		transferProtocolFee(tokenPath, to, amount)
		std.Emit(
			"CollectProtocolFeeFromLaunchPad",
			"prevAddr", prevAddr,
			"prevRealm", prevPkgPath,
			"tokenPath", tokenPath,
			"from", consts.GOV_STAKER_ADDR.String(),
			"to", to.String(),
			"collectedAmount", formatUint(amount),
		)
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

	currentProtocolFeeBalance[tokenPath] -= amount
	token := common.GetTokenTeller(tokenPath)
	checkTransferError(token.Transfer(to, amount))
}
