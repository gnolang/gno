// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_request_invite_04_filetest

package z_request_invite_04_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	user  address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.RequestInvite(cross(cur), 404)
}

// Error:
// board does not exist with ID: 404
