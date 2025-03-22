package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

const bid boards2.BoardID = 404

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	boards2.SetFlaggingThreshold(bid, 1)
}

// Error:
// board does not exist with ID: 404
