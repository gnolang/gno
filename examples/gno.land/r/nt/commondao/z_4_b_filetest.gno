package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
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
