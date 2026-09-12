// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_18_a_filetest

package z_18_a_filetest

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
	testing.SetRealm(testing.NewUserRealm(user))

	// A Charter requires a purpose: an empty purpose is rejected.
	commondao.New(cross(cur), "Foo", "", "", "")
}

// Error:
// DAO purpose is empty
