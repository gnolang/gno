// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_lock_realm_04_filetest

package z_lock_realm_04_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.LockRealm(cross(cur), true)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Should fail because realm is already locked
	boards2.LockRealm(cross(cur), true)
}

// Error:
// realm and members are locked
