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
	"chain/banker"
	"chain/runtime"
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/leon/coinsort"
	"gno.land/p/leon/ctg"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/sys/users"
)

var router *mux.Router

func init() {
	router = mux.NewRouter()

	router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(renderHomepage())
	})

	router.HandleFunc("balances", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(renderBalances(req))
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

This is a simple, readonly realm that allows users to browse native coin balances. Check your coin balance below!

<gno-form path="balances">
	<gno-input name="address" type="text" placeholder="Valid bech32 address (e.g. g1..., cosmos1..., osmo1...)" />
	<gno-input name="coin" type="text" placeholder="Coin (e.g. ugnot)"" />
</gno-form>

Here are a few more ways to use this app:

- ~/r/gnoland/coins:balances?address=g1...~ - show full list of coin balances of an address
	- [Example](/r/gnoland/coins:balances?address=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5)
- ~/r/gnoland/coins:balances?address=g1...&coin=ugnot~ - shows the balance of an address for a specific coin
	- [Example](/r/gnoland/coins:balances?address=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5&coin=ugnot)
- ~/r/gnoland/coins:convert/<bech32_addr>~ - convert a bech32 address to a Gno address
	- [Example](/r/gnoland/coins:convert/cosmos1jg8mtutu9khhfwc4nxmuhcpftf0pajdh6svrgs)
- ~/r/gnoland/coins:supply/<denom>~ - shows the total supply of denom
	- Coming soon!

`, "~", "`", -1)
}

func renderBalances(req *mux.Request) string {
	out := "# Balances\n\n"

	input := req.Query.Get("address")
	coin := req.Query.Get("coin")

	if input == "" && coin == "" {
		out += "Please input a valid address and coin denomination.\n\n"
		return out
	}

	if input == "" {
		out += "Please input a valid bech32 address.\n\n"
		return out
	}

	originalInput := input
	var wasConverted bool

	// Try to validate or convert
	if !address(input).IsValid() {
		addr, err := ctg.ConvertAnyToGno(input)
		if err != nil {
			return out + ufmt.Sprintf("Tried converting `%s` to a Gno address but failed. Please try with a valid bech32 address.\n\n", input)
		}
		input = addr.String()
		wasConverted = true
	}

	if wasConverted {
		out += ufmt.Sprintf("> [!NOTE]\n>  Automatically converted `%s` to its Gno equivalent.\n\n", originalInput)
	}

	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	balances := banker_.GetCoins(address(input))

	if len(balances) == 0 {
		out += "This address currently has no coins."
		return out
	}

	if coin != "" {
		return renderSingleCoinBalance(coin, input, originalInput, wasConverted)
	}

	user, _ := users.ResolveAny(input)
	name := "`" + input + "`"
	if user != nil {
		name = user.RenderLink("")
	}

	out += ufmt.Sprintf("This page shows full coin balances of %s at block #%d\n\n",
		name, runtime.ChainHeight())

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

func renderSingleCoinBalance(denom, addr, origInput string, wasConverted bool) string {
	out := "# Coin balance\n\n"
	banker_ := banker.NewBanker(banker.BankerTypeReadonly)

	if wasConverted {
		out += ufmt.Sprintf("> [!NOTE]\n>  Automatically converted `%s` to its Gno equivalent.\n\n", origInput)
	}

	user, _ := users.ResolveAny(addr)
	name := "`" + addr + "`"
	if user != nil {
		name = user.RenderLink("")
	}

	out += ufmt.Sprintf("%s has `%d%s` at block #%d\n\n",
		name, banker_.GetCoins(address(addr)).AmountOf(denom), denom, runtime.ChainHeight())

	out += "[View full balance list for this address](/r/gnoland/coins:balances?address=" + addr + ")"

	return out
}

func renderConvertedAddress(addr string) string {
	out := "# Address converter\n\n"

	gnoAddress, err := ctg.ConvertAnyToGno(addr)
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
	out += "[[View `ugnot` balance for this address]](/r/gnoland/coins:balances?address=" + gnoAddress.String() + "&coin=ugnot) - "
	out += "[[View full balance list for this address]](/r/gnoland/coins:balances?address=" + gnoAddress.String() + ")"
	return out
}

// Helper functions for sorting and pagination
func getSortField(req *mux.Request) string {
	field := req.Query.Get("sort")
	switch field {
	case "denom", "balance":
		return field
	}
	return "denom"
}

func isSortReversed(req *mux.Request) bool {
	return req.Query.Get("order") != "asc"
}

func renderSortLink(req *mux.Request, field, label string) string {
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
