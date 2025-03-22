package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOriginCaller(owner)

	// Lock the realm without locking realm members
	boards2.LockRealm(false)
}

func main() {
	boards2.LockRealm(true)
	println(boards2.IsRealmLocked())
	println(boards2.AreRealmMembersLocked())
}

// Output:
// true
// true
