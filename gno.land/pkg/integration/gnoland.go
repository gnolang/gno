package integration

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/rogpeppe/go-internal/testscript"
)

type IntegrationConfig struct {
	SkipFailingGenesisTxs bool
	SkipStart             bool
	GenesisBalancesFile   string
	GenesisTxsFile        string
	ChainID               string
	GenesisRemote         string
	RootDir               string
	GenesisMaxVMCycles    int64
	Config                string
}

// NOTE: This is a copy of gnoland actual flags.
// XXX: A lot this make no sense for integration.
func (c *IntegrationConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.SkipFailingGenesisTxs,
		"skip-failing-genesis-txs",
		false,
		"don't panic when replaying invalid genesis txs",
	)
	fs.BoolVar(
		&c.SkipStart,
		"skip-start",
		false,
		"quit after initialization, don't start the node",
	)

	fs.StringVar(
		&c.GenesisBalancesFile,
		"genesis-balances-file",
		"./genesis/genesis_balances.txt",
		"initial distribution file",
	)

	fs.StringVar(
		&c.GenesisTxsFile,
		"genesis-txs-file",
		"./genesis/genesis_txs.txt",
		"initial txs to replay",
	)

	fs.StringVar(
		&c.ChainID,
		"chainid",
		"dev",
		"the ID of the chain",
	)

	fs.StringVar(
		&c.RootDir,
		"root-dir",
		"testdir",
		"directory for config and data",
	)

	fs.StringVar(
		&c.GenesisRemote,
		"genesis-remote",
		"localhost:26657",
		"replacement for '%%REMOTE%%' in genesis",
	)

	fs.Int64Var(
		&c.GenesisMaxVMCycles,
		"genesis-max-vm-cycles",
		10_000_000,
		"set maximum allowed vm cycles per operation. Zero means no limit.",
	)
}

func execTestingGnoland(t *testing.T, logger log.Logger, gnoDataDir, gnoRootDir string, args []string) (*node.Node, error) {
	t.Helper()

	// Setup start config.
	icfg := &IntegrationConfig{}
	{
		fs := flag.NewFlagSet("start", flag.ExitOnError)
		icfg.RegisterFlags(fs)

		// Override default value for flags.
		fs.VisitAll(func(f *flag.Flag) {
			switch f.Name {
			case "root-dir":
				f.DefValue = gnoDataDir
			case "chainid":
				f.DefValue = "tendermint_test"
			case "genesis-balances-file":
				f.DefValue = filepath.Join(gnoRootDir, "gno.land", "genesis", "genesis_balances.txt")
			case "genesis-txs-file":
				f.DefValue = filepath.Join(gnoRootDir, "gno.land", "genesis", "genesis_txs.txt")
			default:
				return
			}

			f.Value.Set(f.DefValue)
		})

		if err := fs.Parse(args); err != nil {
			return nil, fmt.Errorf("unable to parse flags: %w", err)
		}
	}

	// Setup testing config.
	cfg := config.TestConfig().SetRootDir(gnoDataDir)
	{
		cfg.EnsureDirs()
		cfg.Consensus.CreateEmptyBlocks = true
		cfg.Consensus.CreateEmptyBlocksInterval = time.Duration(0)
		cfg.RPC.ListenAddress = "tcp://127.0.0.1:0"
		cfg.P2P.ListenAddress = "tcp://127.0.0.1:0"
	}

	// Prepare genesis.
	if err := setupTestingGenesis(gnoDataDir, cfg, icfg, gnoRootDir); err != nil {
		return nil, err
	}

	// Create application and node.
	return createAppAndNode(cfg, logger, gnoRootDir, icfg)
}

func setupTestingGenesis(gnoDataDir string, cfg *config.Config, icfg *IntegrationConfig, gnoRootDir string) error {
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	genesisFilePath := filepath.Join(gnoDataDir, cfg.Genesis)
	osm.EnsureDir(filepath.Dir(genesisFilePath), 0o700)
	if !osm.FileExists(genesisFilePath) {
		genesisTxs := loadGenesisTxs(icfg.GenesisTxsFile, icfg.ChainID, icfg.GenesisRemote)
		pvPub := priv.GetPubKey()

		gen := &bft.GenesisDoc{
			GenesisTime: time.Now(),
			ChainID:     icfg.ChainID,
			ConsensusParams: abci.ConsensusParams{
				Block: &abci.BlockParams{
					// TODO: update limits.
					MaxTxBytes:   1000000,  // 1MB,
					MaxDataBytes: 2000000,  // 2MB,
					MaxGas:       10000000, // 10M gas
					TimeIotaMS:   100,      // 100ms
				},
			},
			Validators: []bft.GenesisValidator{
				{
					Address: pvPub.Address(),
					PubKey:  pvPub,
					Power:   10,
					Name:    "testvalidator",
				},
			},
		}

		// Load distribution.
		balances := loadGenesisBalances(icfg.GenesisBalancesFile)

		// Load initial packages from examples.
		// XXX: We should be able to config this.
		test1 := crypto.MustAddressFromString(test1Addr)
		txs := []std.Tx{}

		// List initial packages to load from examples.
		// println(filepath.Join(gnoRootDir, "examples"))
		pkgs, err := gnomod.ListPkgs(filepath.Join(gnoRootDir, "examples"))
		if err != nil {
			return fmt.Errorf("listing gno packages: %w", err)
		}

		// Sort packages by dependencies.
		sortedPkgs, err := pkgs.Sort()
		if err != nil {
			return fmt.Errorf("sorting packages: %w", err)
		}

		// Filter out draft packages.
		nonDraftPkgs := sortedPkgs.GetNonDraftPkgs()

		for _, pkg := range nonDraftPkgs {
			// Open files in directory as MemPackage.
			memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)

			var tx std.Tx
			tx.Msgs = []std.Msg{
				vmm.MsgAddPackage{
					Creator: test1,
					Package: memPkg,
					Deposit: nil,
				},
			}

			// XXX: Add fee flag ?
			// Or maybe reduce fee to the minimum ?
			tx.Fee = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
			tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
			txs = append(txs, tx)
		}

		// Load genesis txs from file.
		txs = append(txs, genesisTxs...)

		// Construct genesis AppState.
		gen.AppState = gnoland.GnoGenesisState{
			Balances: balances,
			Txs:      txs,
		}

		writeGenesisFile(gen, genesisFilePath)
	}

	return nil
}

func createAppAndNode(cfg *config.Config, logger log.Logger, gnoRootDir string, icfg *IntegrationConfig) (*node.Node, error) {
	gnoApp, err := gnoland.NewCustomApp(gnoland.CustomAppConfig{
		Logger:                logger,
		GnoRootDir:            gnoRootDir,
		SkipFailingGenesisTxs: icfg.SkipFailingGenesisTxs,
		MaxCycles:             icfg.GenesisMaxVMCycles,
		DB:                    db.NewMemDB(),
	})
	if err != nil {
		return nil, fmt.Errorf("error in creating new app: %w", err)
	}

	cfg.LocalApp = gnoApp
	node, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("error in creating node: %w", err)
	}

	return node, node.Start()
}

func tsValidateError(ts *testscript.TestScript, cmd string, neg bool, err error) {
	if err != nil {
		ts.Logf("%s error: %v\n", cmd, err)
		if !neg {
			ts.Fatalf("unexpected %s command failure", cmd)
		}
	} else {
		if neg {
			ts.Fatalf("unexpected %s command success", cmd)
		}
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
			continue // Skip empty line.
		}

		// Patch the TX.
		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", chainID)
		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", genesisRemote)

		var tx std.Tx
		amino.MustUnmarshalJSON([]byte(txLine), &tx)
		txs = append(txs, tx)
	}

	return txs
}

func loadGenesisBalances(path string) []string {
	// Each balance is in the form: g1xxxxxxxxxxxxxxxx=100000ugnot.
	balances := []string{}
	content := osm.MustReadFile(path)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Remove comments.
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)

		// Skip empty lines.
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

func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
	err := gen.SaveAs(filePath)
	if err != nil {
		panic(err)
	}
}

// XXX: This helper will need to be relocated in the future.
func guessGnoRootDir() string {
	var rootdir string

	// first try to get the root directory from the GNOROOT environment variable.
	if rootdir = os.Getenv("GNOROOT"); rootdir != "" {
		return filepath.Clean(rootdir)
	}

	if gobin, err := exec.LookPath("go"); err == nil {
		// if GNOROOT is not set, try to guess the root directory using the `go list` command.
		cmd := exec.Command(gobin, "list", "-m", "-mod=mod", "-f", "{{.Dir}}", "github.com/gnolang/gno")
		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Errorf("invalid gno directory %q: %w", rootdir, err))
		}

		return strings.TrimSpace(string(out))
	}

	panic("no go binary available, unable to determine gno root-dir path")
}
