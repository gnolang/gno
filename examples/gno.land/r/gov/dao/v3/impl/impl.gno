package impl

import (
	"gno.land/r/gov/dao/v3/memberstore"
)

var (
	law    *Law
	govDAO *GovDAO = NewGovDAO()
)

func init() {
	law = &Law{
		Supermajority: 66.66, // Two thirds
	}
}

func Render(cur realm, in string) string {
	// Same-realm: pass cur to govDAO.Render (also crossing, but same realm
	// so use the literal cur form rather than cross2).
	return govDAO.Render(cur, cur.PkgPath(), in)
}

// AddMember allows T1 and T2 members to freely add T3 members using their invitation points.
func AddMember(cur realm, addr address) {
	caller := cur.Previous()
	if !caller.IsUser() {
		panic("this function must be called by an EOA through msg call or msg run")
	}
	m, t := memberstore.Get(0, cur).GetMember(caller.Address())
	if m == nil {
		panic("caller is not a member")
	}

	if t != memberstore.T1 && t != memberstore.T2 {
		panic("caller is not on T1 or T2. To add members, propose them through proposals")
	}

	m.RemoveInvitationPoint()

	if err := memberstore.Get(0, cur).SetMember(memberstore.T3, addr, memberByTier(memberstore.T3)); err != nil {
		panic(err.Error())
	}
}

// GetInstance returns the singleton *GovDAO. Only the loader realm may
// call it (used during the bootstrap UpdateImpl handoff). The
// IsCurrent() check rejects stale or stashed realm values; PkgPath()
// after the check is the authentic immediate caller.
func GetInstance(_ int, rlm realm) *GovDAO {
	if !rlm.IsCurrent() {
		panic("GetInstance: rlm is not the caller's live cur (stale capture or sibling frame)")
	}
	if rlm.PkgPath() != "gno.land/r/gov/dao/v3/loader" {
		panic("not allowed")
	}

	return govDAO
}
