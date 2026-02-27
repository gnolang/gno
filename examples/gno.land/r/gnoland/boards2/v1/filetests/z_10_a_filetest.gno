package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	moderator address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid boards.ID
	pid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)

	// Invite a moderator to the new board
	boards2.InviteMember(cross, bid, moderator, boards2.RoleModerator)

	// Create a new thread as a moderator
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Flag thread as owner
	boards2.FlagThread(cross, bid, pid, "Reason")

	// Ensure thread is not listed
	println(boards2.Render("test-board"))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › test-board
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #1
// ↳ [Create Thread](/r/gnoland/boards2/v1:test-board/create-thread) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=1) • [Manage Board](?menu=manageBoard)
//
// ---
// ### This board doesn't have any threads
// Do you want to [start a new conversation](/r/gnoland/boards2/v1:test-board/create-thread) in this board?
