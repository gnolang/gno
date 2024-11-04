package simpledao

import (
	"testing"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

// generateProposals generates dummy proposals
func generateProposals(t *testing.T, count int) []*proposal {
	t.Helper()

	var (
		members   = generateMembers(t, count)
		proposals = make([]*proposal, 0, count)
	)

	for i := 0; i < count; i++ {
		proposal := &proposal{
			author:      members[i].Address,
			description: ufmt.Sprintf("proposal %d", i),
			status:      dao.Active,
			tally:       newTally(),
			getTotalVotingPowerFn: func() uint64 {
				return 0
			},
			executor: nil,
		}

		proposals = append(proposals, proposal)
	}

	return proposals
}

func equalProposals(t *testing.T, p1, p2 dao.Proposal) {
	t.Helper()

	uassert.Equal(
		t,
		p1.Author().String(),
		p2.Author().String(),
	)

	uassert.Equal(
		t,
		p1.Description(),
		p2.Description(),
	)

	uassert.Equal(
		t,
		p1.Status().String(),
		p2.Status().String(),
	)

	p1Stats := p1.Stats()
	p2Stats := p2.Stats()

	uassert.Equal(t, p1Stats.YayVotes, p2Stats.YayVotes)
	uassert.Equal(t, p1Stats.NayVotes, p2Stats.NayVotes)
	uassert.Equal(t, p1Stats.AbstainVotes, p2Stats.AbstainVotes)
	uassert.Equal(t, p1Stats.TotalVotingPower, p2Stats.TotalVotingPower)
}

func TestProposal_Data(t *testing.T) {
	t.Parallel()

	t.Run("author", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			author: testutils.TestAddress("address"),
		}

		uassert.Equal(t, p.author, p.Author())
	})

	t.Run("description", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			description: "example proposal description",
		}

		uassert.Equal(t, p.description, p.Description())
	})

	t.Run("status", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			status: dao.ExecutionSuccessful,
		}

		uassert.Equal(t, p.status.String(), p.Status().String())
	})

	t.Run("executor", func(t *testing.T) {
		t.Parallel()

		var (
			numCalled = 0
			cb        = func() error {
				numCalled++

				return nil
			}

			ex = &mockExecutor{
				executeFn: cb,
			}

			p = &proposal{
				executor: ex,
			}
		)

		urequire.NoError(t, p.executor.Execute())
		urequire.NoError(t, p.Executor().Execute())

		uassert.Equal(t, 2, numCalled)
	})

	t.Run("no votes", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			tally: newTally(),
			getTotalVotingPowerFn: func() uint64 {
				return 0
			},
		}

		stats := p.Stats()

		uassert.Equal(t, uint64(0), stats.YayVotes)
		uassert.Equal(t, uint64(0), stats.NayVotes)
		uassert.Equal(t, uint64(0), stats.AbstainVotes)
		uassert.Equal(t, uint64(0), stats.TotalVotingPower)
	})

	t.Run("existing votes", func(t *testing.T) {
		t.Parallel()

		var (
			members    = generateMembers(t, 50)
			totalPower = uint64(len(members)) * 10

			p = &proposal{
				tally: newTally(),
				getTotalVotingPowerFn: func() uint64 {
					return totalPower
				},
			}
		)

		for _, m := range members {
			urequire.NoError(t, p.tally.castVote(m, dao.YesVote))
		}

		stats := p.Stats()

		uassert.Equal(t, totalPower, stats.YayVotes)
		uassert.Equal(t, uint64(0), stats.NayVotes)
		uassert.Equal(t, uint64(0), stats.AbstainVotes)
		uassert.Equal(t, totalPower, stats.TotalVotingPower)
	})
}

func TestSimpleDAO_GetProposals(t *testing.T) {
	t.Parallel()

	t.Run("no proposals", func(t *testing.T) {
		t.Parallel()

		s := New(nil)

		uassert.Equal(t, 0, s.Size())
		proposals := s.Proposals(0, 0)

		uassert.Equal(t, 0, len(proposals))
	})

	t.Run("proper pagination", func(t *testing.T) {
		t.Parallel()

		var (
			numProposals = 20
			halfRange    = numProposals / 2

			s         = New(nil)
			proposals = generateProposals(t, numProposals)
		)

		// Add initial proposals
		for _, proposal := range proposals {
			_, err := s.addProposal(proposal)

			urequire.NoError(t, err)
		}

		uassert.Equal(t, numProposals, s.Size())

		fetchedProposals := s.Proposals(0, uint64(halfRange))
		urequire.Equal(t, halfRange, len(fetchedProposals))

		for index, fetchedProposal := range fetchedProposals {
			equalProposals(t, proposals[index], fetchedProposal)
		}

		// Fetch the other half
		fetchedProposals = s.Proposals(uint64(halfRange), uint64(halfRange))
		urequire.Equal(t, halfRange, len(fetchedProposals))

		for index, fetchedProposal := range fetchedProposals {
			equalProposals(t, proposals[index+halfRange], fetchedProposal)
		}
	})
}

func TestSimpleDAO_GetProposalByID(t *testing.T) {
	t.Parallel()

	t.Run("missing proposal", func(t *testing.T) {
		t.Parallel()

		s := New(nil)

		_, err := s.ProposalByID(0)
		uassert.ErrorIs(t, err, ErrMissingProposal)
	})

	t.Run("proposal found", func(t *testing.T) {
		t.Parallel()

		var (
			s        = New(nil)
			proposal = generateProposals(t, 1)[0]
		)

		// Add the initial proposal
		_, err := s.addProposal(proposal)
		urequire.NoError(t, err)

		// Fetch the proposal
		fetchedProposal, err := s.ProposalByID(0)
		urequire.NoError(t, err)

		equalProposals(t, proposal, fetchedProposal)
	})
}
