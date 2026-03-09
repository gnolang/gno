package main

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var dstBoard boards.BoardID

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	uinit.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))

	dstBoard = boards.CreateBoard(cross, "dst_board")
}

func main() {
	boards.CreateRepost(cross, 0, 0, "First Post in (title)", "Body of the first post. (body)", dstBoard)
}

// Error:
// invalid non-user call
