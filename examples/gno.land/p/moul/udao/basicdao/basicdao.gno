package basicdao

import (
	"time"

	"gno.land/p/demo/avl/list"
	"gno.land/p/moul/udao"
	"gno.land/p/moul/udao/wrap"
)

// BasicProposal implements the Proposal interface
type BasicProposal struct {
	id         uint64
	definition udao.PropDefinition
	state      udao.PropState
	yea        float64
	nay        float64
	abstain    float64
	pending    float64
}

func (p BasicProposal) ID() uint64                         { return p.id }
func (p BasicProposal) Definition() udao.PropDefinition    { return p.definition }
func (p BasicProposal) GetState() udao.PropState           { return p.state }
func (p BasicProposal) GetYeaPercentage() float64          { return p.yea }
func (p BasicProposal) GetNayPercentage() float64          { return p.nay }
func (p BasicProposal) GetAbstainPercentage() float64      { return p.abstain }
func (p BasicProposal) GetPendingVoterPercentage() float64 { return p.pending }

// BasicPropDefinition implements udao.PropDefinition
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
	proposals map[uint64]udao.Proposal
	nextID    uint64
	active    *list.List
	archived  *list.List
}

// NewBasicDAO creates a new BasicDAO instance
func NewBasicDAO() *BasicDAO {
	return &BasicDAO{
		proposals: make(map[uint64]udao.Proposal),
		nextID:    1,
		active:    &list.List{},
		archived:  &list.List{},
	}
}

func (d *BasicDAO) Propose(def udao.PropDefinition) (udao.Proposal, error) {
	prop := BasicProposal{
		id:         d.nextID,
		definition: def,
		state:      udao.Pending,
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

func (d *BasicDAO) GetProposal(proposalID uint64) (udao.Proposal, error) {
	if prop, ok := d.proposals[proposalID]; ok {
		return prop, nil
	}
	return nil, nil
}

func (d *BasicDAO) Execute(proposalID uint64) error {
	if prop, ok := d.proposals[proposalID]; ok {
		if prop.GetState() == udao.Passed {
			newProp := prop.(BasicProposal)
			newProp.state = udao.Executed
			d.proposals[proposalID] = newProp

			// Move from active to archived
			for i := 0; i < d.active.Len(); i++ {
				if p := d.active.Get(i).(udao.Proposal); p.ID() == proposalID {
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

func (d *BasicDAO) ActiveProposals() udao.PropList {
	return wrap.WrapAsPropList(d.active)
}

func (d *BasicDAO) ArchivedProposals() udao.PropList {
	return wrap.WrapAsPropList(d.archived)
}

func (d *BasicDAO) Len() int {
	return len(d.proposals)
}

// Helper methods for managing proposal state
func (d *BasicDAO) UpdateProposalState(proposalID uint64, state udao.PropState) error {
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
