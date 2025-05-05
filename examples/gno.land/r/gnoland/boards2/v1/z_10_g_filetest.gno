package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	moderator = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	// Created a board with a specific flagging threshold
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	cross(boards2.SetFlaggingThreshold)(bid, 2)

	// Invite a moderator to the new board
	cross(boards2.InviteMember)(bid, moderator, boards2.RoleModerator)

	// Create a new thread and flag it as a moderator
	testing.SetRealm(std.NewUserRealm(moderator))
	pid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	cross(boards2.FlagThread)(bid, pid, "")
}

func main() {
	testing.SetRealm(std.NewUserRealm(moderator))

	cross(boards2.FlagThread)(bid, pid, "")
}

// Error:
// item has been already flagged by a current user
