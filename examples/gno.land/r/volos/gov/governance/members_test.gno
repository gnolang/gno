package governance

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/r/volos/gov/xvls"
)

func TestMembership_AddRemove(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		xvls.Mint(cross, bob, 500)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		AddMember(cross, alice)
		AddMember(cross, bob)
	})
	urequire.True(t, volosGovernance.Members().Has(alice))
	urequire.True(t, volosGovernance.Members().Has(bob))

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Burn(cross, bob, 500)
		RemoveMember(cross, bob)
	})
	urequire.False(t, volosGovernance.Members().Has(bob))

	crossThrough(std.NewCodeRealm(staker), func() {
		uassert.AbortsWithMessage(t, "cannot add member: address does not hold any xVLS", func() {
			AddMember(cross, bob)
		})
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		uassert.AbortsWithMessage(t, "cannot remove member: address still holds xVLS", func() {
			RemoveMember(cross, alice)
		})
	})

	notStaker := "gno.land/r/volos/gov/notstaker"
	crossThrough(std.NewCodeRealm(notStaker), func() {
		uassert.AbortsWithMessage(t, "only staker can call this function", func() {
			AddMember(cross, alice)
		})
	})

	crossThrough(std.NewCodeRealm(notStaker), func() {
		uassert.AbortsWithMessage(t, "only staker can call this function", func() {
			RemoveMember(cross, alice)
		})
	})
}

func TestMembers_EdgeCases(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 10000)
		AddMember(cross, alice)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		added := volosGovernance.Members().Add(alice)
		urequire.False(t, added)
	})

	removed := volosGovernance.Members().Remove(bob)
	urequire.False(t, removed)

	removed = volosGovernance.Members().Remove(alice)
	urequire.True(t, removed)
}
