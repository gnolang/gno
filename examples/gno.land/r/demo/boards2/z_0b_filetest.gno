package main

import (
	"std"

	"gno.land/r/demo/boards2"
)

const admin = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOrigCaller(admin)
}

func main() {
	boards2.CreateBoard("")
}

// Error:
// board name is empty
