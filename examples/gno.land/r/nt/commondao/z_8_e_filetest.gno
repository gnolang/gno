// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(std.NewUserRealm(user))

	// Create root DAO with a subDAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	rootID = commondao.New("Root DAO").ID()
}

func main() {
	crossing()

	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	commondao.GetSubDAO(rootID, "not/found")
}

// Error:
// subDAO not found
