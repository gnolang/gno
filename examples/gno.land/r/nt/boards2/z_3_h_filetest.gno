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
	boards2.RenameBoard(name, "short")
}

// Error:
// the minimum allowed board name length is 6 characters
