package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2

func main() {
	// Call realm with a user that has not permission to set realm notice
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetRealmNotice(cross, "Foo")
}

// Error:
// unauthorized
