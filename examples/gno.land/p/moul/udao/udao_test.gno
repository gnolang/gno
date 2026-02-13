package udao

import (
	"testing"
	"time"
)

// MockPropDefinition implements PropDefinition for testing
type MockPropDefinition struct {
	title     string
	body      string
	created   time.Time
	finished  time.Time
	isValid   bool
	reasons   []string
}

func (m MockPropDefinition) Title() string                              { return m.title }
func (m MockPropDefinition) Body() string                               { return m.body }
func (m MockPropDefinition) Created() time.Time                         { return m.created }
func (m MockPropDefinition) Finished() time.Time                        { return m.finished }
func (m MockPropDefinition) CheckConstraints() (bool, []string)         { return m.isValid, m.reasons }

// MockProposal implements Proposal for testing
type MockProposal struct {
	id         uint64
	definition PropDefinition
	state      PropState
	yea        float64
	nay        float64
	abstain    float64
	pending    float64
}

func (m MockProposal) ID() uint64                         { return m.id }
func (m MockProposal) Definition() PropDefinition         { return m.definition }
func (m MockProposal) GetState() PropState                { return m.state }
func (m MockProposal) GetYeaPercentage() float64          { return m.yea }
func (m MockProposal) GetNayPercentage() float64          { return m.nay }
func (m MockProposal) GetAbstainPercentage() float64      { return m.abstain }
func (m MockProposal) GetPendingVoterPercentage() float64 { return m.pending }

// MockPropList implements PropList for testing
type MockPropList struct {
	proposals []Proposal
}

func (m MockPropList) Len() int                                 { return len(m.proposals) }
func (m MockPropList) Get(index int) Proposal                  { return m.proposals[index] }
func (m MockPropList) Slice(start, end int) []Proposal         { return m.proposals[start:end] }

// MockDAO implements DAO for testing
type MockDAO struct {
	proposals map[uint64]Proposal
	nextID    uint64
	active    []Proposal
	archived  []Proposal
}

func NewMockDAO() *MockDAO {
	return &MockDAO{
		proposals: make(map[uint64]Proposal),
		nextID:    1,
		active:    []Proposal{},
		archived:  []Proposal{},
	}
}

func (m *MockDAO) Propose(def PropDefinition) (Proposal, error) {
	prop := MockProposal{
		id:         m.nextID,
		definition: def,
		state:      Pending,
		yea:        0,
		nay:        0,
		abstain:    0,
		pending:    100,
	}
	m.proposals[m.nextID] = prop
	m.active = append(m.active, prop)
	m.nextID++
	return prop, nil
}

func (m *MockDAO) GetProposal(proposalID uint64) (Proposal, error) {
	if prop, ok := m.proposals[proposalID]; ok {
		return prop, nil
	}
	return nil, nil
}

func (m *MockDAO) Execute(proposalID uint64) error {
	if prop, ok := m.proposals[proposalID]; ok {
		if prop.GetState() == Passed {
			// Update state to executed
			updatedProp := prop.(MockProposal)
			updatedProp.state = Executed
			m.proposals[proposalID] = updatedProp
			
			// Move from active to archived
			for i, activeProp := range m.active {
				if activeProp.ID() == proposalID {
					m.active = append(m.active[:i], m.active[i+1:]...)
					m.archived = append(m.archived, updatedProp)
					break
				}
			}
		}
	}
	return nil
}

func (m *MockDAO) ActiveProposals() PropList {
	return MockPropList{proposals: m.active}
}

func (m *MockDAO) ArchivedProposals() PropList {
	return MockPropList{proposals: m.archived}
}

func (m *MockDAO) Len() int {
	return len(m.proposals)
}

func (m *MockDAO) UpdateProposalState(proposalID uint64, state PropState) {
	if prop, ok := m.proposals[proposalID]; ok {
		updatedProp := prop.(MockProposal)
		updatedProp.state = state
		m.proposals[proposalID] = updatedProp
	}
}

// Test functions

func TestPropStateConstants(t *testing.T) {
	states := []PropState{Pending, Active, Passed, Failed, Executed, Cancelled}
	expectedValues := []string{"pending", "active", "passed", "failed", "executed", "cancelled"}
	
	for i, state := range states {
		if string(state) != expectedValues[i] {
			t.Errorf("expected state %s, got %s", expectedValues[i], string(state))
		}
	}
}

func TestMockPropDefinition(t *testing.T) {
	now := time.Now()
	later := now.Add(24 * time.Hour)
	
	def := MockPropDefinition{
		title:    "Test Proposal",
		body:     "This is a test proposal",
		created:  now,
		finished: later,
		isValid:  true,
		reasons:  []string{},
	}
	
	if def.Title() != "Test Proposal" {
		t.Errorf("expected title 'Test Proposal', got '%s'", def.Title())
	}
	if def.Body() != "This is a test proposal" {
		t.Errorf("expected body 'This is a test proposal', got '%s'", def.Body())
	}
	if !def.Created().Equal(now) {
		t.Errorf("expected created time %v, got %v", now, def.Created())
	}
	if !def.Finished().Equal(later) {
		t.Errorf("expected finished time %v, got %v", later, def.Finished())
	}
	
	valid, reasons := def.CheckConstraints()
	if !valid {
		t.Errorf("expected valid=true, got %v", valid)
	}
	if len(reasons) != 0 {
		t.Errorf("expected no reasons, got %v", reasons)
	}
}

func TestMockProposal(t *testing.T) {
	def := MockPropDefinition{
		title:   "Test",
		body:    "Test body",
		isValid: true,
	}
	
	prop := MockProposal{
		id:         1,
		definition: def,
		state:      Active,
		yea:        60.0,
		nay:        30.0,
		abstain:    10.0,
		pending:    0.0,
	}
	
	if prop.ID() != 1 {
		t.Errorf("expected ID=1, got %d", prop.ID())
	}
	if prop.Definition().Title() != def.Title() {
		t.Error("expected definition title to match")
	}
	if prop.Definition().Body() != def.Body() {
		t.Error("expected definition body to match")
	}
	if prop.GetState() != Active {
		t.Errorf("expected state=Active, got %s", prop.GetState())
	}
	if prop.GetYeaPercentage() != 60.0 {
		t.Errorf("expected yea=60.0, got %f", prop.GetYeaPercentage())
	}
	if prop.GetNayPercentage() != 30.0 {
		t.Errorf("expected nay=30.0, got %f", prop.GetNayPercentage())
	}
	if prop.GetAbstainPercentage() != 10.0 {
		t.Errorf("expected abstain=10.0, got %f", prop.GetAbstainPercentage())
	}
	if prop.GetPendingVoterPercentage() != 0.0 {
		t.Errorf("expected pending=0.0, got %f", prop.GetPendingVoterPercentage())
	}
}

func TestMockPropList(t *testing.T) {
	def := MockPropDefinition{title: "Test", body: "Test body", isValid: true}
	props := []Proposal{
		MockProposal{id: 1, definition: def, state: Active},
		MockProposal{id: 2, definition: def, state: Pending},
		MockProposal{id: 3, definition: def, state: Passed},
	}
	
	list := MockPropList{proposals: props}
	
	if list.Len() != 3 {
		t.Errorf("expected len=3, got %d", list.Len())
	}
	
	firstProp := list.Get(0)
	if firstProp.ID() != 1 {
		t.Errorf("expected first prop ID=1, got %d", firstProp.ID())
	}
	
	slice := list.Slice(1, 3)
	if len(slice) != 2 {
		t.Errorf("expected slice len=2, got %d", len(slice))
	}
	if slice[0].ID() != 2 || slice[1].ID() != 3 {
		t.Errorf("unexpected slice IDs: %d, %d", slice[0].ID(), slice[1].ID())
	}
}

func TestMockDAO(t *testing.T) {
	dao := NewMockDAO()
	
	// Test initial state
	if dao.Len() != 0 {
		t.Errorf("expected empty DAO, got len=%d", dao.Len())
	}
	
	// Test proposal creation
	def := MockPropDefinition{
		title:   "Test Proposal",
		body:    "Test body",
		isValid: true,
	}
	
	prop, err := dao.Propose(def)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if prop.ID() != 1 {
		t.Errorf("expected ID=1, got %d", prop.ID())
	}
	if dao.Len() != 1 {
		t.Errorf("expected len=1, got %d", dao.Len())
	}
	
	// Test proposal retrieval
	retrieved, err := dao.GetProposal(1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected to retrieve proposal")
	}
	if retrieved.ID() != 1 {
		t.Errorf("expected ID=1, got %d", retrieved.ID())
	}
	
	// Test active proposals
	active := dao.ActiveProposals()
	if active.Len() != 1 {
		t.Errorf("expected 1 active proposal, got %d", active.Len())
	}
	
	// Test execution
	dao.UpdateProposalState(1, Passed)
	err = dao.Execute(1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
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
	
	// Active should now be empty
	active = dao.ActiveProposals()
	if active.Len() != 0 {
		t.Errorf("expected 0 active proposals after execution, got %d", active.Len())
	}
}
