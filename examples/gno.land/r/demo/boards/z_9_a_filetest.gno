package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var dstBoard boards.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")

	dstBoard = cross(boards.CreateBoard)("dst_board")
}

func main() {
	cross(boards.CreateRepost)(0, 0, "First Post in (title)", "Body of the first post. (body)", dstBoard)
}

// Error:
// src board not exist
