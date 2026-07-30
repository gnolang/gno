// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_a_filetest

package z_20_a_filetest

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
	daoID uint64
	ran   bool
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Single-member council: one YES decides any proposal immediately
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// The execution kind is opt-in: register it by name through governance first.
	pID := commondao.CreateRegisterKindProposal(cross(cur), daoID, "execution")
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	println("execution enabled:", commondao.IsProposalKindEnabled(daoID, "execution"))

	// The closure is authored in this persistent realm, so it survives
	// Propose->Execute. It sets a package var so execution is observable.
	pID := commondao.CreateExecutionProposal(cross(cur), daoID, "Run it", "Body", func(_ int, _ realm) error {
		ran = true
		return nil
	})
	println("created:", pID)

	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("closure ran:", ran)
}

// Output:
// execution enabled: true
// created: 2
// closure ran: true
