// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_repost_02_filetest

package z_create_repost_02_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	srcBID boards.ID
	dstBID boards.ID
	srcTID boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross2(cur), "src-board", false, false)
	dstBID = boards2.CreateBoard(cross2(cur), "dst-board", false, false)

	srcTID = boards2.CreateThread(cross2(cur), srcBID, "Foo", "bar")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.CreateRepost(cross2(cur), srcBID, srcTID, dstBID, "foo", "bar")
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
