package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type varPremineAccounts map[string]std.Coins // name or bech32 to coins.

func (va *varPremineAccounts) Set(value string) error {
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

func generateBalances(kb keys.Keybase, cfg *devCfg) (gnoland.Balances, error) {
	bls := gnoland.NewBalances()
	unlimitedFund := std.Coins{std.NewCoin("ugnot", 10e12)}

	keys, err := kb.List()
	if err != nil {
		return nil, fmt.Errorf("unable to list keys from keybase: %w", err)
	}

	// Automatically set every key from keybase to unlimited fund.
	for _, key := range keys {
		address := key.GetAddress()

		// Check if a predefined amount has been set for this key.
		found := unlimitedFund
		if preDefinedFound, ok := cfg.additionalAccounts[key.GetName()]; ok && preDefinedFound != nil {
			found = preDefinedFound
		} else if preDefinedFound, ok := cfg.additionalAccounts[address.String()]; ok && preDefinedFound != nil {
			found = preDefinedFound
		}

		bls[address] = gnoland.Balance{Amount: found, Address: address}
	}

	if cfg.balancesFile == "" {
		return bls, nil
	}

	file, err := os.Open(cfg.balancesFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open balance file %q: %w", cfg.balancesFile, err)
	}

	blsFile, err := gnoland.GetBalancesFromSheet(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read balances file %q: %w", cfg.balancesFile, err)
	}

	// Left merge keybase balance into loaded file balance.
	blsFile.LeftMerge(bls)
	return blsFile, nil
}

func logAccounts(logger *slog.Logger, kb keys.Keybase, _ *dev.Node) error {
	keys, err := kb.List()
	if err != nil {
		return fmt.Errorf("unable to get keybase keys list: %w", err)
	}

	var tab strings.Builder
	tabw := tabwriter.NewWriter(&tab, 0, 0, 2, ' ', tabwriter.TabIndent)

	fmt.Fprintln(tabw, "KeyName\tAddress\tBalance") // Table header.
	for _, key := range keys {
		if key.GetName() == "" {
			continue // Skip empty key name.
		}

		address := key.GetAddress()
		qres, err := client.NewLocal().ABCIQuery("auth/accounts/"+address.String(), []byte{})
		if err != nil {
			return fmt.Errorf("unable to query account %q: %w", address.String(), err)
		}

		var qret struct{ BaseAccount std.BaseAccount }
		if err = amino.UnmarshalJSON(qres.Response.Data, &qret); err != nil {
			return fmt.Errorf("unable to unmarshal query response: %w", err)
		}

		// Insert row with name, address, and balance amount.
		fmt.Fprintf(tabw, "%s\t%s\t%s\n", key.GetName(),
			address.String(),
			qret.BaseAccount.GetCoins().String())
	}
	// Flush table.
	tabw.Flush()

	headline := fmt.Sprintf("(%d) known accounts", len(keys))
	logger.Info(headline, "table", tab.String())
	return nil
}
