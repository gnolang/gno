package main

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var (
	srcBoard boards.BoardID
	pid      boards.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	uinit.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))

	srcBoard = boards.CreateBoard(cross, "first_board")
	pid = boards.CreateThread(cross, srcBoard, "First Post in (title)", "Body of the first post. (body)")
}

func main() {
	boards.CreateRepost(cross, srcBoard, pid, "First Post in (title)", "Body of the first post. (body)", 0)
}

// Error:
// invalid non-user call
