// Render boards sorted from newest to oldest.
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross, "AAA", true, false)
	boards2.CreateBoard(cross, "BBB", true, false)
	boards2.CreateBoard(cross, "CCC", true, false)
}

func main() {
	println(boards2.Render("?order=desc"))
}

// Output:
// # Boards
// [Create Board](/r/gnoland/boards2/v1:create-board) • [List Admin Users](/r/gnoland/boards2/v1:admin-users) • [Help](/r/gnoland/boards2/v1:help)
//
// ---
// Sort by: [newest first](/r/gnoland/boards2/v1:?order=asc)
//
// **[CCC](/r/gnoland/boards2/v1:CCC)**
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #3
// **0 threads**
//
// **[BBB](/r/gnoland/boards2/v1:BBB)**
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #2
// **0 threads**
//
// **[AAA](/r/gnoland/boards2/v1:AAA)**
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #1
// **0 threads**
