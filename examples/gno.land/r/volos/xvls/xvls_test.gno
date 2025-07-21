package xvls

import (
	"std"
	"testing"
	"gno.land/p/demo/urequire"
	"gno.land/p/demo/uassert"
)

func crossThrough(rlm std.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}

func TestMintAndBurnByStaker(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice")
	
	crossThrough(std.NewCodeRealm(StakerContract), func() {
		Mint(cross, alice, 1000)
	})
	urequire.Equal(t, int64(1000), BalanceOf(alice))
	urequire.Equal(t, int64(1000), TotalSupply())
	
	crossThrough(std.NewCodeRealm(StakerContract), func() {
		Burn(cross, alice, 400)
	})
	urequire.Equal(t, int64(600), BalanceOf(alice))
	urequire.Equal(t, int64(600), TotalSupply())
}

func TestUnauthorizedMintNoEffect(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice5")
	bob := std.DerivePkgAddr("bob5")
	
	crossThrough(std.NewCodeRealm(StakerContract), func() {
		Mint(cross, alice, 100)
	})
	
	urequire.Equal(t, int64(100), BalanceOf(alice))
	urequire.Equal(t, int64(700), TotalSupply()) 
	
	crossThrough(std.NewUserRealm(bob), func() {
		uassert.AbortsWithMessage(t, "unauthorized", func() {
			Mint(cross, alice, 50)
		})
	})
	
	urequire.Equal(t, int64(100), BalanceOf(alice))
	urequire.Equal(t, int64(700), TotalSupply())
}

func TestUnauthorizedBurnNoEffect(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice6")
	bob := std.DerivePkgAddr("bob6")
	
	crossThrough(std.NewCodeRealm(StakerContract), func() {
		Mint(cross, alice, 100)
	})
	
	urequire.Equal(t, int64(100), BalanceOf(alice))
	urequire.Equal(t, int64(800), TotalSupply()) 
	
	crossThrough(std.NewUserRealm(bob), func() {
		uassert.AbortsWithMessage(t, "unauthorized", func() {
			Burn(cross, alice, 50)
		})
	})
	
	urequire.Equal(t, int64(100), BalanceOf(alice))
	urequire.Equal(t, int64(800), TotalSupply())
}

func TestNonTransferability(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice3")
	bob := std.DerivePkgAddr("bob3")
	crossThrough(std.NewCodeRealm(StakerContract), func() {
		Mint(cross, alice, 500)
	})
	
	urequire.Equal(t, int64(500), BalanceOf(alice))
	urequire.Equal(t, int64(1300), TotalSupply()) 
	
	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "xVLS is non-transferable", func() {
			Transfer(cross, bob, 100)
		})
	})
	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "xVLS is non-transferable", func() {
			TransferFrom(cross, alice, bob, 100)
		})
	})
}

func TestVotingSupply(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice4")
	bob := std.DerivePkgAddr("bob4")
	crossThrough(std.NewCodeRealm(StakerContract), func() {
		Mint(cross, alice, 1000)
		Mint(cross, bob, 500)
	})
	
	urequire.Equal(t, int64(2800), TotalSupply()) 
	urequire.Equal(t, int64(2800), VotingSupply())
} 
