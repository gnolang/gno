// Package governance/members manages DAO member storage for Volos governance.
//
// This file is intended to be used by the staker contract to manage the member set.
// Since the staker is the one that mints and burns the xVLS tokens, and every xVLS token holder is a member,
// staker is in charge of managing the member set.
//
// The gov realm also has the power to kick members which drains the xVLS from the member.
package governance

import (
	"std"

	"gno.land/r/volos/gov/xvls"
)

// AddMember adds an address to the DAO's member set if they hold any xVLS. Panics if not.
func AddMember(cur realm, addr std.Address) {
	authorizer.DoByPrevious("add_member", func() error {
		if xvls.BalanceOf(addr) <= 0 {
			panic(ErrNoXVLS)
		}
		volosGovernance.Members().Add(addr)
		return nil
	})
}

// RemoveMember removes an address from the DAO's member set if they have 0 xVLS. Panics if not.
func RemoveMember(cur realm, addr std.Address) {
	authorizer.DoByPrevious("remove_member", func() error {
		if xvls.BalanceOf(addr) != 0 {
			panic(ErrStillHasXVLS)
		}
		volosGovernance.Members().Remove(addr)
		return nil
	})
}

// KickMember removes an address from the DAO's member set.
// This is used to kick members that are voted through a proposal to be removed.
// Members are left with their xVLS balance which they can withdraw or keep in hopes of being re-added.
func KickMember(cur realm, addr std.Address) {
	authorizer.DoByCurrent("kick_member", func() error {
		volosGovernance.Members().Remove(addr)
		return nil
	})
}
