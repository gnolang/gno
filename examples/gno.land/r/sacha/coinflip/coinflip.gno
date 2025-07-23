package coinflip

import (
	"std"
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
)

var latestBets = map[string]bool{}

func flip(choice int) {
	if choice != 0 && choice != 1 {
		panic("Invalid choice")
	}

	result := int(std.ChainHeight() % 2)
	caller := std.OriginCaller().String()

	latestBets[caller] = (result == choice)
}

func Heads(cur realm) {
	flip(1)
}

func Tails(cur realm) {
	flip(0)
}

func GetLatestResults() string {
	out := ""
	for address, result := range latestBets {
		out += "- **" + address + "**: "
		if result == true {
		    out += "won"
		} else {
		    out += "lost"
		}
		out += "\n"
	}
	return out;
}

func Render(_ string) string {
	out := "# "
	out += md.Link("Heads", txlink.Call("Heads")) + " or "
	out += md.Link("Tails", txlink.Call("Tails")) + " ?\n"

	out += "## Latests Results:\n\n"
	out += GetLatestResults()

	return out
}
