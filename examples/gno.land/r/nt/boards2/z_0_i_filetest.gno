package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	boards2.CreateBoard("test123")
}

// Error:
// unauthorized
