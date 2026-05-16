// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_repost_04_filetest

package z_create_repost_04_filetest

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

	// Create a board with a thread
	origBID := boards2.CreateBoard(cross2(cur), "origin-board", false, false)
	origTID := boards2.CreateThread(cross2(cur), origBID, "title", "text")

	// Create a second board and repost a thread using an empty title
	srcBID = boards2.CreateBoard(cross2(cur), "source-board", false, false)
	srcTID = boards2.CreateRepost(cross2(cur), origBID, origTID, srcBID, "original title", "original text")

	// Create a third board to try reposting the repost
	dstBID = boards2.CreateBoard(cross2(cur), "destination-board", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateRepost(cross2(cur), srcBID, srcTID, dstBID, "repost title", "repost text")
}

// Error:
// reposting a thread that is a repost is not allowed
