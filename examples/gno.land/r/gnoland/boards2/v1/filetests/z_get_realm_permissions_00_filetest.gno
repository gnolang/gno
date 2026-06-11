// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_get_realm_permissions_00_filetest

package z_get_realm_permissions_00_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	perms := boards2.GetRealmPermissions()
	println(perms != nil)
}

// Output:
// true
