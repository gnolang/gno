package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/permissions"
)

const (
	owner  address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	member address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid boards.ID
	pid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")

	// Invite a member using a role with permission to delete threads
	boards2.InviteMember(cross, bid, member, permissions.RoleAdmin)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(member))

	boards2.DeleteThread(cross, bid, pid)

	// Ensure thread doesn't exist
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread not found
