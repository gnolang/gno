package protocol_fee

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

var (
	// by default, devOps will get 0% of the protocol fee (which means gov/staker will get 100% of the protocol fee)
	devOpsPct uint64 = 0 // 0%

	// gns amount distributed to devOps and gov/staker
	gnsToDevOps    uint64
	gnsToGovStaker uint64

	// accumulated amount distributed to gov/staker by token path
	accuToGovStaker = avl.NewTree()
)

// DistributeProtocolFee distributes the protocol fee to devOps and gov/staker.
func DistributeProtocolFee() {
	assertOnlyNotHalted()

	tokens := common.ListRegisteredTokens()
	if len(tokens) == 0 {
		return
	}

	for _, token := range tokens {
		// default distribute protocol fee percent
		// govStaker 100%
		// ...

		balance := common.BalanceOf(token, consts.PROTOCOL_FEE_ADDR)
		if balance > 0 {
			toDevOps := balance * devOpsPct / 10000 // default 0%
			toGovStaker := balance - toDevOps       // default 100%

			if token == consts.GNS_PATH {
				gnsToDevOps = toDevOps
				gnsToGovStaker = toGovStaker
			}

			addAccuToGovStaker(token, toGovStaker)

			tokenTeller := common.GetTokenTeller(token)
			if toDevOps > 0 {
				tokenTeller.Transfer(consts.DEV_OPS, toDevOps)
			}

			if toGovStaker > 0 {
				tokenTeller.Transfer(consts.GOV_STAKER_ADDR, toGovStaker)
			}
		}
	}
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

// GetLastTransferToDevOps returns the last transfer to devOps.
func GetLastTransferToDevOps() uint64 {
	return gnsToDevOps
}

// GetAccuTransferToGovStaker returns the accuToGovStaker.
func GetAccuTransferToGovStaker() *avl.Tree {
	return accuToGovStaker
}

// GetAccuTransferToGovStakerByTokenPath returns the accumulated transfer to gov/staker by token path.
func GetAccuTransferToGovStakerByTokenPath(path string) uint64 {
	amountI, exists := accuToGovStaker.Get(path)
	if !exists {
		return 0
	}

	return amountI.(uint64)
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
func addAccuToGovStaker(path string, amount uint64) {
	before := GetAccuTransferToGovStakerByTokenPath(path)
	after := before + amount
	accuToGovStaker.Set(path, after)
}
