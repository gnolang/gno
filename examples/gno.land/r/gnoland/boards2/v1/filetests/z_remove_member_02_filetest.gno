package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RemoveMember(cross, 0, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // Operate on realm DAO instead of individual boards
}

// Error:
// member not found
