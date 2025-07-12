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

var (
	dao    *pcommondao.CommonDAO
	subDAO *pcommondao.CommonDAO
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New("Foo")
	dao.Members().Add(user)

	options := commondao.GetOptions(dao.ID())
	options.SetAllowDissolutionProposals(true)

	// Create a SubDAO to be dissolveed
	subDAO = commondao.NewSubDAO("Bar", dao.ID())
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	pID := commondao.CreateDissolutionProposal(cross, subDAO.ID())

	p := dao.ActiveProposals().Get(pID)
	if p == nil {
		panic("expected proposal to be created in the parent DAO")
	}

	// Make sure proposal doesn't exists in the SubDAO
	println(subDAO.ActiveProposals().Get(pID) == nil)

	println(p.Status() == pcommondao.StatusActive)
	println(p.Creator() == user)
	println(p.Definition().Title() == ("Dissolve DAO: " + dao.Name()))
	println(p.VotingDeadline())
	println(p.Definition().Body())
}

// Output:
// true
// true
// true
// false
// 2009-02-20 23:31:30 +0000 UTC m=+1235172690.000000001
// **DAO:**
// [Bar](/r/nt/commondao:3)
