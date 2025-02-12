package referral

import (
	"std"

	"gno.land/p/gnoswap/consts"
)

// validCallers is a lookup table of addresses that are authorized to modify referral data.
// This includes governance contracts, router, position manager, and staker contracts.
var validCallers = map[std.Address]bool{
	consts.GOV_GOVERNANCE_ADDR: true,
	consts.GOV_STAKER_ADDR:     true,
	consts.ROUTER_ADDR:         true,
	consts.POSITION_ADDR:       true,
	consts.STAKER_ADDR:         true,
	consts.LAUNCHPAD_ADDR:      true,
}

// isValidCaller checks if the given address has permission to modify referral data.
// Only specific pre-authorized addresses defined in validCallers map are allowed to
// register, update, or remove referrals.
func isValidCaller(caller std.Address) error {
	if validCallers[caller] {
		return nil
	}
	return NewError(ErrUnauthorized, caller)
}
