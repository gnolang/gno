package integration

import (
	"bytes"
	"context"
	"errors"
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
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

const nodeMaxLifespan = time.Second * 120

var defaultUserBalance = std.Coins{std.NewCoin(ugnot.Denom, 10e8)}

type envKey int

const (
	envKeyGenesis envKey = iota
	envKeyLogger
	envKeyPkgsLoader
	envKeyPrivValKey
	envKeyExecCommand
	envKeyExecBin
	envKeyBase
	envKeyStdinBuffer
)

type commandkind int

const (
	// commandKindBin builds and uses an integration binary to run the testscript
	// in a separate process. This should be used for any external package that
	// wants to use test scripts.
	commandKindBin commandkind = iota
	// commandKindTesting uses the current testing binary to run the testscript
	// in a separate process. This command cannot be used outside this package.
	commandKindTesting
	// commandKindInMemory runs testscripts in memory.
	commandKindInMemory
)

type tNodeProcess struct {
	NodeProcess
	cfg         *gnoland.InMemoryNodeConfig
	nGnoKeyExec uint // Counter for execution of gnokey.
}

// NodesManager manages access to the nodes map with synchronization.
type NodesManager struct {
	nodes map[string]*tNodeProcess
	mu    sync.RWMutex

	sequentialMu sync.RWMutex
}

// NewNodesManager creates a new instance of NodesManager.
func NewNodesManager() *NodesManager {
	return &NodesManager{
		nodes: make(map[string]*tNodeProcess),
	}
}

func (nm *NodesManager) IsNodeRunning(sid string) bool {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	_, ok := nm.nodes[sid]
	return ok
}

// Get retrieves a node by its SID.
func (nm *NodesManager) Get(sid string) (*tNodeProcess, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	node, exists := nm.nodes[sid]
	return node, exists
}

// Set adds or updates a node in the map.
func (nm *NodesManager) Set(sid string, node *tNodeProcess) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.nodes[sid] = node
}

// Delete removes a node from the map.
func (nm *NodesManager) Delete(sid string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	delete(nm.nodes, sid)
}

func SetupGnolandTestscript(t *testing.T, p *testscript.Params) error {
	t.Helper()

	gnoRootDir := gnoenv.RootDir()

	nodesManager := NewNodesManager()

	defaultPK, err := GeneratePrivKeyFromMnemonic(DefaultAccount_Seed, "", 0, 0)
	require.NoError(t, err)

	var buildOnce sync.Once
	var gnolandBin string

	// Store the original setup scripts for potential wrapping
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		// If there's an original setup, execute it
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		cmd, isSet := env.Values[envKeyExecCommand].(commandkind)
		switch {
		case !isSet:
			cmd = commandKindBin // fallback on commandKindBin
			fallthrough
		case cmd == commandKindBin:
			buildOnce.Do(func() {
				t.Logf("building the gnoland integration node")
				start := time.Now()
				gnolandBin = buildGnoland(t, gnoRootDir)
				t.Logf("time to build the node: %v", time.Since(start).String())
			})

			env.Values[envKeyExecBin] = gnolandBin
		}

		// Store the resolved command kind so setupNode can read it later.
		env.Values[envKeyExecCommand] = cmd

		tmpdir, dbdir := t.TempDir(), t.TempDir()
		gnoHomeDir := filepath.Join(tmpdir, "gno")

		kb, err := keys.NewKeyBaseFromDir(gnoHomeDir)
		if err != nil {
			return err
		}

		kb.ImportPrivKey(DefaultAccount_Name, defaultPK, "")
		env.Setenv(DefaultAccount_Name+"_user_seed", DefaultAccount_Seed)
		env.Setenv(DefaultAccount_Name+"_user_addr", DefaultAccount_Address)

		// New private key
		env.Values[envKeyPrivValKey] = ed25519.GenPrivKey()

		// Set gno dbdir
		env.Setenv("GNO_DBDIR", dbdir)

		// Setup account store
		env.Values[envKeyBase] = kb

		// Generate node short id
		var sid string
		{
			works := env.Getenv("WORK")
			sum := crc32.ChecksumIEEE([]byte(works))
			sid = strconv.FormatUint(uint64(sum), 16)
			env.Setenv("SID", sid)
		}

		// Track new user balances added via the `adduser`
		// command and packages added with the `loadpkg` command.
		// This genesis will be use when node is started.

		genesis := gnoland.DefaultGenState()
		genesis.Balances = LoadDefaultGenesisBalanceFile(t, gnoRootDir)
		genesis.Auth.Params.InitialGasPrice = std.GasPrice{Gas: 0, Price: std.Coin{Amount: 0, Denom: "ugnot"}}
		genesis.Txs = []gnoland.TxWithMetadata{}
		LoadDefaultGenesisParamFile(t, gnoRootDir, &genesis)

		env.Values[envKeyGenesis] = &genesis
		env.Values[envKeyPkgsLoader] = NewPkgsLoader()
		env.Values[envKeyStdinBuffer] = new(strings.Builder)

		env.Setenv("GNOROOT", gnoRootDir)
		env.Setenv("GNOHOME", gnoHomeDir)

		env.Defer(func() {
			// Gracefully stop the node, if any
			n, exist := nodesManager.Get(sid)
			if !exist {
				return
			}

			if err := n.Stop(); err != nil {
				err = fmt.Errorf("unable to stop the node gracefully: %w", err)
				env.T().Fatal(err.Error())
			}
		})

		return nil
	}

	cmds := map[string]func(ts *testscript.TestScript, neg bool, args []string){
		"gnoland":     gnolandCmd(t, nodesManager, gnoRootDir),
		"gnokey":      gnokeyCmd(nodesManager),
		"adduser":     adduserCmd(nodesManager),
		"adduserfrom": adduserfromCmd(nodesManager),
		"patchpkg":    patchpkgCmd(),
		"loadpkg":     loadpkgCmd(gnoRootDir),
		"scanf":       loadpkgCmd(gnoRootDir),
		"input":       inputCmd(),
	}

	// Initialize cmds map if needed
	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}

	// Register gnoland command
	for cmd, call := range cmds {
		if _, exist := p.Cmds[cmd]; exist {
			panic(fmt.Errorf("unable register %q: command already exist", cmd))
		}

		p.Cmds[cmd] = call
	}

	return nil
}

func gnolandCmd(t *testing.T, nodesManager *NodesManager, gnoRootDir string) func(ts *testscript.TestScript, neg bool, args []string) {
	t.Helper()

	defaultPK, err := GeneratePrivKeyFromMnemonic(DefaultAccount_Seed, "", 0, 0)
	require.NoError(t, err)

	return func(ts *testscript.TestScript, neg bool, args []string) {
		sid := getNodeSID(ts)

		cmd, cmdargs := "", []string{}
		if len(args) > 0 {
			cmd, cmdargs = args[0], args[1:]
		}

		var err error
		switch cmd {
		case "":
			err = errors.New("no command provided")
		case "start":
			if nodesManager.IsNodeRunning(sid) {
				err = fmt.Errorf("node already started")
				break
			}

			// XXX: this is a bit hacky, we should consider moving
			// gnoland into his own package to be able to use it
			// directly or use the config command for this.
			fs := flag.NewFlagSet("start", flag.ContinueOnError)
			nonVal := fs.Bool("non-validator", false, "set up node as a non-validator")
			lockTransfer := fs.Bool("lock-transfer", false, "lock transfer ugnot")
			noParallel := fs.Bool("no-parallel", false, "don't run this node in parallel with other testing nodes")
			sysNamesEnabled := fs.Bool("sysnames-enabled", false, "enable namespace enforcement")
			if err := fs.Parse(cmdargs); err != nil {
				ts.Fatalf("unable to parse `gnoland start` flags: %s", err)
			}

			pkgs := ts.Value(envKeyPkgsLoader).(*PkgsLoader)
			defaultFee := std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))
			pkgsTxs, err := pkgs.GenerateTxs(defaultPK, defaultFee, nil)
			if err != nil {
				ts.Fatalf("unable to load packages txs: %s", err)
			}

			cfg := TestingMinimalNodeConfig(gnoRootDir)
			tsGenesis := ts.Value(envKeyGenesis).(*gnoland.GnoGenesisState)
			genesis := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
			genesis.Txs = append(genesis.Txs, append(pkgsTxs, tsGenesis.Txs...)...)
			genesis.Balances = append(genesis.Balances, tsGenesis.Balances...)
			if *lockTransfer {
				genesis.Bank.Params.RestrictedDenoms = []string{"ugnot"}
			}
			genesis.VM.Params = tsGenesis.VM.Params
			if *sysNamesEnabled {
				genesis.VM.Params.SysNamesEnabled = true
			}
			genesis.VM.RealmParams = append(genesis.VM.RealmParams, tsGenesis.VM.RealmParams...)

			cfg.Genesis.AppState = genesis
			if *nonVal {
				pv := bft.NewMockPV()
				pvPubKey := pv.PubKey()
				cfg.Genesis.Validators = []bft.GenesisValidator{
					{
						Address: pvPubKey.Address(),
						PubKey:  pvPubKey,
						Power:   10,
						Name:    "none",
					},
				}
			}

			if *noParallel {
				// The reason for this is that a direct Lock() on the RWMutex
				// can too easily create "splits", which are inefficient;
				// for instance: 10 parallel tests, one sequential test, 10 parallel tests.
				// Instead, TryLock() does not "request" the lock to be
				// transferred to the caller, so any incoming RLock() will be
				// given if there are other RLocks.
				// There is probably a better way to do this without using this hack;
				// however, this should be done if -no-parallel is actually
				// adopted in a variety of tests.
				for !nodesManager.sequentialMu.TryLock() {
					time.Sleep(time.Millisecond * 10)
				}
				ts.Defer(nodesManager.sequentialMu.Unlock)
			} else {
				nodesManager.sequentialMu.RLock()
				ts.Defer(nodesManager.sequentialMu.RUnlock)
			}

			ctx, cancel := context.WithTimeout(context.Background(), nodeMaxLifespan)
			ts.Defer(cancel)

			start := time.Now()

			dbdir := ts.Getenv("GNO_DBDIR")
			priv := ts.Value(envKeyPrivValKey).(ed25519.PrivKeyEd25519)
			nodep := setupNode(ts, ctx, &ProcessNodeConfig{
				ValidatorKey: priv,
				Verbose:      false,
				DBDir:        dbdir,
				RootDir:      gnoRootDir,
				TMConfig:     cfg.TMConfig,
				Genesis:      NewMarshalableGenesisDoc(cfg.Genesis),
			})

			nodesManager.Set(sid, &tNodeProcess{NodeProcess: nodep, cfg: cfg})
			ts.Setenv("RPC_ADDR", nodep.Address())

			// Load user infos
			loadUserEnv(ts, nodep.Address())

			fmt.Fprintf(ts.Stdout(), "node started successfully, took %s\n", time.Since(start).String())

		case "restart":
			node, exists := nodesManager.Get(sid)
			if !exists {
				err = fmt.Errorf("node must be started before being restarted")
				break
			}

			if err = node.Stop(); err != nil {
				err = fmt.Errorf("unable to stop the node gracefully: %w", err)
				break
			}

			ctx, cancel := context.WithTimeout(context.Background(), nodeMaxLifespan)
			ts.Defer(cancel)

			priv := ts.Value(envKeyPrivValKey).(ed25519.PrivKeyEd25519)
			dbdir := ts.Getenv("GNO_DBDIR")
			nodep := setupNode(ts, ctx, &ProcessNodeConfig{
				ValidatorKey: priv,
				DBDir:        dbdir,
				RootDir:      gnoRootDir,
				TMConfig:     node.cfg.TMConfig,
				Genesis:      NewMarshalableGenesisDoc(node.cfg.Genesis),
			})

			ts.Setenv("RPC_ADDR", nodep.Address())
			nodesManager.Set(sid, &tNodeProcess{NodeProcess: nodep, cfg: node.cfg})

			// Load user infos
			loadUserEnv(ts, nodep.Address())

			fmt.Fprintln(ts.Stdout(), "node restarted successfully")

		case "stop":
			node, exists := nodesManager.Get(sid)
			if !exists {
				err = fmt.Errorf("node not started cannot be stopped")
				break
			}

			if err = node.Stop(); err != nil {
				err = fmt.Errorf("unable to stop the node gracefully: %w", err)
				break
			}

			fmt.Fprintln(ts.Stdout(), "node stopped successfully")
			nodesManager.Delete(sid)

		default:
			err = fmt.Errorf("not supported command: %q", cmd)
			// XXX: support gnoland other commands
		}

		tsValidateError(ts, strings.TrimSpace("gnoland "+cmd), neg, err)
	}
}

func gnokeyCmd(nodes *NodesManager) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		gnoHomeDir := ts.Getenv("GNOHOME")

		sid := getNodeSID(ts)

		args, err := unquote(args)
		if err != nil {
			tsValidateError(ts, "gnokey", neg, err)
		}

		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(ts.Stdout()))
		io.SetErr(commands.WriteNopCloser(ts.Stderr()))
		cmd := keyscli.NewRootCmd(io, client.DefaultBaseOptions)

		// Use stdin buffer if available, otherwise default to newline
		if stdinBuf, ok := ts.Value(envKeyStdinBuffer).(*strings.Builder); ok && stdinBuf.Len() > 0 {
			io.SetIn(strings.NewReader(stdinBuf.String()))
			stdinBuf.Reset() // Clear buffer after use
		} else {
			io.SetIn(strings.NewReader("\n"))
		}
		defaultArgs := []string{
			"-home", gnoHomeDir,
			"-insecure-password-stdin=true",
		}

		if n, ok := nodes.Get(sid); ok {
			if raddr := n.Address(); raddr != "" {
				defaultArgs = append(defaultArgs, "-remote", raddr)
			}

			n.nGnoKeyExec++
		}

		args = append(defaultArgs, args...)

		defer func() {
			if r := recover(); r != nil {
				switch val := r.(type) {
				case error:
					err = val
				case string:
					err = fmt.Errorf("error: %s", val)
				default:
					err = fmt.Errorf("unknown error: %#v", val)
				}

				tsValidateError(ts, "gnokey", neg, err)
			}
		}()

		err = cmd.ParseAndRun(context.Background(), args)
		tsValidateError(ts, "gnokey", neg, err)
	}
}

func adduserCmd(nodesManager *NodesManager) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		gnoHomeDir := ts.Getenv("GNOHOME")

		sid := getNodeSID(ts)
		if nodesManager.IsNodeRunning(sid) {
			tsValidateError(ts, "adduser", neg, errors.New("adduser must be used before starting node"))
			return
		}

		if len(args) == 0 {
			ts.Fatalf("new user name required")
		}

		kb, err := keys.NewKeyBaseFromDir(gnoHomeDir)
		if err != nil {
			ts.Fatalf("unable to get keybase")
		}

		coins := defaultUserBalance
		if len(args) > 1 {
			// parse coins from string
			coins, err = std.ParseCoins(args[1])
			if err != nil {
				ts.Fatalf("unable to parse coins: %s", err)
			}
		}

		balance, err := createAccount(ts, kb, args[0], coins)
		if err != nil {
			ts.Fatalf("error creating account %s: %s", args[0], err)
		}

		genesis := ts.Value(envKeyGenesis).(*gnoland.GnoGenesisState)
		genesis.Balances = append(genesis.Balances, balance)
	}
}

func adduserfromCmd(nodesManager *NodesManager) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		gnoHomeDir := ts.Getenv("GNOHOME")

		sid := getNodeSID(ts)
		if nodesManager.IsNodeRunning(sid) {
			tsValidateError(ts, "adduserfrom", neg, errors.New("adduserfrom must be used before starting node"))
			return
		}

		var account, index uint64
		var err error

		switch len(args) {
		case 2:
		case 4:
			index, err = strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				ts.Fatalf("invalid index number %s", args[3])
			}
			fallthrough
		case 3:
			account, err = strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				ts.Fatalf("invalid account number %s", args[2])
			}
		default:
			ts.Fatalf("to create account from metadatas, user name and mnemonic are required ( account and index are optional )")
		}

		kb, err := keys.NewKeyBaseFromDir(gnoHomeDir)
		if err != nil {
			ts.Fatalf("unable to get keybase")
		}

		balance, err := createAccountFrom(ts, kb, args[0], args[1], defaultUserBalance, uint32(account), uint32(index))
		if err != nil {
			ts.Fatalf("error creating wallet %s", err)
		}

		genesis := ts.Value(envKeyGenesis).(*gnoland.GnoGenesisState)
		genesis.Balances = append(genesis.Balances, balance)

		fmt.Fprintf(ts.Stdout(), "Added %s(%s) to genesis", args[0], balance.Address)
	}
}

func patchpkgCmd() func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		args, err := unquote(args)
		if err != nil {
			tsValidateError(ts, "patchpkg", neg, err)
		}

		if len(args) != 2 {
			ts.Fatalf("`patchpkg`: should have exactly 2 arguments")
		}

		pkgs := ts.Value(envKeyPkgsLoader).(*PkgsLoader)
		replace, with := args[0], args[1]
		pkgs.SetPatch(replace, with)
	}
}

func loadpkgCmd(gnoRootDir string) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		workDir := ts.Getenv("WORK")
		examplesDir := filepath.Join(gnoRootDir, "examples")

		pkgs := ts.Value(envKeyPkgsLoader).(*PkgsLoader)

		var dir, path string
		switch len(args) {
		case 2:
			path = args[0]
			dir = filepath.Clean(args[1])
		case 1:
			dir = filepath.Clean(args[0])
		case 0:
			ts.Fatalf("`loadpkg`: no arguments specified")
		default:
			ts.Fatalf("`loadpkg`: too many arguments specified")
		}

		if dir == "all" {
			ts.Logf("warning: loading all packages")
			if err := pkgs.LoadAllPackagesFromDir(examplesDir); err != nil {
				ts.Fatalf("unable to load packages from %q: %s", examplesDir, err)
			}

			return
		}

		if !strings.HasPrefix(dir, workDir) {
			dir = filepath.Join(examplesDir, dir)
		}

		if err := pkgs.LoadPackage(examplesDir, dir, path); err != nil {
			ts.Fatalf("`loadpkg` unable to load package(s) from %q: %s", args[0], err)
		}

		ts.Logf("%q package was added to genesis", args[0])
	}
}

func loadUserEnv(ts *testscript.TestScript, remote string) error {
	const path = "auth/accounts"

	// List all accounts
	kb := ts.Value(envKeyBase).(keys.Keybase)
	accounts, err := kb.List()
	if err != nil {
		ts.Fatalf("query accounts: unable to list keys: %s", err)
	}

	cli, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return fmt.Errorf("unable create rpc client %q: %w", remote, err)
	}

	batch := cli.NewBatch()
	for _, account := range accounts {
		accountPath := filepath.Join(path, account.GetAddress().String())
		if err := batch.ABCIQuery(accountPath, []byte{}); err != nil {
			return fmt.Errorf("unable to create query request: %w", err)
		}
	}

	batchRes, err := batch.Send(context.Background())
	if err != nil {
		return fmt.Errorf("unable to query accounts: %w", err)
	}

	if len(batchRes) != len(accounts) {
		ts.Fatalf("query accounts: len(res) != len(accounts)")
	}

	for i, res := range batchRes {
		account := accounts[i]
		name := account.GetName()
		qres := res.(*ctypes.ResultABCIQuery)

		if err := qres.Response.Error; err != nil {
			ts.Fatalf("query account %q error: %s", account.GetName(), err.Error())
		}

		var qret struct{ BaseAccount std.BaseAccount }
		if err = amino.UnmarshalJSON(qres.Response.Data, &qret); err != nil {
			ts.Fatalf("query account %q unarmshal error: %s", account.GetName(), err.Error())
		}

		strAccountNumber := strconv.Itoa(int(qret.BaseAccount.GetAccountNumber()))
		ts.Setenv(name+"_account_num", strAccountNumber)
		ts.Logf("[%q] account number: %s", name, strAccountNumber)

		strAccountSequence := strconv.Itoa(int(qret.BaseAccount.GetSequence()))
		ts.Setenv(name+"_account_seq", strAccountSequence)
		ts.Logf("[%q] account sequence: %s", name, strAccountNumber)
	}

	return nil
}

type tsLogWriter struct {
	ts *testscript.TestScript
}

func (l *tsLogWriter) Write(p []byte) (n int, err error) {
	l.ts.Logf(string(p))
	return len(p), nil
}

func setupNode(ts *testscript.TestScript, ctx context.Context, cfg *ProcessNodeConfig) NodeProcess {
	pcfg := ProcessConfig{
		Node:   cfg,
		Stdout: &tsLogWriter{ts},
		Stderr: ts.Stderr(),
	}

	// Setup coverdir provided
	if coverdir := ts.Getenv("GOCOVERDIR"); coverdir != "" {
		pcfg.CoverDir = coverdir
	}

	val := ts.Value(envKeyExecCommand)

	switch cmd := val.(commandkind); cmd {
	case commandKindInMemory:
		nodep, err := RunInMemoryProcess(ctx, pcfg)
		if err != nil {
			ts.Fatalf("unable to start in memory node: %s", err)
		}

		return nodep

	case commandKindTesting:
		if !testing.Testing() {
			ts.Fatalf("unable to invoke testing process while not testing")
		}

		return runTestingNodeProcess(&testingTS{ts}, ctx, pcfg)

	case commandKindBin:
		bin := ts.Value(envKeyExecBin).(string)
		nodep, err := RunNodeProcess(ctx, pcfg, bin)
		if err != nil {
			ts.Fatalf("unable to start process node: %s", err)
		}

		return nodep

	default:
		ts.Fatalf("unknown command kind: %+v", cmd)
	}

	return nil
}

// createAccount creates a new account with the given name and adds it to the keybase.
func createAccount(ts *testscript.TestScript, kb keys.Keybase, accountName string, coins std.Coins) (gnoland.Balance, error) {
	var balance gnoland.Balance
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return balance, fmt.Errorf("error creating entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return balance, fmt.Errorf("error generating mnemonic: %w", err)
	}

	return createAccountFrom(ts, kb, accountName, mnemonic, coins, 0, 0)
}

// createAccountFrom creates a new account with the given metadata and adds it to the keybase.
func createAccountFrom(ts *testscript.TestScript, kb keys.Keybase, accountName, mnemonic string, coins std.Coins, account, index uint32) (gnoland.Balance, error) {
	var balance gnoland.Balance

	// check if mnemonic is valid
	if !bip39.IsMnemonicValid(mnemonic) {
		return balance, fmt.Errorf("invalid mnemonic")
	}

	keyInfo, err := kb.CreateAccount(accountName, mnemonic, "", "", account, index)
	if err != nil {
		return balance, fmt.Errorf("unable to create account: %w", err)
	}

	address := keyInfo.GetAddress()
	ts.Setenv(accountName+"_user_seed", mnemonic)
	ts.Setenv(accountName+"_user_addr", address.String())

	return gnoland.Balance{
		Address: address,
		Amount:  coins,
	}, nil
}

func buildGnoland(t *testing.T, rootdir string) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "gnoland-test")

	t.Log("building gnoland integration binary...")

	// Build a fresh gno binary in a temp directory
	gnoArgsBuilder := []string{"build", "-o", bin}

	os.Executable()

	// Forward `-covermode` settings if set
	if coverMode := testing.CoverMode(); coverMode != "" {
		gnoArgsBuilder = append(gnoArgsBuilder,
			"-covermode", coverMode,
		)
	}

	// Append the path to the gno command source
	gnoArgsBuilder = append(gnoArgsBuilder, filepath.Join(rootdir,
		"gno.land", "pkg", "integration", "process"))

	t.Logf("build command: %s", strings.Join(gnoArgsBuilder, " "))

	cmd := exec.Command("go", gnoArgsBuilder...)

	var buff bytes.Buffer
	cmd.Stderr, cmd.Stdout = &buff, &buff
	defer buff.Reset()

	if err := cmd.Run(); err != nil {
		require.FailNowf(t, "unable to build binary", "%q\n%s",
			err.Error(), buff.String())
	}

	return bin
}

// GeneratePrivKeyFromMnemonic generates a crypto.PrivKey from a mnemonic.
func GeneratePrivKeyFromMnemonic(mnemonic, bip39Passphrase string, account, index uint32) (crypto.PrivKey, error) {
	// Generate Seed from Mnemonic
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	// Derive Private Key
	coinType := crypto.CoinType // ensure this is set correctly in your context
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, hdPath.String())
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// Convert to secp256k1 private key
	privKey := secp256k1.PrivKeySecp256k1(derivedPriv)
	return privKey, nil
}

func getNodeSID(ts *testscript.TestScript) string {
	return ts.Getenv("SID")
}

func inputCmd() func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if neg {
			ts.Fatalf("input command does not support negation")
		}

		if len(args) == 0 {
			ts.Fatalf("input requires at least one argument")
		}

		// Get or create stdin buffer
		stdinBuf, ok := ts.Value(envKeyStdinBuffer).(*strings.Builder)
		if !ok {
			ts.Fatalf("stdin buffer not initialized")
		}

		// Join all arguments with spaces and add newline
		content := strings.Join(args, " ") + "\n"
		stdinBuf.WriteString(content)
	}
}

func tsValidateError(ts *testscript.TestScript, cmd string, neg bool, err error) {
	if err != nil {
		fmt.Fprintf(ts.Stderr(), "%q error: %+v\n", cmd, err)
		if !neg {
			ts.Fatalf("unexpected %q command failure: %s", cmd, err)
		}
	} else {
		if neg {
			ts.Fatalf("unexpected %q command success", cmd)
		}
	}
}
