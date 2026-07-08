// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_11_filetest

// A comment's inline children follow the page's ?order: newest-first when the
// thread is sorted desc, oldest-first otherwise — not always oldest-first.
package z_ui_thread_11_filetest

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
	cid := boards2.CreateReply(cross(cur), bid, tid, 0, "parent") // #2
	for i := 0; i < 3; i++ {
		boards2.CreateReply(cross(cur), bid, tid, cid, ufmt.Sprintf("child-%d", i)) // #3,#4,#5
	}
}

func main(cur realm) {
	desc := boards2.Render("test-board/1?order=desc")
	asc := boards2.Render("test-board/1") // default = oldest first

	// desc view: the comment's children render newest-first (child-2 before child-0).
	descOK := strings.Index(desc, "child-2") < strings.Index(desc, "child-0")
	// asc view: oldest-first (child-0 before child-2).
	ascOK := strings.Index(asc, "child-0") < strings.Index(asc, "child-2")
	println(descOK && ascOK)
}

// Output:
// true
