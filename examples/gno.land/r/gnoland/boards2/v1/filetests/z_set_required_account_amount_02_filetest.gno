// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_required_account_amount_02_filetest

package z_set_required_account_amount_02_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetRequiredAccountAmount(cross(cur), 1_000_000)
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
