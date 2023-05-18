package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
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
	vmm "github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type gnolandCfg struct {
	skipFailingGenesisTxs bool
	skipStart             bool
	genesisBalancesFile   string
	genesisTxsFile        string
	chainID               string
	genesisRemote         string
	rootDir               string
	maxCycles             int64
}

func main() {
	cfg := &gnolandCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "[flags] [<arg>...]",
			LongHelp:   "Starts the gnoland blockchain node",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return exec(cfg)
		},
	)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)

		os.Exit(1)
	}
}

func (c *gnolandCfg) RegisterFlags(fs *flag.FlagSet) {
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
		&c.maxCycles,
		"max-vm-cycles",
		10*1000*1000,
		"set maximum allowed vm cycles per operation. Zero means no limit.",
	)
}

func exec(c *gnolandCfg) error {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	rootDir := c.rootDir

	cfg := config.LoadOrMakeConfigWithOptions(rootDir, func(cfg *config.Config) {
		cfg.Consensus.CreateEmptyBlocks = false
		cfg.Consensus.CreateEmptyBlocksInterval = 60 * time.Second
	})

	// create priv validator first.
	// need it to generate genesis.json
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	// write genesis file if missing.
	genesisFilePath := filepath.Join(rootDir, cfg.Genesis)
	if !osm.FileExists(genesisFilePath) {
		genDoc := makeGenesisDoc(
			priv.GetPubKey(),
			c.chainID,
			c.genesisBalancesFile,
			loadGenesisTxs(c.genesisTxsFile, c.chainID, c.genesisRemote),
		)
		writeGenesisFile(genDoc, genesisFilePath)
	}

	// create application and node.
	gnoApp, err := gnoland.NewApp(rootDir, c.skipFailingGenesisTxs, logger, c.maxCycles)
	if err != nil {
		return fmt.Errorf("error in creating new app: %w", err)
	}

	cfg.LocalApp = gnoApp

	gnoNode, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		return fmt.Errorf("error in creating node: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Node created.")

	if c.skipStart {
		fmt.Fprintln(os.Stderr, "'--skip-start' is set. Exiting.")

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
) *bft.GenesisDoc {
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
	balances := loadGenesisBalances(genesisBalancesFile)
	// debug: for _, balance := range balances { fmt.Println(balance) }

	// Load initial packages from examples.
	test1 := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	txs := []std.Tx{}
	for _, path := range []string{
		"p/demo/ufmt",
		"p/demo/avl",
		"p/demo/grc/exts",
		"p/demo/grc/grc20",
		"p/demo/grc/grc721",
		"p/demo/grc/grc1155",
		"p/demo/maths",
		"p/demo/blog",
		"r/demo/users",
		"r/demo/foo20",
		"r/demo/foo1155",
		"r/demo/boards",
		"r/demo/banktest",
		"r/demo/types",
		"r/demo/markdown_test",
		"r/gnoland/blog",
		"r/gnoland/faucet",
		"r/system/validators",
		"r/system/names",
		"r/system/rewards",
		"r/demo/deep/very/deep",
	} {
		// open files in directory as MemPackage.
		fsPath := filepath.Join("..", "examples", "gno.land", path)
		importPath := "gno.land/" + path
		memPkg := gno.ReadMemPackage(fsPath, importPath)
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
	return gen
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
) []std.Tx {
	txs := []std.Tx{}
	txsBz := osm.MustReadFile(path)
	txsLines := strings.Split(string(txsBz), "\n")
	for _, txLine := range txsLines {
		if txLine == "" {
			continue // skip empty line
		}

		// patch the TX
		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", chainID)
		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", genesisRemote)

		var tx std.Tx
		amino.MustUnmarshalJSON([]byte(txLine), &tx)
		txs = append(txs, tx)
	}

	return txs
}

func loadGenesisBalances(path string) []string {
	// each balance is in the form: g1xxxxxxxxxxxxxxxx=100000ugnot
	balances := []string{}
	content := osm.MustReadFile(path)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// remove comments.
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)

		// skip empty lines.
		if line == "" {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			panic("invalid genesis_balance line: " + line)
		}

		balances = append(balances, line)
	}
	return balances
}
