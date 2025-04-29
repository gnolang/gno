package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	srcBoard boards.BoardID
	pid      boards.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")

	srcBoard = cross(boards.CreateBoard)("first_board")
	pid = cross(boards.CreateThread)(srcBoard, "First Post in (title)", "Body of the first post. (body)")
}

func main() {
	cross(boards.CreateRepost)(srcBoard, pid, "First Post in (title)", "Body of the first post. (body)", 0)
}

// Error:
// dst board not exist
