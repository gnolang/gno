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
	cross(users.Register)("gnouser123")
	// create board via registered user
	bid := cross(boards.CreateBoard)("test_board")

	// create post via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(std.Coins{{"ugnot", 9000000}})

	pid := cross(boards.CreateThread)(bid, "First Post (title)", "Body of the first post. (body)")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// please register, otherwise minimum fee 100000000 is required if anonymous
