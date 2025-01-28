package commondao

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
)

func TestProposalDefaults(t *testing.T) {
	id := uint64(1)
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	votingPeriod := time.Minute * 10

	p, err := NewProposal(id, creator, testPropDef{votingPeriod: votingPeriod})

	uassert.NoError(t, err)
	uassert.Equal(t, p.ID(), id)
	uassert.NotEqual(t, p.Definition(), nil)
	uassert.Equal(t, p.Quorum(), DefaultQuorum)
	uassert.True(t, p.Status() == StatusActive)
	uassert.Equal(t, p.Creator(), creator)
	uassert.False(t, p.CreatedAt().IsZero())
	uassert.NotEqual(t, p.VotingRecord(), nil)
	uassert.Empty(t, p.StatusReason())
	uassert.True(t, p.VotingDeadline() == p.CreatedAt().Add(votingPeriod))
}

type testPropDef struct {
	votingPeriod                time.Duration
	tallyResult                 bool
	validationErr, executionErr error
}

func (testPropDef) Title() string                          { return "" }
func (testPropDef) Body() string                           { return "" }
func (testPropDef) Quorum() float64                        { return 0 }
func (d testPropDef) VotingPeriod() time.Duration          { return d.votingPeriod }
func (d testPropDef) Validate() error                      { return d.validationErr }
func (d testPropDef) Tally(ReadOnlyVotingRecord, int) bool { return d.tallyResult }
func (d testPropDef) Execute() error                       { return d.executionErr }
