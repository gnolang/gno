package main

// SEND: 1000000ugnot

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var board *boards.Board

func init() {
	testing.SetRealm(runtime.NewUserRealm(chain.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")

	_ = boards.CreateBoard(cross, "test_board_1")
	_ = boards.CreateBoard(cross, "test_board_2")
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
