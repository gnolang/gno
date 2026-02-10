package txfees

import (
	"chain/banker"
	"chain/runtime"
	"strings"
)

func Render(path string) string {
	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	realmAddr := runtime.CurrentRealm().Address()
	balance := banker_.GetCoins(realmAddr).String()

	if strings.TrimSpace(balance) == "" {
		balance = "\\<empty\\>"
	}

	var output string
	output += "# Transaction Fees\n"
	output += "Balance: " + balance + "\n\n"

	output += "Bucket address: " + realmAddr.String() + "\n"
	return output
}
