package basicdao

import (
	"testing"
	"time"

	"gno.land/p/moul/udao"
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
	dao.UpdateProposalState(1, udao.Passed)
	err = dao.Execute(1)
	if err != nil {
		t.Errorf("unexpected error executing proposal: %v", err)
	}


	executed, _ := dao.GetProposal(1)
	if executed.GetState() != udao.Executed {
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
	if prop.GetState() != udao.Pending {
		t.Errorf("expected initial state=Pending, got %s", prop.GetState())
	}

	// Test state transitions
	states := []udao.PropState{udao.Active, udao.Passed, udao.Failed, udao.Executed, udao.Cancelled}
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
	if updated.GetState() != udao.Pending {
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

func TestCheckConstraints(t *testing.T) {
	// Test valid proposal definition
	validDef := BasicPropDefinition{
		title:    "Valid Proposal",
		body:     "This is a valid proposal with content",
		created:  time.Now(),
		finished: time.Now().Add(24 * time.Hour),
	}
	
	valid, reasons := validDef.CheckConstraints()
	if !valid {
		t.Errorf("expected valid proposal to pass constraints, but got reasons: %v", reasons)
	}
	if len(reasons) != 0 {
		t.Errorf("expected no reasons for valid proposal, got %d reasons", len(reasons))
	}
	
	// Test empty title
	emptyTitleDef := BasicPropDefinition{
		title:    "",
		body:     "This has no title",
		created:  time.Now(),
		finished: time.Now().Add(24 * time.Hour),
	}
	
	valid, reasons = emptyTitleDef.CheckConstraints()
	if valid {
		t.Error("expected proposal with empty title to be invalid")
	}
	if len(reasons) == 0 {
		t.Error("expected reasons for invalid proposal")
	}
	found := false
	for _, reason := range reasons {
		if reason == "title cannot be empty" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'title cannot be empty' in reasons")
	}
	
	// Test empty body
	emptyBodyDef := BasicPropDefinition{
		title:    "Valid Title",
		body:     "",
		created:  time.Now(),
		finished: time.Now().Add(24 * time.Hour),
	}
	
	valid, reasons = emptyBodyDef.CheckConstraints()
	if valid {
		t.Error("expected proposal with empty body to be invalid")
	}
	found = false
	for _, reason := range reasons {
		if reason == "body cannot be empty" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'body cannot be empty' in reasons")
	}
	
	// Test invalid time order
	invalidTimeDef := BasicPropDefinition{
		title:    "Invalid Time",
		body:     "This has invalid time order",
		created:  time.Now(),
		finished: time.Now().Add(-24 * time.Hour), // finished before created
	}
	
	valid, reasons = invalidTimeDef.CheckConstraints()
	if valid {
		t.Error("expected proposal with invalid time order to be invalid")
	}
	found = false
	for _, reason := range reasons {
		if reason == "finished time cannot be before created time" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'finished time cannot be before created time' in reasons")
	}
	
	// Test multiple invalid constraints
	multiInvalidDef := BasicPropDefinition{
		title:    "",
		body:     "",
		created:  time.Now(),
		finished: time.Now().Add(-24 * time.Hour),
	}
	
	valid, reasons = multiInvalidDef.CheckConstraints()
	if valid {
		t.Error("expected proposal with multiple invalid constraints to be invalid")
	}
	if len(reasons) != 3 {
		t.Errorf("expected 3 constraint violations, got %d", len(reasons))
	}
}

func TestHelperConstructors(t *testing.T) {
	// Test NewBasicPropDefinition
	now := time.Now()
	later := now.Add(48 * time.Hour)
	
	def := NewBasicPropDefinition("Test Title", "Test Body", now, later)
	if def.Title() != "Test Title" {
		t.Errorf("expected title 'Test Title', got '%s'", def.Title())
	}
	if def.Body() != "Test Body" {
		t.Errorf("expected body 'Test Body', got '%s'", def.Body())
	}
	if !def.Created().Equal(now) {
		t.Errorf("expected created time %v, got %v", now, def.Created())
	}
	if !def.Finished().Equal(later) {
		t.Errorf("expected finished time %v, got %v", later, def.Finished())
	}
	
	valid, reasons := def.CheckConstraints()
	if !valid {
		t.Errorf("expected valid proposal, got reasons: %v", reasons)
	}
	
	// Test NewSimpleProposal
	beforeCreation := time.Now()
	simpleDef := NewSimpleProposal("Simple Title", "Simple Body")
	afterCreation := time.Now()
	
	if simpleDef.Title() != "Simple Title" {
		t.Errorf("expected title 'Simple Title', got '%s'", simpleDef.Title())
	}
	if simpleDef.Body() != "Simple Body" {
		t.Errorf("expected body 'Simple Body', got '%s'", simpleDef.Body())
	}
	
	// Check that created time is within reasonable bounds
	if simpleDef.Created().Before(beforeCreation) || simpleDef.Created().After(afterCreation) {
		t.Errorf("created time %v not within expected range %v to %v", 
			simpleDef.Created(), beforeCreation, afterCreation)
	}
	
	// Check that finished time is approximately 24 hours after created
	expectedFinished := simpleDef.Created().Add(24 * time.Hour)
	timeDiff := simpleDef.Finished().Sub(expectedFinished)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > time.Second {
		t.Errorf("finished time %v not approximately 24 hours after created %v", 
			simpleDef.Finished(), simpleDef.Created())
	}
	
	valid, reasons = simpleDef.CheckConstraints()
	if !valid {
		t.Errorf("expected valid simple proposal, got reasons: %v", reasons)
	}
}

func TestBasicDAOWithHelpers(t *testing.T) {
	dao := NewBasicDAO()
	
	// Use helper to create a simple proposal
	def := NewSimpleProposal("Helper Test", "Testing helper functions")
	
	prop, err := dao.Propose(def)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if prop.ID() != 1 {
		t.Errorf("expected ID=1, got %d", prop.ID())
	}
	
	if prop.Definition().Title() != "Helper Test" {
		t.Errorf("expected title 'Helper Test', got '%s'", prop.Definition().Title())
	}
	
	if prop.Definition().Body() != "Testing helper functions" {
		t.Errorf("expected body 'Testing helper functions', got '%s'", prop.Definition().Body())
	}
	
	// Verify the proposal passes constraints
	valid, reasons := prop.Definition().CheckConstraints()
	if !valid {
		t.Errorf("expected valid proposal, got reasons: %v", reasons)
	}
}
