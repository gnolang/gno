package commondao

import (
	"testing"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = (*subDAOPropDefinition)(nil)
	_ commondao.DefaultTallier     = (*subDAOPropDefinition)(nil)
	_ commondao.ProposalDefinition = (*dissolvePropDefinition)(nil)
	_ commondao.DefaultTallier     = (*dissolvePropDefinition)(nil)
)

// newTestDAO registers a DAO in the realm state so definitions that read
// realm options can be tested.
func newTestDAO(t *testing.T, name string, members ...address) *commondao.CommonDAO {
	t.Helper()

	id := daoID.Next()
	opts := []commondao.Option{
		commondao.WithID(uint64(id)),
		commondao.WithName(name),
	}
	for _, m := range members {
		opts = append(opts, commondao.WithCouncilMember(m))
	}

	dao := commondao.New(opts...)
	daos.Set(id.String(), dao)
	options.Set(id.String(), &Options{
		AllowChildren: true,
	})
	return dao
}

func TestSubDAOPropDefinitionNew(cur realm, t *testing.T) {
	var member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

	dao := newTestDAO(t, "Root", member)

	var members addrset.Set
	members.Add(member)

	urequire.PanicsWithMessage(t, cur, "parent DAO is required", func() {
		newSubDAOPropDefinition(nil, "Sub", "", &members)
	})

	urequire.PanicsWithMessage(t, cur, "DAO name is empty", func() {
		newSubDAOPropDefinition(dao, "  ", "", &members)
	})

	urequire.PanicsWithMessage(t, cur, "a SubDAO requires at least one initial council member", func() {
		newSubDAOPropDefinition(dao, "Sub", "", &addrset.Set{})
	})

	def := newSubDAOPropDefinition(dao, "Sub", "", &members)
	uassert.Equal(t, "New SubDAO: Sub", def.Title())
	uassert.Equal(t, int(commondao.ThresholdSimpleMajority), int(def.Threshold()))
	uassert.NotEmpty(t, def.Body())
	uassert.NoError(t, def.Validate())
}

func TestSubDAOPropDefinitionExecute(cur realm, t *testing.T) {
	var member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

	dao := newTestDAO(t, "Root2", member)

	var members addrset.Set
	members.Add(member)
	def := newSubDAOPropDefinition(dao, "Sub", "", &members)

	urequire.NoError(t, def.Executor()(cross(cur)))

	urequire.Equal(t, 1, dao.ChildrenCount(), "expect one SubDAO")
	var sub *commondao.CommonDAO
	dao.IterateChildren(func(child *commondao.CommonDAO) bool {
		sub = child
		return true
	})
	uassert.Equal(t, "Sub", sub.Name())
	uassert.True(t, sub.Council().Has(member))
	uassert.Equal(t, 1, sub.Council().Size())

	// A second SubDAO with the same name is rejected at validation
	uassert.ErrorContains(t, def.Validate(), "a SubDAO with the same name already exists")
}

func TestDissolvePropDefinitionExecute(cur realm, t *testing.T) {
	var member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

	dao := newTestDAO(t, "Root3", member)

	urequire.PanicsWithMessage(t, cur, "SubDAO is required", func() {
		newDissolvePropDefinition(nil, "")
	})

	def := newDissolvePropDefinition(dao, member)
	uassert.Equal(t, "Dissolve DAO: Root3", def.Title())
	uassert.Equal(t, int(commondao.ThresholdSupermajority), int(def.Threshold()))
	uassert.NoError(t, def.Validate())

	// Dissolution voids in-flight proposals before soft deleting the DAO
	p, err := dao.Propose(member, textPropDefinition{
		title:        "Pending",
		body:         "pending",
		votingPeriod: time.Hour,
	})
	urequire.NoError(t, err)

	urequire.NoError(t, def.Executor()(cross(cur)))

	uassert.True(t, dao.IsDeleted())
	uassert.Equal(t, 0, dao.ActiveProposalsSize())
	uassert.Equal(t, string(commondao.StatusDismissed), string(p.Status()))
	uassert.Equal(t, "DAO dissolved", p.StatusReason())

	// Dissolved DAOs cannot be dissolved again
	uassert.ErrorContains(t, def.Validate(), "has already been dissolved")
}
