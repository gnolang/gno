package main

import (
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner  = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user   = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	caller = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var parentID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(testing.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	parentID = commondao.New(cross, "Parent DAO").ID()
}

func main() {
	// Call with a realm that is not the owner of the parent DAO
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))

	commondao.NewSubDAO(cross, "Foo", parentID)
}

// Error:
// unauthorized
