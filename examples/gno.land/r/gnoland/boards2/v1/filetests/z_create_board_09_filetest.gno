package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross, "TEST123", false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Should fail because board name already exists with a different casing
	boards2.CreateBoard(cross, "test123", false, false)
}

// Error:
// board already exists
