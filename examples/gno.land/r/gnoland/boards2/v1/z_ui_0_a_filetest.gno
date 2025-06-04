// Render default realm view.
// Default realm view must render the list of non private boards.
package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Create board "AAA" with a single thread
	boardID := cross(boards2.CreateBoard)("AAA", false)
	cross(boards2.CreateThread)(boardID, "Foo", "Bar", false)

	// Create 2 more boards
	cross(boards2.CreateBoard)("BBB", false)
	cross(boards2.CreateBoard)("CCC", false)
	cross(boards2.CreateBoard)("DDD", true) // <-- Private board
}

func main() {
	println(boards2.Render(""))
}

// Output:
// # Boards
//
// [Create Board](/r/gnoland/boards2/v1$help&func=CreateBoard&name=&private=false) • [List Admin Users](/r/gnoland/boards2/v1:admin-users) • [Help](/r/gnoland/boards2/v1:help)
//
//
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:?order=desc)
//
// **[AAA](/r/gnoland/boards2/v1:AAA)**
// Created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC, #1
// **1 threads**
//
// **[BBB](/r/gnoland/boards2/v1:BBB)**
// Created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC, #2
// **0 threads**
//
// **[CCC](/r/gnoland/boards2/v1:CCC)**
// Created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC, #3
// **0 threads**
