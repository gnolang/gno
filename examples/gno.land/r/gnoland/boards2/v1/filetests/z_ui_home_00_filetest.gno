// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_home_00_filetest

// Render default realm view.
// Default realm view must render the list of listed boards.
package z_ui_home_00_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create board "AAA" with a single thread
	boardID := boards2.CreateBoard(cross2(cur), "AAA", true, false)
	boards2.CreateThread(cross2(cur), boardID, "Foo", "Bar")

	// Create 2 more boards
	boards2.CreateBoard(cross2(cur), "BBB", true, false)
	boards2.CreateBoard(cross2(cur), "CCC", true, false)
	boards2.CreateBoard(cross2(cur), "DDD", false, false) // <-- Unlisted board
}

func main() {
	println(boards2.Render(""))
}

// Output:
// # Boards
// [Create Board](/r/gnoland/boards2/v1:create-board) • [List Admin Users](/r/gnoland/boards2/v1:admin-users) • [Help](/r/gnoland/boards2/v1:help)
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:?order=desc)
//
// ###### [AAA](/r/gnoland/boards2/v1:AAA)
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) on 2009-02-13 11:31pm UTC, #1
// **1 threads**
//
// ###### [BBB](/r/gnoland/boards2/v1:BBB)
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) on 2009-02-13 11:31pm UTC, #2
// **0 threads**
//
// ###### [CCC](/r/gnoland/boards2/v1:CCC)
// Created by [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) on 2009-02-13 11:31pm UTC, #3
// **0 threads**
