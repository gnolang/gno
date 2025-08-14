package main

// SEND: 1000000ugnot

import (
	"std"
	"strconv"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	bid boards.BoardID
	pid boards.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")

	bid = boards.CreateBoard(cross, "test_board")
	pid = boards.CreateThread(cross, bid, "First Post in (title)", "Body of the first post. (body)")
}

func main() {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	// post 2 not exist
	boards.EditPost(cross, bid, pid, 2, "Edited: First Post in (title)", "Edited: Body of the first post. (body)")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// r/sys/users: current realm/user does not exist in whitelist: CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users } (previous: CodeRealm{ g1zypp3wv4qg9cv9rs57r34jj6z7uuqxjztyeqyk, gno.land/r/gnoland/users/v1 })
