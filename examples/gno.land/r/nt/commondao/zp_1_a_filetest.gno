package main

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner      = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user       = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	title      = "General Proposal"
	body       = "Foo Bar"
	votingDays = uint8(1)
)

var dao *pcommondao.CommonDAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross, "Foo")
	dao.Members().Add(user)
	dao.Members().Add(owner)

	options := commondao.GetOptions(dao.ID())
	options.SetAllowTextProposals(true)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateTextProposal(cross, dao.ID(), title, body, votingDays)

	p := dao.ActiveProposals().Get(pID)
	if p == nil {
		panic("expected proposal to be created")
	}

	println(string(p.Status()))
	println(p.Creator() == user)
	println(p.Definition().Title() == title)
	println(p.VotingDeadline())
	println(p.Definition().Body())
}

// Output:
// active
// true
// true
// 2009-02-14 23:31:30 +0000 UTC
// Foo Bar
