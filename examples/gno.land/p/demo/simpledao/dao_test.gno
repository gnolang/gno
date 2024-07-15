package simpledao

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/gov/dao"
	"gno.land/p/gov/executor"
)

func TestSimpleDAO_Propose(t *testing.T) {
	t.Parallel()

	t.Run("invalid executor", func(t *testing.T) {
		t.Parallel()

		s := &SimpleDAO{}

		_, err := s.Propose("", nil)
		uassert.ErrorIs(
			t,
			err,
			errInvalidExecutor,
		)
	})

	t.Run("caller cannot cover fee", func(t *testing.T) {
		t.Parallel()

		var (
			called = false
			cb     = func() error {
				called = true

				return nil
			}
			ex = executor.NewCallbackExecutor(cb)

			sentCoins = std.NewCoins(
				std.NewCoin(
					"ugnot",
					minProposalFeeValue-1,
				),
			)

			s = &SimpleDAO{
				membStore: NewMembStore(), // empty member store
			}
		)

		// Set the sent coins to be lower
		// than the proposal fee
		std.TestSetOrigSend(sentCoins, std.Coins{})

		_, err := s.Propose("", ex)
		uassert.ErrorIs(
			t,
			err,
			errInsufficientProposalFunds,
		)

		uassert.False(t, called)
	})

	t.Run("proposal added", func(t *testing.T) {
		t.Parallel()

		var (
			called = false
			cb     = func() error {
				called = true

				return nil
			}

			ex          = executor.NewCallbackExecutor(cb)
			description = "Proposal description"

			proposer  = testutils.TestAddress("proposer")
			sentCoins = std.NewCoins(
				std.NewCoin(
					"ugnot",
					minProposalFeeValue, // enough to cover
				),
			)

			s = &SimpleDAO{
				membStore: NewMembStore(), // empty member store
				propStore: NewPropStore(),
			}
		)

		// Set the sent coins to enough
		// to cover the fee
		std.TestSetOrigSend(sentCoins, std.Coins{})
		std.TestSetOrigCaller(proposer)

		// Make sure the proposal was added
		id, err := s.Propose(description, ex)
		uassert.NoError(t, err)
		uassert.False(t, called)

		// Make sure the proposal exists
		prop, err := s.propStore.GetProposalByID(id)
		uassert.NoError(t, err)

		uassert.Equal(t, proposer.String(), prop.GetAuthor().String())
		uassert.Equal(t, description, prop.GetDescription())
		uassert.Equal(t, dao.Active.String(), prop.GetStatus().String())
		uassert.Equal(t, 0, len(prop.GetVotes(0, maxRequestVotes)))
	})
}

func TestSimpleDAO_VoteOnProposal(t *testing.T) {
	t.Parallel()

	t.Run("not govdao member", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			s = &SimpleDAO{
				membStore: NewMembStore(),
			}
		)

		std.TestSetOrigCaller(voter)

		// Attempt to vote on the proposal
		uassert.ErrorContains(
			t,
			s.VoteOnProposal(0, dao.YesVote),
			errMissingMember.Error(),
		)
	})

	t.Run("missing proposal", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			s = &SimpleDAO{
				membStore: NewMembStore(
					WithInitialMembers([]dao.Member{
						{Address: voter},
					}),
				),
				propStore: NewPropStore(),
			}
		)

		std.TestSetOrigCaller(voter)

		// Attempt to vote on the proposal
		uassert.ErrorContains(
			t,
			s.VoteOnProposal(0, dao.YesVote),
			errMissingProposal.Error(),
		)
	})

	t.Run("proposal executed", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			s = &SimpleDAO{
				membStore: NewMembStore(
					WithInitialMembers([]dao.Member{
						{Address: voter},
					}),
				),
				propStore: NewPropStore(),
			}

			prop = &proposal{
				status: dao.ExecutionSuccessful,
			}
		)

		std.TestSetOrigCaller(voter)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorIs(
			t,
			s.VoteOnProposal(id, dao.YesVote),
			errProposalInactive,
		)
	})

	t.Run("proposal expired", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			s = &SimpleDAO{
				membStore: NewMembStore(
					WithInitialMembers([]dao.Member{
						{Address: voter},
					}),
				),
				propStore: NewPropStore(),
			}

			mockExecutor = &mockExecutor{
				isExpiredFn: func() bool {
					return true
				},
			}

			prop = &proposal{
				status:   dao.Active,
				executor: mockExecutor,
			}
		)

		std.TestSetOrigCaller(voter)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorIs(
			t,
			s.VoteOnProposal(id, dao.YesVote),
			errExecutorExpired,
		)
	})

	t.Run("double vote on proposal", func(t *testing.T) {
		t.Parallel()

		var (
			voter  = testutils.TestAddress("voter")
			member = dao.Member{
				Address:     voter,
				VotingPower: 10,
			}

			s = &SimpleDAO{
				membStore: NewMembStore(
					WithInitialMembers([]dao.Member{
						member,
					}),
				),
				propStore: NewPropStore(),
			}

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				votes:    newVotes(),
			}
		)

		std.TestSetOrigCaller(voter)

		// Cast the initial vote
		urequire.NoError(t, prop.votes.castVote(member, dao.YesVote))

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorContains(
			t,
			s.VoteOnProposal(id, dao.YesVote),
			errAlreadyVoted.Error(),
		)
	})

	t.Run("majority accepted", func(t *testing.T) {
		t.Parallel()

		var (
			members = generateMembers(t, 50)

			s = &SimpleDAO{
				membStore: NewMembStore(WithInitialMembers(members)),
				propStore: NewPropStore(),
			}

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				votes:    newVotes(),
			}
		)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		majorityIndex := (len(members)*2)/3 + 1 // 2/3+
		for _, m := range members[:majorityIndex] {
			std.TestSetOrigCaller(m.Address)

			// Attempt to vote on the proposal
			urequire.NoError(
				t,
				s.VoteOnProposal(id, dao.YesVote),
			)
		}

		// Make sure the proposal was accepted
		uassert.Equal(t, dao.Accepted.String(), prop.status.String())
	})

	t.Run("majority rejected", func(t *testing.T) {
		t.Parallel()

		var (
			members = generateMembers(t, 50)

			s = &SimpleDAO{
				membStore: NewMembStore(WithInitialMembers(members)),
				propStore: NewPropStore(),
			}

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				votes:    newVotes(),
			}
		)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		majorityIndex := (len(members)*2)/3 + 1 // 2/3+
		for _, m := range members[:majorityIndex] {
			std.TestSetOrigCaller(m.Address)

			// Attempt to vote on the proposal
			urequire.NoError(
				t,
				s.VoteOnProposal(id, dao.NoVote),
			)
		}

		// Make sure the proposal was not accepted
		uassert.Equal(t, dao.NotAccepted.String(), prop.status.String())
	})

	t.Run("majority abstained", func(t *testing.T) {
		t.Parallel()

		var (
			members = generateMembers(t, 50)

			s = &SimpleDAO{
				membStore: NewMembStore(WithInitialMembers(members)),
				propStore: NewPropStore(),
			}

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				votes:    newVotes(),
			}
		)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		majorityIndex := (len(members)*2)/3 + 1 // 2/3+
		for _, m := range members[:majorityIndex] {
			std.TestSetOrigCaller(m.Address)

			// Attempt to vote on the proposal
			urequire.NoError(
				t,
				s.VoteOnProposal(id, dao.AbstainVote),
			)
		}

		// Make sure the proposal was not accepted
		uassert.Equal(t, dao.NotAccepted.String(), prop.status.String())
	})

	t.Run("everyone voted, undecided", func(t *testing.T) {
		t.Parallel()

		var (
			members = generateMembers(t, 50)

			s = &SimpleDAO{
				membStore: NewMembStore(WithInitialMembers(members)),
				propStore: NewPropStore(),
			}

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				votes:    newVotes(),
			}
		)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// The first half votes yes
		for _, m := range members[:len(members)/2] {
			std.TestSetOrigCaller(m.Address)

			// Attempt to vote on the proposal
			urequire.NoError(
				t,
				s.VoteOnProposal(id, dao.YesVote),
			)
		}

		// The other half votes no
		for _, m := range members[len(members)/2:] {
			std.TestSetOrigCaller(m.Address)

			// Attempt to vote on the proposal
			urequire.NoError(
				t,
				s.VoteOnProposal(id, dao.NoVote),
			)
		}

		// Make sure the proposal is not active,
		// since everyone voted, and it was undecided
		uassert.Equal(t, dao.NotAccepted.String(), prop.status.String())
	})

	t.Run("proposal undecided", func(t *testing.T) {
		t.Parallel()

		var (
			members = generateMembers(t, 50)

			s = &SimpleDAO{
				membStore: NewMembStore(WithInitialMembers(members)),
				propStore: NewPropStore(),
			}

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				votes:    newVotes(),
			}
		)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// The first quarter votes yes
		for _, m := range members[:len(members)/4] {
			std.TestSetOrigCaller(m.Address)

			// Attempt to vote on the proposal
			urequire.NoError(
				t,
				s.VoteOnProposal(id, dao.YesVote),
			)
		}

		// The second quarter votes no
		for _, m := range members[len(members)/4 : len(members)/2] {
			std.TestSetOrigCaller(m.Address)

			// Attempt to vote on the proposal
			urequire.NoError(
				t,
				s.VoteOnProposal(id, dao.NoVote),
			)
		}

		// Make sure the proposal is still active,
		// since there wasn't quorum reached on any decision
		uassert.Equal(t, dao.Active.String(), prop.status.String())
	})
}

func TestSimpleDAO_ExecuteProposal(t *testing.T) {
	t.Parallel()

	t.Run("caller cannot cover fee", func(t *testing.T) {
		t.Parallel()

		var (
			sentCoins = std.NewCoins(
				std.NewCoin(
					"ugnot",
					minExecuteFeeValue-1,
				),
			)

			s = &SimpleDAO{
				membStore: NewMembStore(), // empty member store
			}
		)

		// Set the sent coins to be lower
		// than the execute fee
		std.TestSetOrigSend(sentCoins, std.Coins{})

		uassert.ErrorIs(
			t,
			s.ExecuteProposal(0),
			errInsufficientExecuteFunds,
		)
	})

	t.Run("missing proposal", func(t *testing.T) {
		t.Parallel()

		var (
			sentCoins = std.NewCoins(
				std.NewCoin(
					"ugnot",
					minExecuteFeeValue,
				),
			)

			s = &SimpleDAO{
				membStore: NewMembStore(), // empty member store
				propStore: NewPropStore(), // empty prop store
			}
		)

		// Set the sent coins to be enough
		// so the execution can take place
		std.TestSetOrigSend(sentCoins, std.Coins{})

		uassert.ErrorContains(
			t,
			s.ExecuteProposal(0),
			errMissingProposal.Error(),
		)
	})

	t.Run("proposal not accepted", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			s = &SimpleDAO{
				membStore: NewMembStore(
					WithInitialMembers([]dao.Member{
						{Address: voter},
					}),
				),
				propStore: NewPropStore(),
			}

			prop = &proposal{
				status: dao.NotAccepted,
			}
		)

		std.TestSetOrigCaller(voter)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorIs(
			t,
			s.ExecuteProposal(id),
			errProposalNotAccepted,
		)
	})

	t.Run("proposal already executed", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name   string
			status dao.ProposalStatus
		}{
			{
				"execution was successful",
				dao.ExecutionSuccessful,
			},
			{
				"execution failed",
				dao.ExecutionFailed,
			},
		}

		for _, testCase := range testTable {
			testCase := testCase

			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				var (
					voter = testutils.TestAddress("voter")

					s = &SimpleDAO{
						membStore: NewMembStore(
							WithInitialMembers([]dao.Member{
								{Address: voter},
							}),
						),
						propStore: NewPropStore(),
					}

					prop = &proposal{
						status: testCase.status,
					}
				)

				std.TestSetOrigCaller(voter)

				// Add an initial proposal
				id, err := s.propStore.addProposal(prop)
				urequire.NoError(t, err)

				// Attempt to vote on the proposal
				uassert.ErrorIs(
					t,
					s.ExecuteProposal(id),
					errProposalExecuted,
				)
			})
		}
	})

	t.Run("proposal expired", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			s = &SimpleDAO{
				membStore: NewMembStore(
					WithInitialMembers([]dao.Member{
						{Address: voter},
					}),
				),
				propStore: NewPropStore(),
			}

			mockExecutor = &mockExecutor{
				isExpiredFn: func() bool {
					return true
				},
			}

			prop = &proposal{
				status:   dao.Accepted,
				executor: mockExecutor,
			}
		)

		std.TestSetOrigCaller(voter)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorIs(
			t,
			s.ExecuteProposal(id),
			errExecutorExpired,
		)
	})

	t.Run("execution error", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			s = &SimpleDAO{
				membStore: NewMembStore(
					WithInitialMembers([]dao.Member{
						{Address: voter},
					}),
				),
				propStore: NewPropStore(),
			}

			execError = errors.New("exec error")

			mockExecutor = &mockExecutor{
				executeFn: func() error {
					return execError
				},
			}

			prop = &proposal{
				status:   dao.Accepted,
				executor: mockExecutor,
			}
		)

		std.TestSetOrigCaller(voter)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorContains(
			t,
			s.ExecuteProposal(id),
			execError.Error(),
		)

		uassert.Equal(t, dao.ExecutionFailed.String(), prop.status.String())
	})

	t.Run("successful execution", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			s = &SimpleDAO{
				membStore: NewMembStore(
					WithInitialMembers([]dao.Member{
						{Address: voter},
					}),
				),
				propStore: NewPropStore(),
			}

			called       = false
			mockExecutor = &mockExecutor{
				executeFn: func() error {
					called = true

					return nil
				},
			}

			prop = &proposal{
				status:   dao.Accepted,
				executor: mockExecutor,
			}
		)

		std.TestSetOrigCaller(voter)

		// Add an initial proposal
		id, err := s.propStore.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.NoError(t, s.ExecuteProposal(id))
		uassert.Equal(t, dao.ExecutionSuccessful.String(), prop.status.String())
		uassert.True(t, called)
	})
}
