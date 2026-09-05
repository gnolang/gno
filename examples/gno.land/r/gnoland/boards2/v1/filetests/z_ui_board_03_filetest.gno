// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_board_03_filetest

// Render readonly board.
package z_ui_board_03_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	boardName         = "TestBoard"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a readonly board and then add a thread
	boardID := boards2.CreateBoard(cross(cur), boardName, false, false)
	boards2.CreateThread(cross(cur), boardID, "A", "Body")
	boards2.FreezeBoard(cross(cur), boardID)
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// > [!WARNING] Info
// > Creating new threads and commenting are disabled within this board
//
// Created by [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) on 2009-02-13 11:31pm UTC, #2
// [List Members](/r/gnoland/boards2/v1:TestBoard/members) • [Unfreeze Board](/r/gnoland/boards2/v1$help&func=UnfreezeBoard&boardID=2&replyID=&threadID=)
//
// ---
// Sort by: [newest first](/r/gnoland/boards2/v1:TestBoard?order=desc)
//
// ###### [A](/r/gnoland/boards2/v1:TestBoard/1)
// Created by [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
