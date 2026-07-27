// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_5_c_filetest

package z_5_c_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	daoID = commondao.New(cross(cur), "Foo", string(user))
	commondao.UpdateOptions(cross(cur), daoID, commondao.AllowTextProposals(true))

	testing.SetRealm(testing.NewUserRealm(user))
	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
}

func main(cur realm) {
	// The caller is not a member of the proposal's electorate
	commondao.Vote(cross(cur), daoID, pID, pcommondao.ChoiceYes, "")
}

// Error:
// account is not a member of the proposal's electorate
