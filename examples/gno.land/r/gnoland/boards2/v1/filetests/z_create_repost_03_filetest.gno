// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_repost_03_filetest

package z_create_repost_03_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	"gno.land/p/nt/ufmt/v0"
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
	srcBID = boards2.CreateBoard(cross2(cur), "src-board", false, false)
	dstBID = boards2.CreateBoard(cross2(cur), "dst-board", false, false)

	srcTID = boards2.CreateThread(cross2(cur), srcBID, "original title", "original text")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Success case
	tID := boards2.CreateRepost(cross2(cur), srcBID, srcTID, dstBID, "repost title", "repost text")
	p := ufmt.Sprintf("dst-board/%s", tID)
	out := boards2.Render(p)

	println(strings.Contains(out, "original text"))
	println(strings.Contains(out, "repost title"))
	println(strings.Contains(out, "repost text"))
}

// Output:
// true
// true
// true
