package main

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/users"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/init"
)

var (
	alice = testutils.TestAddress("g1alice")
)

func init() {
	testing.SetRealm(std.NewUserRealm(alice))
	c := std.OriginCaller()
	init.InitWithUsers(c)

	pReq := users.ProposeNewRelease("gno.land/r/gnoland/users/v2", "This is a note!")

	cross(dao.MustCreateProposal)(pReq)
}

func main() {
	testing.SetRealm(std.NewUserRealm(alice))

	cross(dao.MustVoteOnProposal)(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})

	cross(dao.ExecuteProposal)(dao.ProposalID(0))
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
