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
	commondao.Invite(cross, user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(std.NewUserRealm(user))

	// Create root DAO with a subDAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New(cross, "Foo")
	daoID = dao.ID()

	options := commondao.GetOptions(daoID)
	options.SetAllowMembersUpdate(true)
}

func main() {
	testing.SetRealm(std.NewUserRealm("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"))

	commondao.CreateMembersUpdateProposal(cross, daoID, "", "")
}

// Error:
// caller is not a DAO member
