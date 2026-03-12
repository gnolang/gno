package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	bid := boards2.CreateBoard(cross, "test123", false, false)
	println("ID =", bid)
}

// Output:
// ID = 1
