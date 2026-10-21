// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_1_a_filetest

package z_1_a_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	commondao.Invite(cross(cur), user)

	println(commondao.IsInvited(user))
}

// Output:
// true
