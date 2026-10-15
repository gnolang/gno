// PKGPATH: gno.land/r/gnoland/boards2/v0/filetests/z_create_board_08_filetest

package z_create_board_08_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v0"
)

const (
	owner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	name          = "TestBoard"
)

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross(cur), name, false, false)

	// Unlisted board should not be rendered
	println(boards2.Render(""))

	// Unlisted board can be rendered by path
	println("\n==================")
	println(boards2.Render(name))
}

// Output:
// # Boards
// [Create Board](/r/gnoland/boards2/v0:create-board) • [List Admin Users](/r/gnoland/boards2/v0:admin-users) • [Help](/r/gnoland/boards2/v0:help)
//
// ---
// Sort by: [newest first](/r/gnoland/boards2/v0:?order=desc)
//
// ###### [OpenDiscussions](/r/gnoland/boards2/v0:OpenDiscussions)
// Created by [g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx](/u/g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) on 2009-02-13 11:31pm UTC, #1
// **0 threads**
//
//
//
// ==================
// # [Boards](/r/gnoland/boards2/v0) › TestBoard
// Created by [g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx](/u/g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) on 2009-02-13 11:31pm UTC, #2
// ↳ [Create Thread](/r/gnoland/boards2/v0:TestBoard/create-thread) • [Request Invite](/r/gnoland/boards2/v0$help&func=RequestInvite&boardID=2) • [Manage Board](?menu=manageBoard)
//
// ---
// ### This board doesn't have any threads
// Do you want to [start a new conversation](/r/gnoland/boards2/v0:TestBoard/create-thread) in this board?
