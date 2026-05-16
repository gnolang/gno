// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_16_filetest

package z_create_reply_16_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid, tid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross2(cur), bid, "Foo", "bar")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateReply(cross2(cur), bid, tid, 0, "<gno-form><gno-select name=\"foo\" value=\"\" /></gno-form>")
}

// Error:
// Gno-Flavored Markdown forms are not allowed in replies
