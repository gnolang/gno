package udao

import (
	"time"

	"gno.land/p/demo/avl/list"
)

// BasicProposal implements the Proposal interface
type BasicProposal struct {
	id         uint64
	definition PropDefinition
	state      PropState
	yea        float64
	nay        float64
	abstain    float64
	pending    float64
}

func (p BasicProposal) ID() uint64                         { return p.id }
func (p BasicProposal) Definition() PropDefinition         { return p.definition }
func (p BasicProposal) GetState() PropState                { return p.state }
func (p BasicProposal) GetYeaPercentage() float64          { return p.yea }
func (p BasicProposal) GetNayPercentage() float64          { return p.nay }
func (p BasicProposal) GetAbstainPercentage() float64      { return p.abstain }
func (p BasicProposal) GetPendingVoterPercentage() float64 { return p.pending }

// BasicPropDefinition implements PropDefinition
type BasicPropDefinition struct {
	title    string
	body     string
	created  time.Time
	finished time.Time
}

func (d BasicPropDefinition) Title() string       { return d.title }
func (d BasicPropDefinition) Body() string        { return d.body }
func (d BasicPropDefinition) Created() time.Time  { return d.created }
func (d BasicPropDefinition) Finished() time.Time { return d.finished }
func (d BasicPropDefinition) CheckConstraints() (valid bool, reasons []string) {
	panic("not implemented")
}

// BasicDAO provides a minimal implementation of the DAO interface
type BasicDAO struct {
	proposals map[uint64]Proposal
	nextID    uint64
	active    *list.List
	archived  *list.List
}

// NewBasicDAO creates a new BasicDAO instance
func NewBasicDAO() *BasicDAO {
	return &BasicDAO{
		proposals: make(map[uint64]Proposal),
		nextID:    1,
		active:    &list.List{},
		archived:  &list.List{},
	}
}

func (d *BasicDAO) Propose(def PropDefinition) (Proposal, error) {
	prop := BasicProposal{
		id:         d.nextID,
		definition: def,
		state:      Pending,
		yea:        0,
		nay:        0,
		abstain:    0,
		pending:    100,
	}
	d.proposals[d.nextID] = prop
	d.active.Append(prop)
	d.nextID++
	return prop, nil
}

func (d *BasicDAO) GetProposal(proposalID uint64) (Proposal, error) {
	if prop, ok := d.proposals[proposalID]; ok {
		return prop, nil
	}
	return nil, nil
}

func (d *BasicDAO) Execute(proposalID uint64) error {
	if prop, ok := d.proposals[proposalID]; ok {
		if prop.GetState() == Passed {
			newProp := prop.(BasicProposal)
			newProp.state = Executed
			d.proposals[proposalID] = newProp

			// Move from active to archived
			for i := 0; i < d.active.Len(); i++ {
				if p := d.active.Get(i).(Proposal); p.ID() == proposalID {
					d.active.Delete(i)
					d.archived.Append(newProp)
					break
				}
			}
			return nil
		}
	}
	return nil
}

func (d *BasicDAO) ActiveProposals() PropList {
	return WrapAsPropList(d.active)
}

func (d *BasicDAO) ArchivedProposals() PropList {
	return WrapAsPropList(d.archived)
}

func (d *BasicDAO) Len() int {
	return len(d.proposals)
}

// Helper methods for managing proposal state
func (d *BasicDAO) UpdateProposalState(proposalID uint64, state PropState) error {
	if prop, ok := d.proposals[proposalID]; ok {
		newProp := prop.(BasicProposal)
		newProp.state = state
		d.proposals[proposalID] = newProp
		return nil
	}
	return nil
}

func (d *BasicDAO) UpdateVotingPercentages(proposalID uint64, yea, nay, abstain, pending float64) error {
	if prop, ok := d.proposals[proposalID]; ok {
		newProp := prop.(BasicProposal)
		newProp.yea = yea
		newProp.nay = nay
		newProp.abstain = abstain
		newProp.pending = pending
		d.proposals[proposalID] = newProp
		return nil
	}
	return nil
}
