package main

import (
	"std"
	"testing"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetOriginCaller(owner)
	commondao.Invite(user)

	// Create a DAO
	testing.SetOriginCaller(user)
	daoID = commondao.New("Foo").ID()
}

func main() {
	commondao.Get(daoID)
}

// Error:
// caller is not a realm
