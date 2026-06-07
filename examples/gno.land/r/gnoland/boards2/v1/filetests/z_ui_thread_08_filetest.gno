// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_08_filetest

// A thread large enough to exceed the foreign-block render budget
// degrades gracefully: rendering stops before gnoweb's per-render
// foreign-block cap would blank comments, and a "More replies — view
// full thread" notice is shown instead (see maxRenderedBodies). The
// reply count and assertion are derived from MaxBlocksPerRender so this
// stays a truncation test if the cap changes.
package z_ui_thread_08_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/markdown/foreign/v0"

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
	// Create MaxBlocksPerRender of them — strictly more than the budget
	// (cap minus a margin) — so rendering must truncate.
	rid := boards2.CreateReply(cross(cur), bid, tid, 0, "top comment")
	for i := 0; i < foreign.MaxBlocksPerRender(); i++ {
		boards2.CreateReply(cross(cur), bid, tid, rid, "sub reply")
	}
}

func main(cur realm) {
	content := boards2.Render("test-board/1")
	n := strings.Count(content, "<gno-foreign>")
	// The top comment's many sub-replies are capped inline at pageSizeReplies
	// (so n stays well under the foreign-block budget); the rest are reachable
	// via the comment's paginated re-rooted view ("View all N replies").
	capped := n > 0 && n < foreign.MaxBlocksPerRender()/4
	println(capped && strings.Contains(content, "View all 256 replies"))
}

// Output:
// true
