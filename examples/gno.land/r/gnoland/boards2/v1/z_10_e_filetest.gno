package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	pid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	cross(boards2.FlagThread)(bid, pid, "")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.FlagThread)(bid, pid, "")
}

// Error:
// item flag count threshold exceeded: 1
