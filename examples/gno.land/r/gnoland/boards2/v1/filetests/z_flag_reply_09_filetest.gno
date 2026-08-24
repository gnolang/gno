// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_flag_reply_09_filetest

// A flag reason is user-supplied (only trimmed at write). The flagging-reasons
// table must escape it so a reason can't inject markdown (links/images) or HTML
// into the view other moderators see.
package z_flag_reply_09_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	bid      boards.ID
	rid, tid boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "bar")
	rid = boards2.CreateReply(cross(cur), bid, tid, 0, "body")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.FlagReply(cross(cur), bid, tid, rid, "![pwn](https://evil/x.png) [click](https://evil)")

	out := boards2.Render("test-board/1/2/flagging-reasons")
	// Reason text is shown but escaped — no live markdown image/link syntax.
	ok := strings.Contains(out, "pwn") &&
		!strings.Contains(out, "![pwn]") &&
		!strings.Contains(out, "](https://evil)")
	println(ok)
}

// Output:
// true
