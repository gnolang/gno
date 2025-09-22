package main

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	srcBoard boards.BoardID
	pid      boards.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")

	srcBoard = boards.CreateBoard(cross, "first_board")
	pid = boards.CreateThread(cross, srcBoard, "First Post in (title)", "Body of the first post. (body)")
}

func main() {
	boards.CreateRepost(cross, srcBoard, pid, "First Post in (title)", "Body of the first post. (body)", 0)
}

// Error:
// invalid non-user call
