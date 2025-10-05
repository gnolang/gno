package coinflip

import (
	"chain"
	"strconv"

	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"
)

var (
	githubUsername    string
	pocInnovationName string
)

func init() {
	githubUsername = "dujardin09"
	pocInnovationName = "PoCInnovation"
}

func renderResult(address address, result bool, amount chain.Coin) string {
	addressFmt := "**" + address.String()[:30] + "**: "
	resultFmt := "lost"
	amountFmt := " **" + amount.String() + "** "
	emoji := "📉"

	data := users.ResolveAddress(address)
	if data != nil {
		addressFmt = "**" + data.RenderLink("") + "**: "
	}

	if result == true {
		resultFmt = "won"
		emoji = "📈"
	}

	return addressFmt + resultFmt + amountFmt + emoji
}

func renderLatestResults() string {
	size := latestBets.Size()

	if size == 0 {
		return "No results yet"
	}

	entries := latestBets.Entries()

	var out string
	for i := size - 1; i >= 0; i-- {
		bet := entries[i].(Bet)
		out += md.BulletItem(renderResult(bet.Address, bet.Result, bet.Amount))
	}
	return out
}

func renderGame() string {
	out := "# "
	out += md.Link("Heads", txlink.Call("Heads")) + " or "
	out += md.Link("Tails", txlink.Call("Tails")) + " ?\n"

	out += md.H3("Latests Results:\n\n")
	out += renderLatestResults()

	return out
}

func renderRight() string {
	out := md.Image("Coin Flip Illustration", "https://ipfs.io/ipfs/QmVZxCKz3J5iMHrACNLfFWpfpfcKT2aSmZMfvZsCQnMWbf")
	out += "\n"
	out += renderBalance()
	return out
}

func renderFooter() string {
	out := md.HorizontalRule()
	out += md.BulletList([]string{
		md.Link("Home", "home"),
		md.Link("GitHub: @"+githubUsername, "https://github.com/"+githubUsername),
		md.Link("PoC-Innovation", "https://github.com/"+pocInnovationName),
	})
	return out
}

func renderBalance() string {
	balance := strconv.Itoa(int(GetRealmBalance(cross)))
	out := md.H2("Balance: " + balance + denom + "\n")

	return out
}

func Render(_ string) string {
	out := md.Columns([]string{
		renderGame(),
		renderRight(),
	}, false)

	out += renderFooter()

	return out
}
