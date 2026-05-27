// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_thread_06_filetest

// Open board: Test creating a new thread as a non member user that has no GNOT
package z_create_thread_06_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, true)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Non members should be able to create threads only if they have enough GNOT
	boards2.CreateThread(cross(cur), bid, "Title", "Body")
}

// Error:
// caller is not allowed to create threads: account amount is lower than 3000 GNOT
