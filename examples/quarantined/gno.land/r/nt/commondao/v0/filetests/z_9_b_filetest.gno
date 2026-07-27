// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_9_b_filetest

package z_9_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Council is {user} only
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Resign(cross(cur), daoID)
}

// Error:
// council update would remove every council member
