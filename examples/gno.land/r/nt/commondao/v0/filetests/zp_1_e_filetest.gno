package main

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var dao *pcommondao.CommonDAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross, "Foo")

	options := commondao.GetOptions(dao.ID())
	options.SetAllowTextProposals(true)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateTextProposal(cross, dao.ID(), "", "", 1)
}

// Error:
// caller is not a DAO member
