package commondao

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao/v0"
)

// newCouncilUpdatePropDefinition creates a new proposal definition for
// adding/removing council members.
func newCouncilUpdatePropDefinition(dao *commondao.CommonDAO, add, remove addrset.Set) councilUpdatePropDefinition {
	if dao == nil {
		panic("DAO is required")
	}

	if add.Size() == 0 && remove.Size() == 0 {
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
	toAdd, toRemove addrset.Set
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

	if p.toAdd.Size() > 0 {
		items := make([]string, 0, p.toAdd.Size())
		p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr address) bool {
			items = append(items, addr.String())
			return false
		})

		b.WriteString(md.Paragraph(
			md.Bold("Council Members to Add:") + "\n" + md.BulletList(items),
		))
	}

	if p.toRemove.Size() > 0 {
		items := make([]string, 0, p.toRemove.Size())
		p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr address) bool {
			items = append(items, addr.String())
			return false
		})

		b.WriteString(md.Paragraph(
			md.Bold("Council Members to Remove:") + "\n" + md.BulletList(items),
		))
	}

	return b.String()
}

func (p councilUpdatePropDefinition) Validate() (err error) {
	// Membership is intentionally not validated: adds and removes are
	// idempotent, so updates passed concurrently merge cleanly instead of
	// failing on addresses that another update already added or removed.
	p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr address) bool {
		if p.toRemove.Has(addr) {
			err = errors.New("address is added and removed at once: " + addr.String())
			return true
		}
		return false
	})
	return err
}

func (councilUpdatePropDefinition) Tally(ctx commondao.VotingContext) (bool, error) {
	outcome := commondao.TallyDefault(ctx.VotingRecord, ctx.Council, commondao.ThresholdSupermajority)
	return outcome == commondao.OutcomePassed, nil
}

func (p councilUpdatePropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

func (p councilUpdatePropDefinition) execute(realm) error {
	add := make([]address, 0, p.toAdd.Size())
	p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr address) bool {
		add = append(add, addr)
		return false
	})

	remove := make([]address, 0, p.toRemove.Size())
	p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr address) bool {
		remove = append(remove, addr)
		return false
	})

	return p.dao.UpdateCouncil(add, remove)
}
