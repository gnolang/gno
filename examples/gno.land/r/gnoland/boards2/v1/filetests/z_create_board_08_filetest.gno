// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_board_08_filetest

package z_create_board_08_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
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
// [Create Board](/r/gnoland/boards2/v1:create-board) • [List Admin Users](/r/gnoland/boards2/v1:admin-users) • [Help](/r/gnoland/boards2/v1:help)
//
// ---
// Sort by: [newest first](/r/gnoland/boards2/v1:?order=desc)
//
// ###### [OpenDiscussions](/r/gnoland/boards2/v1:OpenDiscussions)
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) on 2009-02-13 11:31pm UTC, #1
// **0 threads**
//
//
//
// ==================
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) on 2009-02-13 11:31pm UTC, #2
// ↳ [Create Thread](/r/gnoland/boards2/v1:TestBoard/create-thread) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=2) • [Manage Board](?menu=manageBoard)
//
// ---
// ### This board doesn't have any threads
// Do you want to [start a new conversation](/r/gnoland/boards2/v1:TestBoard/create-thread) in this board?
