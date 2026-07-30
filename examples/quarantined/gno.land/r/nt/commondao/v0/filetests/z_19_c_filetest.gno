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

	// Governance-deregister the text kind: new text proposals are blocked
	// (see z_19_a for the terminal rejection)
	pID := commondao.CreateDeregisterKindProposal(cross(cur), daoID, "text")
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	println("after deregister:", commondao.HasProposalKind(daoID, "text"))

	// The DAO page drops the deregistered kind's create link
	out := commondao.Render(strconv.FormatUint(daoID, 10))
	println("create link after deregister:", strings.Contains(out, "CreateTextProposal"))

	// ...and the settings kinds row drops the kind name (sorted row would
	// read "subdao, text, treasury\-clawback" when registered)
	out = commondao.Render(strconv.FormatUint(daoID, 10) + "/settings")
	println("settings row lists text after deregister:", strings.Contains(out, "subdao, text"))

	// Re-register the kind by name through a second governance proposal
	pID := commondao.CreateRegisterKindProposal(cross(cur), daoID, "text")
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("after re-register:", commondao.HasProposalKind(daoID, "text"))

	// ...and shows it again once the kind is re-registered
	out = commondao.Render(strconv.FormatUint(daoID, 10))
	println("create link after re-register:", strings.Contains(out, "CreateTextProposal"))

	// Text proposals can be created again
	tID := commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
	println("created:", tID)
}

// Output:
// after deregister: false
// create link after deregister: false
// settings row lists text after deregister: false
// after re-register: true
// create link after re-register: true
// created: 3
