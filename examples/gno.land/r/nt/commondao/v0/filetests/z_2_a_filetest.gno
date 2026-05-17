// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_2_a_filetest

package z_2_a_filetest

import (
	"chain"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "Foo"
)

var realmAddr = chain.PackageAddress("gno.land/r/demo/test")

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross2(cur), user)
}

func main(cur realm) {
	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))

	dao := commondao.New(cross2(cur), name)
	if dao == nil {
		panic("expected DAO to be created")
	}

	println(dao.Name() == name)
	println(commondao.IsOwner(realmAddr, dao.ID()))
	println(commondao.IsInvited(user))
	println(commondao.GetOptions(cross2(cur), dao.ID()) != nil)
}

// Output:
// true
// true
// false
// true
