package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	users "gno.land/r/gnoland/users/v1"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
)

// Test updating a name via GovDAO
var c std.Address = std.OriginCaller()

func init() {
	daov3init.InitWithUsers(c)

	alice := testutils.TestAddress("alice")

	// Register alice
	testing.SetOriginCaller(alice)
	testing.SetRealm(std.NewUserRealm(alice))
	users.Register(cross, "alice123")

	// Prop to change name
	testing.SetOriginCaller(c)
	testing.SetRealm(std.NewUserRealm(c))
	pr := users.ProposeNewName(alice, "alice_new123")
	dao.MustCreateProposal(cross, pr)
}

func main() {
	// testing.SetOriginCaller(c)

	// println("--")
	// println(dao.Render(""))
	// println("--")
	// println(dao.Render("0"))
	// println("--")

	// dao.MustVoteOnProposal(cross, dao.VoteRequest{
	// 	Option:     dao.YesVote,
	// 	ProposalID: dao.ProposalID(0),
	// })

	// println("--")
	// println(dao.Render("0"))
	// println("--")

	// dao.ExecuteProposal(cross, dao.ProposalID(0))

	// println("--")
	// println(dao.Render("0"))

	// data, _ := susers.ResolveName("alice_new123")
	// println(data.Addr())
}

