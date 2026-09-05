// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_realm_notice_01_filetest

package z_set_realm_notice_01_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func init(cur realm) {
	// Set an initial message so it can be cleared
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.SetRealmNotice(cross(cur), "This is a test realm message")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetRealmNotice(cross(cur), "")

	println(boards2.Notice == "")
}

// Output:
// true
