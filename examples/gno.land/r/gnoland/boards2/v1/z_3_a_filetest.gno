package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name    = "foo123"
	newName = "bar123"
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)(name)

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.RenameBoard)(name, newName)

	// Ensure board is renamed by the default board owner
	bid2, _ := cross(boards2.GetBoardIDFromName)(newName)
	println("IDs match =", bid == bid2)
}

// Output:
// IDs match = true
