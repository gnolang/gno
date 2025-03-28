package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moderator = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	testing.SetOriginCaller(owner)

	// Created a board with a specific flagging threshold
	bid = boards2.CreateBoard("test-board")
	boards2.SetFlaggingThreshold(bid, 2)

	// Invite a moderator to the new board
	boards2.InviteMember(bid, moderator, boards2.RoleModerator)

	// Create a new thread and flag it as a moderator
	testing.SetOriginCaller(moderator)
	pid = boards2.CreateThread(bid, "Foo", "bar")
	boards2.FlagThread(bid, pid, "")
}

func main() {
	boards2.FlagThread(bid, pid, "")
}

// Error:
// item has been already flagged by a current user
