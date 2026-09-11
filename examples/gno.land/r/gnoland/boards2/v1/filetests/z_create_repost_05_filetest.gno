// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_repost_05_filetest

// Creating a thread requires the user to have 3000 GNOT, but reposing a thread
// also creates a thread. Test that this fails when the user has insufficient GNOT.
package z_create_repost_05_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	srcBID boards.ID
	dstBID boards.ID
	srcTID boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross(cur), "src-board", false, false)
	dstBID = boards2.CreateBoard(cross(cur), "dst-board", false, true)

	srcTID = boards2.CreateThread(cross(cur), srcBID, "Foo", "bar")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Non members should be able to repost threads only if they have enough GNOT
	boards2.CreateRepost(cross(cur), srcBID, srcTID, dstBID, "foo", "bar")
}

// Error:
// caller is not allowed to repost threads: account amount is lower than 3000 GNOT
