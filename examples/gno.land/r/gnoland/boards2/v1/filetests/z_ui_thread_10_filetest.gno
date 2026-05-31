// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_10_filetest

// No post or comment renders more than pageSizeReplies (10) direct children
// inline. A comment with more shows the cap inline plus a "View all N replies"
// link to its re-rooted view, which paginates the children on its own ?page=
// (a distinct path from the thread, so no pager collision).
package z_ui_thread_10_filetest

import (
	"strings"
	"testing"

	"gno.land/p/nt/ufmt/v0"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid := boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid := boards2.CreateThread(cross(cur), bid, "Foo", "OP body")
	cid := boards2.CreateReply(cross(cur), bid, tid, 0, "parent comment") // reply #2
	// 12 direct children under the comment — past the inline cap of 10.
	for i := 0; i < 12; i++ {
		boards2.CreateReply(cross(cur), bid, tid, cid, ufmt.Sprintf("child-%d", i))
	}
}

func main(cur realm) {
	thread := boards2.Render("test-board/1")          // threaded view
	reroot := boards2.Render("test-board/1/2")         // re-root at the comment, page 1
	reroot2 := boards2.Render("test-board/1/2?page=2") // re-root, page 2
	threadDesc := boards2.Render("test-board/1?order=desc")

	ok := strings.Contains(thread, "View all 12 replies") && // breadth cap inline + link
		strings.Contains(reroot, "page 1 of 2") && // re-root paginates the children
		strings.Contains(reroot2, "page 2 of 2") && // and advances to the tail
		// In a desc thread the "View all" link carries the order so the
		// re-root opens newest-first too; in the default (asc) view it doesn't.
		strings.Contains(threadDesc, "/r/gnoland/boards2/v1:test-board/1/2?order=desc") &&
		!strings.Contains(thread, "test-board/1/2?order=desc")
	println(ok)
}

// Output:
// true
