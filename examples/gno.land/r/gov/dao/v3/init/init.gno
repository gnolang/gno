package init

import (
	"std"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

func Init() {
	dao.UpdateImpl(dao.UpdateRequest{
		DAO:         impl.NewGovDAO(),
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}

func InitWithUsers(addrs ...std.Address) {
	memberstore.Get().SetTier(memberstore.T1)
	for _, a := range addrs {
		memberstore.Get().SetMember(memberstore.T1, a, &memberstore.Member{InvitationPoints: 3})
	}

	dao.UpdateImpl(dao.UpdateRequest{
		DAO:         impl.NewGovDAO(),
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}
