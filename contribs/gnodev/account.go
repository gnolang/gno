package main

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

type varAccounts map[string]std.Coins // name or bech32 -> coins

func (va *varAccounts) Set(value string) error {
	if *va == nil {
		*va = map[string]std.Coins{}
	}
	accounts := *va

	user, amount, found := strings.Cut(value, ":")
	accounts[user] = nil
	if !found {
		return nil
	}

	coins, err := std.ParseCoins(amount)
	if err != nil {
		return fmt.Errorf("unable to parse coins from %q: %w", user, err)
	}

	// Add the parsed amount into user
	accounts[user] = coins
	return nil
}

func (va varAccounts) String() string {
	accs := make([]string, 0, len(va))
	for user, balance := range va {
		accs = append(accs, fmt.Sprintf("%s(%s)", user, balance.String()))
	}

	return strings.Join(accs, ",")
}
