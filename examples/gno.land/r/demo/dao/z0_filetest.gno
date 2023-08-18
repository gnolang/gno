package main

// SEND: 200000000ugnot

import (
	"gno.land/p/demo/testutils"
	contribdao "gno.land/r/demo/dao/contrib_dao"
	evaldao "gno.land/r/demo/dao/eval_dao"
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
	payee1 := testutils.TestAddress("payee1")
	println("payee1:", payee1)

	payee1Balance := banker.GetCoins(payee1)
	println("payee1 balance before send is: ", payee1Balance)

	payee2 := testutils.TestAddress("payee2")
	println("payee2:", payee2)

	payee3 := testutils.TestAddress("payee3")
	println("payee3:", payee3)

	// start building Dao
	// register user
	std.TestSetOrigCaller(member3)
	users.Register("", "member3", "my profile")
	std.TestSetOrigCaller(member2)
	users.Register("", "member2", "my profile")
	std.TestSetOrigCaller(member1)
	users.Register("", "member1", "my profile")

	println("token of member1 is: ", evaldao.GetToken(member1))

	// create DAO
	evaldao.CreateDAO()
	// add member
	evaldao.AddMember(member1, 1, "this is member1")
	evaldao.AddMember(member2, 1, "this is member2")
	evaldao.AddMember(member3, 1, "this is member3")
	//println(evaldao.RenderGroup())

	println("weight of member1 is: ", evaldao.GetMemberVotingPower(0))
	println("weight of member2 is: ", evaldao.GetMemberVotingPower(1))
	println("weight of member2 is: ", evaldao.GetMemberVotingPower(2))

	// member count
	println("member count is: ", evaldao.GetTotalVotingPower())

	// -----------------------------------------------
	// -----------------------------------------------
	// -----------------------------------------------

	// propose
	amount := std.Coins{{"ugnot", 200000000}}
	propID := evaldao.Propose("this is title", "this is detail", "this is summary", "this is url", payee1, amount, "send_proposal")
	println("proposed, ID is: ", propID)
	println(evaldao.RenderGvr(propID))

	std.TestSkipHeights(1) //skip current block, into voting period
	// vote
	err := evaldao.CastVote(propID, "for")
	if err != nil {
		println(err.Error())
	}
	println(evaldao.RenderGvr(propID))

	// same account vote again, this is forbidden
	err = evaldao.CastVote(propID, "against")
	if err != nil {
		println(err.Error())
	}
	println(evaldao.RenderGvr(propID))

	// switch user and vote
	std.TestSetOrigCaller(member2)
	err = evaldao.CastVote(propID, "against")
	if err != nil {
		println(err.Error())
	}
	println(evaldao.RenderGvr(propID))

	// switch user and vote
	std.TestSetOrigCaller(member3)
	err = evaldao.CastVote(propID, "for")
	if err != nil {
		println(err.Error())
	}

	std.TestSkipHeights(5) // skip voting period
	println(evaldao.RenderGvr(propID))

	// execute
	evaldao.Execute(propID)

	payee1Balance = banker.GetCoins(payee1)
	println("payee1 balance after send is: ", payee1Balance)

	// ---------------------------------------------------------------------
	// ---------------------------------------------------------------------
	// ---------------------------------------------------------------------

	// create contrib DAO
	// register user
	std.TestSetOrigCaller(payee3)
	users.Register("", "payee3", "my profile")
	std.TestSetOrigCaller(payee2)
	users.Register("", "payee2", "my profile")
	std.TestSetOrigCaller(payee1)
	users.Register("", "payee1", "my profile")

	// create DAO
	contribdao.CreateDAO()
	// add member
	contribdao.AddMember(payee1, 1, "this is payee1")
	contribdao.AddMember(payee2, 1, "this is payee2")
	contribdao.AddMember(payee3, 1, "this is payee3")
	//println(evaldao.RenderGroup())

	println("weight of payee1 is: ", contribdao.GetMemberVotingPower(0))
	println("weight of payee2 is: ", contribdao.GetMemberVotingPower(1))
	println("weight of payee3 is: ", contribdao.GetMemberVotingPower(2))

	// member count
	println("member count is: ", contribdao.GetTotalVotingPower())
	println(contribdao.RenderGvr(""))

	// propose all over again

	// calculate tier membership, propose, vote, execute, etc
}

// Output:
