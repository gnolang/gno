// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_06_filetest

package z_create_reply_06_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	bid      boards.ID
	tid, rid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "thread", "thread")
	rid = boards2.CreateReply(cross(cur), bid, tid, 0, "reply1")

	// Hide reply by flagging it so sub reply can't be submitted
	boards2.FlagReply(cross(cur), bid, tid, rid, "reason")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateReply(cross(cur), bid, tid, rid, "reply1.1")
}

// Error:
// replying to a hidden or frozen reply is not allowed
