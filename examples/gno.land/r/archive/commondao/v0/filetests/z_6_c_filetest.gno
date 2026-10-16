// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_6_c_filetest

package z_6_c_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	daoID = commondao.New(cross(cur), "Foo").ID()
}

func main(cur realm) {
	commondao.Execute(cross(cur), daoID, 1)
}

// Error:
// caller is not a DAO member
