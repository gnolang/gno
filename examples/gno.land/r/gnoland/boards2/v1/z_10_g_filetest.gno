package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	moderator = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	// Created a board with a specific flagging threshold
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	boards2.SetFlaggingThreshold(cross, bid, 2)

	// Invite a moderator to the new board
	boards2.InviteMember(cross, bid, moderator, boards2.RoleModerator)

	// Create a new thread and flag it as a moderator
	testing.SetRealm(runtime.NewUserRealm(moderator))
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
	boards2.FlagThread(cross, bid, pid, "")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(moderator))

	boards2.FlagThread(cross, bid, pid, "")
}

// Error:
// item has been already flagged by g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj
