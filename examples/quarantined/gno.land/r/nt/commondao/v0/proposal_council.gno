package commondao

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao/v0"
)

// newCouncilUpdatePropDefinition creates a new proposal definition for
// adding/removing council members.
func newCouncilUpdatePropDefinition(dao *commondao.CommonDAO, add, remove []address) councilUpdatePropDefinition {
	if dao == nil {
		panic("DAO is required")
	}

	if len(add) == 0 && len(remove) == 0 {
		panic("no council members were specified to be added or removed")
	}

	return councilUpdatePropDefinition{
		dao:      dao,
		toAdd:    add,
		toRemove: remove,
	}
}

// councilUpdatePropDefinition defines a proposal type for adding/removing
// council members. Adds and removes apply as idempotent set operations, so
// concurrently passed updates merge in execution order; an update whose
// final set would empty a non-empty council fails at execution.
type councilUpdatePropDefinition struct {
	dao             *commondao.CommonDAO
	toAdd, toRemove []address
}

func (councilUpdatePropDefinition) Title() string               { return "Council Update" }
func (councilUpdatePropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// CapExempt exempts council updates from the active proposals cap: a
// council member who fills the cap must never be able to block their own
// removal. Exempt proposals are bounded to one active one per creator.
func (councilUpdatePropDefinition) CapExempt() {}

// Threshold returns the tally threshold: council self-mutation requires a
// supermajority.
func (councilUpdatePropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

func (p councilUpdatePropDefinition) Body() string {
	var b strings.Builder

	if len(p.toAdd) > 0 {
		b.WriteString(md.Paragraph(
			md.Bold("Council Members to Add:") + "\n" + md.BulletList(addrStrings(p.toAdd)),
		))
	}

	if len(p.toRemove) > 0 {
		b.WriteString(md.Paragraph(
			md.Bold("Council Members to Remove:") + "\n" + md.BulletList(addrStrings(p.toRemove)),
		))
	}

	return b.String()
}

func (p councilUpdatePropDefinition) Validate() error {
	// Membership is intentionally not validated: adds and removes are
	// idempotent, so updates passed concurrently merge cleanly instead of
	// failing on addresses that another update already added or removed.
	for _, a := range p.toAdd {
		for _, r := range p.toRemove {
			if a == r {
				return errors.New("address is added and removed at once: " + a.String())
			}
		}
	}
	return nil
}

func (p councilUpdatePropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

func (p councilUpdatePropDefinition) execute(realm) error {
	return p.dao.UpdateCouncil(p.toAdd, p.toRemove)
}

// addrStrings converts addresses for markdown list rendering.
func addrStrings(addrs []address) []string {
	items := make([]string, 0, len(addrs))
	for _, a := range addrs {
		items = append(items, a.String())
	}
	return items
}
