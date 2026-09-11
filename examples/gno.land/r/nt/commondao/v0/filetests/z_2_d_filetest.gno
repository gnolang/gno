// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_2_d_filetest

package z_2_d_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx") // govdao t1 multisig
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)
}

func main(cur realm) {
	// A realm relaying for the invited user is rejected: it would redeem
	// `user`'s invite and take the council seat for itself.
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/realm_a"))

	commondao.New(cross(cur), "Foo", "Purpose", "", string(owner))
}

// Error:
// only a user can create a DAO
