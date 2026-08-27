// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_i_filetest

package z_20_i_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID  uint64
	textID uint64
	execID uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Single-member council: one YES decides any proposal immediately.
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// Enable the execution kind through governance.
	{
		pID := commondao.CreateRegisterKindProposal(cross(cur), daoID, "execution")
		commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), daoID, pID)
	}

	// A second proposal, passed but left unexecuted: the re-entrant target.
	textID = commondao.CreateTextProposal(cross(cur), daoID, "Text", "Body", 7)
	commondao.Vote(cross(cur), daoID, textID, pdao.ChoiceYes, "")

	// An execution proposal whose closure crosses back in as the DAO's own
	// sub and calls Execute on the passed text proposal — a re-entrant
	// Execute. The closure is authored here so it survives Propose->Execute.
	execID = commondao.CreateExecutionProposal(cross(cur), daoID, "Re-enter", "Body", func(_ int, sub realm) error {
		commondao.Execute(cross(sub), daoID, textID)
		return nil
	})
	commondao.Vote(cross(cur), daoID, execID, pdao.ChoiceYes, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Executing the execution proposal runs its closure, which re-enters
	// Execute while the realm-global latch is raised. The nested Execute
	// panics; thrown across the cross(sub) boundary it arrived through, that
	// aborts the whole transaction.
	commondao.Execute(cross(cur), daoID, execID)
}

// Error:
// commondao: re-entrant Execute is not allowed
