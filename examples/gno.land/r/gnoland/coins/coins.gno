// Package coins provides simple helpers to retrieve information about coins
// on the Gno.land blockchain.
//
// The primary goal of this realm is to allow users to check their token balances without
// relying on external tools or services. This is particularly valuable for new networks
// that aren't yet widely supported by public explorers or wallets. By using this realm,
// users can always access their balance information directly through the gnodev.
//
// While currently focused on basic balance checking functionality, this realm could
// potentially be extended to support other banker-related workflows in the future.
// However, we aim to keep it minimal and focused on its core purpose.
//
// This is a "Render-only realm" - it exposes only a Render function as its public
// interface and doesn't maintain any state of its own. This pattern allows for
// simple, stateless information retrieval directly through the blockchain's
// rendering capabilities.
//
// Example usage:
//
//	/r/gnoland/coins:ugnot - shows the total supply of ugnot
//	/r/gnoland/coins:ugnot/g1... - shows the ugnot balance of a specific address
package coins

import (
	"errors"
	"std"
	"strings"
)

var (
	errEmptyDenom     = errors.New("empty denom")
	errInvalidAddress = errors.New("invalid address")
	errInvalidPath    = errors.New("invalid path")
)

// TotalSupply returns the total supply of the specified denomination
func TotalSupply(denom string) int64 {
	if denom == "" {
		panic(errEmptyDenom)
	}
	// make the coin's denom to follow the correct format
	qualifiedDenom := std.CurrentRealm().CoinDenom(denom)
	return std.NewBanker(std.BankerTypeReadonly).TotalCoin(qualifiedDenom)
}

// AddressBalance returns the balance of the specified token for the given address
func AddressBalance(denom string, address std.Address) int64 {
	if denom == "" {
		panic(errEmptyDenom)
	}
	if !address.IsValid() {
		panic(errInvalidAddress)
	}
	return std.NewBanker(std.BankerTypeReadonly).
		GetCoins(address).
		AmountOf(denom)
}

// Path parsing logic:
// - Empty path: Show homepage with general info and usage instructions
// - "<denom>": Show total supply and info about the specified coin
// - "<denom>/<address>": Show specific address balance for the given coin
func Render(path string) string {
	if path == "" {
		return renderHomepage()
	}

	if !strings.Contains(path, "/") {
		denom := path
		totalSupply := TotalSupply(denom)
		return renderCoinInfo(denom, totalSupply)
	}

	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		panic(errInvalidPath)
	}

	denom, addr := parts[0], parts[1]
	return renderAddressBalance(denom, std.Address(addr))

	// TODO: implement an explanatory homepage that shows usage instructions
	// TODO: implement a coin info page showing total supply, and other metrics if we have them
	// TODO: implement an address page showing balance for specified coin
	// TODO: create navigation links between pages to enable intuitive exploration without typing URLs
}

func renderHomepage() string {
	return `# Gno.land Coins
`
}

func renderCoinInfo(denom string, totalSupply int64) string {
	return `# Coin Info

## Denom
` + denom + `

## Total Supply
` + std.NewCoin(denom, totalSupply).String() + `\n`
}

func renderAddressBalance(denom string, address std.Address) string {
	balance := AddressBalance(denom, address)
	return `# Address Balance

## Address
` + address.String() + `

## Balance
` + std.NewCoin(denom, balance).String() + `

[Back to Coin Info](/r/gnoland/coins:` + denom + `)
`
}
