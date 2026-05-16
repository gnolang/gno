// PKGPATH: gno.land/r/archive/boards/filetests/z_5_d_filetest

package z_5_d_filetest

// SEND: 1000000ugnot

import (
	"chain"
	"strconv"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

const admin = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)
	// create board via registered user
	bid := boards.CreateBoard(cross2(cur), "test_board")
	pid := boards.CreateThread(cross2(cur), bid, "First Post (title)", "Body of the first post. (body)")

	// create reply via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(chain.Coins{{"ugnot", 9000000}})
	boards.CreateReply(cross2(cur), bid, pid, pid, "Reply of the first post")

	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// please register, otherwise minimum fee 100000000 is required if anonymous
