package main

import (
	"std"

	pcommondao "gno.land/p/gnoland/commondao"

	"gno.land/r/gnoland/commondao"
)

const (
	owner         = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user          = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newMembers    = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq,g147ah9520z0r6jh9mjr6c75rv6l8aypzvcd3f7d"
	removeMembers = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
)

var dao *pcommondao.CommonDAO

func init() {
	std.TestSetOriginCaller(owner)
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
	commondao.Invite(user)

	std.TestSetOriginCaller(user)
	dao = commondao.New("Foo")
	dao.Members().Add(user)
	dao.Members().Add(owner)

	std.TestSetRealm(std.NewUserRealm(user))
}

func main() {
	pID := commondao.CreateMembersUpdateProposal(dao.ID(), "", newMembers, removeMembers)

	p, found := dao.GetActiveProposal(pID)
	if !found {
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
