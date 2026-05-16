// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_repost_01_filetest

package z_create_repost_01_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	srcBID boards.ID
	dstBID boards.ID
	srcTID boards.ID = 1024
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross2(cur), "src-board", false, false)
	dstBID = boards2.CreateBoard(cross2(cur), "dst-board", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Repost should fail if source thread doesn't exist
	boards2.CreateRepost(cross2(cur), srcBID, srcTID, dstBID, "foo", "bar")
}

// Error:
// thread does not exist with ID: 1024
