package vls

import (
	"testing"
	"std"
	"gno.land/p/demo/urequire"
)

// Helper to create addresses for testing
func addr(name string) std.Address {
	return std.DerivePkgAddr("test/" + name)
}

func TestMintAndBalance(cur realm, t *testing.T) {
	alice := addr("alice")
	testing.SetRealm(std.NewCodeRealm(VolosDAO))
	Mint(cur, VolosDAOAddress, alice, 1000)
	urequire.Equal(t, int64(1000), BalanceOf(alice))
	urequire.Equal(t, int64(1000), TotalSupply())
}

func TestTransfer(cur realm, t *testing.T) {
	alice := addr("alice2")
	bob := addr("bob2")
	testing.SetRealm(std.NewCodeRealm(VolosDAO))
	Mint(cur, VolosDAOAddress, alice, 1000)

	testing.SetRealm(std.NewUserRealm(alice))
	Transfer(cur, bob, 200)

	urequire.Equal(t, int64(800), BalanceOf(alice))
	urequire.Equal(t, int64(200), BalanceOf(bob))
}

func TestApproveAndTransferFrom(cur realm, t *testing.T) {
	alice := addr("alice3")
	bob := addr("bob3")
	testing.SetRealm(std.NewCodeRealm(VolosDAO))
	Mint(cur, VolosDAOAddress, alice, 1000)

	testing.SetRealm(std.NewUserRealm(alice))
	Approve(cur, bob, 300)
	urequire.Equal(t, int64(300), Allowance(alice, bob))

	testing.SetRealm(std.NewUserRealm(bob))
	TransferFrom(cur, alice, bob, 150)

	urequire.Equal(t, int64(850), BalanceOf(alice))
	urequire.Equal(t, int64(150), BalanceOf(bob))
	urequire.Equal(t, int64(150), Allowance(alice, bob))
}

func TestBurn(cur realm, t *testing.T) {
	alice := addr("alice4")
	testing.SetRealm(std.NewCodeRealm(VolosDAO))
	Mint(cur, VolosDAOAddress, alice, 500)
	Burn(cur, VolosDAOAddress, alice, 200)
	urequire.Equal(t, int64(300), BalanceOf(alice))
	urequire.Equal(t, int64(3300), TotalSupply())
}

func TestDelegationAndVotingPower(cur realm, t *testing.T) {
	alice := addr("alice5")
	bob := addr("bob5")
	testing.SetRealm(std.NewCodeRealm(VolosDAO))
	Mint(cur, VolosDAOAddress, alice, 1000)
	Mint(cur, VolosDAOAddress, bob, 500)

	testing.SetRealm(std.NewUserRealm(bob))
	Delegate(cur, bob)

	testing.SetRealm(std.NewUserRealm(alice))
	Delegate(cur, bob) 

	urequire.Equal(t, int64(0), GetVotingPower(alice))
	urequire.Equal(t, int64(1500), GetVotingPower(bob))
}

func TestUpdateGovernance(cur realm, t *testing.T) {
	newGov := addr("newgov")
	testing.SetRealm(std.NewCodeRealm(VolosDAO))
	UpdateGovernance(cur, newGov)
	urequire.Equal(t, newGov, VolosDAOAddress)
}
