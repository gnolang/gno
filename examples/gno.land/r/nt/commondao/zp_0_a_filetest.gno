package main

import (
	"std"
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner         = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user          = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newMembers    = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq,g147ah9520z0r6jh9mjr6c75rv6l8aypzvcd3f7d"
	removeMembers = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
)

var dao *pcommondao.CommonDAO

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	dao = commondao.New("Foo")
	dao.Members().Add(user)
	dao.Members().Add(owner)
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	pID := cross(commondao.CreateMembersUpdateProposal)(dao.ID(), "", newMembers, removeMembers)

	p := dao.ActiveProposals().Get(pID)
	if p == nil {
		panic("expected proposal to be created")
	}

	println(p.Status() == pcommondao.StatusActive)
	println(p.Creator() == user)
	println(p.Definition().Title() == "Members Update")
	println("")
	println(p.Definition().Body())
}

// Output:
// true
// true
// true
//
// **Members to Add:**
// - g147ah9520z0r6jh9mjr6c75rv6l8aypzvcd3f7d
// - g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
//
// **Members to Remove:**
// - g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
