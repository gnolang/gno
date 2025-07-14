package governance

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

// Test suite for GovernanceState
// These tests verify the immutability guarantees and correct behavior of state transitions

func TestNewGovernanceState(t *testing.T) {
	// Test: Creating a new governance state with initial configuration
	// Expected: State should be initialized with empty collections and provided config

	config := Config{
		VotingStartDelay:              86400,
		VotingPeriod:                  604800,
		VotingWeightSmoothingDuration: 86400,
		Quorum:                        50,
		ProposalCreationThreshold:     1_000_000_000,
		ExecutionDelay:                86400,
		ExecutionWindow:               2592000,
	}

	state := NewGovernanceState(config)

	// Verify initial state
	uassert.Equal(t, state.GetProposalCounter(), uint64(0), "proposal counter should start at 0")

	// Verify config is set correctly
	currentConfig, exists := state.GetConfig(0)
	uassert.True(t, exists, "current config should exist")
	uassert.Equal(t, currentConfig.Quorum, uint64(50), "config quorum should be 50")

	// Verify config version 1 exists
	configV1, exists := state.GetConfig(1)
	uassert.True(t, exists, "config version 1 should exist")
	uassert.Equal(t, configV1.Quorum, uint64(50), "config v1 quorum should be 50")
}

func TestWithNewProposal(t *testing.T) {
	// Test: Adding a new proposal should create new state with incremented counter
	// Expected: New state contains proposal, old state remains unchanged

	state := createTestState()
	proposer := std.Address("g1proposer")

	proposal := ProposalInfo{
		Proposer:     proposer,
		ProposalType: Text,
		State: ProposalState{
			Created:   true,
			CreatedAt: uint64(time.Now().Unix()),
			Upcoming:  true,
		},
		Yea:             u256.Zero(),
		Nay:             u256.Zero(),
		MaxVotingWeight: u256.NewUint(1000000),
		ConfigVersion:   1,
		QuorumAmount:    500000,
		Title:           "Test Proposal",
		Description:     "Test Description",
	}

	// Add proposal
	newState := state.WithNewProposal(proposer, proposal)

	// Verify old state unchanged
	uassert.Equal(t, state.GetProposalCounter(), uint64(0), "old state counter should remain 0")
	_, exists := state.GetProposal(1)
	uassert.False(t, exists, "old state should not have proposal")

	// Verify new state
	uassert.Equal(t, newState.GetProposalCounter(), uint64(1), "new state counter should be 1")
	retrievedProposal, exists := newState.GetProposal(1)
	uassert.True(t, exists, "new state should have proposal")
	uassert.Equal(t, retrievedProposal.Title, "Test Proposal", "proposal title should match")

	// Verify proposer mapping
	proposalId, exists := newState.GetProposalByProposer(proposer)
	uassert.True(t, exists, "proposer should have proposal")
	uassert.Equal(t, proposalId, uint64(1), "proposer should have proposal 1")
}

func TestWithUpdatedProposal(t *testing.T) {
	// Test: Updating a proposal should create new state with modified proposal
	// Expected: Only the specified proposal is updated, others remain unchanged

	state := createTestStateWithProposal()

	// Update proposal to active state
	newState := state.WithUpdatedProposal(1, func(p ProposalInfo) ProposalInfo {
		p.State.Active = true
		p.State.Upcoming = false
		return p
	})

	// Verify old state unchanged
	oldProposal, _ := state.GetProposal(1)
	uassert.True(t, oldProposal.State.Upcoming, "old state should have upcoming true")
	uassert.False(t, oldProposal.State.Active, "old state should have active false")

	// Verify new state
	newProposal, exists := newState.GetProposal(1)
	uassert.True(t, exists, "proposal should exist in new state")
	uassert.False(t, newProposal.State.Upcoming, "new state should have upcoming false")
	uassert.True(t, newProposal.State.Active, "new state should have active true")
}

func TestWithArchivedProposal(t *testing.T) {
	// Test: Archiving a proposal should move it from active to archived
	// Expected: Proposal removed from active, added to archived, proposer mapping cleared

	state := createTestStateWithProposal()
	proposer := std.Address("g1proposer")

	// Archive the proposal
	newState := state.WithArchivedProposal(1)

	// Verify proposal moved
	_, exists := newState.GetProposal(1)
	uassert.False(t, exists, "proposal should not be in active proposals")

	archivedProposal, exists := newState.GetArchivedProposal(1)
	uassert.True(t, exists, "proposal should be in archived proposals")
	uassert.Equal(t, archivedProposal.Title, "Test Proposal", "archived proposal should have correct data")

	// Verify proposer mapping cleared
	_, exists = newState.GetProposalByProposer(proposer)
	uassert.False(t, exists, "proposer should not have active proposal")
}

func TestWithVote(t *testing.T) {
	// Test: Recording a vote should update vote records and proposal counts
	// Expected: Vote recorded, proposal vote counts updated, accumulated totals updated

	state := createTestStateWithProposal()
	voter := std.Address("g1voter")
	proposalId := uint64(1)
	voteWeight := uint64(100)

	// Cast YES vote
	newState := state.WithVote(proposalId, voter, true, voteWeight)

	// Verify vote recorded
	vote, exists := newState.GetVote(proposalId, voter)
	uassert.True(t, exists, "vote should exist")
	uassert.True(t, vote, "vote should be YES")

	// Verify user vote details
	userVote, exists := newState.GetUserVote(voter, proposalId)
	uassert.True(t, exists, "user vote should exist")
	uassert.True(t, userVote.Yes, "user vote should be YES")
	uassert.Equal(t, userVote.Weight, voteWeight, "vote weight should match")

	// Verify accumulated votes
	yesVotes, noVotes := newState.GetAccumulatedVotes(proposalId)
	uassert.Equal(t, yesVotes, voteWeight, "yes votes should equal vote weight")
	uassert.Equal(t, noVotes, uint64(0), "no votes should be 0")

	// Verify proposal updated
	proposal, exists := newState.GetProposal(proposalId)
	uassert.True(t, exists, "proposal should exist")
	uassert.Equal(t, proposal.Yea.Uint64(), voteWeight, "proposal yes votes should be updated")
	uassert.Equal(t, proposal.Nay.Uint64(), uint64(0), "proposal no votes should be 0")

	// Cast NO vote from another voter
	voter2 := std.Address("g1voter2")
	finalState := newState.WithVote(proposalId, voter2, false, voteWeight*2)

	// Verify both votes exist
	vote2, exists := finalState.GetVote(proposalId, voter2)
	uassert.True(t, exists, "vote2 should exist")
	uassert.False(t, vote2, "vote2 should be NO")

	// Verify accumulated totals
	yesVotes, noVotes = finalState.GetAccumulatedVotes(proposalId)
	uassert.Equal(t, yesVotes, voteWeight, "yes votes should remain same")
	uassert.Equal(t, noVotes, voteWeight*2, "no votes should be double weight")
}

func TestWithNewConfig(t *testing.T) {
	// Test: Adding new configuration version
	// Expected: New config becomes current, version counter increments

	state := createTestState()

	newConfig := Config{
		VotingStartDelay:              172800,  // 2 days
		VotingPeriod:                  1209600, // 14 days
		VotingWeightSmoothingDuration: 172800,
		Quorum:                        60, // 60%
		ProposalCreationThreshold:     2_000_000_000,
		ExecutionDelay:                172800,
		ExecutionWindow:               5184000, // 60 days
	}

	newState := state.WithNewConfig(newConfig)

	// Verify new config is current
	currentConfig, _ := newState.GetConfig(0)
	uassert.Equal(t, currentConfig.Quorum, uint64(60), "current config should have new quorum")

	// Verify new version exists
	configV2, exists := newState.GetConfig(2)
	uassert.True(t, exists, "config version 2 should exist")
	uassert.Equal(t, configV2.Quorum, uint64(60), "config v2 should have new quorum")

	// Verify old version still exists
	configV1, exists := newState.GetConfig(1)
	uassert.True(t, exists, "config version 1 should still exist")
	uassert.Equal(t, configV1.Quorum, uint64(50), "config v1 should have old quorum")
}

func TestCanPropose(t *testing.T) {
	// Test: Checking if address can propose
	// Expected: Can propose if no active proposal, cannot if has active proposal

	state := createTestState()
	proposer := std.Address("g1proposer")

	// Should be able to propose initially
	uassert.True(t, state.CanPropose(proposer), "should be able to propose initially")

	// Add proposal
	proposal := createTestProposal2(proposer)
	newState := state.WithNewProposal(proposer, proposal)

	// Should not be able to propose with active proposal
	uassert.False(t, newState.CanPropose(proposer), "should not be able to propose with active proposal")

	// Other addresses should still be able to propose
	otherProposer := std.Address("g1other")
	uassert.True(t, newState.CanPropose(otherProposer), "other address should be able to propose")
}

func TestHasVoted(t *testing.T) {
	// Test: Checking if address has voted
	// Expected: Returns true only if vote exists

	state := createTestStateWithProposal()
	voter := std.Address("g1voter")
	proposalId := uint64(1)

	// Should not have voted initially
	uassert.False(t, state.HasVoted(proposalId, voter), "should not have voted initially")

	// Cast vote
	newState := state.WithVote(proposalId, voter, true, 100)

	// Should have voted after casting
	uassert.True(t, newState.HasVoted(proposalId, voter), "should have voted after casting")

	// Other voter should not have voted
	otherVoter := std.Address("g1other")
	uassert.False(t, newState.HasVoted(proposalId, otherVoter), "other voter should not have voted")
}

func TestIterateProposals(t *testing.T) {
	// Test: Iterating over active proposals
	// Expected: All active proposals visited, archived proposals not included

	state := createTestState()

	// Add multiple proposals
	for i := 1; i <= 3; i++ {
		proposer := std.Address("g1proposer" + formatUint(uint64(i)))
		proposal := createTestProposal2(proposer)
		proposal.Title = "Proposal " + formatUint(uint64(i))
		state = state.WithNewProposal(proposer, proposal)
	}

	// Archive proposal 2
	state = state.WithArchivedProposal(2)

	// Iterate and collect
	var collectedIds []uint64
	var collectedTitles []string

	state.IterateProposals(func(id uint64, p ProposalInfo) bool {
		collectedIds = append(collectedIds, id)
		collectedTitles = append(collectedTitles, p.Title)
		return false // continue iteration
	})

	// Verify only active proposals collected
	uassert.Equal(t, len(collectedIds), 2, "should have 2 active proposals")
	uassert.True(t, containsId(collectedIds, 1), "should contain proposal 1")
	uassert.True(t, containsId(collectedIds, 3), "should contain proposal 3")
	uassert.False(t, containsId(collectedIds, 2), "should not contain archived proposal 2")
}

func TestIterateUserVotes(t *testing.T) {
	// Test: Iterating over user's votes
	// Expected: All votes by specific user returned

	state := createTestState()
	voter := std.Address("g1voter")

	// Create multiple proposals and vote on them
	for i := uint64(1); i <= 3; i++ {
		proposer := std.Address("g1proposer" + formatUint(i))
		proposal := createTestProposal2(proposer)
		state = state.WithNewProposal(proposer, proposal)

		// Vote yes on odd, no on even
		voteYes := i%2 == 1
		state = state.WithVote(i, voter, voteYes, i*100)
	}

	// Add vote from different voter (should not be included)
	otherVoter := std.Address("g1other")
	state = state.WithVote(1, otherVoter, false, 50)

	// Iterate and collect
	var collectedProposals []uint64
	var collectedWeights []uint64

	state.IterateUserVotes(voter, func(proposalId uint64, vote voteWithWeight) bool {
		collectedProposals = append(collectedProposals, proposalId)
		collectedWeights = append(collectedWeights, vote.Weight)
		return false
	})

	// Verify all user votes collected
	uassert.Equal(t, len(collectedProposals), 3, "should have 3 votes")
	uassert.True(t, containsId(collectedProposals, 1), "should have vote on proposal 1")
	uassert.True(t, containsId(collectedProposals, 2), "should have vote on proposal 2")
	uassert.True(t, containsId(collectedProposals, 3), "should have vote on proposal 3")
}

func TestStateImmutability(t *testing.T) {
	// Test: Comprehensive immutability test
	// Expected: Original state never modified by any operation

	originalState := createTestState()
	originalCounter := originalState.GetProposalCounter()

	// Perform multiple operations
	state1 := originalState.WithNewProposal(std.Address("g1p1"), createTestProposal2(std.Address("g1p1")))
	state2 := state1.WithNewProposal(std.Address("g1p2"), createTestProposal2(std.Address("g1p2")))
	state3 := state2.WithVote(1, std.Address("g1v1"), true, 100)
	state4 := state3.WithUpdatedProposal(1, func(p ProposalInfo) ProposalInfo {
		p.State.Active = true
		return p
	})
	state5 := state4.WithArchivedProposal(1)
	state6 := state5.WithNewConfig(Config{Quorum: 75})

	// Verify original state unchanged
	uassert.Equal(t, originalState.GetProposalCounter(), originalCounter, "original counter unchanged")
	_, exists := originalState.GetProposal(1)
	uassert.False(t, exists, "original state has no proposals")

	// Verify each state has expected changes
	uassert.Equal(t, state1.GetProposalCounter(), uint64(1), "state1 has 1 proposal")
	uassert.Equal(t, state2.GetProposalCounter(), uint64(2), "state2 has 2 proposals")

	vote, exists := state3.GetVote(1, std.Address("g1v1"))
	uassert.True(t, exists && vote == true, "state3 has vote")

	proposal, exists := state4.GetProposal(1)
	uassert.True(t, exists, "state4 has proposal")
	uassert.True(t, proposal.State.Active, "state4 has active proposal")

	_, exists = state5.GetProposal(1)
	uassert.False(t, exists, "state5 has archived proposal")

	config, _ := state6.GetConfig(0)
	uassert.Equal(t, config.Quorum, uint64(75), "state6 has new config")
}

// Helper functions

func createTestState() *GovernanceState {
	config := Config{
		VotingStartDelay:              86400,
		VotingPeriod:                  604800,
		VotingWeightSmoothingDuration: 86400,
		Quorum:                        50,
		ProposalCreationThreshold:     1_000_000_000,
		ExecutionDelay:                86400,
		ExecutionWindow:               2592000,
	}
	return NewGovernanceState(config)
}

func createTestStateWithProposal() *GovernanceState {
	state := createTestState()
	proposer := std.Address("g1proposer")
	proposal := createTestProposal2(proposer)
	return state.WithNewProposal(proposer, proposal)
}

func createTestProposal2(proposer std.Address) ProposalInfo {
	return ProposalInfo{
		Proposer:     proposer,
		ProposalType: Text,
		State: ProposalState{
			Created:   true,
			CreatedAt: uint64(time.Now().Unix()),
			Upcoming:  true,
		},
		Yea:             u256.Zero(),
		Nay:             u256.Zero(),
		MaxVotingWeight: u256.NewUint(1000000),
		ConfigVersion:   1,
		QuorumAmount:    500000,
		Title:           "Test Proposal",
		Description:     "Test Description",
		PossibleAddressWithWeight: map[std.Address]uint64{
			std.Address("g1voter"):  100,
			std.Address("g1voter2"): 200,
		},
	}
}

func containsId(ids []uint64, target uint64) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
