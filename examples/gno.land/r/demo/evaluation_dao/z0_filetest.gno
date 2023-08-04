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
	test1 := testutils.TestAddress("test1")
	println("test1:", test1)
	test2 := testutils.TestAddress("test2")
	println("test2:", test2)
	payee := testutils.TestAddress("payee")
	println("payee:", payee)

	std.TestSetOrigCaller(test1)

	users.Register("", "gnouser", "my profile")
	evaldao.CreateDAO()

	println(evaldao.RenderGvr())

	propID := evaldao.Propose(evaldao.ProposalID(), "this is title", "this is detail", "this is summary", "this is url", payee, "send_proposal")
	println("proposed, ID is: ", propID)

	println(evaldao.RenderProposal(propID))
}

// Output:
// test1: g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
// test2: g1w3jhxapjta047h6lta047h6lta047h6laqcyu4
// payee: g1wpshjet9ta047h6lta047h6lta047h6l0kw6lc
// group created:  1
//
// ***gvr name***: eval_governor
// ***gvr version***: 1
//
// proposed, ID is:  0
//
// * **Proposal state is***: Pending
// * **Proposal snapShot is***: 123
// * **Proposal deadLine is***: 128
// * **Current height is***: 123
