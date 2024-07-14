package simpledao

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
	"gno.land/p/gov/dao"
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
			votes:       newVotes(),
			executor:    nil,
		}

		proposals = append(proposals, proposal)
	}

	return proposals
}

func equalProposals(t *testing.T, p1, p2 dao.Proposal) {
	t.Helper()

	uassert.Equal(
		t,
		p1.GetAuthor().String(),
		p2.GetAuthor().String(),
	)

	uassert.Equal(
		t,
		p1.GetDescription(),
		p2.GetDescription(),
	)

	uassert.Equal(
		t,
		p1.GetStatus().String(),
		p2.GetStatus().String(),
	)

	uassert.Equal(
		t,
		len(p1.GetVotes()),
		len(p2.GetVotes()),
	)

	p1Votes := p1.GetVotes()
	for index, v2 := range p2.GetVotes() {
		uassert.Equal(
			t,
			p1Votes[index].Address.String(),
			v2.Address.String(),
		)

		uassert.Equal(
			t,
			p1Votes[index].Option.String(),
			v2.Option.String(),
		)
	}
}

func TestProposal_Data(t *testing.T) {
	t.Parallel()

	t.Run("author", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			author: testutils.TestAddress("address"),
		}

		uassert.Equal(t, p.author, p.GetAuthor())
	})

	t.Run("description", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			description: "example proposal description",
		}

		uassert.Equal(t, p.description, p.GetDescription())
	})

	t.Run("status", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			status: dao.ExecutionSuccessful,
		}

		uassert.Equal(t, p.status.String(), p.GetStatus().String())
	})

	t.Run("no votes", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			votes: newVotes(),
		}

		uassert.Equal(t, 0, len(p.GetVotes()))
	})

	t.Run("existing votes", func(t *testing.T) {
		t.Parallel()

		var (
			members = generateMembers(t, 50)
			p       = &proposal{
				votes: newVotes(),
			}
		)

		for _, m := range members {
			urequire.NoError(t, p.votes.castVote(m, dao.YesVote))
		}

		votes := p.GetVotes()

		urequire.Equal(t, len(members), len(votes))

		for _, v := range votes {
			for _, m := range members {
				if m.Address != v.Address {
					continue
				}

				uassert.Equal(t, dao.YesVote.String(), v.Option.String())
			}
		}
	})

	t.Run("vote by member", func(t *testing.T) {
		t.Parallel()

		var (
			members = generateMembers(t, 50)
			p       = &proposal{
				votes: newVotes(),
			}
		)

		for _, m := range members {
			urequire.NoError(t, p.votes.castVote(m, dao.YesVote))
		}

		votes := p.GetVotes()
		urequire.Equal(t, len(members), len(votes))

		vote, err := p.GetVoteByMember(members[0].Address)
		urequire.NoError(t, err)

		uassert.Equal(t, dao.YesVote.String(), vote.Option.String())
		uassert.Equal(t, members[0].Address.String(), vote.Address.String())
	})

	t.Run("missing vote by member", func(t *testing.T) {
		t.Parallel()

		p := &proposal{
			votes: newVotes(),
		}

		votes := p.GetVotes()
		urequire.Equal(t, 0, len(votes))

		_, err := p.GetVoteByMember(testutils.TestAddress("dummy"))
		uassert.ErrorIs(t, err, errMissingVote)
	})
}

func TestPropStore_GetProposals(t *testing.T) {
	t.Parallel()

	t.Run("no proposals", func(t *testing.T) {
		t.Parallel()

		p := NewPropStore()

		proposals := p.GetProposals(0, 0)

		uassert.Equal(t, 0, len(proposals))
	})

	t.Run("proper pagination", func(t *testing.T) {
		t.Parallel()

		var (
			numProposals = 20
			halfRange    = numProposals / 2

			p         = NewPropStore()
			proposals = generateProposals(t, numProposals)
		)

		// Add initial proposals
		for _, proposal := range proposals {
			_, err := p.addProposal(proposal)

			urequire.NoError(t, err)
		}

		fetchedProposals := p.GetProposals(0, uint64(halfRange))
		urequire.Equal(t, halfRange, len(fetchedProposals))

		for index, fetchedProposal := range fetchedProposals {
			equalProposals(t, proposals[index], fetchedProposal)
		}

		// Fetch the other half
		fetchedProposals = p.GetProposals(uint64(halfRange), uint64(halfRange))
		urequire.Equal(t, halfRange, len(fetchedProposals))

		for index, fetchedProposal := range fetchedProposals {
			equalProposals(t, proposals[index+halfRange], fetchedProposal)
		}
	})
}

func TestPropStore_GetProposalByID(t *testing.T) {
	t.Parallel()

	t.Run("missing proposal", func(t *testing.T) {
		t.Parallel()

		p := NewPropStore()

		_, err := p.GetProposalByID(0)
		uassert.ErrorIs(t, err, errMissingProposal)
	})

	t.Run("proposal found", func(t *testing.T) {
		t.Parallel()

		var (
			p        = NewPropStore()
			proposal = generateProposals(t, 1)[0]
		)

		// Add the initial proposal
		_, err := p.addProposal(proposal)
		urequire.NoError(t, err)

		// Fetch the proposal
		fetchedProposal, err := p.GetProposalByID(0)
		urequire.NoError(t, err)

		equalProposals(t, proposal, fetchedProposal)
	})
}
