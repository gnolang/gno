package protocol_fee

import (
	"std"

	"gno.land/r/gnoswap/v1/common"
)

func getPrev() (string, string) {
	prev := std.PrevRealm()
	return prev.Addr().String(), prev.PkgPath()
}

// assertOnlyNotHalted panics if the contract is halted.
func assertOnlyNotHalted() {
	common.IsHalted()
}
