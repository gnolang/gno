package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.DeleteThread(cross, bid, 404)
}

// Error:
// thread does not exist with ID: 404
