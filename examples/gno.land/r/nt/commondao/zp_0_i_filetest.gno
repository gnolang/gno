package main

import (
	"chain/runtime"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(runtime.NewUserRealm(user))
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New(cross, "Foo")
	dao.Members().Add(user)

	options := commondao.GetOptions(dao.ID())
	options.SetAllowMembersUpdate(true)

	daoID = dao.ID()
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(user))

	commondao.CreateMembersUpdateProposal(cross, daoID, "", owner.String())
}

// Error:
// address is not a DAO member: g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
