// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_lock_realm_05_filetest

package z_lock_realm_05_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func init(cur realm) {
	// Lock the realm without locking realm members
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.LockRealm(cross(cur), false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.LockRealm(cross(cur), true)

	println(boards2.IsRealmLocked())
	println(boards2.AreRealmMembersLocked())
}

// Output:
// true
// true
