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
	alicemember := std.DerivePkgAddr("gno.land/r/volos/gov/alicemember")
	bobmember := std.DerivePkgAddr("gno.land/r/volos/gov/bobmember")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alicemember, 1000)
		xvls.Mint(cross, bobmember, 500)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		AddMember(cross, alicemember)
		AddMember(cross, bobmember)
	})
	urequire.True(t, volosGovernance.Members().Has(alicemember))
	urequire.True(t, volosGovernance.Members().Has(bobmember))

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Burn(cross, bobmember, 500)
		RemoveMember(cross, bobmember)
	})
	urequire.False(t, volosGovernance.Members().Has(bobmember))

	crossThrough(std.NewCodeRealm(staker), func() {
		uassert.AbortsWithMessage(t, "cannot add member: address does not hold any xVLS", func() {
			AddMember(cross, bobmember)
		})
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		uassert.AbortsWithMessage(t, "cannot remove member: address still holds xVLS", func() {
			RemoveMember(cross, alicemember)
		})
	})

	notStaker := "gno.land/r/volos/gov/notstaker"
	crossThrough(std.NewCodeRealm(notStaker), func() {
		uassert.AbortsWithMessage(t, "only staker can call this function", func() {
			AddMember(cross, alicemember)
		})
	})

	crossThrough(std.NewCodeRealm(notStaker), func() {
		uassert.AbortsWithMessage(t, "only staker can call this function", func() {
			RemoveMember(cross, alicemember)
		})
	})
}

func TestMembers_EdgeCases(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alicemember := std.DerivePkgAddr("gno.land/r/volos/gov/alicemember")
	bobmember := std.DerivePkgAddr("gno.land/r/volos/gov/bobmember")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alicemember, 10000)
		AddMember(cross, alicemember)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		added := volosGovernance.Members().Add(alicemember)
		urequire.False(t, added)
	})

	removed := volosGovernance.Members().Remove(bobmember)
	urequire.False(t, removed)

	removed = volosGovernance.Members().Remove(alicemember)
	urequire.True(t, removed)
}
