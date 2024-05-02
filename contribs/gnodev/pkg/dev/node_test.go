package dev

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	mock "github.com/gnolang/gno/contribs/gnodev/internal/mock"

	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// XXX: We should probably use txtar to test this package.

var nodeTestingAddress = crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

// TestNewNode_NoPackages tests the NewDevNode method with no package.
func TestNewNode_NoPackages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewTestingLogger(t)

	// Call NewDevNode with no package should work
	cfg := DefaultNodeConfig(gnoenv.RootDir())
	node, err := NewDevNode(ctx, logger, &emitter.NoopServer{}, cfg)
	require.NoError(t, err)

	assert.Len(t, node.ListPkgs(), 0)

	require.NoError(t, node.Close())
}

// TestNewNode_WithPackage tests the NewDevNode with a single package.
func TestNewNode_WithPackage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const (
		// foobar package
		testGnoMod = "module gno.land/r/dev/foobar\n"
		testFile   = `package foobar
func Render(_ string) string { return "foo" }
`
	)

	// Generate package
	pkgpath := generateTestingPackage(t, "gno.mod", testGnoMod, "foobar.gno", testFile)
	logger := log.NewTestingLogger(t)

	// Call NewDevNode with no package should work
	cfg := DefaultNodeConfig(gnoenv.RootDir())
	cfg.PackagesPathList = []PackagePath{pkgpath}
	node, err := NewDevNode(ctx, logger, &emitter.NoopServer{}, cfg)
	require.NoError(t, err)
	assert.Len(t, node.ListPkgs(), 1)

	// Test rendering
	render, err := testingRenderRealm(t, node, "gno.land/r/dev/foobar")
	require.NoError(t, err)
	assert.Equal(t, render, "foo")

	require.NoError(t, node.Close())
}

func TestNodeAddPackage(t *testing.T) {
	// Setup a Node instance
	const (
		// foo package
		fooGnoMod = "module gno.land/r/dev/foo\n"
		fooFile   = `package foo
func Render(_ string) string { return "foo" }
`
		// bar package
		barGnoMod = "module gno.land/r/dev/bar\n"
		barFile   = `package bar
func Render(_ string) string { return "bar" }
`
	)

	// Generate package foo
	foopkg := generateTestingPackage(t, "gno.mod", fooGnoMod, "foo.gno", fooFile)

	// Call NewDevNode with no package should work
	node, emitter := newTestingDevNode(t, foopkg)
	assert.Len(t, node.ListPkgs(), 1)

	// Test render
	render, err := testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)
	require.Equal(t, render, "foo")

	// Generate package bar
	barpkg := generateTestingPackage(t, "gno.mod", barGnoMod, "bar.gno", barFile)
	err = node.UpdatePackages(barpkg.Path)
	require.NoError(t, err)
	assert.Len(t, node.ListPkgs(), 2)

	// Render should fail as the node hasn't reloaded
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/bar")
	require.Error(t, err)

	err = node.Reload(context.Background())
	require.NoError(t, err)
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtReload)

	// After a reload, render should succeed
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/bar")
	require.NoError(t, err)
	require.Equal(t, render, "bar")
}

func TestNodeUpdatePackage(t *testing.T) {
	// Setup a Node instance
	const (
		// foo package
		foobarGnoMod = "module gno.land/r/dev/foobar\n"
		fooFile      = `package foobar
func Render(_ string) string { return "foo" }
`
		barFile = `package foobar
func Render(_ string) string { return "bar" }
`
	)

	// Generate package foo
	foopkg := generateTestingPackage(t, "gno.mod", foobarGnoMod, "foo.gno", fooFile)

	// Call NewDevNode with no package should work
	node, emitter := newTestingDevNode(t, foopkg)
	assert.Len(t, node.ListPkgs(), 1)

	// Test that render is correct
	render, err := testingRenderRealm(t, node, "gno.land/r/dev/foobar")
	require.NoError(t, err)
	require.Equal(t, render, "foo")

	// Override `foo.gno` file with bar content
	err = os.WriteFile(filepath.Join(foopkg.Path, "foo.gno"), []byte(barFile), 0o700)
	require.NoError(t, err)

	err = node.Reload(context.Background())
	require.NoError(t, err)

	// Check reload event
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtReload)

	// After a reload, render should succeed
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/foobar")
	require.NoError(t, err)
	require.Equal(t, render, "bar")

	assert.Nil(t, emitter.NextEvent())
}

func TestNodeReset(t *testing.T) {
	const (
		// foo package
		foobarGnoMod = "module gno.land/r/dev/foo\n"
		fooFile      = `package foo
var str string = "foo"
func UpdateStr(newStr string) { str = newStr } // method to update 'str' variable
func Render(_ string) string { return str }
`
	)

	// Generate package foo
	foopkg := generateTestingPackage(t, "gno.mod", foobarGnoMod, "foo.gno", fooFile)

	// Call NewDevNode with no package should work
	node, emitter := newTestingDevNode(t, foopkg)
	assert.Len(t, node.ListPkgs(), 1)

	// Test rendering
	render, err := testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)
	require.Equal(t, render, "foo")

	// Call `UpdateStr` to update `str` value with "bar"
	msg := gnoclient.MsgCall{
		PkgPath:  "gno.land/r/dev/foo",
		FuncName: "UpdateStr",
		Args:     []string{"bar"},
		Send:     "",
	}
	res, err := testingCallRealm(t, node, msg)
	require.NoError(t, err)
	require.NoError(t, res.CheckTx.Error)
	require.NoError(t, res.DeliverTx.Error)
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtTxResult)

	// Check for correct render update
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)
	require.Equal(t, render, "bar")

	// Reset state
	err = node.Reset(context.Background())
	require.NoError(t, err)
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtReset)

	// Test rendering should return initial `str` value
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)
	require.Equal(t, render, "foo")

	assert.Nil(t, emitter.NextEvent())
}

func testingRenderRealm(t *testing.T, node *Node, rlmpath string) (string, error) {
	t.Helper()

	signer := newInMemorySigner(t, node.Config().ChainID())
	cli := gnoclient.Client{
		Signer:    signer,
		RPCClient: node.Client(),
	}

	render, res, err := cli.Render(rlmpath, "")
	if err == nil {
		err = res.Response.Error
	}

	return render, err
}

func testingCallRealm(t *testing.T, node *Node, msgs ...gnoclient.MsgCall) (*core_types.ResultBroadcastTxCommit, error) {
	t.Helper()

	signer := newInMemorySigner(t, node.Config().ChainID())
	cli := gnoclient.Client{
		Signer:    signer,
		RPCClient: node.Client(),
	}

	txcfg := gnoclient.BaseTxCfg{
		GasFee:    "1000000ugnot", // Gas fee
		GasWanted: 2_000_000,      // Gas wanted
	}

	return cli.Call(txcfg, msgs...)
}

func generateTestingPackage(t *testing.T, nameFile ...string) PackagePath {
	t.Helper()
	workdir := t.TempDir()

	if len(nameFile)%2 != 0 {
		require.FailNow(t, "Generate testing packages require paired arguments.")
	}

	for i := 0; i < len(nameFile); i += 2 {
		name := nameFile[i]
		content := nameFile[i+1]

		err := os.WriteFile(filepath.Join(workdir, name), []byte(content), 0o700)
		require.NoError(t, err)
	}

	return PackagePath{
		Path:    workdir,
		Creator: nodeTestingAddress,
	}
}

func newTestingDevNode(t *testing.T, pkgslist ...PackagePath) (*Node, *mock.ServerEmitter) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewTestingLogger(t)

	emitter := &mock.ServerEmitter{}

	// Call NewDevNode with no package should work
	cfg := DefaultNodeConfig(gnoenv.RootDir())
	cfg.PackagesPathList = pkgslist
	node, err := NewDevNode(ctx, logger, emitter, cfg)
	require.NoError(t, err)
	assert.Len(t, node.ListPkgs(), len(pkgslist))

	t.Cleanup(func() { node.Close() })

	return node, emitter
}

func newInMemorySigner(t *testing.T, chainid string) *gnoclient.SignerFromKeybase {
	t.Helper()

	mnemonic := integration.DefaultAccount_Seed
	name := integration.DefaultAccount_Name

	kb := keys.NewInMemory()
	_, err := kb.CreateAccount(name, mnemonic, "", "", uint32(0), uint32(0))
	require.NoError(t, err)

	return &gnoclient.SignerFromKeybase{
		Keybase:  kb,      // Stores keys in memory
		Account:  name,    // Account name
		Password: "",      // Password for encryption
		ChainID:  chainid, // Chain ID for transaction signing
	}
}
