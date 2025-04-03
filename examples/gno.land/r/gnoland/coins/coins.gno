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
	"strconv"

	"gno.land/p/demo/mux"
)

var (
	errEmptyDenom     = errors.New("empty denom")
	errInvalidAddress = errors.New("invalid address")
	errInvalidPath    = errors.New("invalid path")
)

var router *mux.Router

func init() {
	router = mux.NewRouter()

	// Homepage
	router.HandleFunc("/", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(renderHomepage())
	})

	// Coin Info
	router.HandleFunc("/{denom}", func(res *mux.ResponseWriter, req *mux.Request) {
		denom := req.GetVar("denom")
		totalSupply := TotalSupply(denom)
		res.Write(renderCoinInfo(denom, totalSupply))
	})

	// Address Balance
	router.HandleFunc("/{denom}/{addr}", func(res *mux.ResponseWriter, req *mux.Request) {
		denom := req.GetVar("denom")
		addr := req.GetVar("addr")
		res.Write(renderAddressBalance(denom, std.Address(addr)))
	})
}

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
	return router.Render(path)
}

func renderHomepage() string {
	out := "# Gno.land Coins\n\n"
	out += "This realm provides an interface to query coin information on the Gno.land chain.\n\n"

	out += "## Usage\n\n"
	out += "You can query information using the following URL patterns:\n\n"
	out += "* /r/gnoland/coins:ugnot - Check total supply of ugnot\n"
	out += "* /r/gnoland/coins:ugnot/{address} - Check ugnot balance of a specific address\n\n"

	out += "## Supported Coins\n\n"
	out += "* ugnot - The native token of Gno.land\n"

	return out
}

func renderCoinInfo(denom string, totalSupply int64) string {
	out := "# Coin Information\n\n"

	out += "## Denomination\n\n"
	out += denom + "\n\n"

	out += "## Total Supply\n\n"
	out += std.NewCoin(denom, totalSupply).String() + "\n\n"

	out += "## Statistics\n\n"
	out += "* Total Supply: " + std.NewCoin(denom, totalSupply).String() + "\n"
	out += "* Chain ID: " + std.ChainID() + "\n"
	out += "* Current Block Height: " + strconv.Itoa(int(std.ChainHeight())) + "\n"

	return out
}

func renderAddressBalance(denom string, address std.Address) string {
	balance := AddressBalance(denom, address)
	out := "# Address Balance Information\n\n"

	out += "## Address\n\n"
	out += address.String() + "\n\n"

	out += "## Balance\n\n"
	out += std.NewCoin(denom, balance).String() + "\n"

	return out
}
