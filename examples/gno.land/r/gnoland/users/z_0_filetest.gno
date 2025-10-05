package main

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/gnoland/users"
	"gno.land/r/gov/dao"
	ini "gno.land/r/gov/dao/v3/init"
	susers "gno.land/r/sys/users"
)

var alice = testutils.TestAddress("g1alice")

func init() {
	testing.SetRealm(testing.NewUserRealm(alice))
	c := runtime.OriginCaller()
	ini.InitWithUsers(c)
	registerTestUsers(c)

	pReq := users.ProposeNewRelease("gno.land/r/gnoland/users/v2", "This is a note!")

	dao.MustCreateProposal(cross, pReq)
}

const gUsersV1Path = "gno.land/r/gnoland/users/v1"

// Register a namespace for every addresses
// Necessary to test GovDAO Vote
func registerTestUsers(addrs ...address) {
	// Set realm to users admin to register test user
	testing.SetRealm(testing.NewCodeRealm(gUsersV1Path))
	for _, addr := range addrs {
		err := susers.RegisterUser(cross, addr.String()[1:], addr)
		if err != nil {
			panic(err.Error() + " : " + addr.String())
		}
	}
}

func main() {
	testing.SetRealm(testing.NewUserRealm(alice))

	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})

	dao.ExecuteProposal(cross, dao.ProposalID(0))
	println(users.Render(""))
}

// Output:
// # r/gnoland/users
// See the r/gnoland/users changelog below.
//
// | Version | Link | Notes |
// | --- | --- | --- |
// | v2 | [r/gnoland/users v2 (latest)](/r/gnoland/users/v2) | This is a note! |
// | v1 | [r/gnoland/users v1](/r/gnoland/users/v1) | [Original PR](https://github.com/gnolang/gno/pull/3166) |
//
