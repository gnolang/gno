package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	name          = "foo123"
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross, name, false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.RenameBoard(cross, name, "barbaz")
}

// Error:
// unauthorized
