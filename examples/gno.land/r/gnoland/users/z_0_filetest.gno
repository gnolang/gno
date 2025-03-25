package main

import (
	"std"
	"testing"

	"gno.land/r/gnoland/users"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/init"
)

func init() {
	c := std.OriginCaller()
	init.InitWithUsers(c)
	pReq := users.ProposeNewRelease("gno.land/r/gnoland/users/v2", "This is a note!")

	testing.SetOriginPkgAddress(std.Address("g1p84dvfh4wrplyxx4zsmy77a8rxcnjup2j7zv5r")) // gno.land/r/gov/dao
	dao.MustCreateProposal(pReq)
}

func main() {
	testing.SetOriginPkgAddress(std.Address("g1p84dvfh4wrplyxx4zsmy77a8rxcnjup2j7zv5r")) // gno.land/r/gov/dao
	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})

	dao.ExecuteProposal(dao.ProposalID(0))
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
