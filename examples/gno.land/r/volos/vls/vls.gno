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
	VolosDAO        = "gno.land/r/volosgov"
	VolosDAOAddress = std.DerivePkgAddr(VolosDAO)
	Token, ledger   = grc20.NewToken("Volos Governance Token", "VLS", 6)

	// Use MemberAuthority for governance address
	Auth = authz.NewWithMembers(VolosDAOAddress)
)

// Name returns the token name.
func Name() string {
	return Token.GetName()
}

// Symbol returns the token symbol.
func Symbol() string {
	return Token.GetSymbol()
}

// Decimals returns the token decimals.
func Decimals() int {
	return Token.GetDecimals()
}

// TotalSupply returns the total supply of the token.
func TotalSupply() int64 {
	return Token.TotalSupply()
}

// BalanceOf returns the token balance of the given address.
func BalanceOf(addr std.Address) int64 {
	return Token.BalanceOf(addr)
}

// Allowance returns the allowance from owner to spender.
func Allowance(owner, spender std.Address) int64 {
	return Token.Allowance(owner, spender)
}

// Transfer tokens from caller to recipient.
func Transfer(cur realm, to std.Address, amount int64) {
	teller := Token.RealmTeller()
	if err := teller.Transfer(to, amount); err != nil {
		panic(err)
	}
}

// Approve a spender to spend tokens on behalf of the caller.
func Approve(cur realm, spender std.Address, amount int64) {
	teller := Token.RealmTeller()
	println(std.CurrentRealm().Address())
	println(spender)
	println(amount)
	if err := teller.Approve(spender, amount); err != nil {
		panic(err)
	}
}

// Transfer tokens from an owner to a recipient, using the caller's allowance.
func TransferFrom(cur realm, owner, to std.Address, amount int64) {
	teller := Token.RealmTeller()
	if err := teller.TransferFrom(owner, to, amount); err != nil {
		panic(err)
	}
}

// Public wrapper functions that handle authorization
func Mint(cur realm, VolosDAO std.Address, to std.Address, amount int64) {
	if err := Auth.DoByPrevious("mint", func() error {
		return mint(to, amount)
	}); err != nil {
		panic(err)
	}
}

func Burn(cur realm, VolosDAO std.Address, from std.Address, amount int64) {
	if err := Auth.DoByPrevious("burn", func() error {
		return burn(from, amount)
	}); err != nil {
		panic(err)
	}
}

func UpdateGovernance(cur realm, newPkgPath string) {
	if err := Auth.DoByPrevious("update_governance", func() error {
		return updateGovernance(newPkgPath)
	}); err != nil {
		panic(err)
	}
}

func mint(to std.Address, amount int64) error {
	if err := ledger.Mint(to, amount); err != nil {
		return err
	}
	return nil
}

func burn(from std.Address, amount int64) error {
	if err := ledger.Burn(from, amount); err != nil {
		return err
	}
	return nil
}

func updateGovernance(newPkgPath string) error {
	VolosDAOAddress = std.DerivePkgAddr(newPkgPath)
	VolosDAO = newPkgPath
	Auth = authz.NewWithMembers(VolosDAOAddress)
	return nil
}
