// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_board_00_filetest

// Render default board view.
package z_ui_board_00_filetest

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

	// Create a board and then add 3 threads
	boardID := boards2.CreateBoard(cross2(cur), boardName, false, false)

	// Create thread "A" with a single comment
	threadID := boards2.CreateThread(cross2(cur), boardID, "A", "Body")
	boards2.CreateReply(cross2(cur), boardID, threadID, 0, "Body")

	// The other 2 threads are created without comments
	boards2.CreateThread(cross2(cur), boardID, "B", "Body")
	threadID = boards2.CreateThread(cross2(cur), boardID, "C", "Body")

	// Repost thread "C" into a different board
	dstBoardID := boards2.CreateBoard(cross2(cur), "Bar", false, false)
	boards2.CreateRepost(cross2(cur), boardID, threadID, dstBoardID, "Title", "Body")
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) on 2009-02-13 11:31pm UTC, #1
// ↳ [Create Thread](/r/gnoland/boards2/v1:TestBoard/create-thread) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=1) • [Manage Board](?menu=manageBoard)
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:TestBoard?order=desc)
//
// ###### [A](/r/gnoland/boards2/v1:TestBoard/1)
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) `owner` on 2009-02-13 11:31pm UTC
// **1 replies • 0 reposts**
//
// ###### [B](/r/gnoland/boards2/v1:TestBoard/3)
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
//
// ###### [C](/r/gnoland/boards2/v1:TestBoard/4)
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 1 reposts**
