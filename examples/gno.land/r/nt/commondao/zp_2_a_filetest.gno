package main

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner   = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user    = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name    = "Foo SubDAO"
	members = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq\ng1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
)

var dao *pcommondao.CommonDAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross, "Foo")
	dao.Members().Add(user)

	options := commondao.GetOptions(dao.ID())
	options.SetAllowChildren(true)
	options.SetAllowSubDAOPorposals(true)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateSubDAOProposal(cross, dao.ID(), name, members)

	p := dao.ActiveProposals().Get(pID)
	if p == nil {
		panic("expected proposal to be created")
	}

	println(string(p.Status()))
	println(p.Creator() == user)
	println(p.Definition().Title() == ("New SubDAO: " + name))
	println(p.VotingDeadline())
	println(p.Definition().Body())
}

// Output:
// active
// true
// true
// 2009-02-20 23:31:30 +0000 UTC
// **Parent DAO:**
// [Foo](/r/nt/commondao:2)
//
// **SubDAO Name:**
// Foo SubDAO
//
// **Members:**
// - g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
// - g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
