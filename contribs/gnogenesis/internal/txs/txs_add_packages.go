package txs

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/crypto/keys"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errInvalidPackageDir = errors.New("invalid package directory")

// Keep in sync with gno.land/cmd/start.go
var genesisDeployFee = std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))

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
		"The package deployer key name or address",
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
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.txsCfg.GenesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Make sure the package dir is set
	if len(args) == 0 {
		return errInvalidPackageDir
	}

	signer, err := signerWithConfig(cfg, io, genesis.ChainID)
	if err != nil {
		return fmt.Errorf("unable to load signer, %w", err)
	}

	info, err := signer.Info()
	if err != nil {
		return fmt.Errorf("unable to get signer info, %w", err)
	}

	parsedTxs := make([]gnoland.TxWithMetadata, 0)
	for _, path := range args {
		// Generate transactions from the packages (recursively)
		txs, err := gnoland.LoadPackagesFromDir(path, info.GetAddress(), genesisDeployFee)
		if err != nil {
			return fmt.Errorf("unable to load txs from directory, %w", err)
		}

		if err := signTxs(txs, signer); err != nil {
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

func signTxs(txs []gnoland.TxWithMetadata, signer *gnoclient.SignerFromKeybase) error {
	for index, tx := range txs {
		signBytes, err := tx.Tx.GetSignBytes(signer.ChainID, 0, 0)
		if err != nil {
			return fmt.Errorf("unable to load txs from directory, %w", err)
		}
		signature, publicKey, err := signer.Keybase.Sign(signer.Account, signer.Password, signBytes)
		txs[index].Tx.Signatures = []std.Signature{
			{
				PubKey:    publicKey,
				Signature: signature,
			},
		}
		if err != nil {
			return fmt.Errorf("unable sign tx %w", err)
		}
	}

	return nil
}

func signerWithConfig(cfg *addPkgCfg, io commands.IO, chainID string) (*gnoclient.SignerFromKeybase, error) {
	var (
		keyname = integration.DefaultAccount_Name
		pass    string
		kb      keys.Keybase
		err     error
	)

	if cfg.keyName != "" {
		keyname = cfg.keyName
		kb, err = keys.NewKeyBaseFromDir(cfg.gnoHome)
		if err != nil {
			return nil, fmt.Errorf("unable to load keybase: %w", err)
		}
		pass, err = io.GetPassword("Enter password.", cfg.insecurePasswordStdin)
		if err != nil {
			return nil, fmt.Errorf("cannot read password: %w", err)
		}
	} else {
		kb = keys.NewInMemory()
		kb.CreateAccount(integration.DefaultAccount_Name, integration.DefaultAccount_Seed, "", "", 0, 0)
	}

	return &gnoclient.SignerFromKeybase{
		Account:  keyname,
		ChainID:  chainID,
		Keybase:  kb,
		Password: pass,
	}, nil
}
