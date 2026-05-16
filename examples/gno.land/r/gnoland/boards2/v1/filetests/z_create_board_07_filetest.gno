// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_board_07_filetest

package z_create_board_07_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross2(cur), "test123", false, false)
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
