package main

import (
	"std"
	"strings"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID uint64
	name  = strings.Repeat("A", 61)
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	commondao.Invite(cross, user)

	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New("Foo")
	dao.Members().Add(user)

	options := commondao.GetOptions(dao.ID())
	options.SetAllowChildren(true)
	options.SetAllowSubDAOPorposals(true)

	daoID = dao.ID()
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	commondao.CreateSubDAOProposal(cross, daoID, name, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}

// Error:
// DAO name is too long, max length is 60 characters
