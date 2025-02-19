package commondao

import (
	"std"
)

func Invite(invitee std.Address) {
	// TODO: Use an authority to restrict inviters when `p/moul/authz` is available
	caller := std.PrevRealm().Addr()
	invites.Set(invitee.String(), caller.String())
}

// IsInvited check if an address has an invitation to the realm.
func IsInvited(addr std.Address) bool {
	return checkIsInvited(addr)
}

func checkIsInvited(addr std.Address) bool {
	_, found := invites.Get(addr.String())
	return found
}
