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
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board", false)
	pid = cross(boards2.CreateThread)(bid, "Foo", "bar", false)

	// Invite a member using a role with permission to flag threads
	cross(boards2.InviteMember)(bid, moderator, boards2.RoleModerator)
}

func main() {
	testing.SetRealm(std.NewUserRealm(moderator))

	cross(boards2.FlagThread)(bid, pid, "")

	// Ensure that original thread content not visible
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread with ID: 1 has been flagged as inappropriate
