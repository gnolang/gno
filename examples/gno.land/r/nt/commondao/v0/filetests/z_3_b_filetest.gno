package main

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner    = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newRealm = address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var daoID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/tests"))
	daoID = commondao.New(cross, "Foo").ID()
}

func main() {
	// Use a caller that is not the owner of the DAO
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.TransferOwnership(cross, daoID, newRealm)
}

// Error:
// unauthorized
