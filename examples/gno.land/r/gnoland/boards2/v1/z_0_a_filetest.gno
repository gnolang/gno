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
	bid := boards2.CreateBoard("test123")
	println("ID =", bid)
}

// Output:
// ID = 1
