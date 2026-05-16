// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_flag_reply_04_filetest

package z_flag_reply_04_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	bid      boards.ID
	rid, tid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross2(cur), bid, "Foo", "bar")
	rid = boards2.CreateReply(cross2(cur), bid, tid, 0, "body")
	boards2.FlagReply(cross2(cur), bid, tid, rid, "Reason")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FlagReply(cross2(cur), bid, tid, rid, "Reason")
}

// Error:
// flagging hidden comments or replies is not allowed
