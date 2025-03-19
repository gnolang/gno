package protocol_fee

import (
	"std"

	phalt "gno.land/p/gnoswap/halt"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/halt"
)

func getPrev() (string, string) {
	prev := std.PreviousRealm()
	return prev.Address().String(), prev.PkgPath()
}

// assertOnlyNotHalted panics if the contract is halted.
func assertOnlyNotHalted() {
	// Protocol fee halting is handled in the router.
	// So, we only need to check if withdrawals are allowed here.
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		if err := halt.IsHalted(phalt.OpTypeWithdraw); err != nil {
			panic(err)
		}
	}
}

// assertOnlyPoolRouterStaker panics if the caller is not the pool, router, or staker contract.
func assertOnlyPoolRouterStaker() {
	caller := std.PreviousRealm().Address()

	poolOnlyErr := common.PoolOnly(caller)
	routerOnlyErr := common.RouterOnly(caller)
	stakerOnlyErr := common.StakerOnly(caller)

	if poolOnlyErr != nil && routerOnlyErr != nil && stakerOnlyErr != nil {
		panic(errNoPermission)
	}
}

// assertOnlyGovStaker panics if the caller is not the gov/staker contract.
func assertOnlyGovStaker() {
	caller := std.PreviousRealm().Address()
	if err := common.GovStakerOnly(caller); err != nil {
		panic(err.Error())
	}
}
