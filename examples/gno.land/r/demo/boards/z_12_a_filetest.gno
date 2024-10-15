// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

func main() {
	users.Register("", "gnouser", "my profile")
	// create a post via registered user
	bid1 := boards.CreateBoard("test_board1")
	pid := boards.CreateThread(bid1, "First Post (title)", "Body of the first post. (body)")
	bid2 := boards.CreateBoard("test_board2")

	// create a repost via anon user
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test2)
	std.TestSetOrigSend(std.Coins{{"ugnot", 9000000}}, nil)

	rid := boards.CreateRepost(bid1, pid, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// please register, otherwise minimum fee 100000000 is required if anonymous
