package commondao

import (
	"chain"
	"testing"

	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = (*treasurySpendPropDefinition)(nil)
	_ commondao.DefaultTallier     = (*treasurySpendPropDefinition)(nil)
	_ commondao.ProposalDefinition = (*treasuryClawbackPropDefinition)(nil)
	_ commondao.DefaultTallier     = (*treasuryClawbackPropDefinition)(nil)
	_ commondao.ProposalDefinition = (*treasuryFreezePropDefinition)(nil)
	_ commondao.DefaultTallier     = (*treasuryFreezePropDefinition)(nil)
)

// newTestTree builds an in-memory root -> mid -> leaf DAO chain.
func newTestTree() (root, mid, leaf *commondao.CommonDAO) {
	root = commondao.New(commondao.WithID(101), commondao.WithName("Root"))
	mid = commondao.New(commondao.WithID(102), commondao.WithName("Mid"), commondao.WithParent(root))
	leaf = commondao.New(commondao.WithID(103), commondao.WithName("Leaf"), commondao.WithParent(mid))
	return
}

func TestTreasurySpendPropDefinition(cur realm, t *testing.T) {
	var member address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

	dao := commondao.New(
		commondao.WithID(100),
		commondao.WithName("Foo"),
		commondao.WithAddress(member), // any valid, unfunded address
	)
	coin := chain.NewCoin("ugnot", 100)

	urequire.PanicsWithMessage(t, cur, "DAO is required", func() {
		newTreasurySpendPropDefinition(nil, member, coin)
	})
	urequire.PanicsWithMessage(t, cur, "invalid recipient address", func() {
		newTreasurySpendPropDefinition(dao, "invalid", coin)
	})
	urequire.PanicsWithMessage(t, cur, "coin denomination is empty", func() {
		newTreasurySpendPropDefinition(dao, member, chain.Coin{Amount: 1})
	})
	urequire.PanicsWithMessage(t, cur, "spend amount must be positive", func() {
		newTreasurySpendPropDefinition(dao, member, chain.NewCoin("ugnot", 0))
	})

	def := newTreasurySpendPropDefinition(dao, member, coin)
	uassert.Equal(t, "Treasury Spend", def.Title())
	uassert.Equal(t, int(commondao.ThresholdSupermajority), int(def.Threshold()))
	uassert.NotEmpty(t, def.Body())

	// The treasury is unfunded
	uassert.ErrorContains(t, def.Validate(), "insufficient treasury balance")

	// Frozen treasuries reject spends before the balance check
	dao.SetTreasuryFrozen(true)
	uassert.ErrorContains(t, def.Validate(), "DAO treasury is frozen")
	dao.SetTreasuryFrozen(false)

	// Dissolved DAOs reject spends
	dao.SetDeleted(true)
	uassert.ErrorContains(t, def.Validate(), "DAO has already been dissolved")
}

func TestTreasuryClawbackPropDefinition(cur realm, t *testing.T) {
	root, mid, leaf := newTestTree()

	urequire.PanicsWithMessage(t, cur, "DAO is required", func() {
		newTreasuryClawbackPropDefinition(nil, leaf)
	})
	urequire.PanicsWithMessage(t, cur, "target DAO is required", func() {
		newTreasuryClawbackPropDefinition(root, nil)
	})

	def := newTreasuryClawbackPropDefinition(root, leaf)
	uassert.Equal(t, "Treasury Clawback", def.Title())
	uassert.Equal(t, int(commondao.ThresholdSimpleMajority), int(def.Threshold()))
	uassert.NotEmpty(t, def.Body())

	// Any proper ancestor qualifies, not just the direct parent
	uassert.NoError(t, newTreasuryClawbackPropDefinition(root, mid).Validate())
	uassert.NoError(t, newTreasuryClawbackPropDefinition(root, leaf).Validate())
	uassert.NoError(t, newTreasuryClawbackPropDefinition(mid, leaf).Validate())

	// A DAO can never target itself, and descendants and strangers have
	// no clawback authority
	uassert.ErrorContains(t,
		newTreasuryClawbackPropDefinition(root, root).Validate(),
		"a DAO cannot target its own treasury")
	uassert.ErrorContains(t,
		newTreasuryClawbackPropDefinition(leaf, root).Validate(),
		"DAO is not an ancestor of the target DAO")
	uassert.ErrorContains(t,
		newTreasuryClawbackPropDefinition(commondao.New(commondao.WithID(999)), leaf).Validate(),
		"DAO is not an ancestor of the target DAO")
}

func TestTreasuryFreezePropDefinition(cur realm, t *testing.T) {
	root, mid, leaf := newTestTree()

	urequire.PanicsWithMessage(t, cur, "DAO is required", func() {
		newTreasuryFreezePropDefinition(nil, leaf, true)
	})
	urequire.PanicsWithMessage(t, cur, "target DAO is required", func() {
		newTreasuryFreezePropDefinition(root, nil, true)
	})

	freeze := newTreasuryFreezePropDefinition(root, leaf, true)
	uassert.Equal(t, "Treasury Freeze", freeze.Title())
	uassert.Equal(t, int(commondao.ThresholdSimpleMajority), int(freeze.Threshold()))
	uassert.NotEmpty(t, freeze.Body())

	unfreeze := newTreasuryFreezePropDefinition(root, leaf, false)
	uassert.Equal(t, "Treasury Unfreeze", unfreeze.Title())
	uassert.NotEmpty(t, unfreeze.Body())

	uassert.NoError(t, freeze.Validate())
	uassert.NoError(t, newTreasuryFreezePropDefinition(mid, leaf, true).Validate())
	uassert.ErrorContains(t,
		newTreasuryFreezePropDefinition(leaf, leaf, false).Validate(),
		"a DAO cannot target its own treasury")
	uassert.ErrorContains(t,
		newTreasuryFreezePropDefinition(leaf, mid, true).Validate(),
		"DAO is not an ancestor of the target DAO")
}
