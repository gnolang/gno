package commondao

import (
	"errors"
	"std"
	"strings"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/commondao"
)

// TODO: Support optional parent DAO mechanics to add/remove direct SubDAO members

// NewMembersPropDefinition creates a new proposal definition for adding/removing DAO members.
func NewMembersPropDefinition(dao *commondao.CommonDAO, add, remove addrset.Set) MembersPropDefinition {
	if dao == nil {
		panic("DAO is required")
	}

	if dao.Members().Size() == 0 {
		panic("a DAO with at least one member is required to create member update proposals")
	}

	if add.Size() == 0 && remove.Size() == 0 {
		panic("no members were specified to be added or removed")
	}

	return MembersPropDefinition{
		dao:      dao,
		toAdd:    add,
		toRemove: remove,
	}
}

// MembersUpdateProposal defines a proposal type for adding/removing DAO members.
type MembersPropDefinition struct {
	dao             *commondao.CommonDAO
	toAdd, toRemove addrset.Set
}

func (MembersPropDefinition) Title() string               { return "Members Update" }
func (MembersPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

func (p MembersPropDefinition) Body() string {
	var b strings.Builder

	if p.toAdd.Size() > 0 {
		b.WriteString("**Members to Add:**\n")
		p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr std.Address) bool {
			b.WriteString("- ")
			b.WriteString(addr.String())
			b.WriteString("\n")
			return false
		})
	}

	if p.toRemove.Size() > 0 {
		b.WriteString("\n**Members to Remove:**\n")
		p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr std.Address) bool {
			b.WriteString("- ")
			b.WriteString(addr.String())
			b.WriteString("\n")
			return false
		})
	}

	return b.String()
}

func (p MembersPropDefinition) Validate() error {
	if p.dao == nil {
		return ErrInvalidProposal
	}

	var (
		err     error
		members = p.dao.Members()
	)

	p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr std.Address) bool {
		if members.Has(addr) {
			err = errors.New("address is already a DAO member: " + addr.String())
			return true
		}
		return false
	})

	if err != nil {
		return err
	}

	p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr std.Address) bool {
		if !members.Has(addr) {
			err = errors.New("address is not a DAO member: " + addr.String())
			return true
		}
		return false
	})

	return err
}

func (MembersPropDefinition) Tally(r commondao.ReadonlyVotingRecord, members commondao.MemberSet) (bool, error) {
	// When DAO has one or two members succeed when there is a YES vote, otherwise
	// tally requires at least three votes to be able to tally by 2/3s super majority
	if members.Size() < 3 {
		return r.VoteCount(commondao.ChoiceYes) > 0, nil
	}

	if !commondao.IsQuorumReached(commondao.QuorumTwoThirds, r, members) {
		return false, commondao.ErrNoQuorum
	}

	c, success := commondao.SelectChoiceBySuperMajority(r, members.Size())
	if success {
		return c == commondao.ChoiceYes, nil
	}
	return false, nil
}

func (p MembersPropDefinition) Execute() error {
	assertIsValidProposalDefinition(p)

	members := p.dao.Members()

	p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr std.Address) bool {
		members.Add(addr)
		return false
	})

	p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr std.Address) bool {
		members.Remove(addr)
		return false
	})

	return nil
}
