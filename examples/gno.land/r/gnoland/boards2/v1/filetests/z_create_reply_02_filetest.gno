// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_02_filetest

package z_create_reply_02_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test123", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateReply(cross2(cur), bid, 404, 0, "comment")
}

// Error:
// thread does not exist with ID: 404
