// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_7_d_filetest

package z_7_d_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var parentID uint64

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// The origin must be the invited user
	testing.SetRealm(testing.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao/v0"))
	parentDAO := commondao.New(cross(cur), "Parent DAO", commondao.AllowChildren(false)) // Disallow SubDAOs
	parentID = parentDAO.ID()
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao/v0"))

	commondao.NewSubDAO(cross(cur), "Foo", parentID)
}

// Error:
// SubDAO support is not enabled
