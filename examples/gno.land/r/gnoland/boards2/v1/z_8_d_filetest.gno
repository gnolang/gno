package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var bid boards2.BoardID

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")
}

func main() {
	boards2.CreateThread(bid, "", "bar")
}

// Error:
// title is empty
