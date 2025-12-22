package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	title         = "Test Thread"
	body          = "Test body"
	path          = "test-board/1"
)

var bid boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	tid := boards2.CreateThread(cross, bid, title, body)

	// Ensure that returned ID is right
	println(tid == 1)

	// Thread should not be frozen by default
	println(boards2.IsThreadFrozen(bid, tid))

	// Render content must contains thread's title and body
	println(boards2.Render(path))
}

// Output:
// true
// false
// # Boards › test-board
// ## Test Thread
//
// Test body
//
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board/1/reply) • [Edit](/r/gnoland/boards2/v1:test-board/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=1&threadID=1)
