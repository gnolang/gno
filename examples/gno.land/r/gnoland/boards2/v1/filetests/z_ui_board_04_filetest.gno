// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_board_04_filetest

// Render default board view when there are no threads.
package z_ui_board_04_filetest

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

	boards2.CreateBoard(cross2(cur), boardName, false, false)
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
// ### This board doesn't have any threads
// Do you want to [start a new conversation](/r/gnoland/boards2/v1:TestBoard/create-thread) in this board?
