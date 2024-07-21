package simpledao

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

// generateMembers generates dummy govdao members
func generateMembers(t *testing.T, count int) []membstore.Member {
	t.Helper()

	members := make([]membstore.Member, 0, count)

	for i := 0; i < count; i++ {
		members = append(members, membstore.Member{
			Address:     testutils.TestAddress(ufmt.Sprintf("member %d", i)),
			VotingPower: 10,
		})
	}

	return members
}

func TestSimpleDAO_Propose(t *testing.T) {
	t.Parallel()

	t.Run("invalid executor", func(t *testing.T) {
		t.Parallel()

		s := New(nil)

		_, err := s.Propose(dao.ProposalRequest{})
		uassert.ErrorIs(
			t,
			err,
			ErrInvalidExecutor,
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
			ex = &mockExecutor{
				executeFn: cb,
			}

			sentCoins = std.NewCoins(
				std.NewCoin(
					"ugnot",
					minProposalFeeValue-1,
				),
			)

			ms = &mockMemberStore{
				isMemberFn: func(_ std.Address) bool {
					return false
				},
			}
			s = New(ms)
		)

		// Set the sent coins to be lower
		// than the proposal fee
		std.TestSetOrigSend(sentCoins, std.Coins{})

		_, err := s.Propose(dao.ProposalRequest{
			Executor: ex,
		})
		uassert.ErrorIs(
			t,
			err,
			ErrInsufficientProposalFunds,
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

			ex = &mockExecutor{
				executeFn: cb,
			}
			description = "Proposal description"

			proposer  = testutils.TestAddress("proposer")
			sentCoins = std.NewCoins(
				std.NewCoin(
					"ugnot",
					minProposalFeeValue, // enough to cover
				),
			)

			ms = &mockMemberStore{
				isMemberFn: func(addr std.Address) bool {
					return addr == proposer
				},
			}
			s = New(ms)
		)

		// Set the sent coins to be enough
		// to cover the fee
		std.TestSetOrigSend(sentCoins, std.Coins{})
		std.TestSetOrigCaller(proposer)

		// Make sure the proposal was added
		id, err := s.Propose(dao.ProposalRequest{
			Description: description,
			Executor:    ex,
		})
		uassert.NoError(t, err)
		uassert.False(t, called)

		// Make sure the proposal exists
		prop, err := s.ProposalByID(id)
		uassert.NoError(t, err)

		uassert.Equal(t, proposer.String(), prop.Author().String())
		uassert.Equal(t, description, prop.Description())
		uassert.Equal(t, dao.Active.String(), prop.Status().String())

		stats := prop.VotingStats()

		uassert.Equal(t, uint64(0), stats.YayVotes)
		uassert.Equal(t, uint64(0), stats.NayVotes)
		uassert.Equal(t, uint64(0), stats.AbstainVotes)
		uassert.Equal(t, uint64(0), stats.TotalVotingPower)
	})
}

func TestSimpleDAO_VoteOnProposal(t *testing.T) {
	t.Parallel()

	t.Run("not govdao member", func(t *testing.T) {
		t.Parallel()

		var (
			voter    = testutils.TestAddress("voter")
			fetchErr = errors.New("fetch error")

			ms = &mockMemberStore{
				memberFn: func(_ std.Address) (membstore.Member, error) {
					return membstore.Member{
						Address: voter,
					}, fetchErr
				},
			}
			s = New(ms)
		)

		std.TestSetOrigCaller(voter)

		// Attempt to vote on the proposal
		uassert.ErrorContains(
			t,
			s.VoteOnProposal(0, dao.YesVote),
			fetchErr.Error(),
		)
	})

	t.Run("missing proposal", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")
			ms    = &mockMemberStore{
				memberFn: func(a std.Address) (membstore.Member, error) {
					if a != voter {
						return membstore.Member{}, errors.New("not found")
					}

					return membstore.Member{
						Address: voter,
					}, nil
				},
			}

			s = New(ms)
		)

		std.TestSetOrigCaller(voter)

		// Attempt to vote on the proposal
		uassert.ErrorContains(
			t,
			s.VoteOnProposal(0, dao.YesVote),
			ErrMissingProposal.Error(),
		)
	})

	t.Run("proposal executed", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			ms = &mockMemberStore{
				memberFn: func(a std.Address) (membstore.Member, error) {
					if a != voter {
						return membstore.Member{}, errors.New("not found")
					}

					return membstore.Member{
						Address: voter,
					}, nil
				},
			}
			s = New(ms)

			prop = &proposal{
				status: dao.ExecutionSuccessful,
			}
		)

		std.TestSetOrigCaller(voter)

		// Add an initial proposal
		id, err := s.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorIs(
			t,
			s.VoteOnProposal(id, dao.YesVote),
			ErrProposalInactive,
		)
	})

	t.Run("proposal expired", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			ms = &mockMemberStore{
				memberFn: func(a std.Address) (membstore.Member, error) {
					if a != voter {
						return membstore.Member{}, errors.New("not found")
					}

					return membstore.Member{
						Address: voter,
					}, nil
				},
			}
			s = New(ms)

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
		id, err := s.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorIs(
			t,
			s.VoteOnProposal(id, dao.YesVote),
			ErrExecutorExpired,
		)
	})

	t.Run("double vote on proposal", func(t *testing.T) {
		t.Parallel()

		var (
			voter  = testutils.TestAddress("voter")
			member = membstore.Member{
				Address:     voter,
				VotingPower: 10,
			}

			ms = &mockMemberStore{
				memberFn: func(a std.Address) (membstore.Member, error) {
					if a != voter {
						return membstore.Member{}, errors.New("not found")
					}

					return member, nil
				},
			}
			s = New(ms)

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				tally:    newTally(),
			}
		)

		std.TestSetOrigCaller(voter)

		// Cast the initial vote
		urequire.NoError(t, prop.tally.castVote(member, dao.YesVote))

		// Add an initial proposal
		id, err := s.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorContains(
			t,
			s.VoteOnProposal(id, dao.YesVote),
			ErrAlreadyVoted.Error(),
		)
	})

	t.Run("majority accepted", func(t *testing.T) {
		t.Parallel()

		var (
			members = generateMembers(t, 50)

			ms = &mockMemberStore{
				memberFn: func(address std.Address) (membstore.Member, error) {
					for _, m := range members {
						if m.Address == address {
							return m, nil
						}
					}

					return membstore.Member{}, errors.New("not found")
				},

				totalPowerFn: func() uint64 {
					power := uint64(0)

					for _, m := range members {
						power += m.VotingPower
					}

					return power
				},
			}
			s = New(ms)

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				tally:    newTally(),
			}
		)

		// Add an initial proposal
		id, err := s.addProposal(prop)
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

			ms = &mockMemberStore{
				memberFn: func(address std.Address) (membstore.Member, error) {
					for _, m := range members {
						if m.Address == address {
							return m, nil
						}
					}

					return membstore.Member{}, errors.New("member not found")
				},

				totalPowerFn: func() uint64 {
					power := uint64(0)

					for _, m := range members {
						power += m.VotingPower
					}

					return power
				},
			}
			s = New(ms)

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				tally:    newTally(),
			}
		)

		// Add an initial proposal
		id, err := s.addProposal(prop)
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

			ms = &mockMemberStore{
				memberFn: func(address std.Address) (membstore.Member, error) {
					for _, m := range members {
						if m.Address == address {
							return m, nil
						}
					}

					return membstore.Member{}, errors.New("member not found")
				},

				totalPowerFn: func() uint64 {
					power := uint64(0)

					for _, m := range members {
						power += m.VotingPower
					}

					return power
				},
			}
			s = New(ms)

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				tally:    newTally(),
			}
		)

		// Add an initial proposal
		id, err := s.addProposal(prop)
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

			ms = &mockMemberStore{
				memberFn: func(address std.Address) (membstore.Member, error) {
					for _, m := range members {
						if m.Address == address {
							return m, nil
						}
					}

					return membstore.Member{}, errors.New("member not found")
				},

				totalPowerFn: func() uint64 {
					power := uint64(0)

					for _, m := range members {
						power += m.VotingPower
					}

					return power
				},
			}
			s = New(ms)

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				tally:    newTally(),
			}
		)

		// Add an initial proposal
		id, err := s.addProposal(prop)
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

			ms = &mockMemberStore{
				memberFn: func(address std.Address) (membstore.Member, error) {
					for _, m := range members {
						if m.Address == address {
							return m, nil
						}
					}

					return membstore.Member{}, errors.New("member not found")
				},

				totalPowerFn: func() uint64 {
					power := uint64(0)

					for _, m := range members {
						power += m.VotingPower
					}

					return power
				},
			}
			s = New(ms)

			prop = &proposal{
				status:   dao.Active,
				executor: &mockExecutor{},
				tally:    newTally(),
			}
		)

		// Add an initial proposal
		id, err := s.addProposal(prop)
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

			ms = &mockMemberStore{
				isMemberFn: func(_ std.Address) bool {
					return false
				},
			}
			s = New(ms)
		)

		// Set the sent coins to be lower
		// than the execute fee
		std.TestSetOrigSend(sentCoins, std.Coins{})

		uassert.ErrorIs(
			t,
			s.ExecuteProposal(0),
			ErrInsufficientExecuteFunds,
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

			ms = &mockMemberStore{
				isMemberFn: func(_ std.Address) bool {
					return true
				},
			}

			s = New(ms)
		)

		// Set the sent coins to be enough
		// so the execution can take place
		std.TestSetOrigSend(sentCoins, std.Coins{})

		uassert.ErrorContains(
			t,
			s.ExecuteProposal(0),
			ErrMissingProposal.Error(),
		)
	})

	t.Run("proposal not accepted", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			ms = &mockMemberStore{
				isMemberFn: func(_ std.Address) bool {
					return true
				},
			}
			s = New(ms)

			prop = &proposal{
				status: dao.NotAccepted,
			}
		)

		std.TestSetOrigCaller(voter)

		// Add an initial proposal
		id, err := s.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorIs(
			t,
			s.ExecuteProposal(id),
			ErrProposalNotAccepted,
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

					ms = &mockMemberStore{
						isMemberFn: func(_ std.Address) bool {
							return true
						},
					}
					s = New(ms)

					prop = &proposal{
						status: testCase.status,
					}
				)

				std.TestSetOrigCaller(voter)

				// Add an initial proposal
				id, err := s.addProposal(prop)
				urequire.NoError(t, err)

				// Attempt to vote on the proposal
				uassert.ErrorIs(
					t,
					s.ExecuteProposal(id),
					ErrProposalExecuted,
				)
			})
		}
	})

	t.Run("proposal expired", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			ms = &mockMemberStore{
				isMemberFn: func(_ std.Address) bool {
					return true
				},
			}

			s = New(ms)

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
		id, err := s.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.ErrorIs(
			t,
			s.ExecuteProposal(id),
			ErrExecutorExpired,
		)
	})

	t.Run("execution error", func(t *testing.T) {
		t.Parallel()

		var (
			voter = testutils.TestAddress("voter")

			ms = &mockMemberStore{
				isMemberFn: func(_ std.Address) bool {
					return true
				},
			}

			s = New(ms)

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
		id, err := s.addProposal(prop)
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

			ms = &mockMemberStore{
				isMemberFn: func(_ std.Address) bool {
					return true
				},
			}
			s = New(ms)

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
		id, err := s.addProposal(prop)
		urequire.NoError(t, err)

		// Attempt to vote on the proposal
		uassert.NoError(t, s.ExecuteProposal(id))
		uassert.Equal(t, dao.ExecutionSuccessful.String(), prop.status.String())
		uassert.True(t, called)
	})
}
