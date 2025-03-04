package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "foo123"
)

func init() {
	std.TestSetOriginCaller(owner)
	boards2.CreateBoard(name)
}

func main() {
	boards2.RenameBoard(name, "")
}

// Error:
// board name is empty
