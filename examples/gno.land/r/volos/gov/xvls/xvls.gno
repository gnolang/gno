// Package xvls implements the xVLS token for the Volos DAO.
//
// xVLS is a non-transferable governance token representing staked VLS. It is minted
// directly to the chosen delegatee (as specified by the staker contract) and burned
// when VLS is unstaked. Only the staker contract can mint or burn xVLS. xVLS is used
// for voting power in governance. Voting supply can exclude balances held by special
// contracts (e.g., launchpad) if needed.
//
// This contract is GRC20-compatible in interface, but disables transfer and transferFrom.
package xvls

import (
	"std"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/moul/authz"
)

var (
	StakerContract        = "gno.land/r/volos/gov/staker"
	StakerContractAddress = std.DerivePkgAddr(StakerContract)
	token, ledger         = grc20.NewToken("Staked Volos Governance Token", "xVLS", 6)

	Auth = authz.NewWithMembers(StakerContractAddress)

	// Launchpad address (if you want to exclude it from voting supply)
	Launchpad = std.Address("") // set if/when needed
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

// TotalSupply returns the total supply of xVLS.
func TotalSupply() int64 {
	return token.TotalSupply()
}

// VotingSupply returns the total supply of xVLS eligible for voting.
// If Launchpad is set, its balance is excluded from voting supply.
func VotingSupply() int64 {
	total := token.TotalSupply()
	if Launchpad != std.Address("") {
		return total - token.BalanceOf(Launchpad)
	}
	return total
}

// BalanceOf returns the xVLS balance of an address.
func BalanceOf(addr std.Address) int64 {
	return token.BalanceOf(addr)
}

// Mint mints xVLS to the specified address. Only callable by the staker contract.
func Mint(cur realm, to std.Address, amount int64) {
	if err := Auth.DoByPrevious("mint", func() error {
		if err := ledger.Mint(to, amount); err != nil {
			panic(err)
		}
		return nil
	}); err != nil {
		panic(err)
	}

	caller := std.PreviousRealm().Address()
	emitMint(caller, to, amount)
}

// Burn burns xVLS from the specified address. Only callable by the staker contract.
func Burn(cur realm, from std.Address, amount int64) {
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

// Transfer is disabled for xVLS (non-transferable).
func Transfer(cur realm, to std.Address, amount int64) {
	panic(ErrNonTransferable)
}

// TransferFrom is disabled for xVLS (non-transferable).
func TransferFrom(cur realm, from, to std.Address, amount int64) {
	panic(ErrNonTransferable)
}

// Optionally, add a Render function for explorer integration.
func Render(path string) string {
	return token.RenderHome()
}
