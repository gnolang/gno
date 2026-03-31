// Render thread from a readonly board.
// Rendered thread action links should be limited to readonly actions.
package main

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

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a readonly board and then add a thread
	boardID := boards2.CreateBoard(cross, boardName, false, false)
	threadID = boards2.CreateThread(cross, boardID, "Foo", "Body")
	boards2.FreezeBoard(cross, boardID)
}

func main() {
	path := boardName + "/" + threadID.String()
	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › [test\-board](/r/gnoland/boards2/v1:test-board)
// ## Foo
//
// **[g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh)** `owner` · 2009-02-13 11:31pm UTC
// Body
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost)
