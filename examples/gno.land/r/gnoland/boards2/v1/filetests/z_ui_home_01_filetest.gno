// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_home_01_filetest

// Render boards sorted from newest to oldest.
package z_ui_home_01_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross(cur), "AAA", true, false)
	boards2.CreateBoard(cross(cur), "BBB", true, false)
	boards2.CreateBoard(cross(cur), "CCC", true, false)
}

func main() {
	println(boards2.Render("?order=desc"))
}

// Output:
// # Boards
// [Create Board](/r/gnoland/boards2/v1:create-board) • [List Admin Users](/r/gnoland/boards2/v1:admin-users) • [Help](/r/gnoland/boards2/v1:help)
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:?order=asc)
//
// ###### [CCC](/r/gnoland/boards2/v1:CCC)
// Created by [g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx](/u/g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) on 2009-02-13 11:31pm UTC, #3
// **0 threads**
//
// ###### [BBB](/r/gnoland/boards2/v1:BBB)
// Created by [g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx](/u/g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) on 2009-02-13 11:31pm UTC, #2
// **0 threads**
//
// ###### [AAA](/r/gnoland/boards2/v1:AAA)
// Created by [g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx](/u/g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) on 2009-02-13 11:31pm UTC, #1
// **0 threads**
