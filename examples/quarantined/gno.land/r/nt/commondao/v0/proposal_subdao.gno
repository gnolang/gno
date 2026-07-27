package commondao

import (
	"chain/banker"
	"errors"
	"strings"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/commondao/v0"
)

// newSubDAOPropDefinition creates a new proposal definition for adding a SubDAO.
func newSubDAOPropDefinition(parent *commondao.CommonDAO, name, description string, members []address) subDAOPropDefinition {
	if parent == nil {
		panic("parent DAO is required")
	}

	name = strings.TrimSpace(name)
	assertDAONameIsValid(name)

	description = strings.TrimSpace(description)
	assertDAODescriptionIsValid(description)

	if len(members) == 0 {
		panic("a SubDAO requires at least one initial council member")
	}

	return subDAOPropDefinition{
		parent:      parent,
		name:        name,
		description: description,
		members:     members,
	}
}

// subDAOPropDefinition defines a proposal type for adding a SubDAO.
type subDAOPropDefinition struct {
	parent      *commondao.CommonDAO
	name        string
	description string
	members     []address
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
		md.Bold("Parent DAO:") + "\n" + daoMDLink(p.parent),
	))

	b.WriteString(md.Paragraph(
		md.Bold("SubDAO Name:") + "\n" + md.EscapeText(p.name),
	))

	if p.description != "" {
		b.WriteString(md.Paragraph(
			md.Bold("SubDAO Description:") + "\n" + md.EscapeText(p.description),
		))
	}

	b.WriteString(md.Paragraph(
		md.Bold("Council Members:") + "\n" + md.BulletList(addrStrings(p.members)),
	))

	return b.String()
}

func (p subDAOPropDefinition) Validate() (err error) {
	p.parent.IterateChildren(func(subDAO *commondao.CommonDAO) bool {
		if subDAO.Name() == p.name {
			err = errors.New("a SubDAO with the same name already exists")
			return true
		}
		return false
	})
	return err
}

func (p subDAOPropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

func (p subDAOPropDefinition) execute(realm) error {
	createSubDAO(p.parent, p.name, p.description, p.members...)
	return nil
}

// newDissolvePropDefinition creates a new proposal definition for
// dissolving a DAO. Dissolution sweeps any remaining treasury balance:
// sub-DAO sweeps go to the parent (fixed, not nameable — the same place
// a clawback would put the funds); a root DAO has no parent, so its
// dissolution requires an explicit destination. Both rules are
// state-independent: parent pointers never change after construction.
func newDissolvePropDefinition(dao *commondao.CommonDAO, destination address) dissolvePropDefinition {
	if dao == nil {
		panic("SubDAO is required")
	}
	if dao.Parent() != nil {
		if destination != "" {
			panic("sub-DAO dissolution sweeps to the parent DAO; destination must be empty")
		}
	} else {
		if !destination.IsValid() {
			panic("root DAO dissolution requires a valid sweep destination")
		}
	}

	return dissolvePropDefinition{dao, destination}
}

// dissolvePropDefinition defines a proposal type for dissolving a SubDAO.
type dissolvePropDefinition struct {
	dao         *commondao.CommonDAO
	destination address // sweep destination, root DAOs only
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
	b.WriteString(daoMDLink(p.dao))

	return b.String()
}

func (p dissolvePropDefinition) Validate() (err error) {
	if p.dao.IsDeleted() {
		return errors.New("DAO has already been dissolved")
	}
	return nil
}

func (p dissolvePropDefinition) Executor() commondao.ExecFunc {
	return p.execute
}

func (p dissolvePropDefinition) execute(cur realm) error {
	// Sweep any remaining treasury before soft deleting: no council
	// remains afterwards to pass a spend. Sub-DAO funds go to the parent;
	// root funds go to the destination named at proposal creation. A
	// frozen treasury does not block dissolution — the sweep sends the
	// funds where a clawback would.
	if balance := treasuryBalance(p.dao); !balance.IsZero() {
		dest := p.destination
		if parent := p.dao.Parent(); parent != nil {
			dest = parent.Address()
		}

		sub := cur.Sub(subpathOf(p.dao.ID()))
		b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
		b.SendCoins(sub.Address(), dest, balance)
	}

	// Drop the DAO from the home index: a dissolved DAO can no longer be
	// unlisted through SetListed (it rejects deleted DAOs), so unlist it
	// here rather than leave it stuck in the listing.
	setListed(p.dao.ID(), false)

	// Dissolution voids the DAO's in-flight proposals before soft deleting
	// it, so nothing remains pending on a DAO that rejects execution.
	p.dao.Dissolve("DAO dissolved")
	return nil
}
