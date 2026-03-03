package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	moderator address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid boards.ID
	pid boards.ID
)

func init() {
	// Created a board with a specific flagging threshold
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	boards2.SetFlaggingThreshold(cross, bid, 2)

	// Invite a moderator to the new board
	boards2.InviteMember(cross, bid, moderator, boards2.RoleModerator)

	// Create a new thread and flag it as a moderator
	testing.SetRealm(testing.NewUserRealm(moderator))
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
	boards2.FlagThread(cross, bid, pid, "Reason")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(moderator))

	boards2.FlagThread(cross, bid, pid, "Reason")
}

// Error:
// post has been already flagged by g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj
