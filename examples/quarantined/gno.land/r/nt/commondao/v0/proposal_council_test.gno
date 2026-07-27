package commondao

import (
	"testing"

	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = (*councilUpdatePropDefinition)(nil)
	_ commondao.CapExempt          = (*councilUpdatePropDefinition)(nil)
)

func TestCouncilUpdatePropDefinitionNew(cur realm, t *testing.T) {
	var member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

	dao := commondao.New(commondao.WithCouncilMember(member))

	urequire.PanicsWithMessage(t, cur, "DAO is required", func() {
		newCouncilUpdatePropDefinition(nil, nil, nil)
	})

	urequire.PanicsWithMessage(t, cur, "no council members were specified to be added or removed", func() {
		newCouncilUpdatePropDefinition(dao, nil, nil)
	})

	def := newCouncilUpdatePropDefinition(dao, []address{member}, nil)

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
	def := newCouncilUpdatePropDefinition(dao, []address{memberA}, []address{memberB})
	uassert.NoError(t, def.Validate())

	// Adding and removing the same address is rejected
	def = newCouncilUpdatePropDefinition(dao, []address{memberB}, []address{memberB})
	uassert.ErrorContains(t, def.Validate(), "address is added and removed at once")
}

func TestCouncilUpdatePropDefinitionExecute(cur realm, t *testing.T) {
	var (
		memberA address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
		memberB address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
	)

	t.Run("add and remove", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))

		def := newCouncilUpdatePropDefinition(dao, []address{memberB}, []address{memberA})

		urequire.NotPanics(t, cur, func() {
			if err := def.Executor()(cross(cur)); err != nil {
				panic(err)
			}
		})
		uassert.True(t, dao.Council().Has(memberB))
		uassert.False(t, dao.Council().Has(memberA))
		uassert.Equal(t, 1, dao.Council().Size())
	})

	t.Run("emptying the council fails", func(t *testing.T) {
		dao := commondao.New(commondao.WithCouncilMember(memberA))

		def := newCouncilUpdatePropDefinition(dao, nil, []address{memberA})

		err := def.execute(cur)
		uassert.ErrorIs(t, err, commondao.ErrEmptyCouncil)
		uassert.Equal(t, 1, dao.Council().Size())
	})
}

func TestAncestorCouncilUpdatePropDefinition(cur realm, t *testing.T) {
	var member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

	root := commondao.New(commondao.WithID(201), commondao.WithName("Root"))
	mid := commondao.New(commondao.WithID(202), commondao.WithName("Mid"), commondao.WithParent(root))
	leaf := commondao.New(commondao.WithID(203), commondao.WithName("Leaf"), commondao.WithParent(mid), commondao.WithCouncilMember(member))

	urequire.PanicsWithMessage(t, cur, "DAO is required", func() {
		newAncestorCouncilUpdatePropDefinition(nil, leaf, []address{member}, nil)
	})
	urequire.PanicsWithMessage(t, cur, "target DAO is required", func() {
		newAncestorCouncilUpdatePropDefinition(root, nil, []address{member}, nil)
	})
	urequire.PanicsWithMessage(t, cur, "no council members were specified to be added or removed", func() {
		newAncestorCouncilUpdatePropDefinition(root, leaf, nil, nil)
	})

	def := newAncestorCouncilUpdatePropDefinition(root, leaf, []address{member}, nil)
	uassert.Equal(t, "Ancestor Council Update", def.Title())
	uassert.Equal(t, int(commondao.ThresholdSupermajority), int(def.Threshold()))
	uassert.NotEmpty(t, def.Body())

	// Any proper ancestor qualifies (parent or grandparent); self and
	// descendants do not.
	uassert.NoError(t, newAncestorCouncilUpdatePropDefinition(root, leaf, []address{member}, nil).Validate())
	uassert.NoError(t, newAncestorCouncilUpdatePropDefinition(mid, leaf, []address{member}, nil).Validate())
	uassert.ErrorContains(t,
		newAncestorCouncilUpdatePropDefinition(leaf, leaf, []address{member}, nil).Validate(),
		"a DAO cannot target itself")
	uassert.ErrorContains(t,
		newAncestorCouncilUpdatePropDefinition(leaf, root, []address{member}, nil).Validate(),
		"DAO is not an ancestor of the target DAO")

	// A contradictory same-address add+remove is rejected.
	uassert.ErrorContains(t,
		newAncestorCouncilUpdatePropDefinition(root, leaf, []address{member}, []address{member}).Validate(),
		"is added and removed at once")

	// The executor mutates the target's council.
	var newMember address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"
	urequire.NoError(t, newAncestorCouncilUpdatePropDefinition(root, leaf, []address{newMember}, []address{member}).Executor()(cross(cur)))
	uassert.True(t, leaf.Council().Has(newMember))
	uassert.False(t, leaf.Council().Has(member))
}
