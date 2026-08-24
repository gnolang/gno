// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_19_e_filetest

package z_19_e_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner    = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	stranger = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	// Only council members can propose kind changes
	testing.SetRealm(testing.NewUserRealm(stranger))
	commondao.CreateDeregisterKindProposal(cross(cur), daoID, "text")
}

// Error:
// caller is not a council member
