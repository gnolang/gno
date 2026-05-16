// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_5_a_filetest

package z_5_a_filetest

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

type propDef struct{}

func (propDef) Title() string                          { return "" }
func (propDef) Body() string                           { return "" }
func (propDef) VotingPeriod() time.Duration            { return time.Hour }
func (propDef) Validate() error                        { return nil }
func (propDef) Tally(pdao.VotingContext) (bool, error) { return false, nil }

var (
	daoID    uint64
	proposal *pdao.Proposal
	vote     pdao.VoteChoice = pdao.ChoiceYes
)

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross2(cur), user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New(cross2(cur), "Foo")
	daoID = dao.ID()

	// Configure DAO
	dao.Members().Add(user)

	// Create a new proposal
	proposal, _ = dao.Propose(user, propDef{})
}

func main(cur realm) {
	// User must be the caller to Vote()
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Vote(cross2(cur), daoID, proposal.ID(), vote, "")

	record := proposal.VotingRecord()
	if record.Size() != 1 {
		panic("expected a single vote")
	}

	println(record.HasVoted(user))
	record.Iterate(0, record.Size(), false, func(v pdao.Vote) bool {
		println(v.Choice == vote)
		return false
	})
}

// Output:
// true
// true
