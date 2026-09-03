// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_board_01_filetest

// Render board sorting threads from newest to oldest.
package z_ui_board_01_filetest

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

	// Create a board and then add 3 threads
	boardID := boards2.CreateBoard(cross(cur), boardName, false, false)

	boards2.CreateThread(cross(cur), boardID, "A", "Body")
	boards2.CreateThread(cross(cur), boardID, "B", "Body")
	boards2.CreateThread(cross(cur), boardID, "C", "Body")
}

func main() {
	path := boardName + "?order=desc"
	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// Created by [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) on 2009-02-13 11:31pm UTC, #2
// ↳ [Create Thread](/r/gnoland/boards2/v1:TestBoard/create-thread) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=2) • [Manage Board](?menu=manageBoard)
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:TestBoard?order=asc)
//
// ###### [C](/r/gnoland/boards2/v1:TestBoard/3)
// Created by [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
//
// ###### [B](/r/gnoland/boards2/v1:TestBoard/2)
// Created by [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
//
// ###### [A](/r/gnoland/boards2/v1:TestBoard/1)
// Created by [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
