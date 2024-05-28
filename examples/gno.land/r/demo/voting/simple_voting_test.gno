package voting

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

// TestCreateProposal tests the creation of a proposal
func TestCreateProposal(t *testing.T) {
	Init()

	title := "Test Proposal"
	numParticipants := uint64(100)
	maxVotesPerUser := uint64(5)

	proposalID := CreateProposal(title, numParticipants, maxVotesPerUser)

	if proposalID == "" {
		t.Errorf("Expected proposalID, got empty string")
	}

	proposal, exists := proposals.Get(proposalID)
	if !exists {
		t.Errorf("Proposal not found")
	}

	prop := proposal.(*Proposal)
	if prop.title != title {
		t.Errorf("Expected title %s, got %s", title, prop.title)
	}

	if prop.maxVotesPerUser != maxVotesPerUser {
		t.Errorf("Expected maxVotesPerUser %d, got %d", maxVotesPerUser, prop.maxVotesPerUser)
	}

	// if prop.voters.Size() != numParticipants {
	// 	t.Errorf("Expected total voters %d, got %d", numParticipants, prop.voters.Size())
	// }

	if uint64(proposalIndex) != 1 {
		t.Errorf("Expected proposalIndex %d, got %d", 1, uint64(proposalIndex))
	}
}

// TestRegisterVoter tests voter registration
func TestRegisterVoter(t *testing.T) {
	Init()

	address := testutils.TestAddress("dummy")
	std.TestSetPrevAddr(address)

	voterID := RegisterVoter()
	if voterID == "" {
		t.Errorf("Expected voterID, got empty string")
	}

	voter, exists := registeredVoters.Get(address.String())
	if !exists {
		t.Errorf("Voter not found")
	}

	if voter.(string) != voterID {
		t.Errorf("Expected voterID %s, got %s", voterID, voter.(string))
	}
}

// TestVote tests voting on a proposal
func TestVote(t *testing.T) {
	Init()

	title := "Test Proposal"
	numParticipants := uint64(100)
	maxVotesPerUser := uint64(5)

	owner := testutils.TestAddress("owner")
	std.TestSetPrevAddr(owner)

	proposalID := CreateProposal(title, numParticipants, maxVotesPerUser)

	voter := testutils.TestAddress("voter")
	std.TestSetPrevAddr(voter)

	voterID := RegisterVoter()

	err := Vote(proposalID, voterID)
	if err != nil {
		t.Errorf("Vote failed: %v", err)
	}

	votes, err := GetVotes(proposalID)
	if err != nil {
		t.Errorf("GetVotes failed: %v", err)
	}
	if votes != 1 {
		t.Errorf("Expected 1 vote, got %d", votes)
	}
}

// TestVoteLimit tests voting limit per user
func TestVoteLimit(t *testing.T) {
	Init()

	title := "Test Proposal"
	numParticipants := uint64(100)
	maxVotesPerUser := uint64(2)

	owner := testutils.TestAddress("owner")
	std.TestSetPrevAddr(owner)

	proposalID := CreateProposal(title, numParticipants, maxVotesPerUser)

	voter := testutils.TestAddress("voter")

	std.TestSetPrevAddr(voter)

	voterID := RegisterVoter()

	err := Vote(proposalID, voterID)
	if err != nil {
		t.Errorf("Vote failed: %v", err)
	}

	err = Vote(proposalID, voterID)
	if err != nil {
		t.Errorf("Vote failed: %v", err)
	}

	err = Vote(proposalID, voterID)
	if err == nil {
		t.Errorf("Expected vote limit error, got nil")
	}
}
