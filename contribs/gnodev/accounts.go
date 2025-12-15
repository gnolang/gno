package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type varPremineAccounts map[string]std.Coins // name or bech32 to coins.

func (va *varPremineAccounts) Set(value string) error {
	if *va == nil {
		*va = map[string]std.Coins{}
	}
	accounts := *va

	user, amount, found := strings.Cut(value, "=")
	accounts[user] = nil
	if !found {
		return nil
	}

	coins, err := std.ParseCoins(amount)
	if err != nil {
		return fmt.Errorf("unable to parse coins from %q: %w", user, err)
	}

	// Add the parsed amount to the user.
	accounts[user] = coins
	return nil
}

func (va varPremineAccounts) String() string {
	accs := make([]string, 0, len(va))
	for user, balance := range va {
		accs = append(accs, fmt.Sprintf("%s(%s)", user, balance.String()))
	}

	return strings.Join(accs, ",")
}

func generateBalances(bk *address.Book, cfg *AppConfig) (gnoland.Balances, error) {
	bls := gnoland.NewBalances()
	premineBalance := std.Coins{std.NewCoin(ugnot.Denom, 10e12)}

	entries := bk.List()

	// Automatically set every key from keybase to unlimited fund.
	for _, entry := range entries {
		address := entry.Address

		// Check if a predefined amount has been set for this key.

		// Check for address
		if preDefinedFound, ok := cfg.premineAccounts[address.String()]; ok && preDefinedFound != nil {
			bls[address] = gnoland.Balance{Amount: preDefinedFound, Address: address}
			continue
		}

		// Check for name
		found := premineBalance
		for _, name := range entry.Names {
			if preDefinedFound, ok := cfg.premineAccounts[name]; ok && preDefinedFound != nil {
				found = preDefinedFound
				break
			}
		}

		bls[address] = gnoland.Balance{Amount: found, Address: address}
	}

	if cfg.balancesFile == "" {
		return bls, nil
	}

	// Load balance file

	file, err := os.Open(cfg.balancesFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open balance file %q: %w", cfg.balancesFile, err)
	}

	blsFile, err := gnoland.GetBalancesFromSheet(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read balances file %q: %w", cfg.balancesFile, err)
	}

	// Add balance address to AddressBook
	for addr := range blsFile {
		bk.Add(addr, "")
	}

	// Left merge keybase balance into loaded file balance.
	// TL;DR: balance file override every balance at the end
	blsFile.LeftMerge(bls)
	return blsFile, nil
}

func logAccounts(ctx context.Context, logger *slog.Logger, book *address.Book, n *dev.Node) error {
	var tab strings.Builder
	tabw := tabwriter.NewWriter(&tab, 0, 0, 2, ' ', tabwriter.TabIndent)

	entries := book.List()

	fmt.Fprintln(tabw, "KeyName\tAddress\tBalance") // Table header.

	for _, entry := range entries {
		address := entry.Address.String()
		qres, err := n.Client().ABCIQuery(ctx, "auth/accounts/"+address, []byte{})
		if err != nil {
			return fmt.Errorf("unable to query account %q: %w", address, err)
		}

		var qret struct{ BaseAccount std.BaseAccount }
		if err = amino.UnmarshalJSON(qres.Response.Data, &qret); err != nil {
			return fmt.Errorf("unable to unmarshal query response: %w", err)
		}

		if len(entry.Names) == 0 {
			// Insert row with name, address, and balance amount.
			fmt.Fprintf(tabw, "%s\t%s\t%s\n", "_", address, qret.BaseAccount.GetCoins().String())
			continue
		}

		for _, name := range entry.Names {
			// Insert row with name, address, and balance amount.
			fmt.Fprintf(tabw, "%s\t%s\t%s\n", name,
				address,
				qret.BaseAccount.GetCoins().String())
		}
	}

	// Flush table.
	tabw.Flush()

	headline := fmt.Sprintf("(%d) known keys", len(entries))
	logger.Info(headline, "table", tab.String())
	return nil
}
