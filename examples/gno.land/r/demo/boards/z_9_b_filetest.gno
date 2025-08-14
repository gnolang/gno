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
	users.Register(cross, "gnouser123")

	srcBoard = boards.CreateBoard(cross, "first_board")
	pid = boards.CreateThread(cross, srcBoard, "First Post in (title)", "Body of the first post. (body)")
}

func main() {
	boards.CreateRepost(cross, srcBoard, pid, "First Post in (title)", "Body of the first post. (body)", 0)
}

// Error:
// r/sys/users: current realm/user does not exist in whitelist: CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users } (previous: CodeRealm{ g1zypp3wv4qg9cv9rs57r34jj6z7uuqxjztyeqyk, gno.land/r/gnoland/users/v1 })
