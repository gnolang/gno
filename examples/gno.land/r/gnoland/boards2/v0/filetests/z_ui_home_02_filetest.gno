// Render default realm view when there are no boards.
package main

import (
	boards2 "gno.land/r/gnoland/boards2/v0"
)

func main() {
	println(boards2.Render(""))
}

// Output:
// # Boards
// [Create Board](/r/gnoland/boards2/v0:create-board) • [List Admin Users](/r/gnoland/boards2/v0:admin-users) • [Help](/r/gnoland/boards2/v0:help)
//
// ---
// ### Currently there are no boards
// Be the first to [create a new board](/r/gnoland/boards2/v0:create-board)!
