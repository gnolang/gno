package txfees

import (
	"std"
	"strings"
)

func Render(path string) string {
	banker := std.NewBanker(std.BankerTypeReadonly)
	realmAddr := std.CurrentRealm().Address()
	balance := banker.GetCoins(realmAddr).String()

	if strings.TrimSpace(balance) == "" {
		balance = "\\<empty\\>"
	}

	var output string
	output += "# Transaction Fees\n"
	output += "Balance: " + balance + "\n\n"

	output += "Bucket address: " + realmAddr.String() + "\n"
	return output
}
