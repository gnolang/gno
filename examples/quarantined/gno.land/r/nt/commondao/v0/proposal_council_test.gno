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
