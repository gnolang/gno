package main

import (
	"strconv"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))
	for i := 0; i < 100; i++ {
		boards2.CreateBoard(cross, "Board"+strconv.Itoa(i), true, true)
	}
}

// Gas:
// 471239275
