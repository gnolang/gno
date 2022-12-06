package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/bft/config"
	"github.com/gnolang/gno/pkgs/bft/node"
	"github.com/gnolang/gno/pkgs/bft/privval"
	bft "github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/crypto"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/log"
	osm "github.com/gnolang/gno/pkgs/os"
	vmm "github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
)

func main() {
	args := os.Args[1:]
	err := runMain(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

var flags struct {
	skipFailingGenesisTxs bool
	skipStart             bool
	genesisBalancesFile   string
	genesisTxsFile        string
	chainID               string
	genesisRemote         string
	rootDir               string
}

func runMain(args []string) error {
	fs := flag.NewFlagSet("gnoland", flag.ExitOnError)
	fs.BoolVar(&flags.skipFailingGenesisTxs, "skip-failing-genesis-txs", false, "don't panic when replaying invalid genesis txs")
	fs.BoolVar(&flags.skipStart, "skip-start", false, "quit after initialization, don't start the node")
	fs.StringVar(&flags.genesisBalancesFile, "genesis-balances-file", "./gnoland/genesis/genesis_balances.txt", "initial distribution file")
	fs.StringVar(&flags.genesisTxsFile, "genesis-txs-file", "./gnoland/genesis/genesis_txs.txt", "initial txs to replay")
	fs.StringVar(&flags.chainID, "chainid", "dev", "chainid")
	fs.StringVar(&flags.rootDir, "root-dir", "testdir", "directory for config and data")
	fs.StringVar(&flags.genesisRemote, "genesis-remote", "localhost:26657", "replacement for '%%REMOTE%%' in genesis")
	fs.Parse(args)

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	rootDir := flags.rootDir
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
		genDoc := makeGenesisDoc(priv.GetPubKey())
		writeGenesisFile(genDoc, genesisFilePath)
	}

	// create application and node.
	gnoApp, err := gnoland.NewApp(rootDir, flags.skipFailingGenesisTxs, logger)
	if err != nil {
		return fmt.Errorf("error in creating new app: %w", err)
	}
	cfg.LocalApp = gnoApp
	gnoNode, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		return fmt.Errorf("error in creating node: %w", err)
	}
	fmt.Fprintln(os.Stderr, "Node created.")

	if flags.skipStart {
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
func makeGenesisDoc(pvPub crypto.PubKey) *bft.GenesisDoc {
	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Now()
	gen.ChainID = flags.chainID
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
	balances := loadGenesisBalances(flags.genesisBalancesFile)
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
		"p/demo/maths",
		"p/demo/blog",
		"r/demo/users",
		"r/demo/foo20",
		"r/demo/boards",
		"r/demo/banktest",
		"r/demo/types",
		"r/gnoland/blog",
		"r/gnoland/faucet",
		"r/system/validators",
		"r/system/names",
		"r/system/rewards",
	} {
		// open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(filepath.Join(".", "examples", "gno.land", path), "gno.land/"+path)
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
	genesisTxs := loadGenesisTxs(flags.genesisTxsFile)
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

func loadGenesisTxs(path string) []std.Tx {
	txs := []std.Tx{}
	txsBz := osm.MustReadFile(path)
	txsLines := strings.Split(string(txsBz), "\n")
	for _, txLine := range txsLines {
		if txLine == "" {
			continue // skip empty line
		}

		// patch the TX
		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", flags.chainID)
		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", flags.genesisRemote)

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
