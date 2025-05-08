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
	cross(commondao.Invite)(user)

	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	dao := commondao.New("Foo")

	// Disallow SubDAO creation
	options := commondao.GetOptions(dao.ID())
	options.SetAllowSubDAOPorposals(false)

	daoID = dao.ID()
}

func main() {
	cross(commondao.CreateSubDAOProposal)(daoID, "", "")
}

// Error:
// forbidden
