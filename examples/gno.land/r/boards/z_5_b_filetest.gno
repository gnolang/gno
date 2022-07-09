package main

// SEND: 2000000000ugnot

import (
	"std"
	"strconv"

	"gno.land/p/testutils"
	"gno.land/r/boards"
	"gno.land/r/users"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	users.Register("", "gnouser", "my profile")
	// create board via registered user
	bid := boards.CreateBoard("test_board")

	// create post via anon user
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test2)
	std.TestSetOrigSend(std.Coins{{"ugnot", 9000000}}, nil)

	pid := boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// please register, otherwise minimum fee 100 is required if anonymous
