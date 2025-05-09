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
	"std"

	"gno.land/p/demo/mux"
)

var router *mux.Router

func init() {
	router = mux.NewRouter()

	// homepage
	router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(renderHomepage())
	})

	// coin info
	router.HandleFunc("{denom}", func(res *mux.ResponseWriter, req *mux.Request) {
		// denom := req.GetVar("denom")
		// banker := std.NewBanker(std.BankerTypeReadonly)
		// res.Write(renderAddressBalance(banker, denom, denom))
		res.Write("Total supply feature is coming soon. Please check back later!")
	})

	// address balance
	router.HandleFunc("{denom}/{address}", func(res *mux.ResponseWriter, req *mux.Request) {
		denom := req.GetVar("denom")
		addr := req.GetVar("address")
		banker := std.NewBanker(std.BankerTypeReadonly)
		res.Write(renderAddressBalance(banker, denom, addr))
	})
}

func Render(path string) string {
	return router.Render(path)
}

func renderHomepage() string {
	return `# gno.land Coins Explorer

## Usage

- /r/gnoland/coins:<denom> - shows the total supply of denom (coming soon)
- /r/gnoland/coins:<denom>/<address> - shows the denom balance of a specific address

Examples:

- /r/gnoland/coins:ugnot - shows the total supply of ` + "`ugnot`" + ` (coming soon)
- /r/gnoland/coins:ugnot/g1... - shows the ` + "`ugnot`" + ` balance of a specific address

`
}

func renderAddressBalance(banker std.Banker, denom string, addr string) string {
	address := std.Address(addr)
	coins := banker.GetCoins(address)

	for _, coin := range coins {
		if coin.Denom == denom {
			return "Balance: " + coin.String()
		}
	}

	return "Balance: 0 " + denom
}
