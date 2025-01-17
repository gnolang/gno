package main

import (
	"std"

	boards2 "gno.land/r/demo/boards2/v1"
)

const (
	owner     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	boardName = "test1"
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.CreateBoard(boardName)
}

func main() {
	boards2.CreateBoard(boardName)
}

// Error:
// board already exists
