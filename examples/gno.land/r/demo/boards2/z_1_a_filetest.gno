package main

import (
	"std"

	"gno.land/r/demo/boards2"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	boards2.InviteMember(admin, boards2.RoleAdmin)
	println("ok")
}

// Output:
// ok
