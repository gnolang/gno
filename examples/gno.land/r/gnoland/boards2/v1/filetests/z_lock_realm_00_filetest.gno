// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_lock_realm_00_filetest

package z_lock_realm_00_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.LockRealm(cross2(cur), false)

	println(boards2.IsRealmLocked())
	println(boards2.AreRealmMembersLocked())
}

// Output:
// true
// false
