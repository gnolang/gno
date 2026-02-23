package impl

import (
	"chain/runtime"

	"gno.land/r/gov/dao/v3/memberstore"
)

var (
	law    *Law
	govDAO *GovDAO = NewGovDAO()
	gRealm         = runtime.CurrentRealm()
)

func init() {
	law = &Law{
		Supermajority: 66.66, // Two thirds
	}
}

func Render(in string) string {
	return govDAO.Render(gRealm.PkgPath(), in)
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
