package integration

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/rogpeppe/go-internal/testscript"
)

const numTestAccounts int = 4

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

	// Track new user balances added via the `adduser` command. These are added to the genesis
	// state when the node is started.
	var newUserBalances []gnoland.Balance

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
			var logger log.Logger
			{
				logger = log.NewNopLogger()
				if persistWorkDir || os.Getenv("LOG_DIR") != "" {
					logname := fmt.Sprintf("gnoland-%s.log", sid)
					logger, err = getTestingLogger(env, logname)
					if err != nil {
						return fmt.Errorf("unable to setup logger: %w", err)
					}
				}

				env.Values["_logger"] = logger
			}

			// test1 must be created outside of the loop below because it is already included in genesis so
			// attempting to recreate results in it getting overwritten and breaking existing tests that
			// rely on its address being static.
			kb.CreateAccount(DefaultAccount_Name, DefaultAccount_Seed, "", "", 0, 0)
			env.Setenv("USER_SEED_"+DefaultAccount_Name, DefaultAccount_Seed)
			env.Setenv("USER_ADDR_"+DefaultAccount_Name, DefaultAccount_Address)

			// Create test accounts starting from test2.
			for i := 1; i < numTestAccounts; i++ {
				accountName := "test" + strconv.Itoa(i+1)

				balance, err := createAccount(env, kb, accountName)
				if err != nil {
					return fmt.Errorf("error creating account %s: %w", accountName, err)
				}

				newUserBalances = append(newUserBalances, balance)
			}

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

				logger := ts.Value("_logger").(log.Logger) // grab logger
				sid := getNodeSID(ts)                      // grab session id

				var cmd string
				cmd, args = args[0], args[1:]

				var err error
				switch cmd {
				case "start":
					if nodeIsRunning(nodes, sid) {
						err = fmt.Errorf("node already started")
						break
					}

					// Warp up `ts` so we can pass it to other testing method
					t := TSTestingT(ts)

					// Generate config and node
					cfg, _ := TestingNodeConfig(t, gnoRootDir)

					// Add balances for users added via the `adduser` command.
					genesis := cfg.Genesis
					genesisState := gnoland.GnoGenesisState{
						Balances: genesis.AppState.(gnoland.GnoGenesisState).Balances,
						Txs:      genesis.AppState.(gnoland.GnoGenesisState).Txs,
					}
					genesisState.Balances = append(genesisState.Balances, newUserBalances...)
					cfg.Genesis.AppState = genesisState

					n, remoteAddr := TestingInMemoryNode(t, logger, cfg)

					// Register cleanup
					nodes[sid] = &testNode{Node: n}

					// Add default environements
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

						// Unset gnoland environements
						ts.Setenv("RPC_ADDR", "")
						fmt.Fprintln(ts.Stdout(), "node stopped successfully")
					}
				default:
					err = fmt.Errorf("invalid gnoland subcommand: %q", cmd)
				}

				tsValidateError(ts, "gnoland "+cmd, neg, err)
			},
			"gnokey": func(ts *testscript.TestScript, neg bool, args []string) {
				logger := ts.Value("_logger").(log.Logger) // grab logger
				sid := ts.Getenv("SID")                    // grab session id

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
					logger.Info(headerlog, strings.Join(args, " "))
					defer logger.Info(headerlog, "END")
				}

				// Inject default argument, if duplicate
				// arguments, it should be override by the ones
				// user provided.
				args = append(defaultArgs, args...)

				err := cmd.ParseAndRun(context.Background(), args)

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

				newUserBalances = append(newUserBalances, balance)
			},
		},
	}
}

func getNodeSID(ts *testscript.TestScript) string {
	return ts.Getenv("SID")
}

func nodeIsRunning(nodes map[string]*testNode, sid string) bool {
	_, ok := nodes[sid]
	return ok
}

func getTestingLogger(env *testscript.Env, logname string) (log.Logger, error) {
	var path string

	if logdir := os.Getenv("LOG_DIR"); logdir != "" {
		if err := os.MkdirAll(logdir, 0o755); err != nil {
			return nil, fmt.Errorf("unable to make log directory %q", logdir)
		}

		var err error
		if path, err = filepath.Abs(filepath.Join(logdir, logname)); err != nil {
			return nil, fmt.Errorf("uanble to get absolute path of logdir %q", logdir)
		}
	} else if workdir := env.Getenv("WORK"); workdir != "" {
		path = filepath.Join(workdir, logname)
	} else {
		return log.NewNopLogger(), nil
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

	logger := log.NewTMLogger(f)
	switch level := os.Getenv("LOG_LEVEL"); strings.ToLower(level) {
	case "error":
		logger.SetLevel(log.LevelError)
	case "debug":
		logger.SetLevel(log.LevelDebug)
	case "info":
		logger.SetLevel(log.LevelInfo)
	case "":
	default:
		return nil, fmt.Errorf("invalid log level %q", level)
	}

	env.T().Log("starting logger: %q", path)
	return logger, nil
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
		Amount:  std.Coins{std.NewCoin("ugnot", 1000000000000000000)},
	}, nil
}
