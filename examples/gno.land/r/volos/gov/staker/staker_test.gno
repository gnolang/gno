package staker

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/r/volos/gov/vls"
	"gno.land/r/volos/gov/xvls"
	"gno.land/r/volos/gov/governance"
)

func crossThrough(rlm std.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}

func TestStakeToSelfAndDelegate(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice")
	bob := std.DerivePkgAddr("bob")

	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)

	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)
		Stake(cross, 600, alice)
	})
	urequire.Equal(t, int64(400), vls.BalanceOf(alice))
	urequire.Equal(t, int64(600), xvls.BalanceOf(alice))

	crossThrough(std.NewUserRealm(alice), func() {
		Stake(cross, 200, bob)
	})
	urequire.Equal(t, int64(200), vls.BalanceOf(alice))
	urequire.Equal(t, int64(600), xvls.BalanceOf(alice))
	urequire.Equal(t, int64(200), xvls.BalanceOf(bob))
}

func TestMultipleStakesAndUnstakes(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice2")

	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)
	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)
		Stake(cross, 500, alice)
		Stake(cross, 200, alice)
	})
	urequire.Equal(t, int64(300), vls.BalanceOf(alice))
	urequire.Equal(t, int64(700), xvls.BalanceOf(alice))

	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 300, alice)
		BeginUnstake(cross, 200, alice)
	})

	testing.SkipHeights(1000000)
	crossThrough(std.NewUserRealm(alice), func() {
		WithdrawUnstaked(cross)
	})
	urequire.Equal(t, int64(800), vls.BalanceOf(alice))
	urequire.Equal(t, int64(200), xvls.BalanceOf(alice))

	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 200, alice)
	})
	urequire.Equal(t, int64(0), xvls.BalanceOf(alice))
}

func TestUnstakeAndWithdrawMultiple(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice3")
	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)
	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)
		Stake(cross, 400, alice)
		Stake(cross, 300, alice)
		BeginUnstake(cross, 200, alice)
		BeginUnstake(cross, 300, alice)
	})

	testing.SkipHeights(1000000)
	crossThrough(std.NewUserRealm(alice), func() {
		WithdrawUnstaked(cross)
	})
	urequire.Equal(t, int64(800), vls.BalanceOf(alice))
	urequire.Equal(t, int64(200), xvls.BalanceOf(alice))
}

func TestUnstakeToDelegatee(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice4")
	bob := std.DerivePkgAddr("bob4")
	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)
	testing.SetRealm(std.NewUserRealm(alice))
	vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)
	crossThrough(std.NewUserRealm(alice), func() {
		Stake(cross, 500, bob)
	})
	urequire.Equal(t, int64(500), vls.BalanceOf(alice))
	urequire.Equal(t, int64(500), xvls.BalanceOf(bob))
	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 200, bob)
	})
	urequire.Equal(t, int64(300), xvls.BalanceOf(bob))

	testing.SkipHeights(1000000)
	crossThrough(std.NewUserRealm(alice), func() {
		WithdrawUnstaked(cross)
	})
	urequire.Equal(t, int64(700), vls.BalanceOf(alice))
}

func TestUnstakeErrors(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice5")
	bob := std.DerivePkgAddr("bob5")
	invalid := std.Address("")

	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 100)

	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 100)
		Stake(cross, 100, alice)
	})

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "insufficient balance", func() {
			BeginUnstake(cross, 200, alice)
		})
	})

	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 100, alice)
	})
	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "cooldown not finished", func() {
			WithdrawUnstaked(cross)
		})
	})

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "amount must be positive", func() {
			Stake(cross, 0, alice)
		})
	})

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "invalid delegatee address", func() {
			Stake(cross, 10, invalid)
		})
	})

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "amount must be positive", func() {
			BeginUnstake(cross, 0, alice)
		})
	})

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "invalid delegatee address", func() {
			BeginUnstake(cross, 10, invalid)
		})
	})

	crossThrough(std.NewUserRealm(bob), func() {
		uassert.AbortsWithMessage(t, "cooldown not finished", func() {
			WithdrawUnstaked(cross)
		})
	})
}

func TestStaker_GovernanceMembershipIntegration(cur realm, t *testing.T) {
	staker := std.DerivePkgAddr("gno.land/r/volos/gov/staker")
	alice := std.DerivePkgAddr("alice_gov")

	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)

	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, staker, 1000)
		Stake(cross, 600, alice)
	})
	urequire.True(t, governance.MemberSet().Has(alice))

	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 200, alice)
	})
	urequire.True(t, governance.MemberSet().Has(alice))

	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 400, alice)
	})
	urequire.False(t, governance.MemberSet().Has(alice))

	crossThrough(std.NewUserRealm(alice), func() {
		Stake(cross, 100, alice)
	})
	urequire.True(t, governance.MemberSet().Has(alice))
}
