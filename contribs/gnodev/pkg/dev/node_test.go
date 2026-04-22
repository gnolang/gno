package dev

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mock "github.com/gnolang/gno/contribs/gnodev/internal/mock/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	tm2events "github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestingLoader builds a packages.Loader backed by an in-memory fetcher
// seeded with the given MemPackages. Resolving any of those paths yields a
// *packages.Package whose ToMemPackage returns the embedded mempkg.
func newTestingLoader(mps ...*std.MemPackage) *packages.Loader {
	return packages.New(packages.Config{
		Fetcher: pkgdownload.NewInMemoryFetcher(mps...),
		Logger:  log.NewNoopLogger().With("group", "loader"),
	})
}

// pkgsFromMem constructs []*packages.Package from in-memory MemPackages for tests.
// Uses pkgdownload.InMemoryFetcher so Resolve() produces packages with an
// embedded MemPackage, avoiding filesystem writes.
func pkgsFromMem(t *testing.T, mps ...*std.MemPackage) []*packages.Package {
	t.Helper()
	l := newTestingLoader(mps...)
	out := make([]*packages.Package, 0, len(mps))
	for _, mp := range mps {
		p, err := l.Resolve(mp.Path)
		require.NoError(t, err)
		out = append(out, p)
	}
	return out
}

// resolvePaths resolves each path using the given loader, skipping
// paths that fail to resolve. Used by the testing Reload closure so the
// node's loaded package set follows its .Paths() list.
func resolvePaths(l *packages.Loader, paths []string) []*packages.Package {
	out := make([]*packages.Package, 0, len(paths))
	for _, p := range paths {
		pkg, err := l.Resolve(p)
		if err != nil {
			continue
		}
		out = append(out, pkg)
	}
	return out
}

// TestNewNode_NoPackages tests the NewDevNode method with no package.
func TestNewNode_NoPackages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.NewTestingLogger(t)

	// Call NewDevNode with no package should work
	cfg := DefaultNodeConfig(gnoenv.RootDir(), "gno.land")
	cfg.Logger = logger
	cfg.Reload = func() ([]*packages.Package, error) { return nil, nil }
	node, err := NewDevNode(ctx, cfg)
	require.NoError(t, err)

	assert.Len(t, node.ListPkgs(), 0)
	assert.Len(t, node.Paths(), 0)

	require.NoError(t, node.Close())
}

// TestNewNode_WithPackage tests the NewDevNode with a single package.
func TestNewNode_WithLoader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pkg := std.MemPackage{
		Name: "foobar",
		Path: "gno.land/r/dev/foobar",
		Files: []*std.MemFile{
			{
				Name: "foobar.gno",
				Body: `package foobar
func Render(_ string) string { return "foo" }
`,
			},
		},
	}

	pkg.SetFile("gnomod.toml", gnolang.GenGnoModLatest(pkg.Path))
	pkg.Sort()

	logger := log.NewTestingLogger(t)

	cfg := DefaultNodeConfig(gnoenv.RootDir(), "gno.land")
	initialPkgs := pkgsFromMem(t, &pkg)
	cfg.Reload = func() ([]*packages.Package, error) { return initialPkgs, nil }
	cfg.Logger = logger

	node, err := NewDevNode(ctx, cfg, pkg.Path)
	require.NoError(t, err)
	assert.Len(t, node.ListPkgs(), 1)
	assert.Len(t, node.Paths(), 1)

	// Test rendering
	render, err := testingRenderRealm(t, node, pkg.Path)
	require.NoError(t, err)
	assert.Equal(t, render, "foo")

	require.NoError(t, node.Close())
}

func TestNodeAddPackage(t *testing.T) {
	// Setup a Node instance
	fooPkg := std.MemPackage{
		Name: "foo",
		Path: "gno.land/r/dev/foo",
		Files: []*std.MemFile{
			{
				Name: "foo.gno",
				Body: `package foo
func Render(_ string) string { return "foo" }
`,
			},
			{
				Name: "gnomod.toml",
				Body: gnolang.GenGnoModLatest("gno.land/r/dev/foo"),
			},
		},
	}

	barPkg := std.MemPackage{
		Name: "bar",
		Path: "gno.land/r/dev/bar",
		Files: []*std.MemFile{
			{
				Name: "bar.gno",
				Body: `package bar
func Render(_ string) string { return "bar" }
`,
			},
			{
				Name: "gnomod.toml",
				Body: gnolang.GenGnoModLatest("gno.land/r/dev/bar"),
			},
		},
	}

	// Generate package foo
	cfg, holder := newTestingNodeConfig(t, &fooPkg, &barPkg)

	// Call NewDevNode with no package should work
	node, emitter := newTestingDevNodeWithConfigAndHolder(t, cfg, holder, fooPkg.Path)
	assert.Len(t, node.ListPkgs(), 1)
	assert.Len(t, node.Paths(), 1)

	// Test render
	render, err := testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)
	require.Equal(t, render, "foo")

	// Render should fail as the node hasn't reloaded
	_, err = testingRenderRealm(t, node, "gno.land/r/dev/bar")
	require.Error(t, err)

	// Add bar package
	node.AddPackagePaths(barPkg.Path)

	err = node.Reload(context.Background())
	require.NoError(t, err)
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtReload)

	// After a reload, render should succeed
	render, err = testingRenderRealm(t, node, barPkg.Path)
	require.NoError(t, err)
	require.Equal(t, render, "bar")
}

func TestNodeUpdatePackage(t *testing.T) {
	foorbarPkg := std.MemPackage{
		Name: "foobar",
		Path: "gno.land/r/dev/foobar",
	}
	modfile := &std.MemFile{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(foorbarPkg.Path)}

	fooFiles := []*std.MemFile{
		{
			Name: "foo.gno",
			Body: `package foobar
func Render(_ string) string { return "foo" }
`,
		},
	}

	barFiles := []*std.MemFile{
		{
			Name: "bar.gno",
			Body: `package foobar
func Render(_ string) string { return "bar" }
`,
		},
	}

	// Update foobar content with bar content
	foorbarPkg.Files = append(fooFiles, modfile)
	foorbarPkg.Sort()

	node, emitter := newTestingDevNode(t, &foorbarPkg)
	assert.Len(t, node.ListPkgs(), 1)
	assert.Len(t, node.Paths(), 1)

	// Test that render is correct
	render, err := testingRenderRealm(t, node, foorbarPkg.Path)
	require.NoError(t, err)
	require.Equal(t, render, "foo")

	// Update foobar content with bar content
	foorbarPkg.Files = append(barFiles, modfile)
	foorbarPkg.Sort()

	err = node.Reload(context.Background())
	require.NoError(t, err)

	// Check reload event
	assert.Equal(t, events.EvtReload, emitter.NextEvent().Type())

	// After a reload, render should succeed
	render, err = testingRenderRealm(t, node, foorbarPkg.Path)
	require.NoError(t, err)
	require.Equal(t, render, "bar")

	assert.Equal(t, mock.EvtNull, emitter.NextEvent().Type())
}

func TestNodeReset(t *testing.T) {
	fooPkg := std.MemPackage{
		Name: "foo",
		Path: "gno.land/r/dev/foo",
		Files: []*std.MemFile{
			{
				Name: "foo.gno",
				Body: `package foo
var str string = "foo"

func UpdateStr(cur realm, newStr string) { // method to update 'str' variable
        str = newStr
}

func Render(_ string) string { return str }
`,
			},
			{
				Name: "gnomod.toml",
				Body: gnolang.GenGnoModLatest("gno.land/r/dev/foo"),
			},
		},
	}

	node, emitter := newTestingDevNode(t, &fooPkg)
	assert.Len(t, node.ListPkgs(), 1)
	assert.Len(t, node.Paths(), 1)

	// Test rendering
	render, err := testingRenderRealm(t, node, fooPkg.Path)
	require.NoError(t, err)
	require.Equal(t, render, "foo")

	// Call `UpdateStr` to update `str` value with "bar"
	msg := vm.MsgCall{
		PkgPath: fooPkg.Path,
		Func:    "UpdateStr",
		Args:    []string{"bar"},
		Send:    nil,
	}
	res, err := testingCallRealm(t, node, msg)
	require.NoError(t, err)
	require.NoError(t, res.CheckTx.Error)
	require.NoError(t, res.DeliverTx.Error)
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtTxResult)

	// Check for correct render update
	render, err = testingRenderRealm(t, node, fooPkg.Path)
	require.NoError(t, err)
	require.Equal(t, render, "bar")

	// Reset state
	err = node.Reset(context.Background())
	require.NoError(t, err)
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtReset)

	// Test rendering should return initial `str` value
	render, err = testingRenderRealm(t, node, fooPkg.Path)
	require.NoError(t, err)
	require.Equal(t, render, "foo")

	assert.Equal(t, mock.EvtNull, emitter.NextEvent().Type())
}

func TestTxGasFailure(t *testing.T) {
	fooPkg := std.MemPackage{
		Name: "foo",
		Path: "gno.land/r/dev/foo",
		Files: []*std.MemFile{
			{
				Name: "foo.gno",
				Body: `package foo
import "strconv"

var i int

func Inc(cur realm) {  // method to increment i
        i++
}

func Render(_ string) string { return strconv.Itoa(i) }
`,
			},
			{
				Name: "gnomod.toml",
				Body: gnolang.GenGnoModLatest("gno.land/r/dev/foo"),
			},
		},
	}

	node, emitter := newTestingDevNode(t, &fooPkg)
	assert.Len(t, node.ListPkgs(), 1)
	assert.Len(t, node.Paths(), 1)

	// Test rendering
	render, err := testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)
	require.Equal(t, "0", render)

	// Call `Inc` to update counter
	msg := vm.MsgCall{
		PkgPath: "gno.land/r/dev/foo",
		Func:    "Inc",
		Args:    nil,
		Send:    nil,
	}

	res, err := testingCallRealm(t, node, msg)
	require.NoError(t, err)
	require.NoError(t, res.CheckTx.Error)
	require.NoError(t, res.DeliverTx.Error)
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtTxResult)

	// Check for correct render update
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)
	require.Equal(t, "1", render)

	// Not Enough gas wanted
	callCfg := gnoclient.BaseTxCfg{
		GasFee: ugnot.ValueString(10000), // Gas fee

		// Ensure sufficient gas is provided for the transaction to be committed.
		// However, avoid providing too much gas to allow the
		// transaction to succeed (OutOfGasError).
		GasWanted: 1_000_000,
	}

	_, err = testingCallRealmWithConfig(t, node, callCfg, msg)
	require.Error(t, err)
	require.ErrorAs(t, err, &std.OutOfGasError{})

	// Transaction should be committed regardless the error
	require.Equal(t, emitter.NextEvent().Type(), events.EvtTxResult,
		"(probably) not enough gas for the transaction to be committed")

	// Reload the node
	err = node.Reload(context.Background())
	require.NoError(t, err)
	assert.Equal(t, events.EvtReload, emitter.NextEvent().Type())

	// Check for correct render update
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)

	// Assert that the previous transaction hasn't succeeded during genesis reload
	require.Equal(t, "1", render)
}

func TestTxTimestampRecover(t *testing.T) {
	const fooFile = `
package foo

import (
	"strconv"
	"strings"
	"time"
)

var times = []time.Time{
	time.Now(), // Evaluate at genesis
}

func SpanTime(cur realm) {
	times = append(times, time.Now())
}

func Render(_ string) string {
	var strs strings.Builder

	strs.WriteRune('[')
	for i, t := range times {
		if i > 0 {
			strs.WriteRune(',')
		}
		strs.WriteString(strconv.Itoa(int(t.UnixNano())))
	}
	strs.WriteRune(']')

	return strs.String()
}
`
	fooPkg := std.MemPackage{
		Name: "foo",
		Path: "gno.land/r/dev/foo",
		Files: []*std.MemFile{
			{
				Name: "foo.gno",
				Body: fooFile,
			},
		},
	}

	// Add a hard deadline of 20 seconds to avoid potential deadlock and fail early
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// XXX(gfanton): Setting this to `false` somehow makes the time block
	// drift from the time spanned by the VM.
	cfg, holder := newTestingNodeConfig(t, &fooPkg)
	cfg.TMConfig.Consensus.SkipTimeoutCommit = false
	cfg.TMConfig.Consensus.TimeoutCommit = 500 * time.Millisecond
	cfg.TMConfig.Consensus.TimeoutPropose = 100 * time.Millisecond
	cfg.TMConfig.Consensus.CreateEmptyBlocks = true

	node, emitter := newTestingDevNodeWithConfigAndHolder(t, cfg, holder, fooPkg.Path)

	render, err := testingRenderRealm(t, node, fooPkg.Path)
	require.NoError(t, err)
	require.NotEmpty(t, render)

	parseJSONTimesList := func(t *testing.T, render string) []time.Time {
		t.Helper()

		var times []time.Time
		var nanos []int64

		err := json.Unmarshal([]byte(render), &nanos)
		require.NoError(t, err)

		for _, nano := range nanos {
			sec, nsec := nano/int64(time.Second), nano%int64(time.Second)
			times = append(times, time.Unix(sec, nsec))
		}

		return times
	}

	// We need to make sure that blocks are separated by at least 1 second
	// (minimal time between blocks). We can ensure this by listening for
	// new blocks and comparing timestamps
	cc := make(chan types.EventNewBlock)
	node.Node.EventSwitch().AddListener("test-timestamp", func(evt tm2events.Event) {
		newBlock, ok := evt.(types.EventNewBlock)
		if !ok {
			return
		}

		select {
		case cc <- newBlock:
		default:
		}
	})

	// wait for first block for reference
	var refHeight, refTimestamp int64

	select {
	case <-ctx.Done():
		require.FailNow(t, ctx.Err().Error())
	case res := <-cc:
		refTimestamp = res.Block.Time.Unix()
		refHeight = res.Block.Height
	}

	// number of span to process
	const nevents = 3

	// Span multiple time
	for range nevents {
		t.Logf("waiting for a block greater than height(%d) and unix(%d)", refHeight, refTimestamp)
		for {
			var block types.EventNewBlock
			select {
			case <-ctx.Done():
				require.FailNow(t, ctx.Err().Error())
			case block = <-cc:
			}

			t.Logf("got a block height(%d) and unix(%d)",
				block.Block.Height, block.Block.Time.Unix())

			// Ensure we consume every block before tx block
			if refHeight >= block.Block.Height {
				continue
			}

			// Ensure new block timestamp is before previous reference timestamp
			if newRefTimestamp := block.Block.Time.Unix(); newRefTimestamp > refTimestamp {
				refTimestamp = newRefTimestamp
				break // break the loop
			}
		}

		t.Logf("found a valid block(%d)! continue", refHeight)

		// Span a new time
		msg := vm.MsgCall{
			PkgPath: fooPkg.Path,
			Func:    "SpanTime",
		}

		res, err := testingCallRealm(t, node, msg)

		require.NoError(t, err)
		require.NoError(t, res.CheckTx.Error)
		require.NoError(t, res.DeliverTx.Error)
		assert.Equal(t, emitter.NextEvent().Type(), events.EvtTxResult)

		// Set the new height from the tx as reference
		refHeight = res.Height
	}

	// Render JSON times list
	render, err = testingRenderRealm(t, node, fooPkg.Path)
	require.NoError(t, err)

	// Parse times list
	timesList1 := parseJSONTimesList(t, render)
	t.Logf("list of times: %+v", timesList1)

	// Ensure times are correctly expending.
	for i, t2 := range timesList1 {
		if i == 0 {
			continue
		}

		t1 := timesList1[i-1]
		require.Greater(t, t2.UnixNano(), t1.UnixNano())
	}

	// Reload the node
	err = node.Reload(context.Background())
	require.NoError(t, err)
	assert.Equal(t, emitter.NextEvent().Type(), events.EvtReload)

	// Fetch time list again from render
	render, err = testingRenderRealm(t, node, fooPkg.Path)
	require.NoError(t, err)

	timesList2 := parseJSONTimesList(t, render)

	// Times list should be identical from the original list
	require.Len(t, timesList2, len(timesList1))
	for i := 0; i < len(timesList1); i++ {
		t1nsec, t2nsec := timesList1[i].UnixNano(), timesList2[i].UnixNano()
		assert.Equal(t, t1nsec, t2nsec,
			"comparing times1[%d](%d) == times2[%d](%d)", i, t1nsec, i, t2nsec)
	}
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

func testingCallRealm(t *testing.T, node *Node, msgs ...vm.MsgCall) (*core_types.ResultBroadcastTxCommit, error) {
	t.Helper()

	defaultCfg := gnoclient.BaseTxCfg{
		GasFee:    ugnot.ValueString(1000000), // Gas fee
		GasWanted: 3_000_000,                  // Gas wanted
	}

	return testingCallRealmWithConfig(t, node, defaultCfg, msgs...)
}

func testingCallRealmWithConfig(t *testing.T, node *Node, bcfg gnoclient.BaseTxCfg, msgs ...vm.MsgCall) (*core_types.ResultBroadcastTxCommit, error) {
	t.Helper()

	signer := newInMemorySigner(t, node.Config().ChainID())
	cli := gnoclient.Client{
		Signer:    signer,
		RPCClient: node.Client(),
	}

	// Set Caller in the msgs
	caller, err := signer.Info()
	require.NoError(t, err)
	vmMsgs := make([]vm.MsgCall, 0, len(msgs))
	for _, msg := range msgs {
		vmMsgs = append(vmMsgs, vm.NewMsgCall(caller.GetAddress(), msg.Send, msg.PkgPath, msg.Func, msg.Args))
	}

	return cli.Call(bcfg, vmMsgs...)
}

// nodeHolder lets the Reload closure read live node state (its .Paths())
// once the *Node is wired in. Before the node is created, the holder's
// node pointer is nil and Reload falls back to the bootstrap paths.
type nodeHolder struct {
	node           *Node
	bootstrapPaths []string
}

func newTestingNodeConfig(t *testing.T, pkgs ...*std.MemPackage) (*NodeConfig, *nodeHolder) {
	t.Helper()
	gnoroot := gnoenv.RootDir()

	// Ensure that a gnomod.toml exists
	for _, pkg := range pkgs {
		if mod := pkg.GetFile("gnomod.toml"); mod != nil {
			continue
		}

		pkg.SetFile("gnomod.toml", gnolang.GenGnoModLatest(pkg.Path))
		pkg.Sort()
	}

	memPkgs := pkgs // captured by closure so in-place mutations of Files
	// (see TestNodeUpdatePackage) are observed on each Reload.
	holder := &nodeHolder{}

	cfg := DefaultNodeConfig(gnoenv.RootDir(), "gno.land")
	cfg.TMConfig = integration.DefaultTestingTMConfig(gnoroot)
	cfg.Reload = func() ([]*packages.Package, error) {
		// Reload is invoked while Node holds muNode as a write lock, so
		// we must NOT call node.Paths() (would deadlock). Read n.paths
		// directly; the node is only set after construction when the
		// initial reset has already completed.
		paths := holder.bootstrapPaths
		if holder.node != nil {
			paths = holder.node.paths
		}
		// Build a fresh Loader + InMemoryFetcher on every reload so that
		// in-place mutations of MemPackage.Files are picked up.
		l := newTestingLoader(memPkgs...)
		return resolvePaths(l, paths), nil
	}
	return cfg, holder
}

func newTestingDevNode(t *testing.T, pkgs ...*std.MemPackage) (*Node, *mock.ServerEmitter) {
	t.Helper()

	cfg, holder := newTestingNodeConfig(t, pkgs...)
	paths := make([]string, len(pkgs))
	for i, pkg := range pkgs {
		paths[i] = pkg.Path
	}

	return newTestingDevNodeWithConfigAndHolder(t, cfg, holder, paths...)
}

func newTestingDevNodeWithConfig(t *testing.T, cfg *NodeConfig, pkgpaths ...string) (*Node, *mock.ServerEmitter) {
	t.Helper()
	return newTestingDevNodeWithConfigAndHolder(t, cfg, nil, pkgpaths...)
}

func newTestingDevNodeWithConfigAndHolder(t *testing.T, cfg *NodeConfig, holder *nodeHolder, pkgpaths ...string) (*Node, *mock.ServerEmitter) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	logger := log.NewTestingLogger(t)
	emitter := &mock.ServerEmitter{}

	cfg.Emitter = emitter
	cfg.Logger = logger

	if holder != nil {
		holder.bootstrapPaths = append([]string(nil), pkgpaths...)
	}

	node, err := NewDevNode(ctx, cfg, pkgpaths...)
	require.NoError(t, err)
	require.Equal(t, emitter.NextEvent().Type(), events.EvtReset)

	if holder != nil {
		holder.node = node
	}

	t.Cleanup(func() {
		node.Close()
		cancel()
	})

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
