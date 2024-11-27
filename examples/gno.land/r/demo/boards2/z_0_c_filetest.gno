package main

import (
	"std"

	"gno.land/r/demo/boards2"
)

const (
	admin     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	boardName = "test1"
)

func init() {
	std.TestSetOrigCaller(admin)
	boards2.CreateBoard(boardName)
}

func main() {
	boards2.CreateBoard(boardName)
}

// Error:
// board already exists
