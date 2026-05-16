// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_7_c_filetest

package z_7_c_filetest

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
	commondao.Invite(cross2(cur), user)

	// The origin must be the invited user
	testing.SetRealm(testing.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	parentID = commondao.New(cross2(cur), "Parent DAO").ID()
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))

	commondao.NewSubDAO(cross2(cur), "", parentID)
}

// Error:
// DAO name is empty
