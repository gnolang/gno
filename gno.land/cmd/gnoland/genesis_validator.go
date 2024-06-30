package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gno.land/pkg/valset"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const valsRealm = "gno.land/r/sys/vals"

var errMissingSysPoCDeployment = errors.New("missing r/sys/vals deployment for PoC")

type validatorCfg struct {
	commonCfg

	address string
}

// newValidatorCmd creates the genesis validator subcommand
func newValidatorCmd(io commands.IO) *commands.Command {
	cfg := &validatorCfg{
		commonCfg: commonCfg{},
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "validator",
			ShortUsage: "validator <subcommand> [flags]",
			ShortHelp:  "validator set management in genesis.json",
			LongHelp:   "Manipulates the genesis.json validator set",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newValidatorAddCmd(cfg, io),
		newValidatorRemoveCmd(cfg, io),
	)

	return cmd
}

func (c *validatorCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.address,
		"address",
		"",
		"the gno bech32 address of the validator",
	)
}

// alignChainValset aligns the validator set in the genesis state with any on-chain valset protocol
func alignChainValset(genesisPath string, genesis *types.GenesisDoc) error {
	// Construct the config path
	var (
		nodeDir    = filepath.Join(filepath.Dir(genesisPath), defaultNodeDir)
		configPath = constructConfigPath(nodeDir)

		cfg = config.DefaultConfig()
		err error
	)

	// Check if there is an existing config file
	if osm.FileExists(configPath) {
		// Attempt to grab the config from disk
		cfg, err = config.LoadConfig(nodeDir)
		if err != nil {
			return fmt.Errorf("unable to load config file, %w", err)
		}
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)

	switch cfg.ValsetProtocol {
	case config.ProofOfContribution:
		// Find the /r/sys/vals deploy transaction
		pkg := findSysValsDeployment(state.Txs)
		if pkg == nil {
			return errMissingSysPoCDeployment
		}

		// Modify the deploy transaction to include the current
		// genesis.json validator set
		if err := valset.ModifyPoCDeployment(pkg, genesis.Validators); err != nil {
			return fmt.Errorf("unable to modify PoC deployment, %w", err)
		}

		// Update the app state
		genesis.AppState = state
	default:
		// No on-chain valset protocol
		return nil
	}

	return nil
}

// findSysValsDeployment finds the package deployment for `r/sys/vals`,
// among the given transaction list. Returns nil if no deployment was found
func findSysValsDeployment(txs []std.Tx) *std.MemPackage {
	addPkgType := vm.MsgAddPackage{}.Type()

	for _, tx := range txs {
		for _, msg := range tx.Msgs {
			// Make sure the transaction is a deploy-tx
			if msg.Type() != addPkgType {
				continue
			}

			// Cast the message
			addPkg := msg.(vm.MsgAddPackage)

			// Check if the message is a Realm
			// deployment for r/sys/vals
			if addPkg.Package.Path != valsRealm {
				continue
			}

			return addPkg.Package
		}
	}

	return nil
}
