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
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
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

func setupKeybase(logger *slog.Logger, cfg *devCfg) (keys.Keybase, error) {
	kb := keys.NewInMemory()
	if cfg.home != "" {
		// Load home keybase into our inMemory keybase
		kbHome, err := keys.NewKeyBaseFromDir(cfg.home)
		if err != nil {
			return nil, fmt.Errorf("unable to load keybase from dir %q: %w", cfg.home, err)
		}

		keys, err := kbHome.List()
		if err != nil {
			return nil, fmt.Errorf("unable to list keys from keybase %q: %w", cfg.home, err)
		}

		for _, key := range keys {
			name := key.GetName()
			armor, err := kbHome.Export(name)
			if err != nil {
				return nil, fmt.Errorf("unable to export key %q: %w", name, err)
			}

			if err := kb.Import(name, armor); err != nil {
				return nil, fmt.Errorf("unable to import key %q: %w", name, err)
			}
		}
	}

	// Add additional users to our keybase
	for user := range cfg.additionalUsers {
		info, err := createAccount(kb, user)
		if err != nil {
			return nil, fmt.Errorf("unable to create user %q: %w", user, err)
		}

		logger.Info("additional user", "name", info.GetName(), "addr", info.GetAddress())
	}

	// Next, make sure that we have a default address to load packages
	info, err := kb.GetByNameOrAddress(cfg.genesisCreator)
	switch {
	case err == nil: // user already have a default user
		break
	case keyerror.IsErrKeyNotFound(err):
		// If the key isn't found, create a default one
		creatorName := fmt.Sprintf("_default#%.10s", DefaultCreatorAddress.String())
		if ok, _ := kb.HasByName(creatorName); ok {
			return nil, fmt.Errorf("unable to create creator account, delete %q from your keybase", creatorName)
		}

		info, err = kb.CreateAccount(creatorName, DefaultCreatorSeed, "", "", 0, 0)
		if err != nil {
			return nil, fmt.Errorf("unable to create default %q account: %w", DefaultCreatorName, err)
		}
	default:
		return nil, fmt.Errorf("unable to get address %q from keybase: %w", info.GetAddress(), err)
	}

	logger.Info("default creator", "name", info.GetName(), "addr", info.GetAddress())
	return kb, nil
}

func generateBalances(kb keys.Keybase, cfg *devCfg) (gnoland.Balances, error) {
	bls := gnoland.NewBalances()
	unlimitedFund := std.Coins{std.NewCoin("ugnot", 10e12)}

	keys, err := kb.List()
	if err != nil {
		return nil, fmt.Errorf("unable to list keys from keybase: %w", err)
	}

	// Automatically set every key from keybase to unlimited found (or pre
	// defined found if specified)
	for _, key := range keys {
		found := unlimitedFund
		if preDefinedFound, ok := cfg.additionalUsers[key.GetName()]; ok && preDefinedFound != nil {
			found = preDefinedFound
		}

		address := key.GetAddress()
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

	// Left merge keybase balance into loaded file balance
	blsFile.LeftMerge(bls)
	return blsFile, nil
}

func logAccounts(logger *slog.Logger, kb keys.Keybase, _ *dev.Node) error {
	keys, err := kb.List()
	if err != nil {
		return fmt.Errorf("unable to get keybase keys list: %w", err)
	}

	var tab strings.Builder
	tab.WriteRune('\n')
	tabw := tabwriter.NewWriter(&tab, 0, 0, 2, ' ', tabwriter.TabIndent)

	fmt.Fprintln(tabw, "KeyName\tAddress\tBalance") // Table header
	for _, key := range keys {
		if key.GetName() == "" {
			continue // skip empty key name
		}

		address := key.GetAddress()
		// XXX: use client from node from argument, should be exposed by the node directly
		qres, err := client.NewLocal().ABCIQuery("auth/accounts/"+address.String(), []byte{})
		if err != nil {
			return fmt.Errorf("unable to querry account %q: %w", address.String(), err)
		}

		var qret struct{ BaseAccount std.BaseAccount }
		if err = amino.UnmarshalJSON(qres.Response.Data, &qret); err != nil {
			return fmt.Errorf("unable to unmarshal query response: %w", err)
		}

		// Insert row with name, addr, balance amount
		fmt.Fprintf(tabw, "%s\t%s\t%s\n", key.GetName(),
			address.String(),
			qret.BaseAccount.GetCoins().String())
	}
	// Flush table
	tabw.Flush()

	headline := fmt.Sprintf("(%d) known accounts", len(keys))
	logger.Info(headline, "table", tab.String())
	return nil
}

// createAccount creates a new account with the given name and adds it to the keybase.
func createAccount(kb keys.Keybase, accountName string) (keys.Info, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, fmt.Errorf("error creating entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, fmt.Errorf("error generating mnemonic: %w", err)
	}

	return kb.CreateAccount(accountName, mnemonic, "", "", 0, 0)
}
