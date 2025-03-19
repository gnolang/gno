package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOriginCaller(owner)
	boards2.LockRealm(false)
}

func main() {
	// Should fail because realm is already locked
	boards2.LockRealm(false)
}

// Error:
// realm is locked
