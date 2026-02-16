package commondao

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/moul/md"
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

	b.WriteString(md.Paragraph(
		md.Bold("Parent DAO:") + "\n" + md.Link(p.parent.Name(), daoURL(p.parent.ID())),
	))

	b.WriteString(md.Paragraph(
		md.Bold("SubDAO Name:") + "\n" + p.name,
	))

	if p.members != nil && p.members.Size() > 0 {
		items := make([]string, 0, p.members.Size())
		p.members.IterateByOffset(0, p.members.Size(), func(addr address) bool {
			items = append(items, addr.String())
			return false
		})

		b.WriteString(md.Paragraph(
			md.Bold("Members:") + "\n" + md.BulletList(items),
		))
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

func (subDAOPropDefinition) Tally(ctx commondao.VotingContext) (bool, error) {
	if !commondao.IsQuorumReached(commondao.QuorumFull, ctx.VotingRecord, ctx.Members) {
		return false, commondao.ErrNoQuorum
	}

	c, success := commondao.SelectChoiceBySuperMajority(ctx.VotingRecord, ctx.Members.Size())
	if success {
		return c == commondao.ChoiceYes, nil
	}
	return false, nil
}

func (p subDAOPropDefinition) Execute(realm) error {
	subDAO := createSubDAO(p.parent, p.name, nil)

	if p.members != nil {
		p.members.IterateByOffset(0, p.members.Size(), func(addr address) bool {
			subDAO.Members().Add(addr)
			return false
		})
	}

	p.parent.Children().Append(subDAO)
	return nil
}

// newDissolvePropDefinition creates a new proposal definition for dissolving a SubDAO.
func newDissolvePropDefinition(dao *commondao.CommonDAO) dissolvePropDefinition {
	if dao == nil {
		panic("SubDAO is required")
	}

	return dissolvePropDefinition{dao}
}

// dissolvePropDefinition defines a proposal type for dissolving a SubDAO.
type dissolvePropDefinition struct {
	dao *commondao.CommonDAO
}

func (p dissolvePropDefinition) Title() string             { return "Dissolve DAO: " + p.dao.Name() }
func (dissolvePropDefinition) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

func (p dissolvePropDefinition) Body() string {
	var b strings.Builder

	b.WriteString(md.Bold("DAO:") + "\n")
	b.WriteString(md.Link(p.dao.Name(), daoURL(p.dao.ID())))

	return b.String()
}

func (p dissolvePropDefinition) Validate() (err error) {
	if p.dao.IsDeleted() {
		return errors.New("DAO has been already dissolveed")
	}
	return nil
}

func (dissolvePropDefinition) Tally(ctx commondao.VotingContext) (bool, error) {
	if !commondao.IsQuorumReached(commondao.QuorumFull, ctx.VotingRecord, ctx.Members) {
		return false, commondao.ErrNoQuorum
	}

	c, success := commondao.SelectChoiceBySuperMajority(ctx.VotingRecord, ctx.Members.Size())
	if success {
		return c == commondao.ChoiceYes, nil
	}
	return false, nil
}

func (p dissolvePropDefinition) Execute(realm) error {
	p.dao.SetDeleted(true)
	return nil
}
