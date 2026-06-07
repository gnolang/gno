// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_09_filetest

// The flat "all comments" view (?flat=1) paginates ThreadMeta.AllReplies so
// every comment is reachable, unlike the recursive threaded view which
// truncates a large thread. Each flat page renders one <gno-foreign> block
// per comment, bounded by pageSizeFlat (50). With 60 comments the flat view
// spans 2 pages; the last comment is reachable via newest-first (desc page 1)
// and via page 2.
package z_ui_thread_09_filetest

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/ufmt/v0"

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
	// pageSizeFlat is 50, so 60 comments span 2 flat pages.
	for i := 0; i < 60; i++ {
		boards2.CreateReply(cross(cur), bid, tid, 0, ufmt.Sprintf("comment-%d", i))
	}
}

func main(cur realm) {
	p1 := boards2.Render("test-board/1?flat=1")              // oldest first, page 1
	desc := boards2.Render("test-board/1?flat=1&order=desc") // newest first, page 1
	p2 := boards2.Render("test-board/1?flat=1&page=2")       // oldest first, page 2

	ok := strings.Contains(p1, "All 60 comments") &&
		strings.Contains(p1, "page 1 of 2") &&
		strings.Contains(p1, "comment-0") && // first comment is on page 1
		!strings.Contains(p1, "comment-59") && // last comment is NOT on oldest-first page 1
		strings.Contains(desc, "comment-59") && // last comment reachable newest-first
		strings.Contains(p2, "comment-59") && // and on the last page
		strings.Contains(p2, "page 2 of 2")
	println(ok)
}

// Output:
// true
