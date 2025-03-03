package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	println(commondao.IsInvited(user))
}

// Output:
// false
