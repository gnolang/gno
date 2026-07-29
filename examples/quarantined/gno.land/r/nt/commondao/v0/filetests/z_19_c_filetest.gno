// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_19_c_filetest

package z_19_c_filetest

import (
	"strconv"
	"strings"
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
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// Governance-disable the text kind: new text proposals are blocked
	// (see z_19_a for the terminal rejection)
	pID := commondao.CreateSetProposalKindProposal(cross(cur), daoID, "text", false)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	println("after disable:", commondao.IsProposalKindEnabled(daoID, "text"))

	// The DAO page drops the disabled kind's create link
	out := commondao.Render(strconv.FormatUint(daoID, 10))
	println("create link after disable:", strings.Contains(out, "CreateTextProposal"))

	// Re-enable the kind through a second governance proposal
	pID := commondao.CreateSetProposalKindProposal(cross(cur), daoID, "text", true)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("after re-enable:", commondao.IsProposalKindEnabled(daoID, "text"))

	// ...and shows it again once the kind is re-enabled
	out = commondao.Render(strconv.FormatUint(daoID, 10))
	println("create link after re-enable:", strings.Contains(out, "CreateTextProposal"))

	// Text proposals can be created again
	tID := commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
	println("created:", tID)
}

// Output:
// after disable: false
// create link after disable: false
// after re-enable: true
// create link after re-enable: true
// created: 3
