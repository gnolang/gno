// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_lock_realm_04_filetest

package z_lock_realm_04_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.LockRealm(cross2(cur), true)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Should fail because realm is already locked
	boards2.LockRealm(cross2(cur), true)
}

// Error:
// realm and members are locked
