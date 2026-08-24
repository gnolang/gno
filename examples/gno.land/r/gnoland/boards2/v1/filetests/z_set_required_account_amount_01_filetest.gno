// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_required_account_amount_01_filetest

package z_set_required_account_amount_01_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetRequiredAccountAmount(cross(cur), 0) // Disable

	println(boards2.RequiredAccountAmount)
}

// Output:
// 0
