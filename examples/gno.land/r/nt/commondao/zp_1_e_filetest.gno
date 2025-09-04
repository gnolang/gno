package main

import (
	"chain"
	"chain/runtime"
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var dao *pcommondao.CommonDAO

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(runtime.NewUserRealm(user))
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross, "Foo")

	options := commondao.GetOptions(dao.ID())
	options.SetAllowTextProposals(true)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(user))

	commondao.CreateTextProposal(cross, dao.ID(), "", "", 1)
}

// Error:
// caller is not a DAO member
