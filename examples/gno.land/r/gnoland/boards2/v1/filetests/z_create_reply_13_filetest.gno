// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_13_filetest

// A reply may now contain Markdown block structure (here a blockquote).
// The write-time blacklist used to reject it; now the reply body renders
// inside a <gno-foreign> sandbox, so its structure is contained and
// cannot hijack realm chrome. See BOARD_FOREIGN.md.
package z_create_reply_13_filetest

import (
	"strings"
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

	// Previously rejected ("headings, blockquotes ... not allowed").
	boards2.CreateReply(cross(cur), bid, tid, 0, "> Markdown blockquote")

	// The reply body is wrapped in a gno-foreign sandbox at render time.
	content := boards2.Render("test-board/1")
	println(strings.Contains(content, "Markdown blockquote") && strings.Contains(content, "<gno-foreign>"))
}

// Output:
// true
