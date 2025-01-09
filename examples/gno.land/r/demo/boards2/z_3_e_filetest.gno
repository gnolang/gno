package main

import (
	"std"

	"gno.land/r/demo/boards2"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "foo"
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.CreateBoard(name)
}

func main() {
	boards2.RenameBoard(name, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}

// Error:
// addresses are not allowed as board name
