package protocol_fee

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/halt"
)

// DistributeProtocolFee distributes the protocol fee to devOps and gov/staker.
// only `gov/staker` can execute this function.
// It returns list of token with amount has been sent to gov/staker.
func DistributeProtocolFee(cur realm) map[string]uint64 {
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		// Check if withdrawals are specifically halted, not just if the system is halted
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			panic(err)
		}

		if halt.IsContractHalted(halt.OpTypeProtocolFee) {
			panic("protocol fee contract operations are currently disabled")
		}
	}

	assertOnlyGovStaker()
	sentToDevOpsForEvent := make([]string, 0)
	sentToGovStakerForEvent := make([]string, 0)
	toReturnDistributedToGovStaker := make(map[string]uint64)

	for token, amount := range protocolFeeState.TokenListWithAmount {
		balance := uint64(common.BalanceOf(token, protocolFeeAddr))

		// anyone can just send certain grc20 token to `protocol_fee` contract
		// therefore, we don't need any guard logic to check whether protocol_fee's xxx token balance is equal to `amount`
		// however, amount always should be less than or equal to balance
		if amount > balance {
			panic(addDetailToError(
				errInvalidAmount,
				ufmt.Sprintf("amount: %d should be less than or equal to balance: %d", amount, balance),
			))
		}

		if amount <= 0 {
			continue
		}

		// Distribute only the recorded amount, not the entire balance
		distributeAmount := amount
		if distributeAmount > balance {
			// This should not happen due to the check above, but safeguard anyway
			distributeAmount = balance
		}

		toDevOpsAmount := distributeAmount * protocolFeeState.DevOpsPct / 10000 // default 0%
		toGovStakerAmount := distributeAmount - toDevOpsAmount                  // default 100%

		// Distribute to DevOps
		if err := protocolFeeState.distributeToDevOps(token, toDevOpsAmount); err != nil {
			panic(err)
		}
		if toDevOpsAmount > 0 {
			sentToDevOpsForEvent = append(sentToDevOpsForEvent, makeEventString(token, toDevOpsAmount))
		}

		// Distribute to Gov/Staker
		if err := protocolFeeState.distributeToGovStaker(token, toGovStakerAmount); err != nil {
			panic(err)
		}
		if toGovStakerAmount > 0 {
			sentToGovStakerForEvent = append(sentToGovStakerForEvent, makeEventString(token, toGovStakerAmount))
			toReturnDistributedToGovStaker[token] = toGovStakerAmount
		}
	}

	protocolFeeState.clearTokenListWithAmount()

	prevAddr, prevRealm := getPrev()
	std.Emit(
		"TransferProtocolFee",
		"prevAddr", prevAddr,
		"prevRealm", prevRealm,
		"toDevOps", strings.Join(sentToDevOpsForEvent, ","),
		"toGovStaker", strings.Join(sentToGovStakerForEvent, ","),
	)

	return toReturnDistributedToGovStaker
}

func GetDevOpsPct() uint64 {
	return protocolFeeState.DevOpsPct
}

// SetDevOpsPctByAdmin sets the devOpsPct.
func SetDevOpsPctByAdmin(cur realm, pct uint64) {
	caller := std.PreviousRealm().Address()
	if err := common.AdminOnly(caller); err != nil {
		panic(err)
	}

	protocolFeeState.setDevOpsPct(pct)
}

// SetDevOpsPct sets the devOpsPct.
// Only governance contract can execute this function via proposal
func SetDevOpsPct(cur realm, pct uint64) {
	caller := std.PreviousRealm().Address()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err)
	}

	protocolFeeState.setDevOpsPct(pct)
}

// AddToProtocolFee adds the amount to the tokenListWithAmount
// Only `pool + router + staker` can execute this function.
func AddToProtocolFee(cur realm, tokenPath string, amount uint64) {
	assertOnlyPoolRouterStaker()
	protocolFeeState.TokenListWithAmount[tokenPath] += amount
}

// ClearTokenListWithAmount clears the tokenListWithAmount.
// only `gov/staker` can execute this function.
func ClearTokenListWithAmount(cur realm) {
	assertOnlyGovStaker()
	protocolFeeState.clearTokenListWithAmount()
}

// ClearAccuTransferToGovStaker clears the accuToGovStaker.
// Only gov/staker can execute this function.
// protocol_fee just sends the accumulated amount to gov/staker
// which means gov/staker doesn't known how much protocol_fee has been distributed from last time
// Therefore, gov/staker calls GetAccuTransferToGovStaker to get the accumulated amount and clear it
func ClearAccuTransferToGovStaker(cur realm) {
	assertOnlyNotHalted()
	caller := std.PreviousRealm().Address()
	if err := common.GovStakerOnly(caller); err != nil {
		panic(err)
	}

	protocolFeeState.AccuToGovStaker = avl.NewTree()
}

func makeEventString(tokenPath string, amount uint64) string {
	return tokenPath + "*FEE*" + strconv.FormatUint(amount, 10)
}
