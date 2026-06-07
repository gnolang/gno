// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_reply_04_filetest

// Re-rooting at a depth-2 reply renders its parent as context with a
// "Continue this thread →" drill-up link (the parent is itself a reply with
// children), instead of a "View Thread" jump to the thread root.
package z_ui_reply_04_filetest

import (
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid := boards2.CreateBoard(cross(cur), "test-board", false, false)
	tid := boards2.CreateThread(cross(cur), bid, "Foo", "OP body")
	aid := boards2.CreateReply(cross(cur), bid, tid, 0, "comment A") // #2, parent = thread
	boards2.CreateReply(cross(cur), bid, tid, aid, "reply B")        // #3, parent = A
}

func main(cur realm) {
	// Re-root at reply #3 (B); its parent #2 (A) renders as context and, being
	// a reply with children, links "Continue this thread →" (to A's re-root).
	out := boards2.Render("test-board/1/3")
	ok := strings.Contains(out, "Continue this thread →") &&
		strings.Contains(out, "/r/gnoland/boards2/v1:test-board/1/2") // → A's re-root, not the thread root
	println(ok)
}

// Output:
// true
