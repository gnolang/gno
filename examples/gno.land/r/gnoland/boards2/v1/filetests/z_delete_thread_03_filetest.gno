// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_delete_thread_03_filetest

package z_delete_thread_03_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid boards.ID
	pid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	pid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")
}

func main(cur realm) {
	// Call using a user that has not permission to delete threads
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.DeleteThread(cross(cur), bid, pid)
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
