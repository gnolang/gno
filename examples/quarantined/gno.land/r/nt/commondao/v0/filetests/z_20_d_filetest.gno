// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_d_filetest

package z_20_d_filetest

import (
	"testing"
	"time"

	pdao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

// customDef / customKind is a foreign proposal kind defined in this
// persistent realm and registered on a DAO by value through governance via
// CreateRegisterCustomKindProposal.
type customDef struct{}

func (customDef) Title() string               { return "Custom" }
func (customDef) Body() string                { return "custom" }
func (customDef) VotingPeriod() time.Duration { return time.Hour * 24 }
func (customDef) Threshold() pdao.Threshold   { return pdao.ThresholdSupermajority }

type customKind struct{}

func (customKind) Name() string { return "custom" }
func (customKind) New(_ pdao.ReadonlyCommonDAO, _ any) (pdao.ProposalDefinition, error) {
	return customDef{}, nil
}

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	println("custom enabled before:", commondao.IsProposalKindEnabled(daoID, "custom"))

	// Governance-register the foreign kind by value through the permanent
	// manage-kinds kind (no opt-in enable needed: manage-kinds is seeded).
	pID := commondao.CreateRegisterCustomKindProposal(cross(cur), daoID, customKind{})
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
	println("custom enabled after:", commondao.IsProposalKindEnabled(daoID, "custom"))

	// Inert-to-propose on THIS realm: the kind is registered, but there is
	// no propose path for it (every Create* wrapper hardcodes its own kind
	// and mustPropose is unexported). It is useful only to a downstream
	// realm that authors its own propose wrapper.
}

// Output:
// custom enabled before: false
// custom enabled after: true
