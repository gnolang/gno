// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner    = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newRealm = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var daoID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// Create a new DAO which gives ownership to `commondao_test`
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
	daoID = commondao.New("Foo").ID()
}

func main() {
	// Use a caller that is not the owner of the DAO
	testing.SetRealm(std.NewUserRealm(user))

	cross(commondao.TransferOwnership)(daoID, newRealm)
}

// Error:
// unauthorized
