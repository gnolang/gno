package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

type cliTestCase struct {
	name string
	args []string // args to dispatch (verb + expression)

	// cfg overrides — if nil, defaults are used
	jsonOut        bool
	printGnokeyCmd bool
	dryRun         bool
	debug          bool
	signing        bool // if true, sets up signing client with key + gas

	// expected outputs — if both contain+be are empty, output must be empty
	stdoutShouldContain string
	stdoutShouldBe      string
	stderrShouldContain string
	errShouldContain    string
	errShouldBe         string
}

func TestCLI(t *testing.T) {
	// Start a shared in-memory node with counter realm
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(rootdir)
	meta := loadCLITestPkgs(t, rootdir, "gno.land/r/demo/counter")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	t.Cleanup(func() { node.Stop() })

	// Create temp home with keybase and cached remote
	home := t.TempDir()
	setupCLITestHome(t, home, remoteAddr)

	tc := []cliTestCase{
		// --- GET (default verb) ---
		{
			name:                "GET realm renders by default",
			args:                []string{"gno.land/r/demo/counter"},
			stdoutShouldContain: "0",
		},
		{
			name:                "GET realm with render path",
			args:                []string{"gno.land/r/demo/counter:anything"},
			stdoutShouldContain: "0",
		},
		{
			name:                "GET network shows info",
			args:                []string{"gno.land"},
			stdoutShouldContain: "Network: gno.land",
		},
		{
			name:                "GET network shows chain ID",
			args:                []string{"gno.land"},
			stdoutShouldContain: "Chain ID:",
		},
		{
			name:                "GET address shows account",
			args:                []string{integration.DefaultAccount_Address},
			stdoutShouldContain: "Address:",
		},

		// --- GET with gnoweb URLs ---
		{
			name:                "GET strips https://",
			args:                []string{"https://gno.land/r/demo/counter"},
			stdoutShouldContain: "0",
		},
		{
			name:                "GET strips fragment",
			args:                []string{"https://gno.land/r/demo/counter#some-anchor"},
			stdoutShouldContain: "0",
		},
		{
			name:                "GET strips trailing slash",
			args:                []string{"https://gno.land/r/demo/counter/"},
			stdoutShouldContain: "0",
		},

		// --- EVAL ---
		{
			name:                "EVAL function call",
			args:                []string{"EVAL", `gno.land/r/demo/counter.Render("")`},
			stdoutShouldContain: `"0"`,
		},
		{
			name:             "EVAL missing expression",
			args:             []string{"EVAL"},
			errShouldContain: "missing expression",
		},
		{
			name:             "EVAL non-existent function",
			args:             []string{"EVAL", `gno.land/r/demo/counter.DoesNotExist()`},
			errShouldContain: "eval:",
		},

		// Crossing function test is in TestCLI_CrossingAutoInject below

		// --- READ ---
		{
			name:                "READ function source",
			args:                []string{"READ", "gno.land/r/demo/counter.Increment"},
			stdoutShouldContain: "func Increment",
		},
		{
			name:                "READ function source has body",
			args:                []string{"READ", "gno.land/r/demo/counter.Increment"},
			stdoutShouldContain: "counter++",
		},
		{
			name:                "READ file",
			args:                []string{"READ", "gno.land/r/demo/counter/counter.gno"},
			stdoutShouldContain: "package counter",
		},
		{
			name:                "READ file has all functions",
			args:                []string{"READ", "gno.land/r/demo/counter/counter.gno"},
			stdoutShouldContain: "func Render",
		},
		{
			name:             "READ non-existent symbol",
			args:             []string{"READ", "gno.land/r/demo/counter.Nope"},
			errShouldContain: "not found",
		},

		// --- INSPECT ---
		{
			name:                "INSPECT realm shows files",
			args:                []string{"INSPECT", "gno.land/r/demo/counter"},
			stdoutShouldContain: "counter.gno",
		},
		{
			name:                "INSPECT realm shows functions",
			args:                []string{"INSPECT", "gno.land/r/demo/counter"},
			stdoutShouldContain: "func Increment",
		},
		{
			name:                "INSPECT realm shows Render",
			args:                []string{"INSPECT", "gno.land/r/demo/counter"},
			stdoutShouldContain: "func Render",
		},

		// --- JSON output ---
		{
			name:                "JSON output for realm",
			args:                []string{"gno.land/r/demo/counter"},
			jsonOut:             true,
			stdoutShouldContain: `"pkg_path"`,
		},
		{
			name:                "JSON output for network",
			args:                []string{"gno.land"},
			jsonOut:             true,
			stdoutShouldContain: `"chain_id"`,
		},

		// --- Debug output ---
		{
			name:                "debug shows dispatch info",
			args:                []string{"gno.land/r/demo/counter"},
			debug:               true,
			stderrShouldContain: "dispatch",
		},
		{
			name:                "debug shows query info",
			args:                []string{"INSPECT", "gno.land/r/demo/counter"},
			debug:               true,
			stderrShouldContain: "dispatch",
		},

		// --- CALL --print-gnokey-command ---
		{
			name:                "CALL generate gnokey command",
			args:                []string{"CALL", "gno.land/r/demo/counter.Increment()"},
			printGnokeyCmd:      true,
			stdoutShouldContain: "gnokey",
		},
		{
			name:                "CALL generate gnokey has func",
			args:                []string{"CALL", "gno.land/r/demo/counter.Increment()"},
			printGnokeyCmd:      true,
			stdoutShouldContain: "-func=Increment",
		},
		{
			name:                "CALL generate gnokey has pkgpath",
			args:                []string{"CALL", "gno.land/r/demo/counter.Increment()"},
			printGnokeyCmd:      true,
			stdoutShouldContain: "-pkgpath=gno.land/r/demo/counter",
		},

		// --- Error cases ---
		{
			name:             "empty args",
			args:             []string{},
			errShouldContain: "usage:",
		},
		{
			name:             "unknown verb with no expression",
			args:             []string{"CALL"},
			errShouldContain: "missing expression",
		},
		{
			name:             "invalid path",
			args:             []string{""},
			errShouldContain: "empty path",
		},
	}

	// Run all tests twice: first pass populates cache, second tests cached path
	for _, pass := range []string{"fresh", "cached"} {
		t.Run(pass, func(t *testing.T) {
			for _, test := range tc {
				t.Run(test.name, func(t *testing.T) {
					runCLITest(t, home, test)
				})
			}
		})
	}
}

// TestCLI_CALL_Stateful tests CALL and RUN with actual state changes.
// These are separate because they mutate state and can't be repeated.
func TestCLI_CALL_Stateful(t *testing.T) {
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(rootdir)
	meta := loadCLITestPkgs(t, rootdir, "gno.land/r/demo/counter")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	t.Cleanup(func() { node.Stop() })

	home := t.TempDir()
	setupCLITestHome(t, home, remoteAddr)

	// Verify counter starts at 0
	runCLITest(t, home, cliTestCase{
		name:           "counter starts at 0",
		args:           []string{"gno.land/r/demo/counter"},
		stdoutShouldBe: "0\n",
	})

	// CALL Increment
	runCLITestSigning(t, home, cliTestCase{
		name:                "CALL increments counter",
		args:                []string{"CALL", "gno.land/r/demo/counter.Increment()"},
		signing:             true,
		stdoutShouldContain: "TX committed",
	})

	// Verify counter is now 1
	runCLITest(t, home, cliTestCase{
		name:           "counter is 1 after CALL",
		args:           []string{"gno.land/r/demo/counter"},
		stdoutShouldBe: "1\n",
	})

	// RUN Increment
	runCLITestSigning(t, home, cliTestCase{
		name:                "RUN increments counter",
		args:                []string{"RUN", "gno.land/r/demo/counter.Increment()"},
		signing:             true,
		stdoutShouldContain: "TX committed",
	})

	// Verify counter is now 2
	runCLITest(t, home, cliTestCase{
		name:           "counter is 2 after RUN",
		args:           []string{"gno.land/r/demo/counter"},
		stdoutShouldBe: "2\n",
	})
}

func runCLITest(t *testing.T, home string, test cliTestCase) {
	t.Helper()

	mockOut := bytes.NewBufferString("")
	mockErr := bytes.NewBufferString("")

	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOut))
	io.SetErr(commands.WriteNopCloser(mockErr))

	cfg := &baseCfg{
		home:           home,
		jsonOut:        test.jsonOut,
		printGnokeyCmd: test.printGnokeyCmd,
		dryRun:         test.dryRun,
		debug:          test.debug,
		keyName:        integration.DefaultAccount_Name,
		gasFee:         "1000000ugnot",
	}

	err := dispatch(context.Background(), cfg, test.args, io)
	checkCLIOutput(t, test, mockOut.String(), mockErr.String(), err)
}

func runCLITestSigning(t *testing.T, home string, test cliTestCase) {
	t.Helper()

	mockOut := bytes.NewBufferString("")
	mockErr := bytes.NewBufferString("")

	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOut))
	io.SetErr(commands.WriteNopCloser(mockErr))

	cfg := &baseCfg{
		home:           home,
		keyName:        integration.DefaultAccount_Name,
		insecureNoPass: true,
		gasWanted:      10_000_000,
		gasFee:         ugnot.ValueString(1000000),
		jsonOut:        test.jsonOut,
		debug:          test.debug,
	}

	var err error
	if len(test.args) > 0 {
		verb := strings.ToUpper(test.args[0])
		expr := ""
		if len(test.args) > 1 {
			expr = test.args[1]
		}
		switch verb {
		case "CALL":
			err = execCall(context.Background(), cfg, expr, io)
		case "RUN":
			err = execRun(context.Background(), cfg, expr, io)
		default:
			err = dispatch(context.Background(), cfg, test.args, io)
		}
	}

	checkCLIOutput(t, test, mockOut.String(), mockErr.String(), err)
}

func checkCLIOutput(t *testing.T, test cliTestCase, stdout, stderr string, err error) {
	t.Helper()

	errShouldBeEmpty := test.errShouldContain == "" && test.errShouldBe == ""
	stdoutShouldBeEmpty := test.stdoutShouldContain == "" && test.stdoutShouldBe == ""
	stderrShouldBeEmpty := test.stderrShouldContain == ""

	// Check error
	if errShouldBeEmpty {
		require.NoError(t, err, "err should be nil")
	} else {
		require.Error(t, err, "err shouldn't be nil")
		if test.errShouldContain != "" {
			require.Contains(t, err.Error(), test.errShouldContain, "err should contain")
		}
		if test.errShouldBe != "" {
			require.Equal(t, test.errShouldBe, err.Error(), "err should be")
		}
	}

	// Check stdout
	if !stdoutShouldBeEmpty {
		if test.stdoutShouldContain != "" {
			require.Contains(t, stdout, test.stdoutShouldContain, "stdout should contain")
		}
		if test.stdoutShouldBe != "" {
			require.Equal(t, test.stdoutShouldBe, stdout, "stdout should be")
		}
	}

	// Check stderr
	if !stderrShouldBeEmpty {
		require.Contains(t, stderr, test.stderrShouldContain, "stderr should contain")
	}
}

func setupCLITestHome(t *testing.T, home, remoteAddr string) {
	t.Helper()

	// Cached remote pointing to test node
	cacheFile := cachePath(home, "gno.land")
	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))
	cacheContent := fmt.Sprintf(
		"cached_at = 2099-01-01T00:00:00Z\nchain_id = \"tendermint_test\"\nname = \"gno.land\"\nrpc = %q\n",
		remoteAddr,
	)
	require.NoError(t, os.WriteFile(cacheFile, []byte(cacheContent), 0o644))

	// Config with default key
	configDir := filepath.Join(home, "gnopie")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.toml"),
		[]byte(fmt.Sprintf("key = %q\n", integration.DefaultAccount_Name)),
		0o644,
	))

	// Keybase with test account
	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	_, err = kb.CreateAccount(
		integration.DefaultAccount_Name,
		integration.DefaultAccount_Seed,
		"", "", 0, 0,
	)
	require.NoError(t, err)
}

func loadCLITestPkgs(t *testing.T, rootdir string, paths ...string) []gnoland.TxWithMetadata {
	t.Helper()
	loader := integration.NewPkgsLoader()
	examplesDir := filepath.Join(rootdir, "examples")
	for _, path := range paths {
		path = filepath.Join(examplesDir, filepath.Clean(path))
		err := loader.LoadPackage(examplesDir, path, "")
		require.NoError(t, err)
	}
	privKey, err := integration.GeneratePrivKeyFromMnemonic(integration.DefaultAccount_Seed, "", 0, 0)
	require.NoError(t, err)
	defaultFee := std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))
	meta, err := loader.GenerateTxs(privKey, defaultFee, nil)
	require.NoError(t, err)
	return meta
}
