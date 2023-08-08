package main

// SEND: 200000000ugnot

import (
	"gno.land/p/demo/testutils"
	evaldao "gno.land/r/demo/evaluation_dao"
	"gno.land/r/demo/users"
	"std"
)

func main() {
	// test1 -> test2
	member1 := testutils.TestAddress("member1")
	println("member1:", member1)
	member2 := testutils.TestAddress("member2")
	println("member2:", member2)
	member3 := testutils.TestAddress("member3")
	println("member3:", member3)
	payee := testutils.TestAddress("payee")
	println("payee:", payee)

	std.TestSetOrigCaller(member1)

	users.Register("", "gnouser", "my profile")
	evaldao.CreateDAO()

	propID := evaldao.Propose(evaldao.ProposalID(), "this is title", "this is detail", "this is summary", "this is url", payee, "send_proposal")
	println("proposed, ID is: ", propID)

	println(evaldao.RenderGvr(propID))

	// TODO: vote, render proposal
	std.TestSkipHeights(1) //skip current block, to vote

	err := evaldao.CastVote(propID, "for")
	if err != nil {
		println(err.Error())
	}
	println(evaldao.RenderGvr(propID))

	// same account vote again
	err = evaldao.CastVote(propID, "against")
	if err != nil {
		println(err.Error())
	}
	println(evaldao.RenderGvr(propID))

	std.TestSetOrigCaller(member2)
	err = evaldao.CastVote(propID, "against")
	if err != nil {
		println(err.Error())
	}
	println(evaldao.RenderGvr(propID))

	std.TestSetOrigCaller(member3)
	err = evaldao.CastVote(propID, "for")
	if err != nil {
		println(err.Error())
	}

	std.TestSkipHeights(5) // skip voting period
	println(evaldao.RenderGvr(propID))
}

// Output:
