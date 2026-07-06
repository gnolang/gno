// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_14_filetest

// In a frozen thread, a reply must not show Reply/Edit/Delete action links
// (the backend rejects them as the thread is frozen); only Flag remains —
// matching how the OP and a readonly board already behave.
package z_ui_thread_14_filetest

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
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "OP body")
	rid = boards2.CreateReply(cross(cur), bid, tid, 0, "a reply") // #2
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.FreezeThread(cross(cur), bid, tid)

	out := boards2.Render("test-board/1")
	// Reply #2 keeps Flag but loses Reply/Edit/Delete in a frozen thread.
	ok := strings.Contains(out, "test-board/1/2/flag") &&
		!strings.Contains(out, "test-board/1/2/edit") &&
		!strings.Contains(out, "test-board/1/2/reply")
	println(ok)
}

// Output:
// true
