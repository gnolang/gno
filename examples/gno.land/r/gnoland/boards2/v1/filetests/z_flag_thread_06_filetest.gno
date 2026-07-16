// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_flag_thread_06_filetest

package z_flag_thread_06_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	moderator address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid boards.ID
	pid boards.ID
)

func init(cur realm) {
	// Created a board with a specific flagging threshold
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	boards2.SetFlaggingThreshold(cross(cur), bid, 2)

	// Invite a moderator to the new board
	boards2.InviteMember(cross(cur), bid, moderator, boards2.RoleModerator)

	// Create a new thread and flag it as a moderator
	testing.SetRealm(testing.NewUserRealm(moderator))
	pid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")
	boards2.FlagThread(cross(cur), bid, pid, "Reason")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(moderator))

	boards2.FlagThread(cross(cur), bid, pid, "Reason")
}

// Error:
// post has been already flagged by g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj
