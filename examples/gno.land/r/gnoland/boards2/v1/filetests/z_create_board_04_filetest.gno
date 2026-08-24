// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_board_04_filetest

package z_create_board_04_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	uinit "gno.land/r/sys/users/init"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init(cur realm) {
	// uinit.RegisterUser is genesis-only since the security fix.
	testing.SetHeight(0)
	uinit.RegisterUser(cross(cur), "gnoland", address("g1g3lsfxhvaqgdv4ccemwpnms4fv6t3aq3p5z6u7"))
	testing.SetHeight(123)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross(cur), "gnoland", false, false)
}

// Error:
// board name is a user name registered to a different user
