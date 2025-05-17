package commondao

import (
	"errors"
	"std"
	"strings"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/commondao"
)

// newSubDAOPropDefinition creates a new proposal definition for adding a SubDAO.
func newSubDAOPropDefinition(parent *commondao.CommonDAO, name string, members *addrset.Set) subDAOPropDefinition {
	if parent == nil {
		panic("parent DAO is required")
	}

	name = strings.TrimSpace(name)
	assertDAONameIsValid(name)

	return subDAOPropDefinition{
		parent:  parent,
		name:    name,
		members: members,
	}
}

// subDAOPropDefinition defines a proposal type for adding a SubDAO.
type subDAOPropDefinition struct {
	parent  *commondao.CommonDAO
	name    string
	members *addrset.Set
}

func (p subDAOPropDefinition) Title() string             { return "New SubDAO: " + p.name }
func (subDAOPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

func (p subDAOPropDefinition) Body() string {
	var b strings.Builder

	b.WriteString("**Parent DAO:** ")
	b.WriteString("[" + p.parent.Name() + "](" + daoURL(p.parent) + ")\n")

	b.WriteString("**SubDAO Name:** ")
	b.WriteString(p.name)
	b.WriteString("\n")

	if p.members != nil && p.members.Size() > 0 {
		b.WriteString("**Members:**\n")
		p.members.IterateByOffset(0, p.members.Size(), func(addr std.Address) bool {
			b.WriteString("- ")
			b.WriteString(addr.String())
			b.WriteString("\n")
			return false
		})
	}

	return b.String()
}

func (p subDAOPropDefinition) Validate() (err error) {
	options := getOptions(p.parent.ID())
	if !(options.AllowSubDAOProposals && options.AllowChildren) {
		return errors.New("SubDAO support is not enabled")
	}

	p.parent.Children().ForEach(func(_ int, v any) bool {
		subDAO := v.(*commondao.CommonDAO)
		if subDAO.Name() == p.name {
			err = errors.New("a SubDAO with the same name already exists")
			return true
		}
		return false
	})
	return err
}

func (subDAOPropDefinition) Tally(r commondao.ReadonlyVotingRecord, members commondao.MemberSet) (bool, error) {
	if !commondao.IsQuorumReached(commondao.QuorumFull, r, members) {
		return false, commondao.ErrNoQuorum
	}

	c, success := commondao.SelectChoiceBySuperMajority(r, members.Size())
	if success {
		return c == commondao.ChoiceYes, nil
	}
	return false, nil
}

func (p subDAOPropDefinition) Execute() error {
	crossing()

	subDAO := createSubDAO(p.parent, p.name)

	if p.members != nil {
		p.members.IterateByOffset(0, p.members.Size(), func(addr std.Address) bool {
			subDAO.Members().Add(addr)
			return false
		})
	}

	p.parent.Children().Append(subDAO)
	return nil
}

// newDismissPropDefinition creates a new proposal definition for dismissing a SubDAO.
func newDismissPropDefinition(dao *commondao.CommonDAO) dismissPropDefinition {
	if dao == nil {
		panic("SubDAO is required")
	}

	return dismissPropDefinition{dao}
}

// dismissPropDefinition defines a proposal type for dismissing a SubDAO.
type dismissPropDefinition struct {
	dao *commondao.CommonDAO
}

func (p dismissPropDefinition) Title() string             { return "Dismiss DAO: " + p.dao.Name() }
func (dismissPropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

func (p dismissPropDefinition) Body() string {
	var b strings.Builder

	b.WriteString("**DAO:** ")
	b.WriteString("[" + p.dao.Name() + "](" + daoURL(p.dao) + ")\n")

	return b.String()
}

func (p dismissPropDefinition) Validate() (err error) {
	if p.dao.IsDeleted() {
		return errors.New("DAO has been already dismissed")
	}
	return nil
}

func (dismissPropDefinition) Tally(r commondao.ReadonlyVotingRecord, members commondao.MemberSet) (bool, error) {
	if !commondao.IsQuorumReached(commondao.QuorumFull, r, members) {
		return false, commondao.ErrNoQuorum
	}

	c, success := commondao.SelectChoiceBySuperMajority(r, members.Size())
	if success {
		return c == commondao.ChoiceYes, nil
	}
	return false, nil
}

func (p dismissPropDefinition) Execute() error {
	crossing()

	p.dao.SetDeleted(true)
	return nil
}
