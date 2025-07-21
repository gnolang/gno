package vls

import (
	"std"
	"testing"

	"gno.land/p/demo/urequire"
)

func crossThrough(rlm std.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}

func TestMintAndBalance(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice")
	crossThrough(std.NewCodeRealm(VolosDAO), func() {
		Mint(cross, VolosDAOAddress, alice, 1000)
	})
	urequire.Equal(t, int64(1000), BalanceOf(alice))
	urequire.Equal(t, int64(1000), TotalSupply())
}

func TestTransfer(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice2")
	bob := std.DerivePkgAddr("bob2")
	crossThrough(std.NewCodeRealm(VolosDAO), func() {
		Mint(cross, VolosDAOAddress, alice, 1000)
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Transfer(cross, bob, 200)
	})

	urequire.Equal(t, int64(800), BalanceOf(alice))
	urequire.Equal(t, int64(200), BalanceOf(bob))
}

func TestApproveAndTransferFrom(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice3")
	bob := std.DerivePkgAddr("bob3")
	crossThrough(std.NewCodeRealm(VolosDAO), func() {
		Mint(cross, VolosDAOAddress, alice, 1000)
	})
	testing.SetRealm(std.NewUserRealm(alice))
	crossThrough(std.NewUserRealm(alice), func() {
		Approve(cross, bob, 300)
	})
	urequire.Equal(t, int64(300), Allowance(alice, bob))

	crossThrough(std.NewUserRealm(bob), func() {
		TransferFrom(cross, alice, bob, 150)
	})

	urequire.Equal(t, int64(850), BalanceOf(alice))
	urequire.Equal(t, int64(150), BalanceOf(bob))
	urequire.Equal(t, int64(150), Allowance(alice, bob))
}

func TestBurn(cur realm, t *testing.T) {
	alice := std.DerivePkgAddr("alice4")
	crossThrough(std.NewCodeRealm(VolosDAO), func() {
		Mint(cross, VolosDAOAddress, alice, 500)
		Burn(cross, VolosDAOAddress, alice, 200)
	})
	urequire.Equal(t, int64(300), BalanceOf(alice))
	urequire.Equal(t, int64(3300), TotalSupply())
}

func TestUpdateGovernance(cur realm, t *testing.T) {
	newGov := "newgov"
	crossThrough(std.NewCodeRealm(VolosDAO), func() {
		UpdateGovernance(cross, newGov)
	})
	urequire.Equal(t, newGov, VolosDAO)
}
