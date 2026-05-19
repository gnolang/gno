// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_delete_thread_04_filetest

package z_delete_thread_04_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	member address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid boards.ID
	pid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	pid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")

	// Invite a member using a role with permission to delete threads
	boards2.InviteMember(cross(cur), bid, member, boards2.RoleAdmin)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(member))

	boards2.DeleteThread(cross(cur), bid, pid)

	// Ensure thread doesn't exist
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread not found
