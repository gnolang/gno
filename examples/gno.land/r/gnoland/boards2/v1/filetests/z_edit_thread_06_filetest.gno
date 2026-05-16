// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_edit_thread_06_filetest

package z_edit_thread_06_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	admin address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	title         = "Test Thread"
	body          = "Test body"
	path          = "test-board/1"
)

var (
	bid boards.ID
	pid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test-board", false, false)
	pid = boards2.CreateThread(cross2(cur), bid, "Foo", "bar")

	// Invite a member using a role with permission to edit threads
	boards2.InviteMember(cross2(cur), bid, admin, boards2.RoleAdmin)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(admin))

	boards2.EditThread(cross2(cur), bid, pid, title, body)

	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › [test\-board](/r/gnoland/boards2/v1:test-board)
// ## Test Thread
//
// **[g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh)** `owner` · 2009-02-13 11:31pm UTC
// Test body
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board/1/reply) • [Edit](/r/gnoland/boards2/v1:test-board/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=1&threadID=1)
