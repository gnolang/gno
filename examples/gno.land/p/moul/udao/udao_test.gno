package udao

import (
	"testing"
	"time"
)

// MockProposal implements the Proposal interface for testing
type MockProposal struct {
	id         uint64
	definition MockPropDefinition
	PropStatus
}

func (p MockProposal) ID() uint64                 { return p.id }
func (p MockProposal) Definition() PropDefinition { return p.definition }

// MockPropDefinition implements PropDefinition for testing
type MockPropDefinition struct {
	title       string
	body        string
	created     time.Time
	constraints []Constraint
}

func (d MockPropDefinition) Title() string             { return d.title }
func (d MockPropDefinition) Body() string              { return d.body }
func (d MockPropDefinition) Created() time.Time        { return d.created }
func (d MockPropDefinition) Constraints() []Constraint { return d.constraints }

// MockDAO implements the DAO interface for testing
type MockDAO struct {
	proposals map[uint64]Proposal
	nextID    uint64
}

func NewMockDAO() *MockDAO {
	return &MockDAO{
		proposals: make(map[uint64]Proposal),
		nextID:    1,
	}
}

func (d *MockDAO) Propose(def PropDefinition) (Proposal, error) {
	prop := MockProposal{
		id:         d.nextID,
		definition: def.(MockPropDefinition),
		PropStatus: PropStatus{
			State:                  Pending,
			YeaPercentage:          0,
			NayPercentage:          0,
			AbstainPercentage:      0,
			PendingVoterPercentage: 100,
		},
	}
	d.proposals[d.nextID] = prop
	d.nextID++
	return prop, nil
}

func (d *MockDAO) GetProposal(proposalID uint64) (Proposal, error) {
	if prop, ok := d.proposals[proposalID]; ok {
		return prop, nil
	}
	return nil, nil
}

func (d *MockDAO) Execute(proposalID uint64) error {
	if prop, ok := d.proposals[proposalID]; ok {
		if prop.(MockProposal).State == Passed {
			newProp := prop.(MockProposal)
			newProp.State = Executed
			d.proposals[proposalID] = newProp
			return nil
		}
	}
	return nil
}

func (d *MockDAO) ActiveProposals() PropList {
	// Implementation omitted for brevity
	return nil
}

func (d *MockDAO) ArchivedProposals() PropList {
	// Implementation omitted for brevity
	return nil
}

// Tests

func TestPropose(t *testing.T) {
	dao := NewMockDAO()
	def := MockPropDefinition{
		title:   "Test Proposal",
		body:    "This is a test proposal",
		created: time.Now(),
	}

	prop, err := dao.Propose(def)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if prop.ID() != 1 {
		t.Errorf("Expected proposal ID 1, got %d", prop.ID())
	}

	if prop.Definition().Title() != "Test Proposal" {
		t.Errorf("Expected title 'Test Proposal', got '%s'", prop.Definition().Title())
	}
}

func TestGetProposal(t *testing.T) {
	dao := NewMockDAO()
	def := MockPropDefinition{
		title:   "Test Proposal",
		body:    "This is a test proposal",
		created: time.Now(),
	}

	created, _ := dao.Propose(def)
	retrieved, err := dao.GetProposal(created.ID())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected to retrieve proposal, got nil")
	}

	if retrieved.ID() != created.ID() {
		t.Errorf("Expected proposal ID %d, got %d", created.ID(), retrieved.ID())
	}
}

func TestExecuteProposal(t *testing.T) {
	dao := NewMockDAO()
	def := MockPropDefinition{
		title:   "Test Proposal",
		body:    "This is a test proposal",
		created: time.Now(),
	}

	prop, _ := dao.Propose(def)

	// Manually set the proposal state to Passed
	mockProp := prop.(MockProposal)
	mockProp.State = Passed
	dao.proposals[prop.ID()] = mockProp

	err := dao.Execute(prop.ID())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	executed, _ := dao.GetProposal(prop.ID())
	if executed.(MockProposal).State != Executed {
		t.Errorf("Expected proposal state to be Executed, got %s", executed.(MockProposal).State)
	}
}

// MockConstraint implements the Constraint interface for testing
type MockConstraint struct {
	isValid     bool
	description string
}

func (c MockConstraint) Validate() (bool, string) {
	return c.isValid, c.description
}

func TestValidateConstraints(t *testing.T) {
	validConstraint := MockConstraint{isValid: true, description: "valid"}
	invalidConstraint := MockConstraint{isValid: false, description: "invalid"}

	tests := []struct {
		name          string
		constraints   []Constraint
		expectValid   bool
		expectedCount int
	}{
		{
			name:          "No constraints",
			constraints:   []Constraint{},
			expectValid:   true,
			expectedCount: 0,
		},
		{
			name:          "Single valid constraint",
			constraints:   []Constraint{validConstraint},
			expectValid:   true,
			expectedCount: 0,
		},
		{
			name:          "Single invalid constraint",
			constraints:   []Constraint{invalidConstraint},
			expectValid:   false,
			expectedCount: 1,
		},
		{
			name:          "Mixed constraints",
			constraints:   []Constraint{validConstraint, invalidConstraint},
			expectValid:   false,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := MockProposal{
				id: 1,
				definition: MockPropDefinition{
					constraints: tt.constraints,
				},
			}

			valid, reasons := ValidateConstraints(prop)
			if valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, valid)
			}
			if len(reasons) != tt.expectedCount {
				t.Errorf("Expected %d reasons, got %d", tt.expectedCount, len(reasons))
			}
		})
	}
}
