// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_reply_16_filetest

// A reply may now contain a <gno-form>. The write-time blacklist used to
// reject it; now the reply body renders inside a <gno-foreign> sandbox
// whose inner renderer does not load forms (gnoweb safe-mode omits the
// raw HTML at render). See BOARD_FOREIGN.md.
package z_create_reply_16_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid, tid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Previously rejected ("forms are not allowed in replies").
	boards2.CreateReply(cross(cur), bid, tid, 0, "<gno-form><gno-select name=\"foo\" value=\"\" /></gno-form>")

	// Accepted; the body is wrapped in a gno-foreign sandbox.
	content := boards2.Render("test-board/1")
	println(strings.Contains(content, "<gno-foreign>"))
}

// Output:
// true
