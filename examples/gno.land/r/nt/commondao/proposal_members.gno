package commondao

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao"
)

// newMembersUpdatePropDefinition creates a new proposal definition for adding/removing DAO members.
func newMembersUpdatePropDefinition(dao *commondao.CommonDAO, add, remove addrset.Set) membersUpdatePropDefinition {
	if dao == nil {
		panic("DAO is required")
	}

	if dao.Members().Size() == 0 {
		panic("a DAO with at least one member is required to create member update proposals")
	}

	if add.Size() == 0 && remove.Size() == 0 {
		panic("no members were specified to be added or removed")
	}

	return membersUpdatePropDefinition{
		dao:      dao,
		toAdd:    add,
		toRemove: remove,
	}
}

// membersUpdatePropDefinition defines a proposal type for adding/removing DAO members.
type membersUpdatePropDefinition struct {
	dao             *commondao.CommonDAO
	toAdd, toRemove addrset.Set
}

func (membersUpdatePropDefinition) Title() string               { return "Members Update" }
func (membersUpdatePropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

func (p membersUpdatePropDefinition) Body() string {
	var b strings.Builder

	if p.toAdd.Size() > 0 {
		items := make([]string, 0, p.toAdd.Size())
		p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr address) bool {
			items = append(items, addr.String())
			return false
		})

		b.WriteString(md.Paragraph(
			md.Bold("Members to Add:") + "\n" + md.BulletList(items),
		))
	}

	if p.toRemove.Size() > 0 {
		items := make([]string, 0, p.toRemove.Size())
		p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr address) bool {
			items = append(items, addr.String())
			return false
		})

		b.WriteString(md.Paragraph(
			md.Bold("Members to Remove:") + "\n" + md.BulletList(items),
		))
	}

	return b.String()
}

func (p membersUpdatePropDefinition) Validate() (err error) {
	members := p.dao.Members()

	p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr address) bool {
		if members.Has(addr) {
			err = errors.New("address is already a DAO member: " + addr.String())
			return true
		}
		return false
	})

	if err != nil {
		return err
	}

	p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr address) bool {
		if !members.Has(addr) {
			err = errors.New("address is not a DAO member: " + addr.String())
			return true
		}
		return false
	})

	return err
}

func (membersUpdatePropDefinition) Tally(ctx commondao.VotingContext) (bool, error) {
	// When DAO has one or two members succeed when there is a YES vote, otherwise
	// tally requires at least three votes to be able to tally by 2/3s super majority
	if ctx.Members.Size() < 3 {
		return ctx.VotingRecord.VoteCount(commondao.ChoiceYes) > 0, nil
	}

	if !commondao.IsQuorumReached(commondao.QuorumTwoThirds, ctx.VotingRecord, ctx.Members) {
		return false, commondao.ErrNoQuorum
	}

	c, success := commondao.SelectChoiceBySuperMajority(ctx.VotingRecord, ctx.Members.Size())
	if success {
		return c == commondao.ChoiceYes, nil
	}
	return false, nil
}

func (p membersUpdatePropDefinition) Execute(realm) error {
	members := p.dao.Members()

	p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr address) bool {
		members.Add(addr)
		return false
	})

	p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr address) bool {
		members.Remove(addr)
		return false
	})

	return nil
}
