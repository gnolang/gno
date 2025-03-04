package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	boards2.CreateBoard("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}

// Error:
// addresses are not allowed as board name
