package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	testing.SetOriginCaller(owner)
}

func main() {
	boards2.CreateReply(404, 1, 0, "comment")
}

// Error:
// board does not exist with ID: 404
