// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_delete_reply_09_filetest

// Hard-deleting a childless DIRECT thread reply removes it from the thread's
// direct-children list (not just the flat AllReplies index), so the threaded
// view doesn't render a ghost comment and the reply count stays accurate.
package z_delete_reply_09_filetest

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
	rb  boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "OP body")
	boards2.CreateReply(cross(cur), bid, tid, 0, "reply-A")      // #2, direct, leaf
	rb = boards2.CreateReply(cross(cur), bid, tid, 0, "reply-B") // #3, direct, leaf
	boards2.CreateReply(cross(cur), bid, tid, 0, "reply-C")      // #4, direct, leaf
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.DeleteReply(cross(cur), bid, tid, rb) // hard delete the leaf

	out := boards2.Render("test-board/1")
	// A and C remain; B must be fully gone (no ghost) from the threaded view.
	ok := strings.Contains(out, "reply-A") &&
		strings.Contains(out, "reply-C") &&
		!strings.Contains(out, "reply-B")
	println(ok)
}

// Output:
// true
