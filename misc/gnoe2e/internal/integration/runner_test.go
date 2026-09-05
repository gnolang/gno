package integration

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPerNodeRPCAddressesAreExported(t *testing.T) {
	dir := t.TempDir()
	script := writeScriptIn(t, dir, "addrs.txtar",
		"exec echo $RPC_ADDR_0 $RPC_ADDR_2\nstdout 'node-zero node-two'\n")

	adapter := NewTestscriptT(testLogger(t), false)
	err := runWithAdapter(adapter, RunConfig{
		ScriptPath: script,
		RPCAddr:    "node-zero",
		RPCAddrs:   []string{"node-zero", "node-one", "node-two"},
		ChainID:    "test",
		Logger:     testLogger(t),
	})

	require.NoError(t, err)
}

// GPAO_ADDR is the oracle's own approver address. A scenario proving no money
// moved has to query this specific account's balance, and it has no other way
// to name it -- $GPAO_KEY_NAME identifies the keybase entry, not the address.
func TestGpaoAddrIsExported(t *testing.T) {
	dir := t.TempDir()
	script := writeScriptIn(t, dir, "addr.txtar",
		"exec echo $GPAO_ADDR\nstdout 'g1gpaoapprover'\n")

	adapter := NewTestscriptT(testLogger(t), false)
	err := runWithAdapter(adapter, RunConfig{
		ScriptPath: script,
		ChainID:    "test",
		GpaoAddr:   "g1gpaoapprover",
		Logger:     testLogger(t),
	})

	require.NoError(t, err)
}

// gnokey injects -remote from RPC_ADDR. A scenario asserting that a package is
// callable on every validator has to override that per invocation, which works
// only because caller arguments are appended after the injected ones and the
// flag parser keeps the last occurrence.
//
// -remote is a root flag, so it has to precede the subcommand. A script writes
// "gnokey -remote $RPC_ADDR_2 query ...", never "gnokey query -remote ...".
func TestGnokeyArgsPutCallerFlagsLast(t *testing.T) {
	args := gnokeyArgs("/home", "injected",
		[]string{"-remote", "explicit", "query", "vm/qfile"})

	require.Equal(t, []string{
		"-home", "/home", "-insecure-password-stdin=true",
		"-remote", "injected",
		"-remote", "explicit", "query", "vm/qfile",
	}, args)
}

// testscriptCmds' registration set is one instruction from silently
// shrinking, and a command dropped from it surfaces only when a scenario hits
// "unknown command". Pinning the exact set here, rather than deriving it from the
// same call under test, makes a dropped -- or silently added -- registration
// fail immediately instead of surfacing as an unrelated-looking bug downstream.
func TestTestscriptCmdsRegistersExpectedCommands(t *testing.T) {
	expected := []string{"gnokey", "sleep", "repeat", "eventually", "http_get", "gpao", "validator"}

	cmds := testscriptCmds(RunConfig{})

	for _, name := range expected {
		fn, ok := cmds[name]
		require.True(t, ok, "missing registration for %q", name)
		require.NotNil(t, fn, "%q registered but nil; would panic on use", name)
	}
	// Catches a silently *added* command too: a later task adding a
	// validator-lifecycle command must update this table deliberately.
	require.Len(t, cmds, len(expected), "unexpected command(s) registered beyond %v", expected)
}

func writeScriptIn(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

// docs/dialect.md lists the variables a scenario may name, and a scenario that
// names one the setup stopped exporting reads an empty string: testscript
// substitutes an unset variable rather than refusing the line, so the assertion
// after it passes or fails for reasons unrelated to what it claims. The set is
// pinned here as a whole, not one variable at a time.
func TestScriptEnvironmentExportsTheDocumentedSet(t *testing.T) {
	dir := t.TempDir()
	script := writeScriptIn(t, dir, "env.txtar",
		"exec echo $RPC_ADDR/$RPC_ADDR_0/$CHAIN_ID/$USER_ADDR/$USER_NAME/$GNOHOME/$GPAO_KEY_NAME/$GPAO_ADDR\n"+
			"stdout 'node-zero/node-zero/test-chain/g1user/test1/"+regexp.QuoteMeta(dir)+"/gpao/g1oracle'\n"+
			"exec sh -c 'test -n \"$GNOROOT\"'\n")

	err := runWithAdapter(NewTestscriptT(testLogger(t), false), RunConfig{
		ScriptPath:  script,
		RPCAddr:     "node-zero",
		RPCAddrs:    []string{"node-zero"},
		ChainID:     "test-chain",
		UserAddr:    "g1user",
		KeyName:     "test1",
		GnoHome:     dir,
		GpaoKeyName: "gpao",
		GpaoAddr:    "g1oracle",
		Logger:      testLogger(t),
	})

	require.NoError(t, err)
}
