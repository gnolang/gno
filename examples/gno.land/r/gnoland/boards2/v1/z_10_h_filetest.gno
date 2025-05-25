package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	// Created a board with flagging threshold greater than 1
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board", false)
	cross(boards2.SetFlaggingThreshold)(bid, 2)

	// Create a thread so the realm owner can flag and hide it with a single flag
	pid = cross(boards2.CreateThread)(bid, "Foo", "bar", false)
	// Also freeze board to make sure that realm owner can still flag the thread
	cross(boards2.FreezeBoard)(bid)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.FlagThread)(bid, pid, "")

	// Ensure that original thread content not visible
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread with ID: 1 has been flagged as inappropriate
