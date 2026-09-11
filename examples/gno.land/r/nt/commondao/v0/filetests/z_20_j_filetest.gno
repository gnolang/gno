// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_j_filetest

// Only a council member may propose arbitrary execution.
//
// /p/ Propose does not check council membership — it checks the DAO is
// live, looks up the kind, takes the latch and counts the cap — so this
// wrapper's gate is the only thing standing between a stranger and an
// ExecFunc closure of their choosing running with the DAO's authority.
// The council would still have to vote it through, but the payload
// author would be unauthenticated and the closure is not displayable.
package z_20_j_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx, genesis council only
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1, owns the DAO below
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// Opt the DAO into the execution kind, so the gate below is the only
	// thing left to refuse the call.
	pID := commondao.CreateRegisterKindProposal(cross(cur), daoID, "execution")
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main(cur realm) {
	// owner sits on the genesis DAO's council, not on this DAO's.
	testing.SetRealm(testing.NewUserRealm(owner))

	commondao.CreateExecutionProposal(cross(cur), daoID, "T", "B",
		func(_ int, sub realm) error { return nil })
}

// Error:
// caller is not a council member
