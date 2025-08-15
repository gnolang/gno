// Package vls implements the Volos Governance Token for the Volos DAO on Gno.land.
//
// This contract defines a GRC20-compatible governance token (VLS) designed for use by the
// Volos DAO. Minting and burning of tokens are strictly restricted to the governance
// contract, ensuring that only authorized proposals or upgrades can affect the total supply.
// The contract enforces all privileged actions through an authority model built on the authz
// package, which ensures that only the correct governance contract can execute sensitive operations.
//
// VLS is the staking token for the Volos protocol. Users must stake VLS in the staker contract
// to participate in governance and receive xVLS, which represents voting power. All logic for
// delegation, voting power, and cooldowns is handled by the staker and xVLS contracts, not here.
// Error messages are defined in a separate errors.gno file for consistency and maintainability.
package vls

import (
	"std"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/moul/authz"
)

var (
	VolosDAO        = "gno.land/r/volos/gov/governance"
	VolosDAOAddress = std.DerivePkgAddr(VolosDAO)
	token, ledger   = grc20.NewToken("Volos Governance Token", "VLS", 6)

	// Use MemberAuthority for governance address
	Auth = authz.NewWithMembers(VolosDAOAddress)
)

// Name returns the token name.
func Name() string {
	return token.GetName()
}

// Symbol returns the token symbol.
func Symbol() string {
	return token.GetSymbol()
}

// Decimals returns the token decimals.
func Decimals() int {
	return token.GetDecimals()
}

// TotalSupply returns the total supply of the token.
func TotalSupply() int64 {
	return token.TotalSupply()
}

// BalanceOf returns the token balance of the given address.
func BalanceOf(addr std.Address) int64 {
	return token.BalanceOf(addr)
}

// Allowance returns the allowance from owner to spender.
func Allowance(owner, spender std.Address) int64 {
	return token.Allowance(owner, spender)
}

// Transfer tokens from caller to recipient.
func Transfer(cur realm, to std.Address, amount int64) {
	teller := token.CallerTeller()
	if err := teller.Transfer(to, amount); err != nil {
		panic(err)
	}

	caller := std.PreviousRealm().Address()
	emitTransfer(caller, caller, to, amount)
}

// Approve a spender to spend tokens on behalf of the caller.
func Approve(cur realm, spender std.Address, amount int64) {
	teller := token.CallerTeller()
	if err := teller.Approve(spender, amount); err != nil {
		panic(err)
	}

	caller := std.PreviousRealm().Address()
	emitApproval(caller, spender, amount)
}

// Transfer tokens from an owner to a recipient, using the caller's allowance.
func TransferFrom(cur realm, owner, to std.Address, amount int64) {
	teller := token.CallerTeller()
	if err := teller.TransferFrom(owner, to, amount); err != nil {
		panic(err)
	}

	caller := std.PreviousRealm().Address()
	emitTransfer(caller, owner, to, amount)
}

// Public wrapper functions that handle authorization
func Mint(cur realm, VolosDAO std.Address, to std.Address, amount int64) {
	if err := Auth.DoByPrevious("mint", func() error {
		if err := ledger.Mint(to, amount); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	caller := std.PreviousRealm().Address()
	emitMint(caller, to, amount)
}

// Burn tokens from the caller's account.
func Burn(cur realm, VolosDAO std.Address, from std.Address, amount int64) {
	if err := Auth.DoByPrevious("burn", func() error {
		if err := ledger.Burn(from, amount); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	caller := std.PreviousRealm().Address()
	emitBurn(caller, from, amount)
}

// Update the governance contract address.
func UpdateGovernance(cur realm, newPkgPath string) {
	oldPkgPath := VolosDAO
	if err := Auth.DoByPrevious("update_governance", func() error {
		VolosDAOAddress = std.DerivePkgAddr(newPkgPath)
		VolosDAO = newPkgPath
		Auth = authz.NewWithMembers(VolosDAOAddress)

		return nil
	}); err != nil {
		panic(err)
	}

	caller := std.PreviousRealm().Address()
	emitGovernanceUpdated(caller, oldPkgPath, newPkgPath)
}

// Render function for explorer integration.
func Render(path string) string {
	return token.RenderHome()
}

// temporary function to mint tokens for testing
func Faucet(cur realm, amount int64) {
	caller := std.PreviousRealm().Address()
	if err := ledger.Mint(caller, amount); err != nil {
		panic(err)
	}

	emitMint(caller, caller, amount)
}
