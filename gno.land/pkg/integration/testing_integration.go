package integration

import (
	"context"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/rogpeppe/go-internal/testscript"
)

// XXX: This should be centralize somewhere.
const (
	test1Addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	test1Seed = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
)

type testNode struct {
	*node.Node
	logger      log.Logger
	nGnoKeyExec uint // Counter for execution of gnokey.
}

// SetupGnolandTestScript prepares the test environment to execute txtar tests
// using a partial InMemory gnoland node. It initializes key storage, sets up the gnoland node,
// and provides custom commands like "gnoland" and "gnokey" for txtar script execution.
//
// The function returns testscript.Params which contain the test setup and command
// executions to be used with the testscript package.
//
// For a detailed explanation of the commands and their behaviors, as well as
// example txtar scripts, refer to the package documentation in doc.go.
func SetupGnolandTestScript(t *testing.T, txtarDir string) testscript.Params {
	t.Helper()

	tmpdir := t.TempDir()

	// `gnoRootDir` should point to the local location of the gno repository.
	// It serves as the gno equivalent of GOROOT.
	gnoRootDir := gnoland.GuessGnoRootDir()

	// `gnoHomeDir` should be the local directory where gnokey stores keys.
	gnoHomeDir := filepath.Join(tmpdir, "gno")

	// `gnoDataDir` should refer to the local location where the gnoland node
	// stores its configuration and data.
	gnoDataDir := filepath.Join(tmpdir, "data")

	// Testscripts run concurrently by default, so we need to be prepared for that.
	var muNodes sync.Mutex
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

			// XXX: Add a command to add custom account.
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
					tsValidateError(ts, "gnoland", neg, fmt.Errorf("syntax: gnoland [start|stop]"))
					return
				}

				sid := getSessionID(ts)

				var cmd string
				cmd, args = args[0], args[1:]

				var err error
				switch cmd {
				case "start":
					if _, ok := nodes[sid]; ok {
						err = fmt.Errorf("node already started")
						break
					}

					logger := log.NewNopLogger()
					if persistWorkDir || os.Getenv("LOG_DIR") != "" {
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
							muNodes.Lock()
							defer muNodes.Unlock()

							if n := nodes[sid]; n != nil {
								if err := n.Stop(); err != nil {
									panic(fmt.Errorf("node %q was unable to stop: %w", sid, err))
								}
							}
						})

						// Get listen address environment.
						// It should have been updated with the right port on start.
						laddr := node.Config().RPC.ListenAddress

						// Add default environements.
						ts.Setenv("RPC_ADDR", laddr)
						ts.Setenv("GNODATA", gnoDataDir)

						const listenerID = "testing_listener"

						// Wait for first block by waiting for `EventNewBlock` event.
						nb := make(chan struct{}, 1)
						node.EventSwitch().AddListener(listenerID, func(ev events.Event) {
							if _, ok := ev.(types.EventNewBlock); ok {
								select {
								case nb <- struct{}{}:
								default:
								}
							}
						})

						if node.BlockStore().Height() == 0 {
							select {
							case <-nb: // ok
							case <-time.After(time.Second * 6):
								ts.Fatalf("timeout while waiting for the node to start")
							}
						}

						node.EventSwitch().RemoveListener(listenerID)

						fmt.Fprintln(ts.Stdout(), "node started successfully")
					}
				case "stop":
					n, ok := nodes[sid]
					if !ok {
						err = fmt.Errorf("node not started cannot be stopped")
						break
					}

					if err = n.Stop(); err == nil {
						delete(nodes, sid)

						// Unset gnoland environements.
						ts.Setenv("RPC_ADDR", "")
						ts.Setenv("GNODATA", "")
						fmt.Fprintln(ts.Stdout(), "node stopped successfully")
					}
				default:
					err = fmt.Errorf("invalid gnoland subcommand: %q", cmd)
				}

				tsValidateError(ts, "gnoland "+cmd, neg, err)
			},
			"gnokey": func(ts *testscript.TestScript, neg bool, args []string) {
				muNodes.Lock()
				defer muNodes.Unlock()

				sid := getSessionID(ts)

				// Setup IO command.
				io := commands.NewTestIO()
				io.SetOut(commands.WriteNopCloser(ts.Stdout()))
				io.SetErr(commands.WriteNopCloser(ts.Stderr()))
				cmd := client.NewRootCmd(io)

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
					n.logger.Info(headerlog, strings.Join(args, " "))
					defer n.logger.Info(headerlog, "END")
				}

				// Inject default argument, if duplicate
				// arguments, it should be override by the ones
				// user provided.
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

	f, err := os.Create(path)
	if err != nil {
		ts.Fatalf("unable to create log file %q: %s", path, err.Error())
	}

	ts.Defer(func() {
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
		ts.Fatalf("invalid log level %q", level)
	}

	ts.Logf("starting logger: %q", path)
	return logger
}
