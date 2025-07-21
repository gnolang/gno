package vls

import (
	"std"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

// Helper to create addresses for testing
func addr(name string) std.Address {
	return std.DerivePkgAddr("test/" + name)
}

func TestMintAndBalance(t uassert.TestingT) {
	alice := addr("alice")
	cross := std.CurrentRealm()

	// Mint 1000 tokens to Alice
	Mint(cross, volosDAOAddress, alice, 1000)
	urequire.Equal(t, int64(1000), BalanceOf(alice))
	urequire.Equal(t, int64(1000), TotalSupply())
}

func TestTransfer(t uassert.TestingT) {
	alice := addr("alice")
	bob := addr("bob")
	cross := std.CurrentRealm()

	// Mint to Alice
	Mint(cross, volosDAOAddress, alice, 1000)

	// Alice transfers 200 to Bob
	std.WithRealm(alice, func() {
		Transfer(cross, bob, 200)
	})

	urequire.Equal(t, int64(800), BalanceOf(alice))
	urequire.Equal(t, int64(200), BalanceOf(bob))
}

func TestApproveAndTransferFrom(t uassert.TestingT) {
	alice := addr("alice")
	bob := addr("bob")
	cross := std.CurrentRealm()

	// Mint to Alice
	Mint(cross, volosDAOAddress, alice, 1000)

	// Alice approves Bob for 300
	std.WithRealm(alice, func() {
		Approve(cross, bob, 300)
	})
	urequire.Equal(t, int64(300), Allowance(alice, bob))

	// Bob transfers 150 from Alice to himself
	std.WithRealm(bob, func() {
		TransferFrom(cross, alice, bob, 150)
	})

	urequire.Equal(t, int64(850), BalanceOf(alice))
	urequire.Equal(t, int64(150), BalanceOf(bob))
	urequire.Equal(t, int64(150), Allowance(alice, bob))
}

func TestBurn(t uassert.TestingT) {
	alice := addr("alice")
	cross := std.CurrentRealm()

	Mint(cross, volosDAOAddress, alice, 500)
	Burn(cross, volosDAOAddress, alice, 200)

	urequire.Equal(t, int64(300), BalanceOf(alice))
	urequire.Equal(t, int64(300), TotalSupply())
}

func TestDelegationAndVotingPower(t uassert.TestingT) {
	alice := addr("alice")
	bob := addr("bob")
	cross := std.CurrentRealm()

	Mint(cross, volosDAOAddress, alice, 1000)
	Mint(cross, volosDAOAddress, bob, 500)

	// Alice delegates to Bob
	std.WithRealm(alice, func() {
		Delegate(cross, bob)
	})

	urequire.Equal(t, int64(0), GetVotingPower(alice))
	urequire.Equal(t, int64(1500), GetVotingPower(bob))

	// Alice transfers 200 to Bob (should update voting power)
	std.WithRealm(alice, func() {
		Transfer(cross, bob, 200)
	})

	urequire.Equal(t, int64(800), BalanceOf(alice))
	urequire.Equal(t, int64(700), BalanceOf(bob))
	urequire.Equal(t, int64(0), GetVotingPower(alice))
	urequire.Equal(t, int64(1500), GetVotingPower(bob)) // Alice's delegation still covers her new balance

	// Alice revokes delegation (delegates to herself)
	std.WithRealm(alice, func() {
		Delegate(cross, alice)
	})

	urequire.Equal(t, int64(800), GetVotingPower(alice))
	urequire.Equal(t, int64(700), GetVotingPower(bob))
}

func TestUpdateGovernance(t uassert.TestingT) {
	newGov := addr("newgov")
	cross := std.CurrentRealm()

	UpdateGovernance(cross, newGov)
	urequire.Equal(t, newGov, volosDAOAddress)
}
