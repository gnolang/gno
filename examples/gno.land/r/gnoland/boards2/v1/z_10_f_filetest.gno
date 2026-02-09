package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/permissions"
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
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")

	// Invite a member using a role with permission to flag threads
	boards2.InviteMember(cross, bid, moderator, permissions.RoleModerator)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(moderator))

	boards2.FlagThread(cross, bid, pid, "")

	// Ensure that original thread content not visible
	println(boards2.Render("test-board/1"))
}

// Output:
// ⚠ Thread has been flagged as inappropriate
