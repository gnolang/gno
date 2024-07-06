package dev

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	mock "github.com/gnolang/gno/contribs/gnodev/internal/mock"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	tm2events "github.com/gnolang/gno/tm2/pkg/events"
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
	cfg.Logger = logger
	node, err := NewDevNode(ctx, cfg)
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
	cfg.PackagesModifier = []PackagePath{pkgpath}
	cfg.Logger = logger
	node, err := NewDevNode(ctx, cfg)
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
	assert.Equal(t, events.EvtReload, emitter.NextEvent().Type())

	// After a reload, render should succeed
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/foobar")
	require.NoError(t, err)
	require.Equal(t, render, "bar")

	assert.Equal(t, mock.EvtNull, emitter.NextEvent().Type())
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
	msg := vm.MsgCall{
		PkgPath: "gno.land/r/dev/foo",
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

	assert.Equal(t, mock.EvtNull, emitter.NextEvent().Type())
}

func TestTxTimestampRecover(t *testing.T) {
	const (
		// foo package
		foobarGnoMod = "module gno.land/r/dev/foo\n"
		fooFile      = `package foo
import (
	"strconv"
	"strings"
	"time"
)

var times = []time.Time{
	time.Now(), // Evaluate at genesis
}

func SpanTime() {
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
	)

	// Add a hard deadline of 20 seconds to avoid potential deadlock and fail early
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

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

	// Generate package foo
	foopkg := generateTestingPackage(t, "gno.mod", foobarGnoMod, "foo.gno", fooFile)

	// Call NewDevNode with no package should work
	cfg := createDefaultTestingNodeConfig(foopkg)

	// XXX(gfanton): Setting this to `false` somehow makes the time block
	// drift from the time spanned by the VM.
	cfg.TMConfig.Consensus.SkipTimeoutCommit = false
	cfg.TMConfig.Consensus.TimeoutCommit = 500 * time.Millisecond
	cfg.TMConfig.Consensus.TimeoutPropose = 100 * time.Millisecond
	cfg.TMConfig.Consensus.CreateEmptyBlocks = true

	node, emitter := newTestingDevNodeWithConfig(t, cfg)

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
	for i := 0; i < nevents; i++ {
		t.Logf("waiting for a bock greater than height(%d) and unix(%d)", refHeight, refTimestamp)
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
			PkgPath: "gno.land/r/dev/foo",
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
	render, err := testingRenderRealm(t, node, "gno.land/r/dev/foo")
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
	render, err = testingRenderRealm(t, node, "gno.land/r/dev/foo")
	require.NoError(t, err)

	timesList2 := parseJSONTimesList(t, render)

	// Times list should be identical from the orignal list
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

	signer := newInMemorySigner(t, node.Config().ChainID())
	cli := gnoclient.Client{
		Signer:    signer,
		RPCClient: node.Client(),
	}

	txcfg := gnoclient.BaseTxCfg{
		GasFee:    ugnot.ValueString(1000000), // Gas fee
		GasWanted: 2_000_000,                  // Gas wanted
	}

	// Set Caller in the msgs
	caller, err := signer.Info()
	require.NoError(t, err)
	vmMsgs := make([]vm.MsgCall, 0, len(msgs))
	for _, msg := range msgs {
		vmMsgs = append(vmMsgs, vm.NewMsgCall(caller.GetAddress(), msg.Send, msg.PkgPath, msg.Func, msg.Args))
	}

	return cli.Call(txcfg, vmMsgs...)
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

func createDefaultTestingNodeConfig(pkgslist ...PackagePath) *NodeConfig {
	cfg := DefaultNodeConfig(gnoenv.RootDir())
	cfg.PackagesModifier = pkgslist
	return cfg
}

func newTestingDevNode(t *testing.T, pkgslist ...PackagePath) (*Node, *mock.ServerEmitter) {
	t.Helper()

	cfg := createDefaultTestingNodeConfig(pkgslist...)
	return newTestingDevNodeWithConfig(t, cfg)
}

func newTestingDevNodeWithConfig(t *testing.T, cfg *NodeConfig) (*Node, *mock.ServerEmitter) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	logger := log.NewTestingLogger(t)
	emitter := &mock.ServerEmitter{}

	cfg.Emitter = emitter
	cfg.Logger = logger

	node, err := NewDevNode(ctx, cfg)
	require.NoError(t, err)
	assert.Len(t, node.ListPkgs(), len(cfg.PackagesModifier))

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
