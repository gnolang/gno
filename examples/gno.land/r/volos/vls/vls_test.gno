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
	testing.SetRealm(std.NewCodeRealm(volosDAO))
	Mint(cur, volosDAOAddress, alice, 1000)
	urequire.Equal(t, int64(1000), BalanceOf(alice))
	urequire.Equal(t, int64(1000), TotalSupply())
}

func TestTransfer(cur realm, t *testing.T) {
	alice := addr("alice")
	bob := addr("bob")
	testing.SetRealm(std.NewCodeRealm(volosDAO))
	Mint(cur, volosDAOAddress, alice, 1000)

	testing.SetRealm(std.NewUserRealm(alice))
	Transfer(cur, bob, 200)

	urequire.Equal(t, int64(800), BalanceOf(alice))
	urequire.Equal(t, int64(200), BalanceOf(bob))
}

func TestApproveAndTransferFrom(cur realm, t *testing.T) {
	alice := addr("alice")
	bob := addr("bob")
	testing.SetRealm(std.NewCodeRealm(volosDAO))
	Mint(cur, volosDAOAddress, alice, 1000)

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
	alice := addr("alice")
	testing.SetRealm(std.NewCodeRealm(volosDAO))
	Mint(cur, volosDAOAddress, alice, 500)
	Burn(cur, volosDAOAddress, alice, 200)
	urequire.Equal(t, int64(300), BalanceOf(alice))
	urequire.Equal(t, int64(300), TotalSupply())
}

func TestDelegationAndVotingPower(cur realm, t *testing.T) {
	alice := addr("alice")
	bob := addr("bob")
	testing.SetRealm(std.NewCodeRealm(volosDAO))
	Mint(cur, volosDAOAddress, alice, 1000)
	Mint(cur, volosDAOAddress, bob, 500)

	testing.SetRealm(std.NewUserRealm(alice))
	Delegate(cur, bob)
	urequire.Equal(t, int64(0), GetVotingPower(alice))
	urequire.Equal(t, int64(1500), GetVotingPower(bob))

	testing.SetRealm(std.NewUserRealm(alice))
	Transfer(cur, bob, 200)
	urequire.Equal(t, int64(800), BalanceOf(alice))
	urequire.Equal(t, int64(700), BalanceOf(bob))
	urequire.Equal(t, int64(0), GetVotingPower(alice))
	urequire.Equal(t, int64(1500), GetVotingPower(bob))

	testing.SetRealm(std.NewUserRealm(alice))
	Delegate(cur, alice)
	urequire.Equal(t, int64(800), GetVotingPower(alice))
	urequire.Equal(t, int64(700), GetVotingPower(bob))
}

func TestUpdateGovernance(cur realm, t *testing.T) {
	newGov := addr("newgov")
	testing.SetRealm(std.NewCodeRealm(volosDAO))
	UpdateGovernance(cur, newGov)
	urequire.Equal(t, newGov, volosDAO)
}
