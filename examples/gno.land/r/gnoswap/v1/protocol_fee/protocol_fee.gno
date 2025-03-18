package protocol_fee

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	phalt "gno.land/p/gnoswap/halt"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/halt"
)

var (
	// by default, devOps will get 0% of the protocol fee (which means gov/staker will get 100% of the protocol fee)
	devOpsPct uint64 = 0 // 0%

	// accumulated amount distributed to gov/staker by token path
	accuToGovStaker = avl.NewTree()
	accuToDevOps    = avl.NewTree()

	// distributedToDevOpsHistory and distributedToGovStakerHistory are used to keep track of the distribution history
	distributedToGovStakerHistory = avl.NewTree()
	distributedToDevOpsHistory    = avl.NewTree()

	protocolFeeAddr, _ = access.GetAddress(access.ROLE_PROTOCOL_FEE)
	govStakerAddr, _   = access.GetAddress(access.ROLE_GOV_STAKER)
	devOpsAddr, _      = access.GetAddress(access.ROLE_DEVOPS)

	tokenListWithAmount = make(map[string]uint64) // tokenPath -> amount
)

// DistributeProtocolFee distributes the protocol fee to devOps and gov/staker.
// only `gov/staker` can execute this function.
// It returns list of token with amount has been sent to gov/staker.
func DistributeProtocolFee() map[string]uint64 {
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		// Check if withdrawals are specifically halted, not just if the system is halted
		if err := halt.IsHalted(phalt.OpTypeWithdraw); err != nil {
			panic(err)
		}

		if halt.IsContractHalted(phalt.OpTypeProtocolFee) {
			panic("protocol fee contract operations are currently disabled")
		}
	}

	assertOnlyGovStaker()
	sentToDevOpsForEvent := make([]string, 0)
	sentToGovStakerForEvent := make([]string, 0)
	toReturnDistributedToGovStaker := make(map[string]uint64)

	for token, amount := range tokenListWithAmount {
		balance := common.BalanceOf(token, protocolFeeAddr)

		// anyone can just send certain grc20 token to `protocol_fee` contract
		// therefore, we don't need any guard logic to check whether protocol_fee's xxx token balance is equal to `amount`
		// however, amount always should be less than or equal to balance
		if amount > balance {
			panic(addDetailToError(
				errInvalidAmount,
				ufmt.Sprintf("amount: %d should be less than or equal to balance: %d", amount, balance),
			))
		}

		if amount > 0 {
			toDevOpsAmount := balance * devOpsPct / 10000 // default 0%
			toGovStakerAmount := balance - toDevOpsAmount // default 100%

			tokenTeller := common.GetTokenTeller(token)
			if toDevOpsAmount > 0 {
				addAccuToDevOps(token, toDevOpsAmount)
				updateDistributedToDevOpsHistory(token, toDevOpsAmount)
				sentToDevOpsForEvent = append(sentToDevOpsForEvent, makeEventString(token, toDevOpsAmount))
				tokenTeller.Transfer(devOpsAddr, toDevOpsAmount)
			}

			if toGovStakerAmount > 0 {
				addAccuToGovStaker(token, toGovStakerAmount)
				updateDistributedToGovStakerHistory(token, toGovStakerAmount)
				sentToGovStakerForEvent = append(sentToGovStakerForEvent, makeEventString(token, toGovStakerAmount))
				tokenTeller.Transfer(govStakerAddr, toGovStakerAmount)

				toReturnDistributedToGovStaker[token] = toGovStakerAmount
			}
		}
	}

	clearTokenListWithAmount()

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
	return devOpsPct
}

// SetDevOpsPctByAdmin sets the devOpsPct.
func SetDevOpsPctByAdmin(pct uint64) {
	caller := std.PreviousRealm().Address()
	if err := common.AdminOnly(caller); err != nil {
		panic(err)
	}

	setDevOpsPct(pct)
}

// SetDevOpsPct sets the devOpsPct.
// Only governance contract can execute this function via proposal
func SetDevOpsPct(pct uint64) {
	caller := std.PreviousRealm().Address()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err)
	}

	setDevOpsPct(pct)
}

// setDevOpsPct sets the devOpsPct.
func setDevOpsPct(pct uint64) {
	if pct > 10000 {
		panic(addDetailToError(
			errInvalidPct,
			ufmt.Sprintf("pct(%d) should not be bigger than 10000", pct),
		))
	}

	prevDevOpsPct := devOpsPct
	devOpsPct = pct

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"SetDevOpsPct",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"newPct", strconv.FormatUint(pct, 10),
		"prevPct", strconv.FormatUint(prevDevOpsPct, 10),
	)
}

// AddToProtocolFee adds the amount to the tokenListWithAmount
// Only `pool + router + staker` can execute this function.
func AddToProtocolFee(tokenPath string, amount uint64) {
	assertOnlyPoolRouterStaker()
	tokenListWithAmount[tokenPath] += amount
}

// ClearTokenListWithAmount clears the tokenListWithAmount.
// only `gov/staker` can execute this function.
func ClearTokenListWithAmount() {
	assertOnlyGovStaker()
	clearTokenListWithAmount()
}

// ClearAccuTransferToGovStaker clears the accuToGovStaker.
// Only gov/staker can execute this function.
// protocol_fee just sends the accumulated amount to gov/staker
// which means gov/staker doesn't known how much protocol_fee has been distributed from last time
// Therefore, gov/staker calls GetAccuTransferToGovStaker to get the accumulated amount and clear it
func ClearAccuTransferToGovStaker() {
	assertOnlyNotHalted()

	caller := std.PreviousRealm().Address()
	if err := common.GovStakerOnly(caller); err != nil {
		panic(err)
	}

	accuToGovStaker = avl.NewTree()
}

// addAccuToGovStaker adds the amount to the accuToGovStaker by token path.
func addAccuToGovStaker(tokenPath string, amount uint64) {
	before := GetAccuTransferToGovStakerByTokenPath(tokenPath)
	after := before + amount
	accuToGovStaker.Set(tokenPath, after)
}

// addAccuToDevOps adds the amount to the accuToDevOps by token path.
func addAccuToDevOps(tokenPath string, amount uint64) {
	before := GetAccuTransferToDevOpsByTokenPath(tokenPath)
	after := before + amount
	accuToDevOps.Set(tokenPath, after)
}

// updateDistributedToGovStakerHistory updates the distributedToGovStakerHistory.
func updateDistributedToGovStakerHistory(tokenPath string, amount uint64) {
	// get the previous amount
	prevAmountI, exists := distributedToGovStakerHistory.Get(tokenPath)
	prevAmount := uint64(0)
	if exists {
		prevAmount = prevAmountI.(uint64)
	}

	// update the amount
	afterAmount := prevAmount + amount
	distributedToGovStakerHistory.Set(tokenPath, afterAmount)
}

// updateDistributedToDevOpsHistory updates the distributedToDevOpsHistory.
func updateDistributedToDevOpsHistory(tokenPath string, amount uint64) {
	// get the previous amount
	prevAmountI, exists := distributedToDevOpsHistory.Get(tokenPath)
	prevAmount := uint64(0)
	if exists {
		prevAmount = prevAmountI.(uint64)
	}

	// update the amount
	afterAmount := prevAmount + amount
	distributedToDevOpsHistory.Set(tokenPath, afterAmount)
}

// clearTokenListWithAmount clears the tokenListWithAmount.
func clearTokenListWithAmount() {
	tokenListWithAmount = make(map[string]uint64)
}

func makeEventString(tokenPath string, amount uint64) string {
	return tokenPath + "*FEE*" + strconv.FormatUint(amount, 10)
}
