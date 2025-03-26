package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moderator = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid      boards2.BoardID
	rid, tid boards2.PostID
)

func init() {
	testing.SetOriginCaller(owner)

	// Created a board with flagging threshold greater than 1
	bid = boards2.CreateBoard("test-board")
	boards2.SetFlaggingThreshold(bid, 2)

	// Create a reply so the realm owner can flag and hide it with a single flag
	tid = boards2.CreateThread(bid, "Foo", "bar")
	rid = boards2.CreateReply(bid, tid, 0, "body")
	// Also freeze board to make sure that realm owner can still flag the reply
	boards2.FreezeBoard(bid)
}

func main() {
	boards2.FlagReply(bid, tid, rid, "")

	// Render content must contain a message about the hidden reply
	content := boards2.Render("test-board/1/2")
	println(strings.Contains(content, "> _⚠ Reply is hidden as it has been flagged as inappropriate_"))
}

// Output:
// true
