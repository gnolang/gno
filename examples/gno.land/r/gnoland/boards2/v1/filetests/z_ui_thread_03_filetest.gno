// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_03_filetest

// Render thread from a readonly board.
// Rendered thread action links should be limited to readonly actions.
package z_ui_thread_03_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	boardName         = "test-board"
)

var threadID boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a readonly board and then add a thread
	boardID := boards2.CreateBoard(cross(cur), boardName, false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Foo", "Body")
	boards2.FreezeBoard(cross(cur), boardID)
}

func main() {
	path := boardName + "/" + threadID.String()
	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › [test\-board](/r/gnoland/boards2/v1:test-board)
// ## Foo
//
// **[g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf)** `owner` · 2009-02-13 11:31pm UTC
//
//
// <gno-foreign>
// Body
// </gno-foreign>
//
//
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost)
