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
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// Create a new DAO which gives ownership to `commondao_test`
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	dao := commondao.New("Foo")
	daoID = dao.ID()

	// Change options to disallow voting
	options := commondao.GetOptions(dao.ID())
	options.SetAllowVoting(false)
}

func main() {
	cross(commondao.Vote)(daoID, 0, "", "")
}

// Error:
// voting is not enabled
