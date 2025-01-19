package main

import (
	"std"

	"gno.land/r/demo/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	bid := boards2.CreateBoard("test1")
	println("ID =", bid)
}

// Output:
// ID = 1
