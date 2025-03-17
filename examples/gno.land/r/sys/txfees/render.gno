package txfees

import (
	"std"
	"strings"
)

func Render(path string) string {
	banker := std.NewBanker(std.BankerTypeRealmSend)
	balance := banker.GetCoins(std.CurrentRealm().Address()).String()

	if strings.TrimSpace(balance) == "" {
		balance = "\\<empty\\>"
	}

	var output string
	output += "# Fee Coinbase\n"
	output += "Balance: " + balance + "\n\n"

	// XXX: Get the actual param value
	output += "Package address: " + std.CurrentRealm().Address().String() + "\n"

	return output
}
