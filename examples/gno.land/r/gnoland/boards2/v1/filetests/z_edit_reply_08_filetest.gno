// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_edit_reply_08_filetest

// Editing a reply to contain a <gno-form> is now accepted: the
// write-time blacklist relaxation applies to edits too (EditReply shares
// the validator). The body renders inside a <gno-foreign> sandbox.
package z_edit_reply_08_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid, tid, rid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")
	rid = boards2.CreateReply(cross(cur), bid, tid, 0, "body")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Previously rejected on edit ("forms are not allowed in replies").
	boards2.EditReply(cross(cur), bid, tid, rid, "<gno-form><gno-select name=\"foo\" value=\"\" /></gno-form>")

	content := boards2.Render("test-board/1")
	println(strings.Contains(content, "<gno-foreign>"))
}

// Output:
// true
