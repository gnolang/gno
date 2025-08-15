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

	// Create a DAO
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/nt/commondao"))
	daoID = commondao.New("Foo").ID()
}

func main() {
	// Use a caller that is not the owner of the DAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/tests/tests"))

	commondao.Get(daoID)
}

// Error:
// unauthorized
