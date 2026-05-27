// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_13_filetest

package z_create_reply_13_filetest

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
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateReply(cross(cur), bid, tid, 0, "> Markdown blockquote")
}

// Error:
// using Markdown headings, blockquotes or horizontal lines is not allowed in replies
