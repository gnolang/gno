package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	boards2.CreateBoard("test123")
}

// Error:
// unauthorized
