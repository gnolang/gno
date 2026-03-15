package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Should fail because board name has a space which is not allowed
	boards2.CreateBoard(cross, "test 123", false, false)
}

// Error:
// board name must start with a letter and have letters, numbers, "-" and "_"
