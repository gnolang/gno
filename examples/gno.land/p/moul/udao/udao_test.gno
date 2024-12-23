package udao

import (
	"testing"
	"time"
)

func TestBasicDAO(t *testing.T) {
	dao := NewBasicDAO()
	if dao.Len() != 0 {
		t.Errorf("expected empty DAO, got len=%d", dao.Len())
	}

	// Test proposal creation
	def := BasicPropDefinition{
		title:    "Test Proposal",
		body:     "This is a test proposal",
		created:  time.Now(),
		finished: time.Now().Add(24 * time.Hour),
	}

	prop, err := dao.Propose(def)
	if err != nil {
		t.Errorf("unexpected error creating proposal: %v", err)
	}
	if prop.ID() != 1 {
		t.Errorf("expected proposal ID=1, got %d", prop.ID())
	}
	if dao.Len() != 1 {
		t.Errorf("expected DAO len=1, got %d", dao.Len())
	}

	// Test proposal retrieval
	retrieved, err := dao.GetProposal(1)
	if err != nil {
		t.Errorf("unexpected error getting proposal: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected to retrieve proposal, got nil")
	}
	if retrieved.ID() != prop.ID() {
		t.Errorf("expected retrieved ID=%d, got %d", prop.ID(), retrieved.ID())
	}

	// Test active proposals list
	active := dao.ActiveProposals()
	if active.Len() != 1 {
		t.Errorf("expected 1 active proposal, got %d", active.Len())
	}
	if active.Get(0).ID() != prop.ID() {
		t.Errorf("expected active proposal ID=%d, got %d", prop.ID(), active.Get(0).ID())
	}

	// Test proposal execution
	dao.UpdateProposalState(1, Passed)
	err = dao.Execute(1)
	if err != nil {
		t.Errorf("unexpected error executing proposal: %v", err)
	}

	executed, _ := dao.GetProposal(1)
	if executed.GetState() != Executed {
		t.Errorf("expected state=Executed, got %s", executed.GetState())
	}

	// Test archived proposals
	archived := dao.ArchivedProposals()
	if archived.Len() != 1 {
		t.Errorf("expected 1 archived proposal, got %d", archived.Len())
	}
	if archived.Get(0).ID() != prop.ID() {
		t.Errorf("expected archived proposal ID=%d, got %d", prop.ID(), archived.Get(0).ID())
	}
}

func TestPropListWrapper(t *testing.T) {
	dao := NewBasicDAO()

	// Create multiple proposals
	for i := 0; i < 3; i++ {
		def := BasicPropDefinition{
			title:    "Test Proposal",
			body:     "This is a test proposal",
			created:  time.Now(),
			finished: time.Now().Add(24 * time.Hour),
		}
		dao.Propose(def)
	}

	active := dao.ActiveProposals()

	// Test Len()
	if active.Len() != 3 {
		t.Errorf("expected len=3, got %d", active.Len())
	}

	// Test Get()
	prop := active.Get(0)
	if prop.ID() != 1 {
		t.Errorf("expected first proposal ID=1, got %d", prop.ID())
	}

	// Test Slice()
	slice := active.Slice(0, 2)
	if len(slice) != 2 {
		t.Errorf("expected slice len=2, got %d", len(slice))
	}
	if slice[0].ID() != 1 || slice[1].ID() != 2 {
		t.Errorf("unexpected slice IDs: %d, %d", slice[0].ID(), slice[1].ID())
	}
}

func TestProposalVoting(t *testing.T) {
	dao := NewBasicDAO()

	// Create a proposal
	def := BasicPropDefinition{
		title:    "Voting Test",
		body:     "Testing voting percentages",
		created:  time.Now(),
		finished: time.Now().Add(24 * time.Hour),
	}
	prop, _ := dao.Propose(def)

	// Test initial voting percentages
	if prop.GetYeaPercentage() != 0 {
		t.Errorf("expected initial yea=0, got %f", prop.GetYeaPercentage())
	}
	if prop.GetNayPercentage() != 0 {
		t.Errorf("expected initial nay=0, got %f", prop.GetNayPercentage())
	}
	if prop.GetPendingVoterPercentage() != 100 {
		t.Errorf("expected initial pending=100, got %f", prop.GetPendingVoterPercentage())
	}

	// Update voting percentages
	err := dao.UpdateVotingPercentages(prop.ID(), 60, 30, 10, 0)
	if err != nil {
		t.Errorf("unexpected error updating voting percentages: %v", err)
	}

	// Verify updated percentages
	updated, _ := dao.GetProposal(prop.ID())
	if updated.GetYeaPercentage() != 60 {
		t.Errorf("expected yea=60, got %f", updated.GetYeaPercentage())
	}
	if updated.GetNayPercentage() != 30 {
		t.Errorf("expected nay=30, got %f", updated.GetNayPercentage())
	}
	if updated.GetAbstainPercentage() != 10 {
		t.Errorf("expected abstain=10, got %f", updated.GetAbstainPercentage())
	}
	if updated.GetPendingVoterPercentage() != 0 {
		t.Errorf("expected pending=0, got %f", updated.GetPendingVoterPercentage())
	}
}

func TestProposalStateTransitions(t *testing.T) {
	dao := NewBasicDAO()

	// Create a proposal
	def := BasicPropDefinition{
		title:    "State Test",
		body:     "Testing state transitions",
		created:  time.Now(),
		finished: time.Now().Add(24 * time.Hour),
	}
	prop, _ := dao.Propose(def)

	// Test initial state
	if prop.GetState() != Pending {
		t.Errorf("expected initial state=Pending, got %s", prop.GetState())
	}

	// Test state transitions
	states := []PropState{Active, Passed, Failed, Executed, Cancelled}
	for _, state := range states {
		err := dao.UpdateProposalState(prop.ID(), state)
		if err != nil {
			t.Errorf("unexpected error updating state to %s: %v", state, err)
		}

		updated, _ := dao.GetProposal(prop.ID())
		if updated.GetState() != state {
			t.Errorf("expected state=%s, got %s", state, updated.GetState())
		}
	}
}

func TestInvalidProposalOperations(t *testing.T) {
	dao := NewBasicDAO()

	// Test getting non-existent proposal
	prop, err := dao.GetProposal(999)
	if err != nil {
		t.Errorf("expected nil error for non-existent proposal, got %v", err)
	}
	if prop != nil {
		t.Error("expected nil proposal for non-existent ID")
	}

	// Test executing non-existent proposal
	err = dao.Execute(999)
	if err != nil {
		t.Errorf("expected nil error for executing non-existent proposal, got %v", err)
	}

	// Test executing non-passed proposal
	def := BasicPropDefinition{
		title:    "Invalid Execute",
		body:     "Testing invalid execution",
		created:  time.Now(),
		finished: time.Now().Add(24 * time.Hour),
	}
	prop, _ = dao.Propose(def)
	err = dao.Execute(prop.ID())
	if err != nil {
		t.Errorf("expected nil error for executing non-passed proposal, got %v", err)
	}

	// Verify state didn't change
	updated, _ := dao.GetProposal(prop.ID())
	if updated.GetState() != Pending {
		t.Errorf("expected state to remain Pending, got %s", updated.GetState())
	}
}

func TestPropListSlicing(t *testing.T) {
	dao := NewBasicDAO()

	// Create 5 proposals
	for i := 0; i < 5; i++ {
		def := BasicPropDefinition{
			title:    "Test Proposal",
			body:     "This is a test proposal",
			created:  time.Now(),
			finished: time.Now().Add(24 * time.Hour),
		}
		dao.Propose(def)
	}

	active := dao.ActiveProposals()

	// Test various slice operations
	testCases := []struct {
		start    int
		end      int
		expected int
	}{
		{0, 2, 2},
		{1, 4, 3},
		{0, 5, 5},
		{2, 3, 1},
		{4, 5, 1},
	}

	for _, tc := range testCases {
		slice := active.Slice(tc.start, tc.end)
		if len(slice) != tc.expected {
			t.Errorf("expected slice[%d:%d] len=%d, got %d",
				tc.start, tc.end, tc.expected, len(slice))
		}

		// Verify IDs are sequential
		for i := range slice {
			expectedID := uint64(tc.start + i + 1)
			if slice[i].ID() != expectedID {
				t.Errorf("expected ID=%d at index %d, got %d",
					expectedID, i, slice[i].ID())
			}
		}
	}
}
