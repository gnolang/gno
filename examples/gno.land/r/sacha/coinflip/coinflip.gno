package coinflip

import (
	"std"
)

type Bet struct {
    Address string
    Result bool
}

var latestBets = []Bet{}

func flip(choice int) {
	if choice != 0 && choice != 1 {
		panic("Invalid choice")
	}

	result := int(std.ChainHeight() % 2)
	caller := std.OriginCaller().String()

	latestBets = append([]Bet{{Address: caller, Result: result == choice}}, latestBets...)
}

func Heads(cur realm) {
	flip(1)
}

func Tails(cur realm) {
	flip(0)
}

func GetLatestResults() string {
	out := ""
	for _, bet := range latestBets {
		out += "- **" + bet.Address + "**: "
		if bet.Result == true {
		    out += "won"
		} else {
		    out += "lost"
		}
		out += "\n"
	}
	return out;
}
