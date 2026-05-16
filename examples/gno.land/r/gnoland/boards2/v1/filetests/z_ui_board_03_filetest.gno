// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_board_03_filetest

// Render readonly board.
package z_ui_board_03_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	boardName         = "TestBoard"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a readonly board and then add a thread
	boardID := boards2.CreateBoard(cross2(cur), boardName, false, false)
	boards2.CreateThread(cross2(cur), boardID, "A", "Body")
	boards2.FreezeBoard(cross2(cur), boardID)
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// > [!WARNING] Info
// > Creating new threads and commenting are disabled within this board
//
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) on 2009-02-13 11:31pm UTC, #1
// [List Members](/r/gnoland/boards2/v1:TestBoard/members) • [Unfreeze Board](/r/gnoland/boards2/v1$help&func=UnfreezeBoard&boardID=1&replyID=&threadID=)
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:TestBoard?order=desc)
//
// ###### [A](/r/gnoland/boards2/v1:TestBoard/1)
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
