package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// Create a DAO
	std.TestSetOriginCaller(user)
	daoID = commondao.New("Foo").ID()
}

func main() {
	commondao.Get(daoID)
}

// Error:
// caller is not a realm
