package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func main() {
	// Call realm with a user that has not permission to lock the realm
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.LockRealm(cross, false)
}

// Error:
// unauthorized
