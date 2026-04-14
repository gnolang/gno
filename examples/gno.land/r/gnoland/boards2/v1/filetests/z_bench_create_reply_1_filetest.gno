package main

import (
	"strconv"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boardID := boards2.CreateBoard(cross, "Board"+strconv.Itoa(0), true, true)
	threadID := boards2.CreateThread(cross, boardID, "Thread"+strconv.Itoa(0), "Body")
	boards2.CreateReply(cross, boardID, threadID, 0, "Comment")
}

// Gas:
// 39912185
