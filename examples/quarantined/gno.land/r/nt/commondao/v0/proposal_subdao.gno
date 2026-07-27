package commondao

import (
	"errors"
	"strings"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao/v0"
)

// newSubDAOPropDefinition creates a new proposal definition for adding a SubDAO.
func newSubDAOPropDefinition(parent *commondao.CommonDAO, name string, members *addrset.Set) subDAOPropDefinition {
	if parent == nil {
		panic("parent DAO is required")
	}

	name = strings.TrimSpace(name)
	assertDAONameIsValid(name)

	if members == nil || members.Size() == 0 {
		panic("a SubDAO requires at least one initial council member")
	}

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

// Threshold returns the tally threshold: the constitution grants SubDAO
// creation at a simple majority of the Council.
func (subDAOPropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSimpleMajority
}

func (p subDAOPropDefinition) Body() string {
	var b strings.Builder

	b.WriteString(md.Paragraph(
		md.Bold("Parent DAO:") + "\n" + md.Link(p.parent.Name(), daoURL(p.parent.ID())),
	))

	b.WriteString(md.Paragraph(
		md.Bold("SubDAO Name:") + "\n" + p.name,
	))

	items := make([]string, 0, p.members.Size())
	p.members.IterateByOffset(0, p.members.Size(), func(addr address) bool {
		items = append(items, addr.String())
		return false
	})

	b.WriteString(md.Paragraph(
		md.Bold("Council Members:") + "\n" + md.BulletList(items),
	))

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
	outcome := commondao.TallyDefault(ctx.VotingRecord, ctx.Council, commondao.ThresholdSimpleMajority)
	return outcome == commondao.OutcomePassed, nil
}

func (p subDAOPropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

func (p subDAOPropDefinition) execute(realm) error {
	members := make([]address, 0, p.members.Size())
	p.members.IterateByOffset(0, p.members.Size(), func(addr address) bool {
		members = append(members, addr)
		return false
	})

	subDAO := createSubDAO(p.parent, p.name, nil, members...)
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

// Threshold returns the tally threshold: the constitution is silent on
// dissolution, so the supermajority default applies.
func (dissolvePropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}

func (p dissolvePropDefinition) Body() string {
	var b strings.Builder

	b.WriteString(md.Bold("DAO:") + "\n")
	b.WriteString(md.Link(p.dao.Name(), daoURL(p.dao.ID())))

	return b.String()
}

func (p dissolvePropDefinition) Validate() (err error) {
	if p.dao.IsDeleted() {
		return errors.New("DAO has been already dissolved")
	}
	return nil
}

func (dissolvePropDefinition) Tally(ctx commondao.VotingContext) (bool, error) {
	outcome := commondao.TallyDefault(ctx.VotingRecord, ctx.Council, commondao.ThresholdSupermajority)
	return outcome == commondao.OutcomePassed, nil
}

func (p dissolvePropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

func (p dissolvePropDefinition) execute(realm) error {
	// Dissolution voids the DAO's in-flight proposals before soft deleting
	// it, so nothing remains pending on a DAO that rejects execution.
	p.dao.Dissolve("DAO dissolved")
	return nil
}
