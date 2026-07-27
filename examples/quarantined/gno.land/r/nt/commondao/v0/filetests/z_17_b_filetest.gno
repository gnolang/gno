// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_17_b_filetest

package z_17_b_filetest

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

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// A DAO cannot modify its own council through the ancestor path
	// (self-mutation goes through CreateCouncilUpdateProposal instead).
	commondao.CreateAncestorCouncilUpdateProposal(cross(cur), daoID, daoID, string(owner), "")
}

// Error:
// a DAO cannot target itself
