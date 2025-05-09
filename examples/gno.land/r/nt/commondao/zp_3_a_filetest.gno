package main

import (
	"std"
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var dao *pcommondao.CommonDAO

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	dao = commondao.New("Foo")
	dao.Members().Add(user)

	options := commondao.GetOptions(dao.ID())
	options.SetAllowDismissalProposals(true)
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	pID := cross(commondao.CreateDismissalProposal)(dao.ID())

	p := dao.ActiveProposals().Get(pID)
	if p == nil {
		panic("expected proposal to be created")
	}

	println(p.Status() == pcommondao.StatusActive)
	println(p.Creator() == user)
	println(p.Definition().Title() == ("Dismiss DAO: " + dao.Name()))
	println(p.VotingDeadline())
	println(p.Definition().Body())
}

// Output:
// true
// true
// true
// 2009-02-20 23:31:30 +0000 UTC m=+1235172690.000000001
// **DAO:** [Foo](/r/nt/commondao:2)
