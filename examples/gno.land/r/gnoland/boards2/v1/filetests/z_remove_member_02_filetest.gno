// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_remove_member_02_filetest

package z_remove_member_02_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RemoveMember(cross(cur), 0, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // Operate on realm DAO instead of individual boards
}

// Error:
// member not found
