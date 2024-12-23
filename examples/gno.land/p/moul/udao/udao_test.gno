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
