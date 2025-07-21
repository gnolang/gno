package xvls

import (
	"std"
	"testing"
	"gno.land/p/demo/urequire"
	"gno.land/p/demo/uassert"
)

func TestMintAndBurnByStaker(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice")
	
	testing.SetRealm(std.NewCodeRealm(StakerContract))
	Mint(cur, alice, 1000)
	urequire.Equal(t, int64(1000), BalanceOf(alice))
	urequire.Equal(t, int64(1000), TotalSupply())
	
	Burn(cur, alice, 400)
	urequire.Equal(t, int64(600), BalanceOf(alice))
	urequire.Equal(t, int64(600), TotalSupply())
}

func TestUnauthorizedMintNoEffect(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice5")
	bob := std.DerivePkgAddr("bob5")
	
	testing.SetRealm(std.NewCodeRealm(StakerContract))
	Mint(cur, alice, 100)
	
	urequire.Equal(t, int64(100), BalanceOf(alice))
	urequire.Equal(t, int64(700), TotalSupply()) 
	
	testing.SetRealm(std.NewUserRealm(bob))
	uassert.AbortsWithMessage(t, "only staker contract can execute: mint", func() {
		Mint(cross, alice, 50)
	})
	
	urequire.Equal(t, int64(100), BalanceOf(alice))
	urequire.Equal(t, int64(700), TotalSupply())
}

func TestUnauthorizedBurnNoEffect(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice6")
	bob := std.DerivePkgAddr("bob6")
	
	testing.SetRealm(std.NewCodeRealm(StakerContract))
	Mint(cur, alice, 100)
	
	urequire.Equal(t, int64(100), BalanceOf(alice))
	urequire.Equal(t, int64(800), TotalSupply()) 
	
	testing.SetRealm(std.NewUserRealm(bob))
	uassert.AbortsWithMessage(t, "only staker contract can execute: burn", func() {
		Burn(cross, alice, 50)
	})
	
	urequire.Equal(t, int64(100), BalanceOf(alice))
	urequire.Equal(t, int64(800), TotalSupply())
}

func TestNonTransferability(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice3")
	bob := std.DerivePkgAddr("bob3")
	testing.SetRealm(std.NewCodeRealm(StakerContract))
	Mint(cur, alice, 500)
	
	urequire.Equal(t, int64(500), BalanceOf(alice))
	urequire.Equal(t, int64(1300), TotalSupply()) 
	
	testing.SetRealm(std.NewUserRealm(alice))
	uassert.AbortsWithMessage(t, "xVLS is non-transferable", func() {
		Transfer(cross, bob, 100)
	})
	uassert.AbortsWithMessage(t, "xVLS is non-transferable", func() {
		TransferFrom(cross, alice, bob, 100)
	})
}

func TestVotingSupply(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice4")
	bob := std.DerivePkgAddr("bob4")
	testing.SetRealm(std.NewCodeRealm(StakerContract))
	Mint(cur, alice, 1000)
	Mint(cur, bob, 500)
	
	urequire.Equal(t, int64(2800), TotalSupply()) 
	urequire.Equal(t, int64(2800), VotingSupply())
} 
