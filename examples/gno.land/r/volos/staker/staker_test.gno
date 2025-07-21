package staker

import (
	"std"
	"testing"
	"gno.land/p/demo/urequire"
	"gno.land/p/demo/uassert"
	"gno.land/r/volos/vls"
	"gno.land/r/volos/xvls"
)

func crossThrough(rlm std.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}

func TestStakeToSelfAndDelegate(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice")
	bob := std.DerivePkgAddr("bob")
	// Mint VLS to Alice
	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)
	// Alice stakes to herself
	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/staker"), 1000)
		Stake(cross, 600, alice)
	})
	urequire.Equal(t, int64(400), vls.BalanceOf(alice))
	urequire.Equal(t, int64(600), xvls.BalanceOf(alice))
	// Alice stakes to Bob (delegation)
	crossThrough(std.NewUserRealm(alice), func() {
		Stake(cross, 200, bob)
	})
	urequire.Equal(t, int64(200), vls.BalanceOf(alice))
	urequire.Equal(t, int64(600), xvls.BalanceOf(alice))
	urequire.Equal(t, int64(200), xvls.BalanceOf(bob))
}

func TestMultipleStakesAndUnstakes(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice2")
	// Mint VLS to Alice
	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 1000)
	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/staker"), 1000)
		Stake(cross, 500, alice)
		Stake(cross, 200, alice)
	})
	urequire.Equal(t, int64(300), vls.BalanceOf(alice))
	urequire.Equal(t, int64(700), xvls.BalanceOf(alice))
	// Begin two unstakes
	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 300, alice)
		BeginUnstake(cross, 200, alice)
	})
	// Fast forward time for first unstake
	testing.SkipHeights(1000000) // advances block time by 5 seconds per height
	crossThrough(std.NewUserRealm(alice), func() {
		WithdrawUnstaked(cross)
	})
	urequire.Equal(t, int64(800), vls.BalanceOf(alice))
	urequire.Equal(t, int64(200), xvls.BalanceOf(alice))
	// Second unstake not matured yet
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
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/staker"), 1000)
		Stake(cross, 400, alice)
		Stake(cross, 300, alice)
		BeginUnstake(cross, 200, alice)
		BeginUnstake(cross, 300, alice)
	})
	// Fast forward for first unstake
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
	vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/staker"), 1000)
	crossThrough(std.NewUserRealm(alice), func() {
		Stake(cross, 500, bob)
	})
	urequire.Equal(t, int64(500), vls.BalanceOf(alice))
	urequire.Equal(t, int64(500), xvls.BalanceOf(bob))
	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 200, bob)
	})
	urequire.Equal(t, int64(300), xvls.BalanceOf(bob))
	// Fast forward
	testing.SkipHeights(1000000)
	crossThrough(std.NewUserRealm(alice), func() {
		WithdrawUnstaked(cross)
	})
	urequire.Equal(t, int64(700), vls.BalanceOf(alice))
}

func TestUnstakeErrors(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice5")
	testing.SetRealm(std.NewCodeRealm(vls.VolosDAO))
	vls.Mint(cross, vls.VolosDAOAddress, alice, 100)
	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/staker"), 100)
		Stake(cross, 100, alice)
	})
	// Try to unstake more than available
	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "insufficient xVLS to unstake from delegatee", func() {
			BeginUnstake(cross, 200, alice)
		})
	})
	// Try to withdraw before cooldown
	crossThrough(std.NewUserRealm(alice), func() {
		BeginUnstake(cross, 100, alice)
	})
	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "no matured unstake to withdraw", func() {
			WithdrawUnstaked(cross)
		})
	})
} 
