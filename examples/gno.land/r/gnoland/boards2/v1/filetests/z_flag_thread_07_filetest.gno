// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_flag_thread_07_filetest

package z_flag_thread_07_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	bid boards.ID
	pid boards.ID
)

func init(cur realm) {
	// Created a board with flagging threshold greater than 1
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test-board", false, false)
	boards2.SetFlaggingThreshold(cross2(cur), bid, 2)

	// Create a thread so the realm owner can flag and hide it with a single flag
	pid = boards2.CreateThread(cross2(cur), bid, "Foo", "bar")
	// Also freeze board to make sure that realm owner can still flag the thread
	boards2.FreezeBoard(cross2(cur), bid)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FlagThread(cross2(cur), bid, pid, "Reason")

	// Ensure that original thread content not visible
	println(boards2.Render("test-board/1"))
}

// Output:
// ⚠ Thread has been flagged as [inappropriate](/r/gnoland/boards2/v1:test-board/1/flagging-reasons)
