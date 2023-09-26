package main

import (
	"context"
	"flag"
	"fmt"
	"hash/crc32"
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
)

func TestTestdata(t *testing.T) {
	testscript.Run(t, setupGnolandTestScript(t, "testdata"))
}

func setupGnolandTestScript(t *testing.T, txtarDir string) testscript.Params {
	t.Helper()

	goModPath, err := exec.Command("go", "env", "GOMOD").CombinedOutput()
	require.NoError(t, err)

	gnoRootDir := filepath.Dir(string(goModPath))
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
						time.Sleep(time.Second)
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
	scfg := &startCfg{}
	{
		fs := flag.NewFlagSet("start", flag.ExitOnError)
		scfg.RegisterFlags(fs)

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
	if err := setupTestingGenesis(gnoDataDir, cfg, scfg, gnoRootDir); err != nil {
		return nil, err
	}

	// Create application and node
	return createAppAndNode(cfg, logger, gnoRootDir, scfg)
}

func getSessionID(ts *testscript.TestScript) string {
	works := ts.Getenv("WORK")
	sum := crc32.ChecksumIEEE([]byte(works))
	return strconv.FormatUint(uint64(sum), 16)
}

func setupTestingGenesis(gnoDataDir string, cfg *config.Config, scfg *startCfg, gnoRootDir string) error {
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	genesisFilePath := filepath.Join(gnoDataDir, cfg.Genesis)
	osm.EnsureDir(filepath.Dir(genesisFilePath), 0o700)
	if !osm.FileExists(genesisFilePath) {
		genesisTxs := loadGenesisTxs(scfg.genesisTxsFile, scfg.chainID, scfg.genesisRemote)
		pvPub := priv.GetPubKey()

		gen := &bft.GenesisDoc{
			GenesisTime: time.Now(),
			ChainID:     scfg.chainID,
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
		balances := loadGenesisBalances(scfg.genesisBalancesFile)

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

func createAppAndNode(cfg *config.Config, logger log.Logger, gnoRootDir string, scfg *startCfg) (*node.Node, error) {
	gnoApp, err := gnoland.NewCustomApp(gnoland.CustomAppConfig{
		Logger:                logger,
		GnoRootDir:            gnoRootDir,
		SkipFailingGenesisTxs: scfg.skipFailingGenesisTxs,
		MaxCycles:             scfg.genesisMaxVMCycles,
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
