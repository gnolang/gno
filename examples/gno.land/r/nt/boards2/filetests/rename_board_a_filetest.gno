package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name    = "foo123"
	newName = "bar123"
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.CreateBoard(name)
}

func main() {
	boards2.RenameBoard(name, newName)

	// Check that board is renamed by the default realm owner
	bid, _ := boards2.GetBoardIDFromName(newName)
	println("ID =", bid)
}

// Output:
// ID = 1
