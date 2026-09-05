// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_board_02_filetest

// Render board view with the manage board menu expanded.
package z_ui_board_02_filetest

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

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross(cur), boardName, false, false)
	boards2.CreateThread(cross(cur), boardID, "A", "Body")
}

func main() {
	path := boardName + "?menu=manageBoard"
	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// Created by [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) on 2009-02-13 11:31pm UTC, #1
// ↳ [Create Thread](/r/gnoland/boards2/v1:TestBoard/create-thread) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=1) • **Manage Board**
// └─ [Invite Member](/r/gnoland/boards2/v1:TestBoard/invite-member) • [List Invite Requests](/r/gnoland/boards2/v1:TestBoard/invites) • [List Members](/r/gnoland/boards2/v1:TestBoard/members) • [List Banned Users](/r/gnoland/boards2/v1:TestBoard/banned-users) • [Freeze Board](/r/gnoland/boards2/v1$help&func=FreezeBoard&boardID=1)
//
// ---
// Sort by: [newest first](/r/gnoland/boards2/v1:TestBoard?menu=manageBoard&order=desc)
//
// ###### [A](/r/gnoland/boards2/v1:TestBoard/1)
// Created by [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
