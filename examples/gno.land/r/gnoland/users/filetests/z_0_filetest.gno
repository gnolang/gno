package main

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnoland/users"
	"gno.land/r/gov/dao"
	ini "gno.land/r/gov/dao/v3/init"
)

var alice = testutils.TestAddress("g1alice")

func init() {
	testing.SetRealm(testing.NewUserRealm(alice))
	c := runtime.OriginCaller()
	ini.InitWithUsers(c)

	pReq := users.ProposeNewRelease("gno.land/r/gnoland/users/v2", "This is a note!")

	dao.MustCreateProposal(cross, pReq)
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
