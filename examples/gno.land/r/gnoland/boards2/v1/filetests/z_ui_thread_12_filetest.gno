// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_12_filetest

// When the render budget is exhausted mid-page in the threaded view, the
// offset page-picker would offer a "next page" that jumps past the replies
// that were skipped when the budget ran out. Instead the picker is suppressed
// and the reader is funneled to the complete, reachable flat view.
package z_ui_thread_12_filetest

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

	// First top-level reply carries a deep chain that alone exhausts the
	// render budget (maxRenderedBodies, 246) on page 1.
	prev := boards2.CreateReply(cross(cur), bid, tid, 0, "top-deep")
	for i := 0; i < 250; i++ {
		prev = boards2.CreateReply(cross(cur), bid, tid, prev, "chain")
	}
	// Plus enough more top-level replies that the thread spans 2 pages, so a
	// picker WOULD render absent the suppression.
	for i := 0; i < 12; i++ {
		boards2.CreateReply(cross(cur), bid, tid, 0, "top-more")
	}
}

func main(cur realm) {
	out := boards2.Render("test-board/1")
	// Budget truncated page 1: the flat "view all comments" link is shown,
	// and the content-skipping page picker ("page 1 of N") is suppressed.
	ok := strings.Contains(out, "view all comments") &&
		!strings.Contains(out, "page 1 of")
	println(ok)
}

// Output:
// true
