package governance

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	u256 "gno.land/p/gnoswap/uint256"
)

func TestVote(t *testing.T) {
	baseTime := uint64(time.Now().Unix())
	voter := testutils.TestAddress("voter")

	newConfig := Config{
		VotingStartDelay: 50,
		VotingPeriod:     100,
	}
	setConfigVersion(1, newConfig)

	testCases := []struct {
		name        string
		setup       func() uint64
		pid         uint64
		yes         bool
		expectError bool
		validate    func(t *testing.T, voteKey string)
	}{
		{
			name: "Successful YES vote",
			setup: func() uint64 {
				pid := uint64(1)
				proposals.Set(formatUint(pid), ProposalInfo{
					ConfigVersion: 1,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime - 60, // Voting period started
					},
					Yea: u256.NewUint(0),
					Nay: u256.NewUint(0),
					PossibleAddressWithWeight: map[std.Address]uint64{
						voter: 100,
					},
				})
				return pid
			},
			yes:         true,
			expectError: false,
			validate: func(t *testing.T, voteKey string) {
				prop, exists := proposals.Get(formatUint(1))
				if !exists {
					t.Error("Proposal was not stored")
					return
				}
				proposal := prop.(ProposalInfo)
				if proposal.Yea.Cmp(u256.NewUint(100)) != 0 {
					t.Errorf("Expected Yea votes to be 100, got %v", proposal.Yea)
				}
				if proposal.Nay.Cmp(u256.NewUint(0)) != 0 {
					t.Errorf("Expected Nay votes to be 0, got %v", proposal.Nay)
				}

				value, exists := votes.Get(voteKey)
				if !exists || !value.(bool) {
					t.Error("Vote record not properly stored")
				}

				// Verify user vote record
				userVote, exists := userVotes.Get(voter.String())
				if !exists {
					t.Error("User vote record not properly stored")
				}

				weight := userVote.(voteWithWeight)
				uassert.True(t, weight.Yes)
				uassert.Equal(t, weight.Weight, uint64(100))
			},
		},
		{
			name: "Successful NO vote",
			setup: func() uint64 {
				pid := uint64(2)
				proposals.Set(formatUint(pid), ProposalInfo{
					ConfigVersion: 1,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime - 60,
					},
					Yea: u256.NewUint(0),
					Nay: u256.NewUint(0),
					PossibleAddressWithWeight: map[std.Address]uint64{
						voter: 100,
					},
				})
				return pid
			},
			yes:         false,
			expectError: false,
			validate: func(t *testing.T, voteKey string) {
				prop, exists := proposals.Get(formatUint(2))
				if !exists {
					t.Error("Proposal was not stored")
					return
				}
				proposal := prop.(ProposalInfo)
				uassert.Equal(t, proposal.Yea.ToString(), "0")
				uassert.Equal(t, proposal.Nay.ToString(), "100")
			},
		},
		{
			name: "Vote before voting period starts",
			setup: func() uint64 {
				pid := uint64(3)
				proposals.Set(formatUint(pid), ProposalInfo{
					ConfigVersion: 1,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime, // Just created, voting hasn't started
					},
					PossibleAddressWithWeight: map[std.Address]uint64{
						voter: 100,
					},
				})
				return pid
			},
			expectError: true,
		},
		{
			name: "Vote on canceled proposal",
			setup: func() uint64 {
				pid := uint64(5)
				proposals.Set(formatUint(pid), ProposalInfo{
					ConfigVersion: 1,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime - 60,
						Canceled:  true,
					},
					PossibleAddressWithWeight: map[std.Address]uint64{
						voter: 100,
					},
				})
				return pid
			},
			expectError: true,
		},
		{
			name: "Double voting attempt",
			setup: func() uint64 {
				pid := uint64(6)
				proposals.Set(formatUint(pid), ProposalInfo{
					ConfigVersion: 1,
					State: ProposalState{
						Created:   true,
						CreatedAt: baseTime - 60,
					},
					PossibleAddressWithWeight: map[std.Address]uint64{
						voter: 100,
					},
				})
				// Pre-existing vote
				voteKey := createVoteKey(pid, voter.String())
				votes.Set(voteKey, true)
				return pid
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetVoteEnv(t)

			pid := tc.setup()

			var voteKey string
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = r.(error)
					}
				}()
				voteKey = Vote(pid, tc.yes)
			}()

			if tc.expectError {
				uassert.Error(t, err)
				return
			}
		})
	}
}

func TestNewVoteState(t *testing.T) {
	voter := testutils.TestAddress("voter")
	voteKey := createVoteKey(1, voter.String())

	// test config
	// configVersions = map[uint64]Config{
	// 	1: {
	// 		VotingStartDelay: 50,
	// 		VotingPeriod:     100,
	// 	},
	// }
	newConfig := Config{
		VotingStartDelay: 50,
		VotingPeriod:     100,
	}
	setConfigVersion(1, newConfig)

	tests := []struct {
		name          string
		proposal      ProposalInfo
		now           uint64
		expectedState *VoteState
		expectError   bool
	}{
		{
			name: "Valid propoal: in voting period",
			proposal: ProposalInfo{
				ConfigVersion: 1,
				State: ProposalState{
					// voting start: 900 + 50 = 950
					// voting end: 950 + 100 = 1050
					CreatedAt: 900,
					Canceled:  false,
				},
				PossibleAddressWithWeight: map[std.Address]uint64{
					voter: 100,
				},
			},
			now: 1000,
			expectedState: &VoteState{
				isVotingPeriod:  true,
				isProposalValid: true,
				hasUserVoted:    false,
				userWeight:      100,
			},
			expectError: false,
		},
		{
			name: "Before voting start",
			proposal: ProposalInfo{
				ConfigVersion: 1,
				State: ProposalState{
					CreatedAt: 900,
					Canceled:  false,
				},
				PossibleAddressWithWeight: map[std.Address]uint64{
					voter: 100,
				},
			},
			now: 940,
			expectedState: &VoteState{
				isVotingPeriod:  false,
				isProposalValid: true,
				hasUserVoted:    false,
				userWeight:      100,
			},
			expectError: false,
		},
		{
			name: "After voting end",
			proposal: ProposalInfo{
				ConfigVersion: 1,
				State: ProposalState{
					CreatedAt: 900,
					Canceled:  false,
				},
				PossibleAddressWithWeight: map[std.Address]uint64{
					voter: 100,
				},
			},
			now: 1060,
			expectedState: &VoteState{
				isVotingPeriod:  false,
				isProposalValid: true,
				hasUserVoted:    false,
				userWeight:      100,
			},
			expectError: false,
		},
		{
			name: "Canceled proposal",
			proposal: ProposalInfo{
				ConfigVersion: 1,
				State: ProposalState{
					CreatedAt: 900,
					Canceled:  true,
				},
				PossibleAddressWithWeight: map[std.Address]uint64{
					voter: 100,
				},
			},
			now: 1000,
			expectedState: &VoteState{
				isVotingPeriod:  true,
				isProposalValid: false,
				hasUserVoted:    false,
				userWeight:      100,
			},
			expectError: false,
		},
		{
			name: "No voting permmisions",
			proposal: ProposalInfo{
				ConfigVersion: 1,
				State: ProposalState{
					CreatedAt: 900,
					Canceled:  false,
				},
				PossibleAddressWithWeight: map[std.Address]uint64{},
			},
			now: 1000,
			expectedState: &VoteState{
				isVotingPeriod:  true,
				isProposalValid: true,
				hasUserVoted:    false,
				userWeight:      0,
			},
			expectError: false,
		},
		{
			name: "Already voted",
			proposal: ProposalInfo{
				ConfigVersion: 1,
				State: ProposalState{
					CreatedAt: 900,
					Canceled:  false,
				},
				PossibleAddressWithWeight: map[std.Address]uint64{
					voter: 100,
				},
			},
			now: 1000,
			expectedState: &VoteState{
				isVotingPeriod:  true,
				isProposalValid: true,
				hasUserVoted:    true,
				userWeight:      100,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedState.hasUserVoted {
				votes.Set(voteKey, true)
			} else {
				votes.Remove(voteKey)
			}

			state, err := newVoteState(tt.proposal, voteKey, voter, tt.now)

			if tt.expectError {
				uassert.Error(t, err)
			}

			if err != nil {
				uassert.NoError(t, err)
			}

			uassert.Equal(t, state.isVotingPeriod, tt.expectedState.isVotingPeriod)
			uassert.Equal(t, state.isProposalValid, tt.expectedState.isProposalValid)
			uassert.Equal(t, state.hasUserVoted, tt.expectedState.hasUserVoted)
			uassert.Equal(t, state.userWeight, tt.expectedState.userWeight)
		})
	}
}

func TestNewVoteExecutor(t *testing.T) {
	voter := testutils.TestAddress("voter")
	proposal := &ProposalInfo{
		Yea: u256.NewUint(0),
		Nay: u256.NewUint(0),
	}
	weight := uint64(100)

	executor := newVoteExecutor(proposal, voter, weight)

	if executor == nil {
		t.Fatal("Expected non-nil VoteExecutor")
	}

	if executor.proposal != proposal {
		t.Error("Proposal reference mismatch")
	}

	uassert.Equal(t, executor.voter, voter)
	uassert.Equal(t, executor.weight, weight)
}

func TestVoteExecutor_Execute(t *testing.T) {
	tests := []struct {
		name        string
		yes         bool
		initYea     uint64
		initNay     uint64
		weight      uint64
		expectYea   uint64
		expectNay   uint64
		expectError bool
	}{
		{
			name:        "Valid YES vote with zero initial votes",
			yes:         true,
			initYea:     0,
			initNay:     0,
			weight:      100,
			expectYea:   100,
			expectNay:   0,
			expectError: false,
		},
		{
			name:        "Valid NO vote with zero initial votes",
			yes:         false,
			initYea:     0,
			initNay:     0,
			weight:      100,
			expectYea:   0,
			expectNay:   100,
			expectError: false,
		},
		{
			name:        "YES vote with existing votes",
			yes:         true,
			initYea:     150,
			initNay:     50,
			weight:      100,
			expectYea:   250,
			expectNay:   50,
			expectError: false,
		},
		{
			name:        "NO vote with existing votes",
			yes:         false,
			initYea:     150,
			initNay:     50,
			weight:      100,
			expectYea:   150,
			expectNay:   150,
			expectError: false,
		},
		{
			name:        "YES vote with maximum safe weight",
			yes:         true,
			initYea:     0,
			initNay:     0,
			weight:      ^uint64(0), // max uint64
			expectYea:   ^uint64(0),
			expectNay:   0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			voter := testutils.TestAddress("voter")
			proposal := &ProposalInfo{
				Yea: u256.NewUint(tt.initYea),
				Nay: u256.NewUint(tt.initNay),
			}

			executor := newVoteExecutor(proposal, voter, tt.weight)

			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = r.(error)
					}
				}()
				err = executor.execute(tt.yes)
			}()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if proposal.Yea.Cmp(u256.NewUint(tt.expectYea)) != 0 {
				t.Errorf("Yea votes mismatch: got %v, want %v",
					proposal.Yea, tt.expectYea)
			}

			if proposal.Nay.Cmp(u256.NewUint(tt.expectNay)) != 0 {
				t.Errorf("Nay votes mismatch: got %v, want %v",
					proposal.Nay, tt.expectNay)
			}
		})
	}
}

func resetVoteEnv(t *testing.T) {
	t.Helper()
	votes = avl.NewTree()
	userVotes = avl.NewTree()
}
