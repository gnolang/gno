package main

// /SEND: 200000000ugnot

import (
	"gno.land/p/demo/testutils"
	"gno.land/r/demo/dao/contrib_dao"
	"gno.land/r/demo/dao/eval_dao"
	"gno.land/r/demo/users"
	"std"
)

func main() {
	// simulate treasury
	treasuryAddr := std.DerivePkgAddr("gno.land/p/demo/gor")
	println("treasury addr: ", treasuryAddr)
	std.TestSetOrigPkgAddr(treasuryAddr)
	std.TestIssueCoins(treasuryAddr, std.Coins{{"ugnot", 1000000000}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 1000000000}}, nil)

	banker := std.GetBanker(std.BankerTypeReadonly)
	treasuryBalance := banker.GetCoins(treasuryAddr)
	println("treasury after issue: ", treasuryBalance)

	// members in evaluation DAO
	member1 := testutils.TestAddress("member1")
	println("member1:", member1)
	member2 := testutils.TestAddress("member2")
	println("member2:", member2)
	member3 := testutils.TestAddress("member3")
	println("member3:", member3)

	// members in contrib DAO
	worker1 := testutils.TestAddress("worker1")
	println("worker1:", worker1)

	worker1Balance := banker.GetCoins(worker1)
	println("worker1 balance before send is: ", worker1Balance)

	worker2 := testutils.TestAddress("worker2")
	println("worker2:", worker2)

	worker3 := testutils.TestAddress("worker3")
	println("worker3:", worker3)

	// start building Dao
	// register user
	std.TestSetOrigCaller(member3)
	users.Register("", "member3", "my profile")
	std.TestSetOrigCaller(member2)
	users.Register("", "member2", "my profile")
	std.TestSetOrigCaller(member1)
	users.Register("", "member1", "my profile")

	// create eval DAO
	eval_dao.CreateDAO("eval_dao")
	// add member
	eval_dao.AddMember(member1, 1, "this is member1")
	eval_dao.AddMember(member2, 1, "this is member2")
	eval_dao.AddMember(member3, 1, "this is member3")
	//println(eval_dao.RenderGroup())

	// create contrib DAO
	// register user first
	std.TestSetOrigCaller(worker3)
	users.Register("", "worker3", "my profile")
	std.TestSetOrigCaller(worker2)
	users.Register("", "worker2", "my profile")
	std.TestSetOrigCaller(worker1)
	users.Register("", "worker1", "my profile")

	contrib_dao.CreateDAO("contrib_dao")
	// add member, initially set to 1
	contrib_dao.AddMember(worker1, 0, "this is worker1")
	contrib_dao.AddMember(worker2, 0, "this is worker2")
	contrib_dao.AddMember(worker3, 0, "this is worker3")

	println("weight of member1 is: ", eval_dao.GetMemberVotingPower(member1))
	println("weight of member2 is: ", eval_dao.GetMemberVotingPower(member2))
	println("weight of member2 is: ", eval_dao.GetMemberVotingPower(member3))

	println("weight of worker1 is: ", contrib_dao.GetMemberVotingPower(worker1))
	println("weight of worker2 is: ", contrib_dao.GetMemberVotingPower(worker2))
	println("weight of worker3 is: ", contrib_dao.GetMemberVotingPower(worker3))
	// member count
	println("eval_dao total weight is: ", eval_dao.GetTotalVotingPower())
	println("contrib_dao total weight is: ", contrib_dao.GetTotalVotingPower())

	// -----------------------------------------------
	// -----------------------------------------------
	// -----------------------------------------------
	println("-----------------------------------------------")

	// propose
	amount := std.Coins{{"ugnot", 200000000}}
	propID := eval_dao.Propose("this is title", "this is detail", "this is summary", "this is url", worker1, amount, "send_proposal")
	println("proposed, ID is: ", propID)
	println(eval_dao.Render(propID))

	//skip current block, into voting period
	std.TestSkipHeights(1)

	// set caller to member1, voter
	std.TestSetOrigCaller(member1)
	// vote
	err := eval_dao.CastVoteWithMemo(propID, "for", "good point")
	if err != nil {
		println(err.Error())
	}
	println(eval_dao.Render(propID))

	// same account vote again, this is forbidden
	err = eval_dao.CastVoteWithMemo(propID, "against", "strongly against")
	if err != nil {
		println(err.Error())
	}
	println(eval_dao.Render(propID))

	// switch user and vote
	std.TestSetOrigCaller(member2)
	err = eval_dao.CastVoteWithMemo(propID, "against", "strongly against")
	if err != nil {
		println(err.Error())
	}
	println(eval_dao.Render(propID))

	// switch user and vote
	std.TestSetOrigCaller(member3)
	err = eval_dao.CastVoteWithMemo(propID, "for", "yep")
	if err != nil {
		println(err.Error())
	}

	std.TestSkipHeights(5) // skip voting period
	println(eval_dao.Render(propID))

	// execute
	eval_dao.Execute(propID)

	worker1Balance = banker.GetCoins(worker1)
	println("worker1 balance after send is: ", worker1Balance)

	println("executing done----------------------------------")
	// ---------------------------------------------------------------------
	// ---------------------------------------------------------------------
	// ---------------------------------------------------------------------

	// member count
	println("total voting power of contrib_dao: ", contrib_dao.GetTotalVotingPower())
	println(contrib_dao.Render(""))

	// propose all over again

	// calculate tier membership, propose, vote, execute, etc
}

// Output:
