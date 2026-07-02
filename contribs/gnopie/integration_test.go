package main

import (
	"bytes"
	"context"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEnv holds a running in-memory node and helpers for CLI testing.
type testEnv struct {
	t          *testing.T
	home       string
	remoteAddr string
}

func setupTestEnv(t *testing.T, pkgs ...string) *testEnv {
	t.Helper()

	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(rootdir)

	// Load packages into genesis
	if len(pkgs) > 0 {
		meta := loadTestPkgs(t, rootdir, pkgs...)
		state := config.Genesis.AppState.(gnoland.GnoGenesisState)
		state.Txs = append(state.Txs, meta...)
		config.Genesis.AppState = state
	}

	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	t.Cleanup(func() { node.Stop() })

	// Create temp home with cached remote pointing to our test node
	home := t.TempDir()
	cacheDir := filepath.Join(home, "gnopie", "cache")
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))

	// Write a cached remote so gnopie discovers our test node via gno.land domain
	cacheFile := cachePath(home, "gno.land")
	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))
	cacheContent := "cached_at = 2099-01-01T00:00:00Z\nchain_id = \"tendermint_test\"\nname = \"gno.land\"\nrpc = \"" + remoteAddr + "\"\n"
	require.NoError(t, os.WriteFile(cacheFile, []byte(cacheContent), 0o644))

	// Set up default key in config
	configDir := filepath.Join(home, "gnopie")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	configContent := "key = \"" + integration.DefaultAccount_Name + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(configContent), 0o644))

	// Set up keybase with test account
	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	_, err = kb.CreateAccount(
		integration.DefaultAccount_Name,
		integration.DefaultAccount_Seed,
		"", "", 0, 0,
	)
	require.NoError(t, err)

	return &testEnv{t: t, home: home, remoteAddr: remoteAddr}
}

// newTestIO creates an IO that captures stdout/stderr and provides empty stdin.
func newTestIO() (commands.IO, *bytes.Buffer, *bytes.Buffer) {
	var outBuf, errBuf bytes.Buffer
	cio := &commands.IOImpl{}
	cio.SetIn(strings.NewReader("\n"))
	cio.SetOut(commands.WriteNopCloser(&outBuf))
	cio.SetErr(commands.WriteNopCloser(&errBuf))
	return cio, &outBuf, &errBuf
}

// run executes a gnopie command and returns stdout and stderr.
func (e *testEnv) run(args ...string) (stdout, stderr string) {
	e.t.Helper()

	io, outBuf, errBuf := newTestIO()
	cfg := &baseCfg{home: e.home}
	err := dispatch(context.Background(), cfg, args, io)
	if err != nil {
		errBuf.WriteString("error: " + err.Error() + "\n")
	}

	return outBuf.String(), errBuf.String()
}

// runOK executes a gnopie command and asserts no error.
func (e *testEnv) runOK(args ...string) string {
	e.t.Helper()
	stdout, stderr := e.run(args...)
	if strings.Contains(stderr, "error:") {
		e.t.Fatalf("gnopie %v failed: %s", args, stderr)
	}
	return stdout
}

func loadTestPkgs(t *testing.T, rootdir string, paths ...string) []gnoland.TxWithMetadata {
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

// --- Integration Tests ---
// Each test demonstrates a gnopie command and serves as a usage example.

func TestGET_Render(t *testing.T) {
	// gnopie gno.land/r/demo/counter
	// → calls Render(""), returns the counter value
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	out := env.runOK("gno.land/r/demo/counter")
	assert.Equal(t, "0\n", out, "counter should start at 0")
}

func TestGET_RenderPath(t *testing.T) {
	// gnopie gno.land/r/demo/counter:somepath
	// → calls Render("somepath")
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	out := env.runOK("gno.land/r/demo/counter:somepath")
	// counter's Render ignores the path argument, still returns "0"
	assert.Equal(t, "0\n", out)
}

func TestEVAL_FunctionCall(t *testing.T) {
	// gnopie EVAL 'gno.land/r/demo/counter.Render("")'
	// → evaluates Render("") via qeval
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	out := env.runOK("EVAL", `gno.land/r/demo/counter.Render("")`)
	assert.Contains(t, out, `"0"`)
}

func TestINSPECT_Realm(t *testing.T) {
	// gnopie INSPECT gno.land/r/demo/counter
	// → shows files, functions, storage
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	out := env.runOK("INSPECT", "gno.land/r/demo/counter")
	assert.Contains(t, out, "Realm: gno.land/r/demo/counter")
	assert.Contains(t, out, "func Increment")
	assert.Contains(t, out, "func Render")
	assert.Contains(t, out, "counter int") // variable
}

func TestINSPECT_Network(t *testing.T) {
	// gnopie gno.land
	// → shows network info (block height, chain ID)
	env := setupTestEnv(t)

	out := env.runOK("gno.land")
	assert.Contains(t, out, "Network: gno.land")
	assert.Contains(t, out, "Chain ID:")
	assert.Contains(t, out, "Block height:")
}

func TestREAD_FunctionSource(t *testing.T) {
	// gnopie READ gno.land/r/demo/counter.Increment
	// → shows the source code of the Increment function
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	out := env.runOK("READ", "gno.land/r/demo/counter.Increment")
	assert.Contains(t, out, "func Increment")
	assert.Contains(t, out, "counter++")
}

func TestREAD_File(t *testing.T) {
	// gnopie READ gno.land/r/demo/counter/counter.gno
	// → shows the full file contents
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	out := env.runOK("READ", "gno.land/r/demo/counter/counter.gno")
	assert.Contains(t, out, "package counter")
	assert.Contains(t, out, "func Increment")
	assert.Contains(t, out, "func Render")
}

func TestJSON_Output(t *testing.T) {
	// gnopie --json gno.land/r/demo/counter
	// → returns JSON with render result
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	io, outBuf, _ := newTestIO()
	cfg := &baseCfg{home: env.home, jsonOut: true}
	err := dispatch(context.Background(), cfg, []string{"gno.land/r/demo/counter"}, io)
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, `"pkg_path"`)
	assert.Contains(t, out, `"result"`)
}

func TestCALL_Increment(t *testing.T) {
	// gnopie CALL gno.land/r/demo/counter.Increment()
	// → signs and broadcasts a transaction that increments the counter
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	// First verify counter is 0
	out := env.runOK("gno.land/r/demo/counter")
	assert.Equal(t, "0\n", out)

	// Execute CALL
	io, outBuf, _ := newTestIO()
	cfg := &baseCfg{
		home:           env.home,
		keyName:        integration.DefaultAccount_Name,
		insecureNoPass: true,
		gasWanted:      10_000_000,
		gasFee:         ugnot.ValueString(1000000),
	}
	err := execCall(context.Background(), cfg, "gno.land/r/demo/counter.Increment()", io)
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "TX committed")

	// Verify counter is now 1
	out = env.runOK("gno.land/r/demo/counter")
	assert.Equal(t, "1\n", out)
}

func TestCALL_GenerateGnokey(t *testing.T) {
	// gnopie CALL --print-gnokey-command gno.land/r/demo/counter.Increment()
	// → prints the equivalent gnokey command
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	io, outBuf, _ := newTestIO()
	cfg := &baseCfg{
		home:           env.home,
		keyName:        integration.DefaultAccount_Name,
		printGnokeyCmd: true,
		gasFee:         "1000000ugnot",
	}
	err := execCall(context.Background(), cfg, "gno.land/r/demo/counter.Increment()", io)
	require.NoError(t, err)

	out := outBuf.String()
	assert.Contains(t, out, "gnokey")
	assert.Contains(t, out, "maketx")
	assert.Contains(t, out, "call")
	assert.Contains(t, out, "-func=Increment")
	assert.Contains(t, out, "-pkgpath=gno.land/r/demo/counter")
}

func TestCrossingFunction_AutoInjectCross(t *testing.T) {
	// gnopie 'gno.land/r/demo/counter.Increment()'
	// → Increment is a crossing function (first param is realm),
	//   gnopie should auto-inject `cross` in qeval
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	// EVAL on a crossing function that modifies state will fail with
	// "invalid non-origin call" but the important thing is it doesn't
	// fail with "missing realm argument"
	_, stderr := env.run("EVAL", "gno.land/r/demo/counter.Increment()")
	assert.NotContains(t, stderr, "missing realm argument",
		"crossing function should have `cross` auto-injected")
}

func TestRUN_Increment(t *testing.T) {
	// gnopie RUN gno.land/r/demo/counter.Increment()
	// → generates code and executes via maketx run
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	// First verify counter is 0
	out := env.runOK("gno.land/r/demo/counter")
	assert.Equal(t, "0\n", out)

	// Execute RUN
	io, outBuf, _ := newTestIO()
	cfg := &baseCfg{
		home:           env.home,
		keyName:        integration.DefaultAccount_Name,
		insecureNoPass: true,
		gasWanted:      10_000_000,
		gasFee:         ugnot.ValueString(1000000),
	}
	err := execRun(context.Background(), cfg, "gno.land/r/demo/counter.Increment()", io)
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "TX committed")

	// Verify counter is now 1
	out = env.runOK("gno.land/r/demo/counter")
	assert.Equal(t, "1\n", out)
}

func TestGnoweb_URL(t *testing.T) {
	// gnopie https://gno.land/r/demo/counter
	// → strips https://, calls Render("")
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	out := env.runOK("https://gno.land/r/demo/counter")
	assert.Equal(t, "0\n", out)
}

func TestGnoweb_URL_WithFragment(t *testing.T) {
	// gnopie https://gno.land/r/demo/counter#some-anchor
	// → strips fragment, calls Render("")
	env := setupTestEnv(t, "gno.land/r/demo/counter")

	out := env.runOK("https://gno.land/r/demo/counter#some-anchor")
	assert.Equal(t, "0\n", out)
}

func TestAddress_Inspect(t *testing.T) {
	// gnopie g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
	// → inspects the test account
	env := setupTestEnv(t)

	out, _ := env.run(integration.DefaultAccount_Address)
	assert.Contains(t, out, "Address:")
	assert.Contains(t, out, integration.DefaultAccount_Address)
}
