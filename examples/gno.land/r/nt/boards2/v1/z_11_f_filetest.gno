package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, "Foo", "bar")

	std.TestSetOriginCaller(user)
}

func main() {
	boards2.EditThread(bid, pid, "Foo", "bar")
}

// Error:
// unauthorized
