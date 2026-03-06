package main

// SEND: 1000000ugnot

import (
	"chain"
	"strconv"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/archive/boards"
	users "gno.land/r/gnoland/users/v1"
)

const admin = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")
	// create board via registered user
	bid := boards.CreateBoard(cross, "test_board")
	pid := boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")

	// create reply via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(chain.Coins{{"ugnot", 9000000}})
	boards.CreateReply(cross, bid, pid, pid, "Reply of the first post")

	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// please register, otherwise minimum fee 100000000 is required if anonymous
