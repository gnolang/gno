// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_edit_reply_03_filetest

package z_edit_reply_03_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test-board", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.EditReply(cross2(cur), bid, 404, 0, "body")
}

// Error:
// thread does not exist with ID: 404
