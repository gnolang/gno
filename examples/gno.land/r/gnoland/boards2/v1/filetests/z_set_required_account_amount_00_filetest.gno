// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_required_account_amount_00_filetest

package z_set_required_account_amount_00_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetRequiredAccountAmount(cross(cur), 1_000_000)

	println(boards2.RequiredAccountAmount)
}

// Output:
// 1000000
