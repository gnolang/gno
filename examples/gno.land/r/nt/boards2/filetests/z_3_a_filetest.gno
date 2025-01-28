package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "foo123"
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.CreateBoard(name)
}

func main() {
	newName := "bar123"
	_, exists := boards2.GetBoardIDFromName(newName)
	println("Exists =", exists)

	boards2.RenameBoard(name, newName)

	bid, _ := boards2.GetBoardIDFromName(newName)
	println("ID =", bid)
}

// Output:
// Exists = false
// ID = 1
