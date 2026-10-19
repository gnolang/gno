package integration

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
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
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
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
	envKeyBase
	envKeyStdinBuffer
	envKeyNodeSlot
)

// nodeSlot records whether a script already holds a node budget token.
// Only the script's own goroutine touches it.
type nodeSlot struct{ held bool }

// Constants governing how many nodes run at once; see [nodeBudget]. The memory
// figures come from running the testdata suite at fixed node counts on a
// 16-core host: peak RSS goes 3.9 GiB at two nodes, 7.2 at six, 12.0 at
// sixteen, while wall time goes 145s, 54s, 43s.
const (
	// nodeMinParallel is the floor the allowance never shrinks below, so that a
	// machine under memory pressure still makes progress. One node stretches
	// the suite to ~240s against ~54s at six, and a machine that cannot hold
	// two nodes cannot build this repo either.
	nodeMinParallel = 2

	// nodeMemCost is the peak RSS one more concurrent node adds: ~580 MiB
	// measured over the range above, rounded up.
	nodeMemCost = 640 << 20

	// nodeMemReserveDiv makes the suite leave a quarter of the memory it may
	// use — the machine's, or its cgroup's — to everything else. A fraction
	// rather than a constant so the same rule holds for an 8 GiB laptop and a
	// 128 GiB builder. It needs to be this generous because the ramp stops only
	// once a further node would not fit, and the nodes already admitted are
	// still growing into theirs: the reserve is what absorbs that overshoot.
	nodeMemReserveDiv = 4

	// nodeStartSettle paces changes of allowance, so that the memory claimed by
	// the node the last change let in is reflected in the reading behind the
	// next one — rather than a burst of scripts all deciding against the same
	// stale figure. Node startup measures p50 320ms, p90 1.2s.
	nodeStartSettle = time.Second

	// nodeBudgetPoll is how often a blocked script re-checks. The conditions
	// are time- and memory-based, so there is nothing to signal on.
	nodeBudgetPoll = 200 * time.Millisecond
)

// traceNodeBudget logs every change of allowance to stderr. The budget tunes
// itself, so this is the way to see what it decided and why on a given machine.
var traceNodeBudget = os.Getenv("GNO_TEST_TRACE_BUDGET") != ""

// nodeBudget decides how many nodes may run at once. testscript runs every
// txtar as a parallel subtest, so left alone the node count is whatever
// `go test -parallel` allows — GOMAXPROCS by default. Each node keeps its own
// store, stdlib byte cache and genesis write batch alive, which makes peak
// memory a function of the host's core count: ~6 GiB over four nodes, ~12 GiB
// over sixteen, enough to take a workstation down while CI's four-core runners
// never notice.
//
// Rather than pin a single number, start at the count CI validates and move
// from there: up towards max while the machine reports room to spare, down
// towards [nodeMinParallel] when it stops doing so. That spends the memory
// actually going free, on the machine at hand, without anyone having to tune a
// flag — and gives it back when something else needs it.
type nodeBudget struct {
	mu      sync.Mutex
	running int

	// limit is the current allowance, between min and max. It is a value that
	// ratchets rather than a decision re-made per node: scripts finish
	// constantly, so a rule phrased against the live count would keep falling
	// back to the floor and never accumulate a ramp.
	limit     int
	lastCheck time.Time

	min, max int

	// reserve is the memory left to the rest of the system. Zero means this
	// platform cannot report free memory, in which case min == max and the
	// budget never ramps.
	reserve uint64

	// readMem is [testutils.ReadMemInfo], swapped out by tests.
	readMem func() (testutils.MemInfo, bool)
}

func newNodeBudget() *nodeBudget {
	// An explicit override pins the count, as does a platform whose free
	// memory we cannot read.
	if n, ok := testutils.MaxParallelOverride(); ok {
		return &nodeBudget{min: n, max: n, limit: n, readMem: testutils.ReadMemInfo}
	}
	mi, ok := testutils.ReadMemInfo()
	if !ok {
		n := testutils.MaxParallel()
		return &nodeBudget{min: n, max: n, limit: n, readMem: testutils.ReadMemInfo}
	}
	return &nodeBudget{
		min: nodeMinParallel,
		max: max(runtime.GOMAXPROCS(0), nodeMinParallel),
		// Start where the static cap would have: the count CI validates, and
		// therefore the one worth assuming before any reading has been taken.
		// Starting at the floor instead would leave any platform with a
		// pessimistic reading — darwin, counting only wholly free pages —
		// permanently slower than the fixed cap it replaced.
		limit:   max(testutils.MaxParallel(), nodeMinParallel),
		reserve: mi.Total / nodeMemReserveDiv,
		readMem: testutils.ReadMemInfo,
	}
}

// acquire blocks until there is room for one more node.
func (b *nodeBudget) acquire() {
	for !b.tryAcquire() {
		time.Sleep(nodeBudgetPoll)
	}
}

func (b *nodeBudget) tryAcquire() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.resize()
	if b.running >= b.limit {
		return false
	}
	b.running++
	return true
}

// resize moves the allowance one step towards what the machine can currently
// take. Called with b.mu held.
//
// Growing and shrinking use different thresholds so the limit settles instead
// of oscillating: room for a whole further node to grow, dipping into the
// reserve to shrink, and nothing in between. Shrinking matters because the
// memory being measured is shared — a second suite, or a browser, starting
// midway through should cost this one its ramp rather than the machine.
func (b *nodeBudget) resize() {
	if b.reserve == 0 || time.Since(b.lastCheck) < nodeStartSettle {
		return
	}
	mi, ok := b.readMem()
	if !ok {
		return
	}
	// Pace the reading itself and not merely changes to the limit: blocked
	// scripts poll five times a second, and on darwin every reading costs a
	// vm_stat subprocess. Sitting inside the hysteresis band must be cheap.
	b.lastCheck = time.Now()
	switch {
	case mi.Available < b.reserve && b.limit > b.min:
		b.limit--
	case mi.Available >= b.reserve+nodeMemCost && b.limit < b.max && b.running >= b.limit:
		// Only when the current allowance is actually saturated. Raising it
		// while slots sit free buys nothing and overshoots, the nodes already
		// admitted still being on their way up to full size.
		b.limit++
	default:
		return
	}
	if traceNodeBudget {
		fmt.Fprintf(os.Stderr, "nodeBudget: limit=%d running=%d max=%d avail=%.2fGiB reserve=%.2fGiB\n",
			b.limit, b.running, b.max, float64(mi.Available)/(1<<30), float64(b.reserve)/(1<<30))
	}
}

func (b *nodeBudget) release() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running--
}

type commandkind int

const (
	// commandKindTesting uses the current testing binary to run the testscript
	// in a separate process. This command cannot be used outside this package.
	commandKindTesting commandkind = iota
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

	// budget bounds how many nodes exist at once. Nodes are the expensive part
	// of a script, so cap those rather than overriding the user's -parallel.
	budget *nodeBudget
}

// NewNodesManager creates a new instance of NodesManager.
func NewNodesManager() *NodesManager {
	return &NodesManager{
		nodes:  make(map[string]*tNodeProcess),
		budget: newNodeBudget(),
	}
}

// acquireSlot reserves the right to run one node, blocking until the suite is
// below its node budget. The token is held for the rest of the script rather
// than released on `gnoland stop`, so that a script which stops and starts
// again never waits on a second token — every holder waiting for one more
// would deadlock once the budget is full.
func (nm *NodesManager) acquireSlot(ts *testscript.TestScript) {
	slot := ts.Value(envKeyNodeSlot).(*nodeSlot)
	if slot.held {
		return
	}
	nm.budget.acquire()
	slot.held = true
	ts.Defer(nm.budget.release)
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

	// Store the original setup scripts for potential wrapping
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		// If there's an original setup, execute it
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		// Default to running nodes in-memory when the caller didn't pick a
		// command kind. setupNode reads this later.
		if _, isSet := env.Values[envKeyExecCommand].(commandkind); !isSet {
			env.Values[envKeyExecCommand] = commandKindInMemory
		}

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
		env.Values[envKeyNodeSlot] = new(nodeSlot)

		env.Setenv("GNOROOT", gnoRootDir)
		env.Setenv("GNOHOME", gnoHomeDir)

		env.Defer(func() {
			// Gracefully stop the node, if any
			n, exist := nodesManager.Get(sid)
			if !exist {
				return
			}

			// Drop the node from the manager so it (and its in-memory store,
			// which retains the per-node stdlib cache) becomes collectable once
			// the script ends. Without this the manager — which lives for the
			// whole TestTestdata run — pins every node, leaking ~50 MB/script.
			nodesManager.Delete(sid)

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
		"genesiscall": genesiscallCmd(defaultPK),
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
			maxGas := fs.Int64("max-gas", 0, "override block max gas (0 = use default)")
			if err := fs.Parse(cmdargs); err != nil {
				ts.Fatalf("unable to parse `gnoland start` flags: %s", err)
			}

			// Before the genesis txs are built, which are themselves held in
			// memory until the node has them. Always taken before
			// sequentialMu, never the other way around.
			nodesManager.acquireSlot(ts)

			pkgs := ts.Value(envKeyPkgsLoader).(*PkgsLoader)
			defaultFee := std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))
			pkgsTxs, err := pkgs.GenerateTxs(defaultPK, defaultFee, nil)
			if err != nil {
				ts.Fatalf("unable to load packages txs: %s", err)
			}

			cfg := TestingMinimalNodeConfig(gnoRootDir)
			if *maxGas > 0 {
				cfg.Genesis.ConsensusParams.Block.MaxGas = *maxGas
			}
			tsGenesis := ts.Value(envKeyGenesis).(*gnoland.GnoGenesisState)
			genesis := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
			genesis.Txs = append(genesis.Txs, append(pkgsTxs, tsGenesis.Txs...)...)
			genesis.Balances = append(genesis.Balances, tsGenesis.Balances...)
			// run_submitters is deliberately NOT merged, and not seeded
			// anywhere in this harness: an empty list means the MsgRun gate is
			// off, which is what every txtar wants except the ones testing the
			// gate itself. Those populate it in-script, so the default here
			// must stay empty or they would be testing a pre-seeded list.
			if *lockTransfer {
				genesis.Bank.Params.RestrictedDenoms = []string{"ugnot"}
			}
			genesis.VM.RealmParams = append(genesis.VM.RealmParams, tsGenesis.VM.RealmParams...)
			// Carry the scalar vm params the genesis params file can set.
			//
			// Those two are the only ones LoadGenesisParamsFile writes into
			// VM.Params today, and it errors on any other key, so the input
			// side is self-policing. The merge here is not: it copies named
			// fields, so a value set in the file but not listed here is
			// silently replaced by the default. That was true of both of these
			// until now, and harmless only because the file happens to set what
			// the defaults already are -- so a test would have passed while a
			// real chain used the file's value. TestGenesisParamsReachTheHarness
			// fails if a third field is added to the loader without being
			// carried here.
			genesis.VM.Params.ChainDomain = tsGenesis.VM.Params.ChainDomain
			genesis.VM.Params.SysNamesPkgPath = tsGenesis.VM.Params.SysNamesPkgPath

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

		case "wait-for-new-block":
			node, exists := nodesManager.Get(sid)
			if !exists {
				err = fmt.Errorf("node not started, cannot wait for new block")
				break
			}
			err = waitForNewBlock(ts, node.Address(), defaultPK)

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
			dir = ResolveExamplePath(examplesDir, dir)
		}

		if err := pkgs.LoadPackage(examplesDir, dir, path); err != nil {
			ts.Fatalf("`loadpkg` unable to load package(s) from %q: %s", args[0], err)
		}

		ts.Logf("%q package was added to genesis", args[0])
	}
}

func genesiscallCmd(defaultPK crypto.PrivKey) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if len(args) < 2 {
			ts.Fatalf("`genesiscall` requires at least 2 arguments: <pkgpath> <func> [args...]")
		}

		pkgPath := args[0]
		funcName := args[1]
		var callArgs []string
		if len(args) > 2 {
			callArgs = args[2:]
		}

		txs := []gnoland.TxWithMetadata{{
			Tx: std.Tx{
				Msgs: []std.Msg{vm.MsgCall{
					Caller:  defaultPK.PubKey().Address(),
					PkgPath: pkgPath,
					Func:    funcName,
					Args:    callArgs,
				}},
				Fee: std.NewFee(2_000_000, std.NewCoin(ugnot.Denom, 1_000_000)),
			},
		}}

		if err := gnoland.SignGenesisTxs(txs, defaultPK, "tendermint_test"); err != nil {
			ts.Fatalf("`genesiscall` unable to sign tx: %s", err)
		}

		genesis := ts.Value(envKeyGenesis).(*gnoland.GnoGenesisState)
		genesis.Txs = append(genesis.Txs, txs...)

		ts.Logf("genesis call %s.%s added", pkgPath, funcName)
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

		var qret gnoland.GnoAccount
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

// waitForNewBlock submits a 1ugnot self-transfer from the default account
// and returns after the containing block is committed. BroadcastTxCommit
// returns the height of the block that included the tx — strictly greater
// than the height at submission, since CheckTx happens after submission.
// Used by txtar tests that need to burn a deterministic number of blocks
// (e.g. throttle-window tests) without relying on auto-empty-block timing.
//
// Built directly against the RPC client (rather than gnoclient) because
// gnoclient imports this package in its tests, which would create a cycle.
func waitForNewBlock(ts *testscript.TestScript, remote string, defaultPK crypto.PrivKey) error {
	cli, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return fmt.Errorf("create rpc client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addr := defaultPK.PubKey().Address()
	qres, err := cli.ABCIQuery(ctx, "auth/accounts/"+addr.String(), []byte{})
	if err != nil {
		return fmt.Errorf("query account: %w", err)
	}
	if qres.Response.Error != nil {
		return fmt.Errorf("query account: %w", qres.Response.Error)
	}
	var acct gnoland.GnoAccount
	if err := amino.UnmarshalJSON(qres.Response.Data, &acct); err != nil {
		return fmt.Errorf("unmarshal account: %w", err)
	}

	tx := std.Tx{
		Msgs: []std.Msg{bank.MsgSend{
			FromAddress: addr,
			ToAddress:   addr,
			Amount:      std.Coins{std.NewCoin(ugnot.Denom, 1)},
		}},
		Fee: std.NewFee(890_000, std.NewCoin(ugnot.Denom, 1_000_000)),
	}
	signBytes, err := tx.GetSignBytes("tendermint_test", acct.BaseAccount.GetAccountNumber(), acct.BaseAccount.GetSequence())
	if err != nil {
		return fmt.Errorf("get sign bytes: %w", err)
	}
	sig, err := defaultPK.Sign(signBytes)
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	tx.Signatures = []std.Signature{{PubKey: defaultPK.PubKey(), Signature: sig}}

	txBytes, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf("marshal tx: %w", err)
	}

	bres, err := cli.BroadcastTxCommit(ctx, txBytes)
	if err != nil {
		return fmt.Errorf("broadcast: %w", err)
	}
	if bres.CheckTx.IsErr() {
		return fmt.Errorf("check tx failed: %s", bres.CheckTx.Log)
	}
	if bres.DeliverTx.IsErr() {
		return fmt.Errorf("deliver tx failed: %s", bres.DeliverTx.Log)
	}

	fmt.Fprintf(ts.Stdout(), "new block at height %d\n", bres.Height)
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
