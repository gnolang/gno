package main

// SEND: 1000000ugnot

import (
	"chain"
	"strconv"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/boards"
	susers "gno.land/r/sys/users"
)

const admin = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	// create board via registered user
	bid := boards.CreateBoard(cross, "test_board")

	// create post via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(chain.Coins{{"ugnot", 9000000}})

	pid := boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// please register, otherwise minimum fee 100000000 is required if anonymous
