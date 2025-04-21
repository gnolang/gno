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
	testing.SetOriginCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user
	testing.SetOriginCaller(user)

	// Create the root DAO
	parentID = commondao.New("Parent DAO").ID()

	// Change the caller to a user that is not an owner of the parent DAO
	testing.SetOriginCaller(caller)
}

func main() {
	commondao.NewSubDAO("Foo", parentID)
}

// Error:
// unauthorized
