package main

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var dstBoard boards.BoardID

func init() {
	users.Register("", "gnouser", "my profile")

	dstBoard = boards.CreateBoard("dst_board")

	boards.CreateRepost(0, 0, "First Post in (title)", "Body of the first post. (body)", dstBoard)
}

func main() {
}

// Error:
// src board not exist
