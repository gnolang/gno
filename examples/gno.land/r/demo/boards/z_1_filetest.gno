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
	users.Register(cross, "gnouser123")

	_ = boards.CreateBoard(cross, "test_board_1")
	_ = boards.CreateBoard(cross, "test_board_2")
}

func main() {
	println(boards.Render(""))
}

// Error:
// r/sys/users: current realm/user does not exist in whitelist: CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users } (previous: CodeRealm{ g1zypp3wv4qg9cv9rs57r34jj6z7uuqxjztyeqyk, gno.land/r/gnoland/users/v1 })
