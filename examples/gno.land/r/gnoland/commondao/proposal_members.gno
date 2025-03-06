package commondao

import (
	"errors"
	"std"
	"strings"
	"time"

	"gno.land/p/gnoland/commondao"
	"gno.land/p/moul/addrset"
)

// TODO: Support optional parent DAO mechanics to add/remove direct SubDAO members

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

type MembersPropDefinition struct {
	dao             *commondao.CommonDAO
	toAdd, toRemove addrset.Set
}

func (MembersPropDefinition) Title() string               { return "Members Update" }
func (MembersPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }
func (MembersPropDefinition) Quorum() float64             { return 0.67 } // two thirds

func (p MembersPropDefinition) Body() string {
	var b strings.Builder

	if p.toAdd.Size() > 0 {
		b.WriteString("**Members to Add:**  \n")
		p.toAdd.IterateByOffset(0, p.toAdd.Size(), func(addr std.Address) bool {
			b.WriteString("- " + addr.String() + "\n")
			return false
		})
	}

	if p.toRemove.Size() > 0 {
		b.WriteString("\n**Members to Remove:**  \n")
		p.toRemove.IterateByOffset(0, p.toRemove.Size(), func(addr std.Address) bool {
			b.WriteString("- " + addr.String() + "\n")
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

func (MembersPropDefinition) Tally(r commondao.ReadOnlyVotingRecord, membersCount int) bool {
	// When DAO has one or two members succeed when there is a YES vote, otherwise
	// tally requires at least three votes to be able to tally by 2/3s super majority
	if membersCount < 3 {
		return r.VoteCount(commondao.ChoiceYes) > 0
	}

	_, success := commondao.SelectChoiceBySuperMajority(r, membersCount)
	return success
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
