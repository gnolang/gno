package integration

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	tm2Log "github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/rogpeppe/go-internal/testscript"
	"go.uber.org/zap/zapcore"
)

const (
	envKeyGenesis int = iota
	envKeyLogger
	envKeyPkgsLoader
)

type tSeqShim struct{ *testing.T }

// noop Parallel method allow us to run test sequentially
func (tSeqShim) Parallel() {}

func (t tSeqShim) Run(name string, f func(testscript.T)) {
	t.T.Run(name, func(t *testing.T) {
		f(tSeqShim{t})
	})
}

func (t tSeqShim) Verbose() bool {
	return testing.Verbose()
}

// RunGnolandTestscripts sets up and runs txtar integration tests for gnoland nodes.
// It prepares an in-memory gnoland node and initializes the necessary environment and custom commands.
// The function adapts the test setup for use with the testscript package, enabling
// the execution of gnoland and gnokey commands within txtar scripts.
//
// Refer to package documentation in doc.go for more information on commands and example txtar scripts.
func RunGnolandTestscripts(t *testing.T, txtarDir string) {
	t.Helper()

	p := setupGnolandTestScript(t, txtarDir)
	if deadline, ok := t.Deadline(); ok && p.Deadline.IsZero() {
		p.Deadline = deadline
	}

	testscript.RunT(tSeqShim{t}, p)
}

type testNode struct {
	*node.Node
	nGnoKeyExec uint // Counter for execution of gnokey.
}

func setupGnolandTestScript(t *testing.T, txtarDir string) testscript.Params {
	t.Helper()

	tmpdir := t.TempDir()

	// `gnoRootDir` should point to the local location of the gno repository.
	// It serves as the gno equivalent of GOROOT.
	gnoRootDir := gnoenv.RootDir()

	// `gnoHomeDir` should be the local directory where gnokey stores keys.
	gnoHomeDir := filepath.Join(tmpdir, "gno")

	// Testscripts run concurrently by default, so we need to be prepared for that.
	nodes := map[string]*testNode{}

	updateScripts, _ := strconv.ParseBool(os.Getenv("UPDATE_SCRIPTS"))
	persistWorkDir, _ := strconv.ParseBool(os.Getenv("TESTWORK"))
	return testscript.Params{
		UpdateScripts: updateScripts,
		TestWork:      persistWorkDir,
		Dir:           txtarDir,
		Setup: func(env *testscript.Env) error {
			kb, err := keys.NewKeyBaseFromDir(gnoHomeDir)
			if err != nil {
				return err
			}

			// create sessions ID
			var sid string
			{
				works := env.Getenv("WORK")
				sum := crc32.ChecksumIEEE([]byte(works))
				sid = strconv.FormatUint(uint64(sum), 16)
				env.Setenv("SID", sid)
			}

			// setup logger
			var logger *slog.Logger
			{
				logger = tm2Log.NewNoopLogger()
				if persistWorkDir || os.Getenv("LOG_PATH_DIR") != "" {
					logname := fmt.Sprintf("txtar-gnoland-%s.log", sid)
					logger, err = getTestingLogger(env, logname)
					if err != nil {
						return fmt.Errorf("unable to setup logger: %w", err)
					}
				}

				env.Values[envKeyLogger] = logger
			}

			// Track new user balances added via the `adduser`
			// command and packages added with the `loadpkg` command.
			// This genesis will be use when node is started.
			genesis := &gnoland.GnoGenesisState{
				Balances: LoadDefaultGenesisBalanceFile(t, gnoRootDir),
				Txs:      []std.Tx{},
			}

			// test1 must be created outside of the loop below because it is already included in genesis so
			// attempting to recreate results in it getting overwritten and breaking existing tests that
			// rely on its address being static.
			kb.CreateAccount(DefaultAccount_Name, DefaultAccount_Seed, "", "", 0, 0)
			env.Setenv("USER_SEED_"+DefaultAccount_Name, DefaultAccount_Seed)
			env.Setenv("USER_ADDR_"+DefaultAccount_Name, DefaultAccount_Address)

			env.Values[envKeyGenesis] = genesis
			env.Values[envKeyPkgsLoader] = newPkgsLoader()

			env.Setenv("GNOROOT", gnoRootDir)
			env.Setenv("GNOHOME", gnoHomeDir)

			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"gnoland": func(ts *testscript.TestScript, neg bool, args []string) {
				if len(args) == 0 {
					tsValidateError(ts, "gnoland", neg, fmt.Errorf("syntax: gnoland [start|stop]"))
					return
				}

				logger := ts.Value(envKeyLogger).(*slog.Logger) // grab logger
				sid := getNodeSID(ts)                           // grab session id

				var cmd string
				cmd, args = args[0], args[1:]

				var err error
				switch cmd {
				case "start":
					if nodeIsRunning(nodes, sid) {
						err = fmt.Errorf("node already started")
						break
					}

					// get packages
					pkgs := ts.Value(envKeyPkgsLoader).(*pkgsLoader)                // grab logger
					creator := crypto.MustAddressFromString(DefaultAccount_Address) // test1
					defaultFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
					pkgsTxs, err := pkgs.LoadPackages(creator, defaultFee, nil)
					if err != nil {
						ts.Fatalf("unable to load packages txs: %s", err)
					}

					// Warp up `ts` so we can pass it to other testing method
					t := TSTestingT(ts)

					// Generate config and node
					cfg := TestingMinimalNodeConfig(t, gnoRootDir)
					genesis := ts.Value(envKeyGenesis).(*gnoland.GnoGenesisState)
					genesis.Txs = append(pkgsTxs, genesis.Txs...)

					// setup genesis state
					cfg.Genesis.AppState = *genesis

					n, remoteAddr := TestingInMemoryNode(t, logger, cfg)

					// Register cleanup
					nodes[sid] = &testNode{Node: n}

					// Add default environments
					ts.Setenv("RPC_ADDR", remoteAddr)

					fmt.Fprintln(ts.Stdout(), "node started successfully")
				case "stop":
					n, ok := nodes[sid]
					if !ok {
						err = fmt.Errorf("node not started cannot be stopped")
						break
					}
					if err = n.Stop(); err == nil {
						delete(nodes, sid)

						// Unset gnoland environments
						ts.Setenv("RPC_ADDR", "")
						fmt.Fprintln(ts.Stdout(), "node stopped successfully")
					}
				default:
					err = fmt.Errorf("invalid gnoland subcommand: %q", cmd)
				}

				tsValidateError(ts, "gnoland "+cmd, neg, err)
			},
			"gnokey": func(ts *testscript.TestScript, neg bool, args []string) {
				logger := ts.Value(envKeyLogger).(*slog.Logger) // grab logger
				sid := ts.Getenv("SID")                         // grab session id

				// Unquote args enclosed in `"` to correctly handle `\n` or similar escapes.
				args, err := unquote(args)
				if err != nil {
					tsValidateError(ts, "gnokey", neg, err)
				}

				// Setup IO command
				io := commands.NewTestIO()
				io.SetOut(commands.WriteNopCloser(ts.Stdout()))
				io.SetErr(commands.WriteNopCloser(ts.Stderr()))
				cmd := keyscli.NewRootCmd(io, client.DefaultBaseOptions)

				io.SetIn(strings.NewReader("\n")) // Inject empty password to stdin.
				defaultArgs := []string{
					"-home", gnoHomeDir,
					"-insecure-password-stdin=true", // There no use to not have this param by default.
				}

				if n, ok := nodes[sid]; ok {
					if raddr := n.Config().RPC.ListenAddress; raddr != "" {
						defaultArgs = append(defaultArgs, "-remote", raddr)
					}

					n.nGnoKeyExec++
					headerlog := fmt.Sprintf("%.02d!EXEC_GNOKEY", n.nGnoKeyExec)

					// Log the command inside gnoland logger, so we can better scope errors.
					logger.Info(headerlog, "args", strings.Join(args, " "))
					defer logger.Info(headerlog, "delimiter", "END")
				}

				// Inject default argument, if duplicate
				// arguments, it should be override by the ones
				// user provided.
				args = append(defaultArgs, args...)

				err = cmd.ParseAndRun(context.Background(), args)
				tsValidateError(ts, "gnokey", neg, err)
			},
			// adduser commands must be executed before starting the node; it errors out otherwise.
			"adduser": func(ts *testscript.TestScript, neg bool, args []string) {
				if nodeIsRunning(nodes, getNodeSID(ts)) {
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

				balance, err := createAccount(ts, kb, args[0])
				if err != nil {
					ts.Fatalf("error creating account %s: %s", args[0], err)
				}

				// Add balance to genesis
				genesis := ts.Value(envKeyGenesis).(*gnoland.GnoGenesisState)
				genesis.Balances = append(genesis.Balances, balance)
			},
			// adduserfrom commands must be executed before starting the node; it errors out otherwise.
			"adduserfrom": func(ts *testscript.TestScript, neg bool, args []string) {
				if nodeIsRunning(nodes, getNodeSID(ts)) {
					tsValidateError(ts, "adduserfrom", neg, errors.New("adduserfrom must be used before starting node"))
					return
				}

				var account, index uint64
				var err error

				switch len(args) {
				case 2:
					// expected user input
					// adduserfrom 'username 'menmonic'
					// no need to do anything

				case 4:
					// expected user input
					// adduserfrom 'username 'menmonic' 'account' 'index'

					// parse 'index' first, then fallghrough to `case 3` to parse 'account'
					index, err = strconv.ParseUint(args[3], 10, 32)
					if err != nil {
						ts.Fatalf("invalid index number %s", args[3])
					}

					fallthrough // parse 'account'
				case 3:
					// expected user input
					// adduserfrom 'username 'menmonic' 'account'

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

				balance, err := createAccountFrom(ts, kb, args[0], args[1], uint32(account), uint32(index))
				if err != nil {
					ts.Fatalf("error creating wallet %s", err)
				}

				// Add balance to genesis
				genesis := ts.Value(envKeyGenesis).(*gnoland.GnoGenesisState)
				genesis.Balances = append(genesis.Balances, balance)

				fmt.Fprintf(ts.Stdout(), "Added %s(%s) to genesis", args[0], balance.Address)
			},
			// `loadpkg` load a specific package from the 'examples' or working directory
			"loadpkg": func(ts *testscript.TestScript, neg bool, args []string) {
				// special dirs
				workDir := ts.Getenv("WORK")
				examplesDir := filepath.Join(gnoRootDir, "examples")

				pkgs := ts.Value(envKeyPkgsLoader).(*pkgsLoader)

				var path, name string
				switch len(args) {
				case 2:
					name = args[0]
					path = filepath.Clean(args[1])
				case 1:
					path = filepath.Clean(args[0])
				case 0:
					ts.Fatalf("`loadpkg`: no arguments specified")
				default:
					ts.Fatalf("`loadpkg`: too many arguments specified")
				}

				// If `all` is specified, fully load 'examples' directory.
				// NOTE: In 99% of cases, this is not needed, and
				// packages should be loaded individually.
				if path == "all" {
					ts.Logf("warning: loading all packages")
					if err := pkgs.LoadAllPackagesFromDir(examplesDir); err != nil {
						ts.Fatalf("unable to load packages from %q: %s", examplesDir, err)
					}

					return
				}

				if !strings.HasPrefix(path, workDir) {
					path = filepath.Join(examplesDir, path)
				}

				if err := pkgs.LoadPackage(examplesDir, path, name); err != nil {
					ts.Fatalf("`loadpkg` unable to load package(s) from %q: %s", args[0], err)
				}

				ts.Logf("%q package was added to genesis", args[0])
			},
		},
	}
}

// `unquote` takes a slice of strings, resulting from splitting a string block by spaces, and
// processes them. The function handles quoted phrases and escape characters within these strings.
func unquote(args []string) ([]string, error) {
	const quote = '"'

	parts := []string{}
	var inQuote bool

	var part strings.Builder
	for _, arg := range args {
		var escaped bool
		for _, c := range arg {
			if escaped {
				// If the character is meant to be escaped, it is processed with Unquote.
				// We use `Unquote` here for two main reasons:
				// 1. It will validate that the escape sequence is correct
				// 2. It converts the escaped string to its corresponding raw character.
				//    For example, "\\t" becomes '\t'.
				uc, err := strconv.Unquote(`"\` + string(c) + `"`)
				if err != nil {
					return nil, fmt.Errorf("unhandled escape sequence `\\%c`: %w", c, err)
				}

				part.WriteString(uc)
				escaped = false
				continue
			}

			// If we are inside a quoted string and encounter an escape character,
			// flag the next character as `escaped`
			if inQuote && c == '\\' {
				escaped = true
				continue
			}

			// Detect quote and toggle inQuote state
			if c == quote {
				inQuote = !inQuote
				continue
			}

			// Handle regular character
			part.WriteRune(c)
		}

		// If we're inside a quote, add a single space.
		// It reflects one or multiple spaces between args in the original string.
		if inQuote {
			part.WriteRune(' ')
			continue
		}

		// Finalize part, add to parts, and reset for next part
		parts = append(parts, part.String())
		part.Reset()
	}

	// Check if a quote is left open
	if inQuote {
		return nil, errors.New("unfinished quote")
	}

	return parts, nil
}

func getNodeSID(ts *testscript.TestScript) string {
	return ts.Getenv("SID")
}

func nodeIsRunning(nodes map[string]*testNode, sid string) bool {
	_, ok := nodes[sid]
	return ok
}

func getTestingLogger(env *testscript.Env, logname string) (*slog.Logger, error) {
	var path string

	if logdir := os.Getenv("LOG_PATH_DIR"); logdir != "" {
		if err := os.MkdirAll(logdir, 0o755); err != nil {
			return nil, fmt.Errorf("unable to make log directory %q", logdir)
		}

		var err error
		if path, err = filepath.Abs(filepath.Join(logdir, logname)); err != nil {
			return nil, fmt.Errorf("unable to get absolute path of logdir %q", logdir)
		}
	} else if workdir := env.Getenv("WORK"); workdir != "" {
		path = filepath.Join(workdir, logname)
	} else {
		return tm2Log.NewNoopLogger(), nil
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("unable to create log file %q: %w", path, err)
	}

	env.Defer(func() {
		if err := f.Close(); err != nil {
			panic(fmt.Errorf("unable to close log file %q: %w", path, err))
		}
	})

	// Initialize the logger
	logLevel, err := zapcore.ParseLevel(strings.ToLower(os.Getenv("LOG_LEVEL")))
	if err != nil {
		return nil, fmt.Errorf("unable to parse log level, %w", err)
	}

	// Build zap logger for testing
	zapLogger := log.NewZapTestingLogger(f, logLevel)
	env.Defer(func() { zapLogger.Sync() })

	env.T().Log("starting logger", path)
	return log.ZapLoggerToSlog(zapLogger), nil
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

type envSetter interface {
	Setenv(key, value string)
}

// createAccount creates a new account with the given name and adds it to the keybase.
func createAccount(env envSetter, kb keys.Keybase, accountName string) (gnoland.Balance, error) {
	var balance gnoland.Balance
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return balance, fmt.Errorf("error creating entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return balance, fmt.Errorf("error generating mnemonic: %w", err)
	}

	var keyInfo keys.Info
	if keyInfo, err = kb.CreateAccount(accountName, mnemonic, "", "", 0, 0); err != nil {
		return balance, fmt.Errorf("unable to create account: %w", err)
	}

	address := keyInfo.GetAddress()
	env.Setenv("USER_SEED_"+accountName, mnemonic)
	env.Setenv("USER_ADDR_"+accountName, address.String())

	return gnoland.Balance{
		Address: address,
		Amount:  std.Coins{std.NewCoin("ugnot", 10e6)},
	}, nil
}

// createAccountFrom creates a new account with the given metadata and adds it to the keybase.
func createAccountFrom(env envSetter, kb keys.Keybase, accountName, mnemonic string, account, index uint32) (gnoland.Balance, error) {
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
	env.Setenv("USER_SEED_"+accountName, mnemonic)
	env.Setenv("USER_ADDR_"+accountName, address.String())

	return gnoland.Balance{
		Address: address,
		Amount:  std.Coins{std.NewCoin("ugnot", 10e6)},
	}, nil
}

type pkgsLoader struct {
	pkgs    []gnomod.Pkg
	visited map[string]struct{}
}

func newPkgsLoader() *pkgsLoader {
	return &pkgsLoader{
		pkgs:    make([]gnomod.Pkg, 0),
		visited: make(map[string]struct{}),
	}
}

func (pl *pkgsLoader) List() gnomod.PkgList {
	return pl.pkgs
}

func (pl *pkgsLoader) LoadPackages(creator bft.Address, fee std.Fee, deposit std.Coins) ([]std.Tx, error) {
	pkgslist, err := pl.List().Sort() // sorts packages by their dependencies.
	if err != nil {
		return nil, fmt.Errorf("unable to sort packages: %w", err)
	}

	txs := make([]std.Tx, len(pkgslist))
	for i, pkg := range pkgslist {
		tx, err := gnoland.LoadPackage(pkg, creator, fee, deposit)
		if err != nil {
			return nil, fmt.Errorf("unable to load pkg %q: %w", pkg.Name, err)
		}
		txs[i] = tx
	}

	return txs, nil
}

func (pl *pkgsLoader) LoadAllPackagesFromDir(path string) error {
	// list all packages from target path
	pkgslist, err := gnomod.ListPkgs(path)
	if err != nil {
		return fmt.Errorf("listing gno packages: %w", err)
	}

	for _, pkg := range pkgslist {
		if !pl.exist(pkg) {
			pl.add(pkg)
		}
	}

	return nil
}

func (pl *pkgsLoader) LoadPackage(modroot string, path, name string) error {
	// Initialize a queue with the root package
	queue := []gnomod.Pkg{{Dir: path, Name: name}}

	for len(queue) > 0 {
		// Dequeue the first package
		currentPkg := queue[0]
		queue = queue[1:]

		if currentPkg.Dir == "" {
			return fmt.Errorf("no path specified for package")
		}

		if currentPkg.Name == "" {
			// Load `gno.mod` information
			gnoModPath := filepath.Join(currentPkg.Dir, "gno.mod")
			gm, err := gnomod.ParseGnoMod(gnoModPath)
			if err != nil {
				return fmt.Errorf("unable to load %q: %w", gnoModPath, err)
			}
			gm.Sanitize()

			// Override package info with mod infos
			currentPkg.Name = gm.Module.Mod.Path
			currentPkg.Draft = gm.Draft
			for _, req := range gm.Require {
				currentPkg.Requires = append(currentPkg.Requires, req.Mod.Path)
			}
		}

		if currentPkg.Draft {
			continue // Skip draft package
		}

		if pl.exist(currentPkg) {
			continue
		}
		pl.add(currentPkg)

		// Add requirements to the queue
		for _, pkgPath := range currentPkg.Requires {
			fullPath := filepath.Join(modroot, pkgPath)
			queue = append(queue, gnomod.Pkg{Dir: fullPath})
		}
	}

	return nil
}

func (pl *pkgsLoader) add(pkg gnomod.Pkg) {
	pl.visited[pkg.Name] = struct{}{}
	pl.pkgs = append(pl.pkgs, pkg)
}

func (pl *pkgsLoader) exist(pkg gnomod.Pkg) (ok bool) {
	_, ok = pl.visited[pkg.Name]
	return
}
