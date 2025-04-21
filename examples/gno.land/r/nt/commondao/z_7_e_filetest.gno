// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootDAO *pcommondao.CommonDAO

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(std.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	rootDAO = commondao.New("Parent DAO")

	// Change options to disallow subDAOs
	options := commondao.GetOptions(rootDAO.ID())
	options.SetAllowChildren(false)
}

func main() {
	crossing()

	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	commondao.NewSubDAO("Foo", rootDAO)
}

// Error:
// SubDAO support is not enabled
