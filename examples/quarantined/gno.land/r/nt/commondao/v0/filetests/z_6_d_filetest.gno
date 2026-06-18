// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_6_d_filetest

package z_6_d_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var dao *pcommondao.CommonDAO

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	dao = commondao.New(cross(cur), "Foo")

	// Configure DAO
	dao.Members().Add(user)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Execute(cross(cur), dao.ID(), 404)
}

// Error:
// proposal not found
