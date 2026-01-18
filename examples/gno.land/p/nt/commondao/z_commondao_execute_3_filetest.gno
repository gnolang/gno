// PKGPATH: gno.land/r/test
package test

import (
	"errors"
	"time"

	"gno.land/p/nt/commondao"
)

const (
	member    address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devxZ
	groupName         = "admin"
)

var (
	dao      *commondao.CommonDAO
	proposal *commondao.Proposal
)

type testPropDef struct{}

func (testPropDef) Title() string               { return "" }
func (testPropDef) Body() string                { return "" }
func (testPropDef) VotingPeriod() time.Duration { return 0 }
func (testPropDef) Execute(cur realm) error     { return nil }

// Tally only succeeds if one YES vote from an admin is found
func (testPropDef) Tally(ctx commondao.VotingContext) (bool, error) {
	if ctx.VotingRecord.Size() < 1 {
		return false, errors.New("proposal requires at least 1 YES vote from an admin to pass")
	}

	admins, _ := ctx.Members.Grouping().Get(groupName)
	return admins.Members().IterateByOffset(0, admins.Members().Size(), func(member address) bool {
		v, found := ctx.VotingRecord.GetVote(member)
		// Proposal passes when a YES vote is found
		return (found && v.Choice == commondao.ChoiceYes)
	}), nil
}

func init() {
	// Create a DAO with member grouping support and a proposal
	storage := commondao.NewMemberStorageWithGrouping()
	dao = commondao.New(commondao.WithMemberStorage(storage))
	proposal = dao.MustPropose(member, testPropDef{})

	// Add a single group with members
	admins, _ := storage.Grouping().Add(groupName)
	admins.Members().Add(member)
	admins.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	// Add a vote to the voting record
	proposal.VotingRecord().AddVote(commondao.Vote{
		Address: member,
		Choice:  commondao.ChoiceYes,
	})
}

func main() {
	err := dao.Execute(proposal.ID())
	if err != nil {
		panic(err)
	}

	println(string(proposal.Status()))
	println(proposal.StatusReason())
}

// Output:
// executed
