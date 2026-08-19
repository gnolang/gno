// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_4_b_filetest

// Test getting the realm admin users using a zero board ID
package z_hub_4_b_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))

	// Invite a realm member, which requires a zero board ID
	boards2.InviteMember(cross(cur), 0, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
}

func main(cur realm) {
	// A zero board ID must list the realm admin users instead of returning
	// nothing, matching how IsMember and the members page resolve it.
	members := boards2.GetMembers(0, 0, 10)

	for _, m := range members {
		println(m.Address(), m.Roles())
	}
}

// Output:
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 slice[("admin" string)]
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh slice[("owner" string)]
