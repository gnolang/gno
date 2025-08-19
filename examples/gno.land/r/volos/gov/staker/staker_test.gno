package staker

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/r/volos/gov/governance"
	"gno.land/r/volos/gov/vls"
	"gno.land/r/volos/gov/xvls"
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

func TestDelegationMapping_BasicTracking(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice_delegation")
	bob := std.DerivePkgAddr("bob_delegation")
	charlie := std.DerivePkgAddr("charlie_delegation")

	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)

	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)

		Stake(cross, 600, alice)
		urequire.Equal(t, int64(600), GetDelegatedAmount(alice, alice))

		Stake(cross, 200, bob)
		urequire.Equal(t, int64(200), GetDelegatedAmount(alice, bob))

		Stake(cross, 200, charlie)
		urequire.Equal(t, int64(200), GetDelegatedAmount(alice, charlie))
	})

	urequire.Equal(t, int64(600), GetDelegatedAmount(alice, alice))
	urequire.Equal(t, int64(200), GetDelegatedAmount(alice, bob))
	urequire.Equal(t, int64(200), GetDelegatedAmount(alice, charlie))

	urequire.Equal(t, int64(0), GetDelegatedAmount(bob, alice))
	urequire.Equal(t, int64(0), GetDelegatedAmount(charlie, alice))
}

func TestDelegationMapping_UnstakeSecurity(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice_security")
	bob := std.DerivePkgAddr("bob_security")
	charlie := std.DerivePkgAddr("charlie_security")

	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)
	vls.Mint(cross, vls.VolosDAOAddress, bob, 1000)

	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)
		Stake(cross, 500, bob)
	})

	urequire.Equal(t, int64(500), xvls.BalanceOf(bob))
	urequire.Equal(t, int64(500), GetDelegatedAmount(alice, bob))

	crossThrough(std.NewUserRealm(bob), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)
		Stake(cross, 300, bob)
	})

	urequire.Equal(t, int64(800), xvls.BalanceOf(bob))
	urequire.Equal(t, int64(300), GetDelegatedAmount(bob, bob))

	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 200, bob)
	})

	urequire.Equal(t, int64(600), xvls.BalanceOf(bob))
	urequire.Equal(t, int64(300), GetDelegatedAmount(alice, bob))

	crossThrough(std.NewUserRealm(bob), func() {
		uassert.AbortsWithMessage(t, "insufficient delegation", func() {
			BeginUnstake(cross, 400, bob)
		})
	})

	crossThrough(std.NewUserRealm(charlie), func() {
		uassert.AbortsWithMessage(t, "insufficient delegation", func() {
			BeginUnstake(cross, 100, bob)
		})
	})

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "insufficient delegation", func() {
			BeginUnstake(cross, 400, bob)
		})
	})
}

func TestDelegationMapping_MultipleStakersToSameDelegatee(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice_multi")
	bob := std.DerivePkgAddr("bob_multi")
	charlie := std.DerivePkgAddr("charlie_multi")

	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)
	vls.Mint(cross, vls.VolosDAOAddress, bob, 1000)

	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)
		Stake(cross, 600, charlie)
	})

	crossThrough(std.NewUserRealm(bob), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)
		Stake(cross, 400, charlie)
	})

	urequire.Equal(t, int64(1000), xvls.BalanceOf(charlie))

	urequire.Equal(t, int64(600), GetDelegatedAmount(alice, charlie))
	urequire.Equal(t, int64(400), GetDelegatedAmount(bob, charlie))
	urequire.Equal(t, int64(0), GetDelegatedAmount(charlie, charlie))

	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 300, charlie)
	})

	urequire.Equal(t, int64(700), xvls.BalanceOf(charlie))
	urequire.Equal(t, int64(300), GetDelegatedAmount(alice, charlie))
	urequire.Equal(t, int64(400), GetDelegatedAmount(bob, charlie))

	crossThrough(std.NewUserRealm(bob), func() {
		uassert.AbortsWithMessage(t, "insufficient delegation", func() {
			BeginUnstake(cross, 500, charlie)
		})
	})

	crossThrough(std.NewUserRealm(bob), func() {
		BeginUnstake(cross, 400, charlie)
	})
	urequire.Equal(t, int64(300), xvls.BalanceOf(charlie))
	urequire.Equal(t, int64(300), GetDelegatedAmount(alice, charlie))
	urequire.Equal(t, int64(0), GetDelegatedAmount(bob, charlie))
}

func TestDelegationMapping_CompleteUnstakeAndRedelegate(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice_complete")
	bob := std.DerivePkgAddr("bob_complete")

	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)

	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 1000)

		Stake(cross, 500, bob)
		urequire.Equal(t, int64(500), GetDelegatedAmount(alice, bob))

		BeginUnstake(cross, 500, bob)
		urequire.Equal(t, int64(0), GetDelegatedAmount(alice, bob))
		urequire.Equal(t, int64(0), xvls.BalanceOf(bob))

		Stake(cross, 300, bob)
		urequire.Equal(t, int64(300), GetDelegatedAmount(alice, bob))
		urequire.Equal(t, int64(300), xvls.BalanceOf(bob))
	})
}

func TestUnstakeLockPeriod_GetterAndSetter(cur realm, t *testing.T) {
	gov := "gno.land/r/volos/gov/governance"

	urequire.Equal(t, int64(7*24*60*60), UnstakeLockPeriod())

	crossThrough(std.NewCodeRealm(gov), func() {
		SetUnstakeLockPeriod(cross, 14*24*60*60)
	})
	urequire.Equal(t, int64(14*24*60*60), UnstakeLockPeriod())

	alice := std.DerivePkgAddr("alice_unauthorized")
	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "unauthorized", func() {
			SetUnstakeLockPeriod(cross, 1*24*60*60)
		})
	})
	urequire.Equal(t, int64(14*24*60*60), UnstakeLockPeriod())

	crossThrough(std.NewCodeRealm(gov), func() {
		SetUnstakeLockPeriod(cross, 7*24*60*60)
	})
	urequire.Equal(t, int64(7*24*60*60), UnstakeLockPeriod())
}
