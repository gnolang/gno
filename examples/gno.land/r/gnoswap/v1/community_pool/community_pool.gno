package community_pool

import (
	"std"
	"strconv"

	"gno.land/r/gnoswap/v1/common"
)

// TransferTokenByAdmin transfers token to the given address.
func TransferTokenByAdmin(tokenPath string, to std.Address, amount uint64) {
	assertOnlyNotHalted()
	assertOnlyAdmin()

	transferToken(tokenPath, to, amount)
}

// TransferToken transfers token to the given address.
// Only governance contract can execute this function via proposal
func TransferToken(tokenPath string, to std.Address, amount uint64) {
	assertOnlyNotHalted()
	assertOnlyGovernance()

	transferToken(tokenPath, to, amount)
}

// transferToken transfers token to the given address.
func transferToken(tokenPath string, to std.Address, amount uint64) {
	teller := common.GetTokenTeller(tokenPath)
	checkErr(teller.Transfer(to, amount))

	prevAddr, prevPkgPath := getPrevAsString()

	std.Emit(
		"TransferToken",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"tokenPath", tokenPath,
		"to", to.String(),
		"amount", strconv.FormatUint(amount, 10),
	)
}

// checkErr panics if the error is not nil.
func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// assertOnlyNotHalted panics if the contract is halted.
func assertOnlyNotHalted() {
	common.IsHalted()
}

// assertOnlyAdmin panics if the caller is not the admin.
func assertOnlyAdmin() {
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err)
	}
}

// assertOnlyGovernance panics if the caller is not the governance.
func assertOnlyGovernance() {
	caller := getPrevAddr()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err)
	}
}

// getPrevAddr returns the address of the caller.
func getPrevAddr() std.Address {
	return std.PrevRealm().Addr()
}

// getPrevAsString returns the address and realm of the caller as a string.
func getPrevAsString() (string, string) {
	prev := std.PrevRealm()
	return prev.Addr().String(), prev.PkgPath()
}
