package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner  = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	caller = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var parentID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(std.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	parentID = commondao.New("Parent DAO").ID()
}

func main() {
	// Call with a realm that is not the owner of the parent DAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/test"))

	commondao.NewSubDAO("Foo", parentID)
}

// Error:
// unauthorized
