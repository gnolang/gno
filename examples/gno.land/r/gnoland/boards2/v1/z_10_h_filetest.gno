package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	bid boards.ID
	pid boards.ID
)

func init() {
	// Created a board with flagging threshold greater than 1
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	boards2.SetFlaggingThreshold(cross, bid, 2)

	// Create a thread so the realm owner can flag and hide it with a single flag
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
	// Also freeze board to make sure that realm owner can still flag the thread
	boards2.FreezeBoard(cross, bid)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FlagThread(cross, bid, pid, "")

	// Ensure that original thread content not visible
	println(boards2.Render("test-board/1"))
}

// Output:
// ⚠ Thread has been flagged as [inappropriate](/r/gnoland/boards2/v1:test-board/1/flagging-reasons)
