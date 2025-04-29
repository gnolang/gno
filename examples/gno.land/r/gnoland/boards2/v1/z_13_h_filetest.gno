package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	moderator = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid      boards2.BoardID
	rid, tid boards2.PostID
)

func init() {
	// Created a board with flagging threshold greater than 1
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	cross(boards2.SetFlaggingThreshold)(bid, 2)

	// Create a reply so the realm owner can flag and hide it with a single flag
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "body")
	// Also freeze board to make sure that realm owner can still flag the reply
	cross(boards2.FreezeBoard)(bid)

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.FlagReply)(bid, tid, rid, "")

	// Render content must contain a message about the hidden reply
	content := boards2.Render("test-board/1/2")
	println(strings.Contains(content, "> _⚠ Reply is hidden as it has been flagged as inappropriate_"))
}

// Output:
// true
