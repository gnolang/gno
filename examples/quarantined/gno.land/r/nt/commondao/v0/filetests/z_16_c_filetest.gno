// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_16_c_filetest

package z_16_c_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
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
	commondao.SetListed(cross(cur), daoID, true)

	// Dissolve the DAO through a passed proposal
	pID := commondao.CreateDissolutionProposal(cross(cur), daoID, user)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main() {
	// Dissolution auto-unlists the DAO from the home index.
	println("listed before dissolve was true; after dissolve:", commondao.IsListed(daoID))
}

// Output:
// listed before dissolve was true; after dissolve: false
