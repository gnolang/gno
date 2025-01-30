package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var bid boards2.BoardID

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test123")

	std.TestSetOrigCaller(user)
}

func main() {
	boards2.CreateThread(bid, "Foo", "bar")
}

// Error:
// unauthorized
