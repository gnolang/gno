// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_delete_thread_00_filetest

package z_delete_thread_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards/v0"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	title         = "Test Thread"
	body          = "Test body"
)

var (
	bid boards.ID
	pid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	pid = boards2.CreateThread(cross(cur), bid, title, body)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.DeleteThread(cross(cur), bid, pid)

	// Ensure thread doesn't exist
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread not found
