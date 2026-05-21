// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_repost_00_filetest

package z_create_repost_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	srcBID boards.ID
	dstBID boards.ID
	srcTID boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross(cur), "src-board", false, false)
	dstBID = boards2.CreateBoard(cross(cur), "dst-board", false, false)

	srcTID = boards2.CreateThread(cross(cur), srcBID, "Foo", "bar")
	boards2.FlagThread(cross(cur), srcBID, srcTID, "idk")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Repost should fail if source thread is flagged
	boards2.CreateRepost(cross(cur), srcBID, srcTID, dstBID, "foo", "bar")
}

// Error:
// thread is hidden
