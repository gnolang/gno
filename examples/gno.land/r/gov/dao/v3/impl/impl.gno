package impl

import (
	"std"

	"gno.land/r/gov/dao/v3/memberstore"
)

var govDAO *GovDAO = NewGovDAO()
var law *Law
var realm = std.CurrentRealm()

func init() {
	law = &Law{
		Supermajority: 66,
	}
}

func Render(in string) string {
	return govDAO.Render(realm.PkgPath(), in)
}

// AddMember allows T1 and T2 members to freely add T3 members using their invitation points.
func AddMember(addr std.Address) {
	caller := std.OriginCaller()
	m, t := memberstore.Get().GetMember(caller)
	if m == nil {
		panic("caller is not a member")
	}

	if t != memberstore.T1 || t != memberstore.T2 {
		panic("caller is not on T1 or T2. To add members, propose them through proposals")
	}

	m.RemoveInvitationPoint()

	if err := memberstore.Get().SetMember(memberstore.T3, addr, memberByTier(memberstore.T3)); err != nil {
		panic(err.Error())
	}
}

func GetInstance() *GovDAO {
	if std.CurrentRealm().PkgPath() != "gno.land/r/gov/dao/v3/loader" {
		panic("not allowed")
	}

	return govDAO
}
