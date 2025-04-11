// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	"gno.land/r/gnoland/commondao"
)

const (
	owner    = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newRealm = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var (
	realm std.Address
	daoID uint64
)

func init() {
	// Init `commondao_test` realm address
	realm = std.CurrentRealm().Address()

	// Invite a user to be able to start creating DAOs
	testing.SetOriginCaller(owner)
	commondao.Invite(user)

	// Create a new DAO which gives ownership to `commondao_test`
	testing.SetOriginCaller(user)
	daoID = commondao.New("Foo").ID()
}

func main() {
	commondao.TransferOwnership(daoID, newRealm)

	println(commondao.IsOwner(realm, daoID))
	println(commondao.IsOwner(newRealm, daoID))
}

// Output:
// false
// true
