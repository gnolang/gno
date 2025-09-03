package init

import (
	"std"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
	"gno.land/r/sys/users"
)

func Init() {
	// This is needed because state is saved between unit tests,
	// and we want to avoid having real members used on tests
	memberstore.Get().DeleteAll()
	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO:         impl.NewGovDAO(),
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}

func InitWithUsers(addrs ...std.Address) {
	// This is needed because state is saved between unit tests,
	// and we want to avoid having real members used on tests
	memberstore.Get().DeleteAll()
	memberstore.Get().SetTier(memberstore.T1)
	for _, a := range addrs {
		if !a.IsValid() {
			panic("invalid address: " + a.String())
		}
		memberstore.Get().SetMember(memberstore.T1, a, &memberstore.Member{InvitationPoints: 3})
	}

	RegisterTestUsers(addrs...)

	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO:         impl.NewGovDAO(),
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}

const gUsersV1Path = "gno.land/r/gnoland/users/v1"

// Register a namespace for every addresses
// Necessary to test GovDAO Vote
// Set the user namespace to Address.String()
func RegisterTestUsers(addrs ...std.Address) {
	// Set realm to users admin to register test user
	testing.SetRealm(std.NewCodeRealm(gUsersV1Path))
	for _, a := range addrs {
		if !a.IsValid() {
			panic("invalid address: " + a.String())
		}
		users.RegisterUser(cross, a.String(), a)
	}
}
