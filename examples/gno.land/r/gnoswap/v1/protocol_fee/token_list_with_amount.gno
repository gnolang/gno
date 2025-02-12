package protocol_fee

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

var (
	tokenListWithAmount = make(map[string]uint64) // tokenPath -> amount
)

// TokenList returns only the list of token path.
// If positive is true, it returns only the token path with amount > 0.
// If positive is false, it returns all the token path.
func TokenList(positive bool) []string {
	tokens := []string{}

	for tokenPath, amount := range tokenListWithAmount {
		if positive && amount == 0 {
			continue
		}

		tokens = append(tokens, tokenPath)
	}

	return tokens
}

// TokenListWithAmount returns the token path and amount.
func TokenListWithAmount() map[string]uint64 {
	return tokenListWithAmount
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

// TransferProtocolFee transfers the protocol fee to devOps and gov/staker.
// only `gov/staker` can execute this function.
// It returns list of token with amount has been sent to gov/staker.
func TransferProtocolFee() map[string]uint64 {
	assertOnlyGovStaker()

	sentToDevOps := []string{}
	sentToGovStaker := []string{}
	toReturn := map[string]uint64{}

	for token, amount := range tokenListWithAmount {
		balance := common.BalanceOf(token, consts.PROTOCOL_FEE_ADDR)

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
			toDevOps := balance * devOpsPct / 10000 // default 0%
			toGovStaker := balance - toDevOps       // default 100%

			tokenTeller := common.GetTokenTeller(token)
			if toDevOps > 0 {
				tokenTeller.Transfer(consts.DEV_OPS, toDevOps)
				sentToDevOps = append(sentToDevOps, makeEventString(token, toDevOps))
			}

			if toGovStaker > 0 {
				tokenTeller.Transfer(consts.GOV_STAKER_ADDR, toGovStaker)
				sentToGovStaker = append(sentToGovStaker, makeEventString(token, toGovStaker))

				toReturn[token] = toGovStaker
			}
		}
	}

	clearTokenListWithAmount()

	prevAddr, prevRealm := getPrev()
	std.Emit(
		"TransferProtocolFee",
		"prevAddr", prevAddr,
		"prevRealm", prevRealm,
		"toDevOps", strings.Join(sentToDevOps, ","),
		"toGovStaker", strings.Join(sentToGovStaker, ","),
	)

	return toReturn
}

// clearTokenListWithAmount clears the tokenListWithAmount.
func clearTokenListWithAmount() {
	tokenListWithAmount = map[string]uint64{}
}

// assertOnlyPoolRouterStaker panics if the caller is not the pool, router, or staker contract.
func assertOnlyPoolRouterStaker() {
	caller := std.PrevRealm().Addr()

	poolOnlyErr := common.PoolOnly(caller)
	routerOnlyErr := common.RouterOnly(caller)
	stakerOnlyErr := common.StakerOnly(caller)

	if poolOnlyErr != nil && routerOnlyErr != nil && stakerOnlyErr != nil {
		panic(errNoPermission)
	}
}

// assertOnlyGovStaker panics if the caller is not the gov/staker contract.
func assertOnlyGovStaker() {
	caller := std.PrevRealm().Addr()
	if err := common.GovStakerOnly(caller); err != nil {
		panic(err.Error())
	}
}

func makeEventString(tokenPath string, amount uint64) string {
	return tokenPath + "*FEE*" + ufmt.Sprintf("%d", amount)
}
