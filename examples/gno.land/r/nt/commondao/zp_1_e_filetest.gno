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
	commondao.Invite(cross, user)

	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross, "Foo")

	options := commondao.GetOptions(dao.ID())
	options.SetAllowTextProposals(true)
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	commondao.CreateTextProposal(cross, dao.ID(), "", "", 1)
}

// Error:
// caller is not a DAO member
