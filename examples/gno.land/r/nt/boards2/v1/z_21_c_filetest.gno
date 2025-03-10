package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const user = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func init() {
	// Call as a user that is not a member of the boards realm
	std.TestSetOriginCaller(user)
}

func main() {
	boards2.SetReplyFee(0)
}

// Error:
// only owners are allowed to change reply fee value
