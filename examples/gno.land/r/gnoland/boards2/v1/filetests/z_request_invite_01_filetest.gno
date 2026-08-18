// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_request_invite_01_filetest

package z_request_invite_01_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))

	boards2.RequestInvite(cross(cur), bid)
}

// Error:
// caller must be user
