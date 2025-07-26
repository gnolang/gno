package coinflip

import (
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
	"strconv"
)

var (
	githubUsername string
	pocInnovationName string
)

func init() {
	githubUsername = "dujardin09"
	pocInnovationName = "PoCInnovation"
}

func renderResult(
	address string,
	result string,
	amount string,
	emoji string,
) string {
	addressFmt := "**" + address[:30] + "**: "
	resultFmt := result + " **" + amount + denom + "** " + emoji
	return addressFmt + resultFmt
}

func renderLatestResults() []string {
	latestsResults := GetLatestResults(cross)
	out := []string{}
	count := 0

	for _, result := range latestsResults {
		out = append(out, renderResult(result[0], result[1], result[3], result[2]))
		count++
		if count == 8 {
			break
		}
	}
	if count == 0 {
		out = append(out, "No results yet, time to flip a coin!")
	}
	return out
}

func renderGame() string {
	out := "# "
	out += md.Link("Heads", txlink.Call("Heads")) + " or "
	out += md.Link("Tails", txlink.Call("Tails")) + " ?\n"

	out += md.H3("Latests Results:\n\n")
	out += md.BulletList(renderLatestResults())

	return out
}

func renderRight() string {
	out := md.Image("Coin Flip Illustration", "https://cdn.dribbble.com/userupload/23852777/file/original-7c6789feec78629e91335795a5baf105.gif")
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
