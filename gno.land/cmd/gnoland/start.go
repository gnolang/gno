package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type startCfg struct {
	skipFailingGenesisTxs bool
	skipStart             bool
	genesisBalancesFile   string
	genesisTxsFile        string
	chainID               string
	genesisRemote         string
	rootDir               string
	genesisMaxVMCycles    int64
	config                string
	nodeConfigPath        string
}

func newStartCmd(io *commands.IO) *commands.Command {
	cfg := &startCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "start",
			ShortUsage: "start [flags]",
			ShortHelp:  "Run the full node",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execStart(cfg, io)
		},
	)
}

func (c *startCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.skipFailingGenesisTxs,
		"skip-failing-genesis-txs",
		false,
		"don't panic when replaying invalid genesis txs",
	)

	fs.BoolVar(
		&c.skipStart,
		"skip-start",
		false,
		"quit after initialization, don't start the node",
	)

	fs.StringVar(
		&c.genesisBalancesFile,
		"genesis-balances-file",
		"./genesis/genesis_balances.txt",
		"initial distribution file",
	)

	fs.StringVar(
		&c.genesisTxsFile,
		"genesis-txs-file",
		"./genesis/genesis_txs.txt",
		"initial txs to replay",
	)

	fs.StringVar(
		&c.chainID,
		"chainid",
		"dev",
		"the ID of the chain",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"testdir",
		"directory for config and data",
	)

	fs.StringVar(
		&c.genesisRemote,
		"genesis-remote",
		"localhost:26657",
		"replacement for '%%REMOTE%%' in genesis",
	)

	fs.Int64Var(
		&c.genesisMaxVMCycles,
		"genesis-max-vm-cycles",
		10_000_000,
		"set maximum allowed vm cycles per operation. Zero means no limit.",
	)

	fs.StringVar(
		&c.config,
		"config",
		"",
		"the flag config file (optional)",
	)

	fs.StringVar(
		&c.nodeConfigPath,
		"tm2-node-config",
		"",
		"the node TOML config file path (optional)",
	)
}

func execStart(c *startCfg, io *commands.IO) error {
	logger := log.NewTMLogger(log.NewSyncWriter(io.Out))
	rootDir := c.rootDir

	var (
		cfg        *config.Config
		loadCfgErr error
	)

	// Set the node configuration
	if c.nodeConfigPath != "" {
		// Load the node configuration
		// from the specified path
		cfg, loadCfgErr = config.LoadConfigFile(c.nodeConfigPath)
	} else {
		// Load the default node configuration
		cfg, loadCfgErr = config.LoadOrMakeConfig(rootDir)
	}

	if loadCfgErr != nil {
		return fmt.Errorf("unable to load node configuration, %w", loadCfgErr)
	}

	// create priv validator first.
	// need it to generate genesis.json
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	// write genesis file if missing.
	genesisFilePath := filepath.Join(rootDir, cfg.Genesis)

	genesisTxs, genesisTxsErr := loadGenesisTxs(c.genesisTxsFile, c.chainID, c.genesisRemote)
	if genesisTxsErr != nil {
		return fmt.Errorf("unable to load genesis txs, %w", genesisTxsErr)
	}

	if !osm.FileExists(genesisFilePath) {
		genDoc, err := makeGenesisDoc(
			priv.GetPubKey(),
			c.chainID,
			c.genesisBalancesFile,
			genesisTxs,
		)
		if err != nil {
			return fmt.Errorf("unable to generate genesis.json, %w", err)
		}

		writeGenesisFile(genDoc, genesisFilePath)
	}

	// create application and node.
	gnoApp, err := gnoland.NewApp(rootDir, c.skipFailingGenesisTxs, logger, c.genesisMaxVMCycles)
	if err != nil {
		return fmt.Errorf("error in creating new app: %w", err)
	}

	cfg.LocalApp = gnoApp

	gnoNode, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		return fmt.Errorf("error in creating node: %w", err)
	}

	fmt.Fprintln(io.Err, "Node created.")

	if c.skipStart {
		fmt.Fprintln(io.Err, "'--skip-start' is set. Exiting.")

		return nil
	}

	if err := gnoNode.Start(); err != nil {
		return fmt.Errorf("error in start node: %w", err)
	}

	// run forever
	osm.TrapSignal(func() {
		if gnoNode.IsRunning() {
			_ = gnoNode.Stop()
		}
	})

	select {} // run forever
}

// Makes a local test genesis doc with local privValidator.
func makeGenesisDoc(
	pvPub crypto.PubKey,
	chainID string,
	genesisBalancesFile string,
	genesisTxs []std.Tx,
) (*bft.GenesisDoc, error) {
	gen := &bft.GenesisDoc{}

	gen.GenesisTime = time.Now()
	gen.ChainID = chainID
	gen.ConsensusParams = abci.ConsensusParams{
		Block: &abci.BlockParams{
			// TODO: update limits.
			MaxTxBytes:   1000000,  // 1MB,
			MaxDataBytes: 2000000,  // 2MB,
			MaxGas:       10000000, // 10M gas
			TimeIotaMS:   100,      // 100ms
		},
	}
	gen.Validators = []bft.GenesisValidator{
		{
			Address: pvPub.Address(),
			PubKey:  pvPub,
			Power:   10,
			Name:    "testvalidator",
		},
	}

	// Load distribution.
	balances, err := loadGenesisBalances(genesisBalancesFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load genesis balances, %w", err)
	}

	// Load initial packages from examples.
	test1 := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	txs := []std.Tx{}

	// List initial packages to load from examples.
	pkgs, err := gnomod.ListPkgs(filepath.Join("..", "examples"))
	if err != nil {
		panic(fmt.Errorf("listing gno packages: %w", err))
	}

	// Sort packages by dependencies.
	sortedPkgs, err := pkgs.Sort()
	if err != nil {
		panic(fmt.Errorf("sorting packages: %w", err))
	}

	// Filter out draft packages.
	nonDraftPkgs := sortedPkgs.GetNonDraftPkgs()

	for _, pkg := range nonDraftPkgs {
		// open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)

		var tx std.Tx
		tx.Msgs = []std.Msg{
			vmm.MsgAddPackage{
				Creator: test1,
				Package: memPkg,
				Deposit: nil,
			},
		}
		tx.Fee = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}

	// load genesis txs from file.
	txs = append(txs, genesisTxs...)

	// construct genesis AppState.
	gen.AppState = gnoland.GnoGenesisState{
		Balances: balances,
		Txs:      txs,
	}
	return gen, nil
}

func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
	err := gen.SaveAs(filePath)
	if err != nil {
		panic(err)
	}
}

func loadGenesisTxs(
	path string,
	chainID string,
	genesisRemote string,
) ([]std.Tx, error) {
	txs := make([]std.Tx, 0)

	if !osm.FileExists(path) {
		// No initial transactions
		return txs, nil
	}

	txsFile, openErr := os.Open(path)
	if openErr != nil {
		return nil, fmt.Errorf("unable to open genesis txs file, %w", openErr)
	}

	scanner := bufio.NewScanner(txsFile)

	for scanner.Scan() {
		txLine := scanner.Text()

		if txLine == "" {
			continue // skip empty line
		}

		// patch the TX
		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", chainID)
		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", genesisRemote)

		var tx std.Tx

		if unmarshalErr := amino.UnmarshalJSON([]byte(txLine), &tx); unmarshalErr != nil {
			return nil, fmt.Errorf("unable to amino unmarshal tx, %w", unmarshalErr)
		}

		txs = append(txs, tx)
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("error encountered while scanning, %w", scanErr)
	}

	return txs, nil
}

func loadGenesisBalances(path string) ([]string, error) {
	// each balance is in the form: g1xxxxxxxxxxxxxxxx=100000ugnot
	balances := make([]string, 0)

	if !osm.FileExists(path) {
		// No initial balances
		return balances, nil
	}

	balancesFile, openErr := os.Open(path)
	if openErr != nil {
		return nil, fmt.Errorf("unable to open genesis balances file, %w", openErr)
	}

	scanner := bufio.NewScanner(balancesFile)

	for scanner.Scan() {
		line := scanner.Text()

		line = strings.TrimSpace(line)

		// remove comments.
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)

		// skip empty lines.
		if line == "" {
			continue
		}

		if len(strings.Split(line, "=")) != 2 {
			return nil, fmt.Errorf("invalid genesis_balance line: %s", line)
		}

		balances = append(balances, line)
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("error encountered while scanning, %w", scanErr)
	}

	return balances, nil
}
