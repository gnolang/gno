// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "Foo"
)

var realm std.Address

func init() {
	// Init `commondao_test` realm address
	realm = std.CurrentRealm().Address()

	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	std.TestSetOriginCaller(user)
}

func main() {
	dao := commondao.New(name)

	if dao == nil {
		panic("expected DAO to be created")
	}

	println(dao.Name() == name)
	println(commondao.IsOwner(realm, dao.ID()))
	println(commondao.IsInvited(user))
}

// Output:
// true
// true
// false
