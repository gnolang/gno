package main

import (
	"std"

	boards2 "gno.land/r/demo/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	boards2.CreateBoard("")
}

// Error:
// board name is empty
