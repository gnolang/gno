// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_04_filetest

// Render thread repost of a deleted thread.
package z_ui_thread_04_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner        address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	srcBoardName         = "test-board"
	dstBoardName         = "test-board-2"
)

var repostThreadID boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross(cur), srcBoardName, false, false)
	threadID := boards2.CreateThread(cross(cur), boardID, "Foo", "Body")

	// Repost thread into a different board
	dstBoardID := boards2.CreateBoard(cross(cur), dstBoardName, false, false)
	repostThreadID = boards2.CreateRepost(cross(cur), boardID, threadID, dstBoardID, "Bar", "Body2")

	// Remove the original thread
	boards2.DeleteThread(cross(cur), boardID, threadID)
}

func main() {
	path := dstBoardName + "/" + repostThreadID.String()
	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › [test\-board\-2](/r/gnoland/boards2/v1:test-board-2)
// ## Bar
//
// **[g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf)** `owner` · 2009-02-13 11:31pm UTC
//
//
// <gno-foreign>
// Body2
// </gno-foreign>
//
//
//
// > ⚠ Source post has been deleted
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board-2/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board-2/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board-2/1/reply) • [Edit](/r/gnoland/boards2/v1:test-board-2/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=3&threadID=1)
