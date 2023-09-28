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

	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/jaekwon/testify/require"
	"github.com/rogpeppe/go-internal/testscript"
)

// XXX: should be centralize somewhere
const (
	test1Addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	test1Seed = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
)

func TestTestdata(t *testing.T) {
	testscript.Run(t, SetupGnolandTestScript(t, "testdata"))
}

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

	fs.StringVar(
		&c.Config,
		"config",
		"",
		"config file (optional)",
	)
}

type testNode struct {
	*node.Node
	logger     log.Logger
	gnokeyexec uint // counter for execution of gnokey
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
	nodes := map[string]*testNode{}

	bTestWork, _ := strconv.ParseBool(os.Getenv("TESTWORK"))
	return testscript.Params{
		TestWork: bTestWork,
		Dir:      txtarDir,
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

					logger := log.NewNopLogger()
					if testing.Verbose() && (os.Getenv("LOG_DIR") != "" || bTestWork) {
						logname := fmt.Sprintf("gnoland-%s.log", sid)
						logger = getTestingLogger(ts, logname)
					}

					dataDir := filepath.Join(gnoDataDir, sid)
					var node *node.Node
					if node, err = execTestingGnoland(t, logger, dataDir, gnoRootDir, args); err == nil {
						nodes[sid] = &testNode{
							Node:   node,
							logger: logger,
						}
						ts.Defer(func() {
							if n := nodes[sid]; n != nil {
								if err := n.Stop(); err != nil {
									panic(fmt.Errorf("node %q was unable to stop: %w", sid, err))
								}
							}
						})

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

					n.gnokeyexec++
					headerlog := fmt.Sprintf("%.02d!EXEC_GNOKEY", n.gnokeyexec)
					// we error to log this even on lower level
					n.logger.Info(headerlog, strings.Join(args, " "))
					defer n.logger.Info(headerlog, "END")
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

func getSessionID(ts *testscript.TestScript) string {
	works := ts.Getenv("WORK")
	sum := crc32.ChecksumIEEE([]byte(works))
	return strconv.FormatUint(uint64(sum), 16)
}

func getTestingLogger(ts *testscript.TestScript, logname string) log.Logger {
	var path string
	if logdir := os.Getenv("LOG_DIR"); logdir != "" {
		if err := os.MkdirAll(logdir, 0o755); err != nil {
			ts.Fatalf("unable to make log directory %q", logdir)
		}

		var err error
		if path, err = filepath.Abs(filepath.Join(logdir, logname)); err != nil {
			ts.Fatalf("uanble to get absolute path of logdir %q", logdir)
		}

	} else if workdir := ts.Getenv("WORK"); workdir != "" {
		path = filepath.Join(workdir, logname)
	} else {
		return log.NewNopLogger()
	}

	// create the file
	f, err := os.Create(path)
	if err != nil {
		ts.Fatalf("unable to create log file %q: %s", path, err.Error())
	}

	ts.Defer(func() {
		if err := f.Close(); err != nil {
			panic(fmt.Errorf("unable to close log file %q: %s", path, err))
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
		panic("invalid log level: " + level)
	}

	ts.Logf("starting logger: %q", path)
	return logger
}
