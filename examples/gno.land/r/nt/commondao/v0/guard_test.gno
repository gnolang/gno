package commondao

import (
	"testing"

	"gno.land/p/nt/bylaws/v0"
	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
)

// TestHelpersRefuseForeignSub calls every crossing write helper as another
// DAO's sub. Each must abort in assertExecutor before it writes; revive
// catches the abort, which leaves a cross call.
func TestHelpersRefuseForeignSub(cur realm, t *testing.T) {
	const want = "commondao: caller is not the DAO's executor"
	var member address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	var newcomer address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"

	dao := commondao.New(commondao.WithID(401), commondao.WithCouncilMember(member))
	target := commondao.New(commondao.WithID(402), commondao.WithParent(dao))
	set := bylaws.New()
	patch := bylaws.Patch{Path: "bylaws/a.md", Ops: []bylaws.Op{{Type: bylaws.OpInsert, Text: "x"}}}
	foreign := cur.Sub(subpathOf(499))

	expectRefused(t, "updateCouncil", want, revive(func() {
		updateCouncil(cross(foreign), dao, dao, []address{newcomer}, nil)
	}))
	expectRefused(t, "setFrozen", want, revive(func() {
		setFrozen(cross(foreign), dao, target, true)
	}))
	expectRefused(t, "manageKind", want, revive(func() {
		manageKind(cross(foreign), dao, false, "execution")
	}))
	expectRefused(t, "addSubDAO", want, revive(func() {
		addSubDAO(cross(foreign), dao, "Child", "", "", nil)
	}))
	expectRefused(t, "dissolveDAO", want, revive(func() {
		dissolveDAO(cross(foreign), dao)
	}))
	expectRefused(t, "applyBylaws", want, revive(func() {
		applyBylaws(cross(foreign), dao.ID(), set, patch)
	}))

	// State each guard protects is untouched.
	uassert.False(t, dao.Council().Has(newcomer), "council")
	uassert.False(t, target.IsTreasuryFrozen(), "frozen")
	uassert.False(t, dao.HasKind("execution"), "kind")
	uassert.Equal(t, 1, dao.ChildrenCount(), "children")
	uassert.False(t, dao.IsDeleted(), "deleted")
	uassert.False(t, set.Has("bylaws/a.md"), "bylaws")
}

func expectRefused(t *testing.T, name, want string, got any) {
	t.Helper()
	if s, _ := got.(string); s != want {
		t.Errorf("%s: revive gave %v, want %q", name, got, want)
	}
}

// TestExecutorsRefuseStaleSub hands every executor that writes through a
// helper a sub of the right DAO that is no longer current. Each must stop
// in assertSub, before the VM's own cross check, and write nothing.
func TestExecutorsRefuseStaleSub(cur realm, t *testing.T) {
	const want = "commondao: executor sub is not current"
	var member address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	var newcomer address = "g1manfred47kzduec920z88wfr64ylksmdcedlf5"

	dao := commondao.New(commondao.WithID(501), commondao.WithCouncilMember(member))
	target := commondao.New(commondao.WithID(502), commondao.WithParent(dao), commondao.WithCouncilMember(member))
	set := bylaws.New()
	patch := bylaws.Patch{Path: "bylaws/a.md", Ops: []bylaws.Op{{Type: bylaws.OpInsert, Text: "x"}}}

	stale := mintSub(cross(cur), dao.ID())
	expectRefused(t, "councilUpdate", want, refusal(func() {
		newCouncilUpdatePropDefinition(dao, []address{newcomer}, nil).execute(0, stale)
	}))
	expectRefused(t, "ancestorCouncilUpdate", want, refusal(func() {
		newAncestorCouncilUpdatePropDefinition(dao, target, []address{newcomer}, nil).execute(0, stale)
	}))
	expectRefused(t, "subDAO", want, refusal(func() {
		newSubDAOPropDefinition(dao, "Child", "A purpose", "", []address{member}).execute(0, stale)
	}))
	expectRefused(t, "dissolve", want, refusal(func() {
		newDissolvePropDefinition(dao, member).execute(0, stale)
	}))
	expectRefused(t, "treasuryFreeze", want, refusal(func() {
		newTreasuryFreezePropDefinition(dao, target, true).execute(0, stale)
	}))
	expectRefused(t, "manageKinds", want, refusal(func() {
		manageKindsProposal{dao: dao, name: "execution"}.execute(0, stale)
	}))
	expectRefused(t, "amendBylaws", want, refusal(func() {
		amendBylawsProposal{daoID: dao.ID(), set: set, patch: patch}.execute(0, stale)
	}))

	uassert.False(t, dao.Council().Has(newcomer), "council")
	uassert.False(t, target.Council().Has(newcomer), "ancestor council")
	uassert.Equal(t, 1, dao.ChildrenCount(), "children")
	uassert.False(t, dao.IsDeleted(), "deleted")
	uassert.False(t, target.IsTreasuryFrozen(), "frozen")
	uassert.False(t, dao.HasKind("execution"), "kind")
	uassert.False(t, set.Has("bylaws/a.md"), "bylaws")
}

// refusal returns fn's panic value: recover catches a plain panic such as
// assertSub's, revive one that leaves a cross call such as the VM's.
func refusal(fn func()) (got any) {
	ex := revive(func() { got = recovered(fn) })
	if ex != nil {
		got = ex
	}
	return got
}

func recovered(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}
