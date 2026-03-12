package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", false, false)
}

// Error:
// addresses are not allowed as board name
