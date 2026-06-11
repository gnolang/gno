// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_7_b_filetest

package z_7_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.NewSubDAO(cross(cur), "Foo", 404)
}

// Error:
// DAO not found
