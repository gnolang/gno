package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	boards2.CreateBoard("short")
}

// Error:
// the minimum allowed board name length is 6 characters
