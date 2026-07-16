package txfees

import (
	"chain/banker"
	"strings"
)

func Render(cur realm, _ string) string {
	banker_ := banker.NewReadonlyBanker()
	realmAddr := cur.Address()
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
