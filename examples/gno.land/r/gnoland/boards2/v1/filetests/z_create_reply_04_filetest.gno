// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_04_filetest

package z_create_reply_04_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	bid boards.ID
	tid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")

	// Hide thread by flagging it so reply can't be submitted
	boards2.FlagThread(cross(cur), bid, tid, "reason")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateReply(cross(cur), bid, tid, 0, "Test reply")
}

// Error:
// thread is hidden
