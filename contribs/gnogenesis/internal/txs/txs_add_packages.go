package txs

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/crypto/keys"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	defaultAccount_Name      = "test1"
	defaultAccount_Seed      = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
	defaultAccount_publicKey = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj"
)

var errInvalidPackageDir = errors.New("invalid package directory")

// Keep in sync with gno.land/cmd/start.go
var genesisDeployFee = std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1)))

type addPkgCfg struct {
	txsCfg                *txsCfg
	keyName               string
	gnoHome               string // default GNOHOME env var, just here to ease testing with parallel tests
	insecurePasswordStdin bool
}

func (c *addPkgCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.keyName,
		"key-name",
		"",
		"The package deployer key name or address contained on gnokey",
	)

	fs.StringVar(
		&c.gnoHome,
		"gno-home",
		os.Getenv("GNOHOME"),
		"the gno home directory",
	)

	fs.BoolVar(
		&c.insecurePasswordStdin,
		"insecure-password-stdin",
		false,
		"the gno home directory",
	)
}

// newTxsAddPackagesCmd creates the genesis txs add packages subcommand
func newTxsAddPackagesCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	cfg := &addPkgCfg{
		txsCfg: txsCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "packages",
			ShortUsage: "txs add packages <package-path ...>",
			ShortHelp:  "imports transactions from the given packages into the genesis.json",
			LongHelp:   "Imports the transactions from a given package directory recursively to the genesis.json",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execTxsAddPackages(cfg, io, args)
		},
	)
}

func execTxsAddPackages(
	cfg *addPkgCfg,
	io commands.IO,
	args []string,
) error {
	var (
		keyName = defaultAccount_Name
		keybase keys.Keybase
		pass    string
	)

	// Load the genesis
	genesis, err := types.GenesisDocFromFile(cfg.txsCfg.GenesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	// Make sure the package dir is set
	if len(args) == 0 {
		return errInvalidPackageDir
	}

	if cfg.keyName != "" {
		keyName = cfg.keyName
		keybase, err = keys.NewKeyBaseFromDir(cfg.gnoHome)
		if err != nil {
			return fmt.Errorf("unable to load keybase: %w", err)
		}
		pass, err = io.GetPassword("Enter password.", cfg.insecurePasswordStdin)
		if err != nil {
			return fmt.Errorf("cannot read password: %w", err)
		}
	} else {
		keybase = keys.NewInMemory()
		_, err := keybase.CreateAccount(defaultAccount_Name, defaultAccount_Seed, "", "", 0, 0)
		if err != nil {
			return fmt.Errorf("unable to create account: %w", err)
		}
	}

	info, err := keybase.GetByNameOrAddress(keyName)
	if err != nil {
		return fmt.Errorf("unable to find key in keybase: %w", err)
	}

	creator := info.GetAddress()
	parsedTxs := make([]gnoland.TxWithMetadata, 0)
	for _, path := range args {
		// Generate transactions from the packages (recursively)
		txs, err := gnoland.LoadPackagesFromDir(path, creator, genesisDeployFee)
		if err != nil {
			return fmt.Errorf("unable to load txs from directory, %w", err)
		}

		if err := signTxs(txs, keybase, genesis.ChainID, keyName, pass); err != nil {
			return fmt.Errorf("unable to sign txs, %w", err)
		}

		parsedTxs = append(parsedTxs, txs...)
	}

	// Save the txs to the genesis.json
	if err := appendGenesisTxs(genesis, parsedTxs); err != nil {
		return fmt.Errorf("unable to append genesis transactions, %w", err)
	}

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.txsCfg.GenesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Saved %d transactions to genesis.json",
		len(parsedTxs),
	)

	return nil
}

func signTxs(txs []gnoland.TxWithMetadata, keybase keys.Keybase, chainID, keyname string, password string) error {
	for index, tx := range txs {
		// Here accountNumber and sequenceNumber are set to 0 because they are considered as 0 on genesis transactions.
		signBytes, err := tx.Tx.GetSignBytes(chainID, 0, 0)
		if err != nil {
			return fmt.Errorf("unable to load txs from directory, %w", err)
		}
		signature, publicKey, err := keybase.Sign(keyname, password, signBytes)
		if err != nil {
			return fmt.Errorf("unable sign tx %w", err)
		}
		txs[index].Tx.Signatures = []std.Signature{
			{
				PubKey:    publicKey,
				Signature: signature,
			},
		}
	}

	return nil
}
