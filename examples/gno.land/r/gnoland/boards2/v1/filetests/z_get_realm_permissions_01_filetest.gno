// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_get_realm_permissions_01_filetest

package z_get_realm_permissions_01_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/test"))

	// Calling as a non Boards2 sub realm should fail
	_ = boards2.GetRealmPermissions(cross2(cur))
}

// Error:
// forbidden, caller should live within "gno.land/r/gnoland/boards2/" namespace
