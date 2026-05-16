// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_8_f_filetest

package z_8_f_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootID uint64

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross2(cur), user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(testing.NewUserRealm(user))

	// Create root DAO without subDAOs
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	rootID = commondao.New(cross2(cur), "Root DAO").ID()
}

func main() {
	// Use a caller that is not the owner of the DAO
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/tests"))

	commondao.NewIterator(rootID)
}

// Error:
// unauthorized
