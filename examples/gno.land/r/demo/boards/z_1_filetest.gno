package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var board *boards.Board

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")

	_ = cross(boards.CreateBoard)("test_board_1")
	_ = cross(boards.CreateBoard)("test_board_2")
}

func main() {
	println(boards.Render(""))
}

// Output:
// These are all the boards of this realm:
//
//  * [/r/demo/boards:test_board_1](/r/demo/boards:test_board_1)
//  * [/r/demo/boards:test_board_2](/r/demo/boards:test_board_2)
//
