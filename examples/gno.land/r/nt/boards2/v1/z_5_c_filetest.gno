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
	boards2.RemoveMember(0, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // Operate on realm DAO instead of individual boards
}

// Error:
// member not found
