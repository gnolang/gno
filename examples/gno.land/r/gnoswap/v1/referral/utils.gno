package referral

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoswap/v1/access"
)

// validCallers is a lookup table of addresses that are authorized to modify referral data.
// This includes governance contracts, router, position manager, and staker contracts.
var validCallers = map[std.Address]bool{
	getAddress(access.ROLE_GOVERNANCE): true,
	getAddress(access.ROLE_GOV_STAKER): true,
	getAddress(access.ROLE_ROUTER):     true,
	getAddress(access.ROLE_POSITION):   true,
	getAddress(access.ROLE_STAKER):     true,
	getAddress(access.ROLE_LAUNCHPAD):  true,
}

// isValidCaller checks if the given address has permission to modify referral data.
// Only specific pre-authorized addresses defined in validCallers map are allowed to
// register, update, or remove referrals.
func isValidCaller(caller std.Address) error {
	if validCallers[caller] {
		return nil
	}
	return ufmt.Errorf("unauthorized caller: %s", caller)
}

func getAddress(role string) std.Address {
	addr, err := access.GetAddress(role)
	if err != nil {
		panic(err.Error())
	}
	return addr
}
