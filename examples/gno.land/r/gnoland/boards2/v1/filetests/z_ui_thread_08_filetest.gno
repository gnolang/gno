// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_08_filetest

// A thread large enough to exceed the foreign-block render budget
// degrades gracefully: rendering stops before gnoweb's
// MaxGnoForeignBlocksPerConvert (100) cap would blank comments, and a
// "More replies — view full thread" notice is shown instead. See
// BOARD_FOREIGN.md (Phase 2) / maxRenderedBodies.
package z_ui_thread_08_filetest

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
	tid = boards2.CreateThread(cross(cur), bid, "Foo", "OP body")

	// One top-level comment with many direct sub-replies (sub-replies
	// are not paginated, so they consume the shared render budget).
	rid := boards2.CreateReply(cross(cur), bid, tid, 0, "top comment")
	for i := 0; i < 95; i++ {
		boards2.CreateReply(cross(cur), bid, tid, rid, "sub reply")
	}
}

func main(cur realm) {
	content := boards2.Render("test-board/1")
	n := strings.Count(content, "<gno-foreign>")
	// Budget bounds the foreign blocks under the 100 cap, and the
	// truncation notice appears instead of silently dropping comments.
	println(n > 0 && n < 100 && strings.Contains(content, "More replies — view full thread"))
}

// Output:
// true
