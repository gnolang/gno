// Render default realm view when there are no boards.
package main

import (
	boards2 "gno.land/r/gnoland/boards2/v1"
)

func main() {
	println(boards2.Render(""))
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
