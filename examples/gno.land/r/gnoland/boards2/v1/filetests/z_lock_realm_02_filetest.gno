// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_lock_realm_02_filetest

package z_lock_realm_02_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2

func main(cur realm) {
	// Call realm with a user that has not permission to lock the realm
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.LockRealm(cross2(cur), false)
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
