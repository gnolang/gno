package commondao

import (
	"testing"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = (*councilUpdatePropDefinition)(nil)
	_ commondao.DefaultTallier     = (*councilUpdatePropDefinition)(nil)
	_ commondao.CapExempt          = (*councilUpdatePropDefinition)(nil)
)

func TestCouncilUpdatePropDefinitionNew(cur realm, t *testing.T) {
	var member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

	dao := commondao.New(commondao.WithCouncilMember(member))

	urequire.PanicsWithMessage(t, cur, "DAO is required", func() {
		newCouncilUpdatePropDefinition(nil, addrset.Set{}, addrset.Set{})
	})

	urequire.PanicsWithMessage(t, cur, "no council members were specified to be added or removed", func() {
		newCouncilUpdatePropDefinition(dao, addrset.Set{}, addrset.Set{})
	})

	var add addrset.Set
	add.Add(member)
	def := newCouncilUpdatePropDefinition(dao, add, addrset.Set{})

	uassert.Equal(t, "Council Update", def.Title())
	uassert.Equal(t, int(commondao.ThresholdSupermajority), int(def.Threshold()))
	uassert.NotEmpty(t, def.Body())
}

func TestCouncilUpdatePropDefinitionValidate(t *testing.T) {
	var (
		memberA address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
		memberB address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
	)

	dao := commondao.New(commondao.WithCouncilMember(memberA))

	// Membership is not validated: adding an existing member or removing a
	// stranger is legal (idempotent set semantics)
	var add, remove addrset.Set
	add.Add(memberA)
	remove.Add(memberB)
	def := newCouncilUpdatePropDefinition(dao, add, remove)
	uassert.NoError(t, def.Validate())

	// Adding and removing the same address is rejected
	var overlap addrset.Set
	overlap.Add(memberB)
	def = newCouncilUpdatePropDefinition(dao, overlap, overlap)
	uassert.ErrorContains(t, def.Validate(), "address is added and removed at once")
}

func TestCouncilUpdatePropDefinitionExecute(cur realm, t *testing.T) {
	var (
		memberA address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
		memberB address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
	)

	t.Run("add and remove", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))

		var add, remove addrset.Set
		add.Add(memberB)
		remove.Add(memberA)
		def := newCouncilUpdatePropDefinition(dao, add, remove)

		urequire.NotPanics(t, cur, func() {
			commondao.MustExecute(def, cur)
		})
		uassert.True(t, dao.Council().Has(memberB))
		uassert.False(t, dao.Council().Has(memberA))
		uassert.Equal(t, 1, dao.Council().Size())
	})

	t.Run("emptying the council fails", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))

		var remove addrset.Set
		remove.Add(memberA)
		def := newCouncilUpdatePropDefinition(dao, addrset.Set{}, remove)

		err := def.execute(cur)
		uassert.ErrorIs(t, err, commondao.ErrEmptyCouncil)
		uassert.Equal(t, 1, dao.Council().Size())
	})
}
