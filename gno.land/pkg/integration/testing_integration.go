// ------------------------------------------------------------------------------------------------
// WARNING: TEMPORARY CODE
//
// This file is a rapid prototype to meet immediate project needs and is not intended for long-term
// use in its current form. It requires review, possible refactoring, and thorough testing.
// Use at your own risk.
// ------------------------------------------------------------------------------------------------

package integration

import (
	"context"
	"flag"
	"fmt"
	"hash/crc32"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/require"
	"github.com/rogpeppe/go-internal/testscript"
)

// XXX: should be centralize somewhere
const (
	test1Addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	test1Seed = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
	test2Addr = "g1m5exxkaqrsxd8ne93psljuakxzhkzcm42yg7ye"
	test2Seed = "bid kangaroo tomorrow raccoon habit fine circle battle question push bounce dust bonus town remember diamond hill busy frozen project movie giant file ceiling"
	test3Addr = "g13hlh3a3kygwq9g3vgjzz5zu4fy7gpkk523ex6l"
	test3Seed = "eyebrow vote bind response vanish sad spoon few bargain quote stone recycle rail bulb force syrup menu zero disagree bread gift clump artist rebel"
)

func TestTestdata(t *testing.T) {
	testscript.Run(t, SetupGnolandTestScript(t, "testdata"))
}

type IntegrationConfig struct {
	skipFailingGenesisTxs bool
	skipStart             bool
	genesisBalancesFile   string
	genesisTxsFile        string
	chainID               string
	genesisRemote         string
	rootDir               string
	genesisMaxVMCycles    int64
	config                string
}

func (c *IntegrationConfig) RegisterFlags(fs *flag.FlagSet) {
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
		"config file (optional)",
	)
}

func SetupGnolandTestScript(t *testing.T, txtarDir string) testscript.Params {
	t.Helper()

	cmd := exec.Command("go", "list", "-m", "-mod=mod", "-f", "{{.Dir}}", "github.com/gnolang/gno")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	gnoRootDir := strings.TrimSpace(string(out))
	gnoHomeDir := filepath.Join(t.TempDir(), "gno")
	gnoDataDir := filepath.Join(t.TempDir(), "data")

	var muNodes sync.Mutex
	nodes := map[string]*node.Node{}
	t.Cleanup(func() {
		for id, n := range nodes {
			if err := n.Stop(); err != nil {
				panic(fmt.Errorf("node %q was unable to stop: %w", id, err))
			}
		}
	})

	return testscript.Params{
		Dir: txtarDir,
		Setup: func(env *testscript.Env) error {
			kb, err := keys.NewKeyBaseFromDir(gnoHomeDir)
			if err != nil {
				return err
			}

			kb.CreateAccount("test1", test1Seed, "", "", 0, 0)
			env.Setenv("USER_SEED_test1", test1Seed)
			env.Setenv("USER_ADDR_test1", test1Addr)
			kb.CreateAccount("test2", test2Seed, "", "", 0, 0)
			env.Setenv("USER_SEED_test2", test2Seed)
			env.Setenv("USER_ADDR_test2", test2Addr)
			kb.CreateAccount("test3", test3Seed, "", "", 0, 0)
			env.Setenv("USER_SEED_test3", test3Seed)
			env.Setenv("USER_ADDR_test3", test3Addr)

			env.Setenv("GNOROOT", gnoRootDir)
			env.Setenv("GNOHOME", gnoHomeDir)

			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"gnoland": func(ts *testscript.TestScript, neg bool, args []string) {
				muNodes.Lock()
				defer muNodes.Unlock()

				if len(args) == 0 {
					tsValidateError(ts, "gnoland", neg, fmt.Errorf("use gnoland [start|stop] command"))
					return
				}

				sid := getSessionID(ts)

				var cmd string
				cmd, args = args[0], args[1:]

				var err error
				switch cmd {
				case "start":
					if _, ok := nodes[sid]; ok {
						err = fmt.Errorf("node %q already started", sid)
						break
					}

					dataDir := filepath.Join(gnoDataDir, sid)
					var node *node.Node
					if node, err = execTestingGnoland(t, dataDir, gnoRootDir, args); err == nil {
						// XXX need mutex ?
						nodes[sid] = node

						// get listen addr environement
						// should have been updated with the right port on start
						laddr := node.Config().RPC.ListenAddress

						// add default environement
						ts.Setenv("RPC_ADDR", laddr)
						ts.Setenv("GNODATA", gnoDataDir)

						// XXX: Use something similar to `require.Eventually` to check for node
						// availability. For now, if this sleep duration is too short, the
						// subsequent command might fail with an [internal error].
						time.Sleep(time.Second * 2)
					}
				case "stop":
					n, ok := nodes[sid]
					if !ok {
						err = fmt.Errorf("node %q not started cannot be stop", sid)
						break
					}

					if err = n.Stop(); err != nil {
						delete(nodes, sid)

						// unset env dirs
						ts.Setenv("RPC_ADDR", "")
						ts.Setenv("GNODATA", "")
					}
				}

				tsValidateError(ts, "gnoland "+cmd, neg, err)
			},
			"gnokey": func(ts *testscript.TestScript, neg bool, args []string) {
				muNodes.Lock()
				defer muNodes.Unlock()

				sid := getSessionID(ts)

				// Setup io command
				io := commands.NewTestIO()
				io.SetOut(commands.WriteNopCloser(ts.Stdout()))
				io.SetErr(commands.WriteNopCloser(ts.Stderr()))
				cmd := client.NewRootCmd(io)

				io.SetIn(strings.NewReader("\n")) // inject empty password to stdin
				defaultArgs := []string{
					"-home", gnoHomeDir,
					"-insecure-password-stdin=true", // there no use to not have this param by default
					/* ideally, we'd like to have this for 'gnokey maketx call ...'.
					"-chainid=tendermint_test",
					"-gas-fee=1ugnot",
					"-gas-wanted=10000000",
					"-broadcast=true",
					*/
				}

				if n, ok := nodes[sid]; ok {
					if raddr := n.Config().RPC.ListenAddress; raddr != "" {
						defaultArgs = append(defaultArgs, "-remote", raddr)
					}
				}

				// inject default argument, if duplicate
				// arguments, it should be override by the ones
				// user provided
				args = append(defaultArgs, args...)

				err := cmd.ParseAndRun(context.Background(), args)
				tsValidateError(ts, "gnokey", neg, err)
			},
		},
	}
}

func execTestingGnoland(t *testing.T, gnoDataDir, gnoRootDir string, args []string) (*node.Node, error) {
	t.Helper()

	// Setup logger
	logger := log.NewNopLogger()

	// Setup start config
	icfg := &IntegrationConfig{}
	{
		fs := flag.NewFlagSet("start", flag.ExitOnError)
		icfg.RegisterFlags(fs)

		// Override default value for flags
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

	// Setup testing config
	cfg := config.TestConfig().SetRootDir(gnoDataDir)
	{
		cfg.EnsureDirs()
		cfg.Consensus.CreateEmptyBlocks = true
		cfg.Consensus.CreateEmptyBlocksInterval = time.Duration(0)
		cfg.RPC.ListenAddress = "tcp://127.0.0.1:0"
		cfg.P2P.ListenAddress = "tcp://127.0.0.1:0"
	}

	// Prepare genesis
	if err := setupTestingGenesis(gnoDataDir, cfg, icfg, gnoRootDir); err != nil {
		return nil, err
	}

	// Create application and node
	return createAppAndNode(cfg, logger, gnoRootDir, icfg)
}

func getSessionID(ts *testscript.TestScript) string {
	works := ts.Getenv("WORK")
	sum := crc32.ChecksumIEEE([]byte(works))
	return strconv.FormatUint(uint64(sum), 16)
}

func setupTestingGenesis(gnoDataDir string, cfg *config.Config, icfg *IntegrationConfig, gnoRootDir string) error {
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	genesisFilePath := filepath.Join(gnoDataDir, cfg.Genesis)
	osm.EnsureDir(filepath.Dir(genesisFilePath), 0o700)
	if !osm.FileExists(genesisFilePath) {
		genesisTxs := loadGenesisTxs(icfg.genesisTxsFile, icfg.chainID, icfg.genesisRemote)
		pvPub := priv.GetPubKey()

		gen := &bft.GenesisDoc{
			GenesisTime: time.Now(),
			ChainID:     icfg.chainID,
			ConsensusParams: abci.ConsensusParams{
				Block: &abci.BlockParams{
					// TODO: update limits.
					MaxTxBytes:   10_000_000,  // 10MB,
					MaxDataBytes: 20_000_000,  // 20MB,
					MaxGas:       100_000_000, // 100M gas
					TimeIotaMS:   100,         // 100ms
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
		balances := loadGenesisBalances(icfg.genesisBalancesFile)

		// Load initial packages from examples.
		// XXX: we should be able to config this
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

			// XXX: add fee flag ?
			// or maybe reduce fee to the minimum ?
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

		writeGenesisFile(gen, genesisFilePath)
	}

	return nil
}

func createAppAndNode(cfg *config.Config, logger log.Logger, gnoRootDir string, icfg *IntegrationConfig) (*node.Node, error) {
	gnoApp, err := gnoland.NewCustomApp(gnoland.CustomAppConfig{
		Logger:                logger,
		GnoRootDir:            gnoRootDir,
		SkipFailingGenesisTxs: icfg.skipFailingGenesisTxs,
		MaxCycles:             icfg.genesisMaxVMCycles,
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

func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
	err := gen.SaveAs(filePath)
	if err != nil {
		panic(err)
	}
}

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
