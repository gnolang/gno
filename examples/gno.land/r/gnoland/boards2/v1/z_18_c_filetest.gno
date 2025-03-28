package main

import (
	"std"
	"testing"

	"gno.land/p/gnoland/commondao"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	perms boards2.Permissions
	bid   boards2.BoardID
)

func init() {
	// Create a new permissions instance without users
	perms = boards2.NewDefaultPermissions(commondao.New())

	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("foobar")
}

func main() {
	boards2.SetPermissions(bid, perms)

	// Owner that setted new board permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false
