// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_flag_reply_01_filetest

package z_flag_reply_01_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FlagReply(cross(cur), 404, 1, 1, "Reason")
}

// Error:
// board does not exist with ID: 404
