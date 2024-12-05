package txs

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errInvalidPackageDir   = errors.New("invalid package directory")
	errInvalidDeployerAddr = errors.New("invalid deployer address")
)

// Keep in sync with gno.land/cmd/start.go
var (
	defaultCreator   = crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // test1
	genesisDeployFee = std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))
)

type addPkgCfg struct {
	txsCfg                *txsCfg
	keyName               string
	gnoHome               string
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
		gnoenv.HomeDir(),
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

	var (
		creator = defaultCreator
		pass    string
	)
	kb, err := keys.NewKeyBaseFromDir(cfg.gnoHome)
	if err != nil {
		return err
	}
	if cfg.keyName != "" {
		info, err := kb.GetByNameOrAddress(cfg.keyName)
		if err != nil {
			return err
		}
		creator = info.GetAddress()
		pass, err = io.GetPassword("Enter password.", cfg.insecurePasswordStdin)
		if err != nil {
			return fmt.Errorf("cannot read password: %w", err)
		}
	}

	parsedTxs := make([]gnoland.TxWithMetadata, 0)
	for _, path := range args {
		// Generate transactions from the packages (recursively)
		txs, err := gnoland.LoadPackagesFromDir(path, creator, genesisDeployFee)
		if err != nil {
			return fmt.Errorf("unable to load txs from directory, %w", err)
		}
		if creator != defaultCreator {
			if err := signTxs(txs, cfg.keyName, genesis.ChainID, kb, pass); err != nil {
				return fmt.Errorf("unable to sign txs, %w", err)
			}
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

func signTxs(txs []gnoland.TxWithMetadata, keyName string, chainID string, kb keys.Keybase, pass string) error {
	for index, tx := range txs {
		signBytes, err := tx.Tx.GetSignBytes(chainID, 0, 0)
		if err != nil {
			return fmt.Errorf("unable to load txs from directory, %w", err)
		}
		signature, publicKey, err := kb.Sign(keyName, pass, signBytes)
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
