package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New(cross, "Foo")

	options := commondao.GetOptions(dao.ID())
	options.SetAllowChildren(true)
	options.SetAllowSubDAOPorposals(true)

	daoID = dao.ID()
}

func main() {
	// Call with a user that is not a member of the DAO
	testing.SetRealm(std.NewUserRealm(user))

	commondao.CreateSubDAOProposal(cross, daoID, "", "")
}

// Error:
// caller is not a DAO member
