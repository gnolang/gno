package main

// SEND: 1000000ugnot

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")
	// create board via registered user
	bid := boards.CreateBoard(cross, "test_board")
	pid := boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")

	// create reply via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(std.Coins{{"ugnot", 9000000}})
	boards.CreateReply(cross, bid, pid, pid, "Reply of the first post")

	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// r/sys/users: current realm/user does not exist in whitelist: CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users } (previous: CodeRealm{ g1zypp3wv4qg9cv9rs57r34jj6z7uuqxjztyeqyk, gno.land/r/gnoland/users/v1 })
