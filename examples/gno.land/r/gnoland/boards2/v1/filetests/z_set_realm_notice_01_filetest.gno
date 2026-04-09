package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init() {
	// Set an initial message so it can be cleared
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.SetRealmNotice(cross, "This is a test realm message")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetRealmNotice(cross, "")

	println(boards2.Notice == "")
}

// Output:
// true
