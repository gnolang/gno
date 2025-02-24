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
	// Use a realm caller
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	// Invite a user to be able to start creating DAOs
	std.TestSetOrigCaller(owner)
	commondao.Invite(user)

	// Create a new DAO which gives ownership to `commondao_test`
	std.TestSetOrigCaller(user)
	daoID = commondao.New("Foo").ID()
}

func main() {
	commondao.Execute(daoID, 1)
}

// Error:
// caller is not a DAO member
