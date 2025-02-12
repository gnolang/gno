package xgns

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

var (
	token  *grc20.Token
	ledger *grc20.PrivateLedger
	admin  *ownable.Ownable
)

func init() {
	admin = ownable.NewWithAddress(std.DerivePkgAddr(consts.GOV_STAKER_PATH))

	token, ledger = grc20.NewToken("XGNS", "xGNS", 6)
}

func TotalSupply() uint64 { return token.TotalSupply() }

// VotingSupply calculates the total supply of tokens eligible for voting.
//
// This function determines the total voting supply by subtracting the amount
// of tokens held by the launchpad contract from the total minted token supply.
// Tokens held by the launchpad contract do not participate in voting.
//
// Returns:
// - uint64: The total supply of tokens available for voting.
//
// Notes:
// - `TotalSupply`: Represents the total amount of xGNS tokens minted.
// - `BalanceOf(consts.LAUNCHPAD_ADDR)`: Retrieves the amount of xGNS tokens held by the launchpad contract.
func VotingSupply() uint64 {
	total := token.TotalSupply() // this is entire amount of xGNS minted

	// this is amount of xGNS held by launchpad
	// this xGNS doesn't participate in voting
	launchpad := token.BalanceOf(consts.LAUNCHPAD_ADDR)
	return total - launchpad
}

// BalanceOf retrieves the token balance of a specified address.
//
// This function resolves the provided address or name and queries the token balance
// associated with the resolved address.
//
// Parameters:
// - owner: The address or name of the user whose balance is being queried.
//
// Returns:
// - uint64: The current token balance of the specified address.
func BalanceOf(owner std.Address) uint64 {
	return token.BalanceOf(owner)
}

// xGNS is non-transferable
// Therefore it doesn't have transfer and transferFrom functions
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return token.RenderHome()
	case c == 2 && parts[0] == "balance":
		balance := token.BalanceOf(std.Address(parts[1]))
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

// Mint increases the balance of a specified address by a given amount.
//
// This function is restricted to be called only by specific authorized contracts:
// - Governance staker contract
//
// If the caller is not one of these contracts, the function will panic with an error.
//
// Parameters:
// - to: The address or name of the user whose balance will be increased.
// - amount: The amount of tokens to be minted.
//
// Errors:
// - Panics if the caller is unauthorized.
// - Propagates any error from the ledger.Mint function.
func Mint(to std.Address, amount uint64) {
	common.IsHalted()

	// only (gov staker) or (launchpad) contract can call Mint
	caller := std.PrevRealm().Addr()
	if caller != consts.GOV_STAKER_ADDR && caller != consts.LAUNCHPAD_ADDR {
		panic(addDetailToError(
			errNoPermission,
			ufmt.Sprintf("only gov/staker(%s) or launchpad(%s) contract can call Mint, called from %s", consts.GOV_STAKER_ADDR.String(), consts.LAUNCHPAD_ADDR.String(), caller.String()),
		))
	}

	checkErr(ledger.Mint(to, amount))
}

// MintByLaunchPad increases the balance of a specified address by a given amount.
//
// This function is restricted to be called only by specific authorized contracts:
// - Launchpad contract
//
// If the caller is not one of these contracts, the function will panic with an error.
//
// Parameters:
// - to: The address or name of the user whose balance will be increased.
// - amount: The amount of tokens to be minted.
//
// Errors:
// - Panics if the caller is unauthorized.
// - Propagates any error from the ledger.Mint function.
func MintByLaunchPad(to std.Address, amount uint64) {
	common.IsHalted()

	caller := std.PrevRealm().Addr()
	if caller != consts.LAUNCHPAD_ADDR {
		panic(addDetailToError(
			errNoPermission,
			ufmt.Sprintf("only launchpad(%s) contract can call MintByLaunchPad, called from %s", consts.LAUNCHPAD_ADDR.String(), caller.String()),
		))
	}

	mint(to, amount)
}

// mint increases the balance of a specified address by a given amount.
func mint(to std.Address, amount uint64) {
	checkErr(ledger.Mint(to, amount))
}

// Burn reduces the balance of a specified address by a given amount.
//
// This function is restricted to be called only by specific authorized contracts:
// - Governance staker contract
// - Launchpad contract
//
// If the caller is not one of these contracts, the function will panic with an error.
//
// Parameters:
// - from: The address or name of the user whose balance will be reduced.
// - amount: The amount of tokens to be burned.
//
// Errors:
// - Panics if the caller is unauthorized.
// - Propagates any error from the ledger.Burn function.
func Burn(from std.Address, amount uint64) {
	common.IsHalted()

	// only (gov staker) or (launchpad) contract can call Mint
	caller := std.PrevRealm().Addr()
	if !(caller == consts.GOV_STAKER_ADDR || caller == consts.LAUNCHPAD_ADDR) {
		panic(addDetailToError(
			errNoPermission,
			ufmt.Sprintf("only gov/staker(%s) or launchpad(%s) contract can call Burn, called from %s", consts.GOV_STAKER_ADDR.String(), consts.LAUNCHPAD_ADDR.String(), caller.String()),
		))
	}

	burn(from, amount)
}

func BurnByLaunchPad(from std.Address, amount uint64) {
	common.IsHalted()

	caller := std.PrevRealm().Addr()
	if caller != consts.LAUNCHPAD_ADDR {
		panic(addDetailToError(
			errNoPermission,
			ufmt.Sprintf("only launchpad(%s) contract can call BurnByLaunchPad, called from %s", consts.LAUNCHPAD_ADDR.String(), caller.String()),
		))
	}
	burn(from, amount)
}

func burn(from std.Address, amount uint64) {
	checkErr(ledger.Burn(from, amount))
}
