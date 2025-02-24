package main

import (
	"std"

	"gno.land/p/nt/commondao"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	bid   = boards2.BoardID(0)                                      // Operate on realm instead of individual boards
)

var perms boards2.Permissions

func init() {
	// Create a new permissions instance without users
	perms = boards2.NewDefaultPermissions(commondao.New())

	std.TestSetOriginCaller(owner)
}

func main() {
	boards2.SetPermissions(bid, perms)

	// Owner that setted new permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false
