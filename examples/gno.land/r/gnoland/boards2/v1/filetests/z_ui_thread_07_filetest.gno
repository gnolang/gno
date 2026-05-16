// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_07_filetest

// Render thread with a title that contains Markdown
package z_ui_thread_07_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	boardName         = "test-board"
)

var threadID boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross2(cur), boardName, false, false)
	threadID = boards2.CreateThread(cross2(cur), boardID, "[Foo](https://foo.com)", "Body")
}

func main() {
	path := boardName + "/" + threadID.String()
	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › [test\-board](/r/gnoland/boards2/v1:test-board)
// ## \[Foo\]\(https://foo\.com\)
//
// **[g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh)** `owner` · 2009-02-13 11:31pm UTC
// Body
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board/1/reply) • [Edit](/r/gnoland/boards2/v1:test-board/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=1&threadID=1)
