package impl

import (
	"chain/runtime"

	"gno.land/r/gov/dao/v3/memberstore"
)

var (
	law    *Law
	govDAO *GovDAO = NewGovDAO()
	// SECURITY XXX: gRealm captures the full runtime.Realm struct at
	// init time. Only used at line ~22 for gRealm.PkgPath() — that's
	// safe (PkgPath is a stable string). But holding a runtime.Realm
	// value in a package var is the same anti-pattern as storing `cur`
	// itself. A future change that adds another use of gRealm (e.g.
	// gRealm.Address() being used for auth, or worse, gRealm being
	// passed somewhere as if it were a live capability) would convert
	// a brittle pattern into an actual leak. Fix: replace with two
	// const strings (gRealmPkgPath, gRealmAddress) computed once and
	// hardcoded, or just inline the runtime call at the single use site.
	gRealm         = runtime.CurrentRealm()
)

func init() {
	law = &Law{
		Supermajority: 66.66, // Two thirds
	}
}

func Render(cur realm, in string) string {
	// Same-realm: pass cur to govDAO.Render (also crossing, but same realm
	// so use the literal cur form rather than cross2).
	return govDAO.Render(cur, gRealm.PkgPath(), in)
}

// AddMember allows T1 and T2 members to freely add T3 members using their invitation points.
func AddMember(cur realm, addr address) {
	caller := runtime.PreviousRealm()
	if !caller.IsUser() {
		panic("this function must be called by an EOA through msg call or msg run")
	}
	m, t := memberstore.Get().GetMember(caller.Address())
	if m == nil {
		panic("caller is not a member")
	}

	if t != memberstore.T1 && t != memberstore.T2 {
		panic("caller is not on T1 or T2. To add members, propose them through proposals")
	}

	m.RemoveInvitationPoint()

	if err := memberstore.Get().SetMember(memberstore.T3, addr, memberByTier(memberstore.T3)); err != nil {
		panic(err.Error())
	}
}

func GetInstance() *GovDAO {
	if runtime.CurrentRealm().PkgPath() != "gno.land/r/gov/dao/v3/loader" {
		panic("not allowed")
	}

	return govDAO
}
