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
package coins

import (
	"net/url"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/leon/coinsort"
	"gno.land/p/leon/ctg"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/realmpath"

	"gno.land/r/sys/users"
)

var router *mux.Router

func init() {
	router = mux.NewRouter()

	router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(renderHomepage())
	})

	router.HandleFunc("balances/{address}", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(renderAllBalances(req.RawPath, req.GetVar("address")))
	})

	router.HandleFunc("convert/{address}", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(renderConvertedAddress(req.GetVar("address")))
	})

	// Coin info
	router.HandleFunc("supply/{denom}", func(res *mux.ResponseWriter, req *mux.Request) {
		// banker := std.NewBanker(std.BankerTypeReadonly)
		// res.Write(renderAddressBalance(banker, denom, denom))
		res.Write("The total supply feature is coming soon.")
	})

	router.NotFoundHandler = func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write("# 404\n\nThat page was not found. Would you like to [**go home**?](/r/gnoland/coins)")
	}
}

func Render(path string) string {
	return router.Render(path)
}

func renderHomepage() string {
	return strings.Replace(`# Gno.land Coins Explorer

This is a simple, readonly realm that allows users to browse native coin balances.
Here are a few examples on how to use it:

- ~/r/gnoland/coins:balances/<address>~ - show full list of coin balances of an address
	- [Example](/r/gnoland/coins:balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5)
- ~/r/gnoland/coins:balances/<address>?coin=ugnot~ - shows the balance of an address for a specific coin
	- [Example](/r/gnoland/coins:balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5?coin=ugnot)
- ~/r/gnoland/coins:convert/<cosmos_address>~ - convert Cosmos address to Gno address
	- [Example](/r/gnoland/coins:convert/cosmos1jg8mtutu9khhfwc4nxmuhcpftf0pajdh6svrgs)
- ~/r/gnoland/coins:supply/<denom>~ - shows the total supply of denom
	- Coming soon!
`, "~", "`", -1)
}

func renderConvertedAddress(addr string) string {
	out := "# Address converter\n\n"

	gnoAddress, err := ctg.ConvertCosmosToGno(addr)
	if err != nil {
		out += err.Error()
		return out
	}

	user, _ := users.ResolveAny(gnoAddress.String())
	name := "`" + gnoAddress.String() + "`"
	if user != nil {
		name = user.RenderLink("")
	}

	out += ufmt.Sprintf("`%s` on Cosmos matches %s on gno.land.\n\n", addr, name)
	out += "[View `ugnot` balance for this address](/r/gnoland/coins:balances/" + gnoAddress.String() + "?coin=ugnot)\n\n"
	out += "[View full balance list for this address](/r/gnoland/coins:balances/" + gnoAddress.String() + ")"
	return out
}

func renderSingleCoinBalance(banker std.Banker, denom string, addr string) string {
	out := "# Single coin balance\n\n"
	if !std.Address(addr).IsValid() {
		out += "Invalid address."
		return out
	}

	user, _ := users.ResolveAny(addr)
	name := "`" + addr + "`"
	if user != nil {
		name = user.RenderLink("")
	}

	out += ufmt.Sprintf("%s has `%d%s` at block #%d\n\n",
		name, banker.GetCoins(std.Address(addr)).AmountOf(denom), denom, std.ChainHeight())

	out += "[View full balance list for this address](/r/gnoland/coins:balances/" + addr + ")"

	return out
}

func renderAllBalances(rawpath, input string) string {
	out := "# Balances\n\n"

	if strings.HasPrefix(input, "cosmos") {
		addr, err := ctg.ConvertCosmosToGno(input)
		if err != nil {
			out += "Tried converting a Cosmos address to a Gno address but failed. Please double-scheck your input."
			return out
		}
		out += ufmt.Sprintf("> [!NOTE]\n>  Automatically converted `%s` to its Gno equivalent.\n\n", input)
		input = addr.String()
	} else {
		if !std.Address(input).IsValid() {
			out += "Invalid address."
			return out
		}
	}

	user, _ := users.ResolveAny(input)
	name := "`" + input + "`"
	if user != nil {
		name = user.RenderLink("")
	}

	banker := std.NewBanker(std.BankerTypeReadonly)
	out += ufmt.Sprintf("This page shows full coin balances of %s at block #%d\n\n",
		name, std.ChainHeight())

	req := realmpath.Parse(rawpath)

	coin := req.Query.Get("coin")
	if coin != "" {
		return renderSingleCoinBalance(banker, coin, input)
	}

	balances := banker.GetCoins(std.Address(input))

	// Determine sorting
	if getSortField(req) == "balance" {
		coinsort.SortByBalance(balances)
	}

	// Create table
	denomColumn := renderSortLink(req, "denom", "Denomination")
	balanceColumn := renderSortLink(req, "balance", "Balance")
	table := mdtable.Table{
		Headers: []string{denomColumn, balanceColumn},
	}

	if isSortReversed(req) {
		for _, b := range balances {
			table.Append([]string{b.Denom, strconv.Itoa(int(b.Amount))})
		}
	} else {
		for i := len(balances) - 1; i >= 0; i-- {
			table.Append([]string{balances[i].Denom, strconv.Itoa(int(balances[i].Amount))})
		}
	}

	out += table.String() + "\n\n"
	return out
}

// Helper functions for sorting and pagination
func getSortField(req *realmpath.Request) string {
	field := req.Query.Get("sort")
	switch field {
	case "denom", "balance": // XXX: add Coins.SortBy{denom,bal} methods
		return field
	}
	return "denom"
}

func isSortReversed(req *realmpath.Request) bool {
	return req.Query.Get("order") != "asc"
}

func renderSortLink(req *realmpath.Request, field, label string) string {
	currentField := getSortField(req)
	currentOrder := req.Query.Get("order")

	newOrder := "desc"
	if field == currentField && currentOrder != "asc" {
		newOrder = "asc"
	}

	query := make(url.Values)
	for k, vs := range req.Query {
		query[k] = append([]string(nil), vs...)
	}

	query.Set("sort", field)
	query.Set("order", newOrder)

	if field == currentField {
		if currentOrder == "asc" {
			label += " ↑"
		} else {
			label += " ↓"
		}
	}

	return md.Link(label, "?"+query.Encode())
}
