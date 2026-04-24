package keyscli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Helpers ───

func newWizardIO(stdin string) (commands.IO, *bytes.Buffer, *bytes.Buffer) {
	var outBuf, errBuf bytes.Buffer
	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(stdin))
	io.SetOut(commands.WriteNopCloser(&outBuf))
	io.SetErr(commands.WriteNopCloser(&errBuf))
	return io, &outBuf, &errBuf
}

func createTestKey(t *testing.T, kbHome, keyName, passphrase string) {
	t.Helper()
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)
	mnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	_, err = kb.CreateAccount(keyName, mnemonic, "", passphrase, 0, 0)
	require.NoError(t, err)
}

func createTestGnoFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, filename)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func createTestPkgDir(t *testing.T, dir, pkgPath string) string {
	t.Helper()
	gnoDir := filepath.Join(dir, "pkg")
	require.NoError(t, os.MkdirAll(gnoDir, 0o755))

	gnomodContent := fmt.Sprintf("module = %q\ngno = \"0.9\"", pkgPath)
	require.NoError(t, os.WriteFile(filepath.Join(gnoDir, "gnomod.toml"), []byte(gnomodContent), 0o644))

	gnoContent := fmt.Sprintf("package %s", filepath.Base(pkgPath))
	require.NoError(t, os.WriteFile(filepath.Join(gnoDir, filepath.Base(pkgPath)+".gno"), []byte(gnoContent), 0o644))

	return gnoDir
}

// ─── Unit Tests: promptKeyOrAddress ───

func TestPromptKeyOrAddress_WithKeys(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()
	createTestKey(t, kbHome, "alice", "test1234")
	createTestKey(t, kbHome, "bob", "test1234")

	io, _, _ := newWizardIO("1\n")

	keyName, err := promptKeyOrAddress(kbHome, io)
	require.NoError(t, err)
	assert.Equal(t, "alice", keyName)
}

func TestPromptKeyOrAddress_NoKeys(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()

	io, _, _ := newWizardIO("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\n")

	keyName, err := promptKeyOrAddress(kbHome, io)
	require.NoError(t, err)
	assert.Equal(t, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", keyName)
}

func TestPromptKeyOrAddress_InvalidAddressRetries(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()

	io, _, errBuf := newWizardIO("invalid-bech32\ng1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\n")

	keyName, err := promptKeyOrAddress(kbHome, io)
	require.NoError(t, err)
	assert.Contains(t, errBuf.String(), "invalid")
	assert.Equal(t, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", keyName)
}

func TestPromptKeyOrAddress_SelectByNumber(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()
	createTestKey(t, kbHome, "alice", "test1234")
	createTestKey(t, kbHome, "bob", "test1234")

	io, _, _ := newWizardIO("2\n")

	keyName, err := promptKeyOrAddress(kbHome, io)
	require.NoError(t, err)
	assert.Equal(t, "bob", keyName)
}

func TestPromptKeyOrAddress_TypeNameDirectly(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()
	createTestKey(t, kbHome, "alice", "test1234")

	io, _, _ := newWizardIO("alice\n")

	keyName, err := promptKeyOrAddress(kbHome, io)
	require.NoError(t, err)
	assert.Equal(t, "alice", keyName)
}

// Core GnowebTxURL coverage lives in root_test.go; no separate tests here.

// ─── Unit Tests: knownNetworks ───

func TestKnownNetworks_ExpectedPresent(t *testing.T) {
	t.Parallel()
	for _, want := range []string{"dev", "staging", "gnoland1"} {
		assert.NotNil(t, findNetworkByChainID(want), "missing network %q", want)
	}
}

func TestFindNetworkByChainID(t *testing.T) {
	t.Parallel()

	n := findNetworkByChainID("staging")
	require.NotNil(t, n)
	assert.Equal(t, "staging", n.ChainID)
	assert.Equal(t, "https://rpc.staging.gno.land:443", n.Remote)

	n = findNetworkByChainID("nonexistent")
	assert.Nil(t, n)
}

func TestFindNetworkByRemote(t *testing.T) {
	t.Parallel()

	n := findNetworkByRemote("https://rpc.gno.land:443")
	require.NotNil(t, n)
	assert.Equal(t, "gnoland1", n.ChainID)

	n = findNetworkByRemote("nonexistent")
	assert.Nil(t, n)
}

// ─── Unit Tests: printAirGapHints ───

func TestPrintAirGapHints(t *testing.T) {
	t.Parallel()

	io, _, errBuf := newWizardIO("")

	printAirGapHints(io, "g1abc123", "staging", "https://rpc.staging.gno.land:443", "./unsigned_call.tx", "alice", "")

	output := errBuf.String()
	assert.Contains(t, output, "Air-Gap Signing Workflow")
	assert.Contains(t, output, "gnokey query auth/accounts/g1abc123")
	assert.Contains(t, output, "-remote https://rpc.staging.gno.land:443")
	assert.Contains(t, output, "gnokey sign -tx-path ./unsigned_call.tx")
	assert.Contains(t, output, "-chainid staging")
	assert.Contains(t, output, "alice")
	assert.Contains(t, output, "gnokey broadcast")
}

func TestPrintAirGapHintsWithGnowebURL(t *testing.T) {
	t.Parallel()

	io, _, errBuf := newWizardIO("")

	printAirGapHints(io, "g1abc123", "dev", "127.0.0.1:26657", "./unsigned_call.tx", "alice", "http://127.0.0.1:8888/r/demo/foo$help&func=Vote")

	output := errBuf.String()
	assert.Contains(t, output, "View after broadcast:")
	assert.Contains(t, output, "http://127.0.0.1:8888/r/demo/foo$help&func=Vote")
}

func TestPrintAirGapHintsWithoutGnowebURL(t *testing.T) {
	t.Parallel()

	io, _, errBuf := newWizardIO("")

	printAirGapHints(io, "g1abc123", "dev", "127.0.0.1:26657", "./unsigned_call.tx", "alice", "")

	output := errBuf.String()
	assert.NotContains(t, output, "View after broadcast:")
}

// ─── Unit Tests: printSummary ───

func TestPrintSummary_Call(t *testing.T) {
	t.Parallel()

	io, _, errBuf := newWizardIO("")

	summary := txSummary{
		Type:       "call",
		KeyName:    "alice",
		KeyAddr:    "g1abc123",
		PkgPath:    "gno.land/r/demo/foo",
		FuncName:   "Vote",
		Args:       []string{"1", `"yes"`},
		GasWanted:  60000,
		GasFee:     "1ugnot",
		GasAutoEst: true,
		GasEstUsed: 50000,
		ChainID:    "dev",
		Remote:     "127.0.0.1:26657",
		GnowebURL:  "http://127.0.0.1:8888/r/demo/foo$help&func=Vote",
	}

	printSummary(io, summary)

	output := errBuf.String()
	assert.Contains(t, output, "Transaction Summary")
	assert.Contains(t, output, "Type:        call")
	assert.Contains(t, output, "Key:         alice (g1abc123)")
	assert.Contains(t, output, "Package:     gno.land/r/demo/foo")
	assert.Contains(t, output, "Function:    Vote")
	assert.Contains(t, output, "Arguments:   1, \"yes\"")
	assert.Contains(t, output, "Gas Wanted:  60000 (auto-estimated: 50000 used × 1.2)")
	assert.Contains(t, output, "Gas Fee:     1ugnot")
	assert.Contains(t, output, "View on gno.land:")
	assert.Contains(t, output, "http://127.0.0.1:8888/r/demo/foo$help&func=Vote")
}

// ─── Integration Tests: Hybrid (Partial Flags) ───

func TestWizard_Hybrid_Call_AllFlagsProvided(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()
	createTestKey(t, kbHome, "test1", "test1234")

	io, outBuf, errBuf := newWizardIO("")

	cfg := &MakeCallCfg{
		RootCfg: &client.MakeTxCfg{
			RootCfg: &client.BaseCfg{
				BaseOptions: client.BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			GasWanted: 2000000,
			GasFee:    "1000000ugnot",
			Broadcast: false,
			ChainID:   "dev",
		},
		PkgPath:  "gno.land/r/demo/foo",
		FuncName: "Bar",
	}

	err := execMakeCall(cfg, []string{"test1"}, io)
	require.NoError(t, err)

	output := outBuf.String()
	assert.Contains(t, output, `"msg"`)
	assert.Contains(t, output, `"fee"`)
	assert.NotContains(t, errBuf.String(), "Transaction Summary")
}

// ─── Integration Tests: --no-interactive ───

func TestWizard_NoInteractive_ErrorsOnMissingFields(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()
	createTestKey(t, kbHome, "test1", "test1234")

	io, _, _ := newWizardIO("")

	cfg := &MakeCallCfg{
		RootCfg: &client.MakeTxCfg{
			RootCfg: &client.BaseCfg{
				BaseOptions: client.BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			NoInteractive: true,
			Broadcast:     false,
			ChainID:       "dev",
		},
		PkgPath: "gno.land/r/demo/foo",
	}

	err := execMakeCall(cfg, []string{"test1"}, io)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "func not specified")
}

// ─── Integration Tests: Broadcast=false non-TTY ───

func TestWizard_NonTTY_BroadcastFalse_NoAirGapHints(t *testing.T) {
	t.Parallel()

	kbHome := t.TempDir()
	createTestKey(t, kbHome, "test1", "test1234")

	io, outBuf, errBuf := newWizardIO("")

	cfg := &MakeCallCfg{
		RootCfg: &client.MakeTxCfg{
			RootCfg: &client.BaseCfg{
				BaseOptions: client.BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			GasWanted: 2000000,
			GasFee:    "1000000ugnot",
			Broadcast: false,
			ChainID:   "dev",
		},
		PkgPath:  "gno.land/r/demo/foo",
		FuncName: "Bar",
	}

	err := execMakeCall(cfg, []string{"test1"}, io)
	require.NoError(t, err)

	output := outBuf.String()
	assert.Contains(t, output, `"msg"`)
	assert.NotContains(t, errBuf.String(), "Air-Gap Signing Workflow")
}

// ─── Network Selection ───

func TestPromptNetwork_ChainID(t *testing.T) {
	t.Parallel()

	io, _, _ := newWizardIO("2\n")

	chainID, remote, err := promptNetwork(io, "", "")
	require.NoError(t, err)
	assert.Equal(t, "staging", chainID)
	assert.Equal(t, "https://rpc.staging.gno.land:443", remote)
}

func TestPromptNetwork_RemoteMismatch(t *testing.T) {
	t.Parallel()

	// items: [keep, dev, staging, gnoland1, test11, manual]. "3" = staging.
	io, _, errBuf := newWizardIO("3\n")

	chainID, remote, err := promptNetwork(io, "dev", "https://rpc.gno.land:443")
	require.NoError(t, err)
	assert.Equal(t, "staging", chainID)
	assert.Equal(t, "https://rpc.staging.gno.land:443", remote)

	output := errBuf.String()
	assert.Contains(t, output, "different known networks")
}

func TestPromptNetwork_ManualEntry(t *testing.T) {
	t.Parallel()

	io, _, _ := newWizardIO("5\nmy-custom-chain\nhttp://my-node:26657\n")

	chainID, remote, err := promptNetwork(io, "", "")
	require.NoError(t, err)
	assert.Equal(t, "my-custom-chain", chainID)
	assert.Equal(t, "http://my-node:26657", remote)
}

func TestPromptNetwork_DefaultDev(t *testing.T) {
	t.Parallel()

	io, _, _ := newWizardIO("\n")

	chainID, remote, err := promptNetwork(io, "", "")
	require.NoError(t, err)
	assert.Equal(t, "dev", chainID)
	assert.Equal(t, "127.0.0.1:26657", remote)
}

// ─── Verify JSON output structure ───

func TestCallTxJSON(t *testing.T) {
	t.Parallel()

	caller := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	msg := vm.MsgCall{
		Caller:  caller,
		PkgPath: "gno.land/r/demo/foo",
		Func:    "Bar",
		Args:    []string{"1", `"hello"`},
	}

	gasfee, _ := std.ParseCoin("1000000ugnot")
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(2000000, gasfee),
		Signatures: nil,
		Memo:       "",
	}

	jsonBz := amino.MustMarshalJSON(tx)
	assert.Contains(t, string(jsonBz), `"msg"`)
	assert.Contains(t, string(jsonBz), `"fee"`)
	assert.Contains(t, string(jsonBz), `gno.land/r/demo/foo`)
	assert.Contains(t, string(jsonBz), `Bar`)
}
