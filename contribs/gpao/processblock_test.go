package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	bftstate "github.com/gnolang/gno/tm2/pkg/bft/state"
	bfttypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// stubRPC serves one canned block and one canned result set. The embedded
// interface is nil on purpose: any method processBlock does not use will panic
// if it is ever called, which keeps the stub honest about its coverage.
type stubRPC struct {
	rpcclient.Client
	block      *ctypes.ResultBlock
	results    *ctypes.ResultBlockResults
	resultsErr error
}

func (s stubRPC) Block(context.Context, *int64) (*ctypes.ResultBlock, error) {
	return s.block, nil
}

func (s stubRPC) BlockResults(context.Context, *int64) (*ctypes.ResultBlockResults, error) {
	return s.results, s.resultsErr
}

func addPkgTx(t *testing.T, path string) bfttypes.Tx {
	t.Helper()
	raw, err := amino.Marshal(std.Tx{
		Msgs: []std.Msg{vm.MsgAddPackage{
			Creator: crypto.AddressFromPreimage([]byte("submitter")),
			Package: &std.MemPackage{Name: "p", Path: path},
		}},
	})
	require.NoError(t, err)
	return bfttypes.Tx(raw)
}

func blockWith(txs ...bfttypes.Tx) *ctypes.ResultBlock {
	return &ctypes.ResultBlock{
		Block: &bfttypes.Block{Data: bfttypes.Data{Txs: txs}},
	}
}

func resultsWith(oks ...bool) *ctypes.ResultBlockResults {
	rs := make([]abci.ResponseDeliverTx, 0, len(oks))
	for _, ok := range oks {
		r := abci.ResponseDeliverTx{}
		if !ok {
			r.ResponseBase.Error = abci.StringError("failed")
		}
		rs = append(rs, r)
	}
	return &ctypes.ResultBlockResults{Results: &bftstate.ABCIResponses{DeliverTxs: rs}}
}

var errStub = errors.New("rpc unavailable")

func newStubOracle(rpc rpcclient.Client) *oracle {
	return &oracle{
		io:         commands.NewTestIO(),
		client:     gnoclient.Client{RPCClient: rpc},
		candidates: make(chan *std.MemPackage, 8),
		seen:       map[string]struct{}{},
		overBudget: map[string]int{},
	}
}

// TestProcessBlockIgnoresFailedTransactions pins the fix for a budget bypass.
//
// A block carries every transaction that was proposed, including ones that
// failed in DeliverTx. Verifying a failed MsgAddPackage means timing bytes the
// chain discarded and then approving a path whose stored contents are something
// else -- so the enable the oracle signs becomes the vehicle for exactly the
// unmetered compile it exists to prevent.
func TestProcessBlockIgnoresFailedTransactions(t *testing.T) {
	good := addPkgTx(t, "gno.land/r/test/good")
	bad := addPkgTx(t, "gno.land/r/test/decoy")

	o := newStubOracle(stubRPC{
		block:   blockWith(good, bad),
		results: resultsWith(true, false),
	})

	require.NoError(t, o.processBlock(context.Background(), 1))

	var queued []string
	for len(o.candidates) > 0 {
		queued = append(queued, (<-o.candidates).Path)
	}
	require.Equal(t, []string{"gno.land/r/test/good"}, queued,
		"only the package from the transaction that SUCCEEDED may be queued")
}

// TestProcessBlockFailsClosedWithoutResults pins that an unverifiable block is
// not acted on at all, rather than verified blindly.
func TestProcessBlockFailsClosedWithoutResults(t *testing.T) {
	tx := addPkgTx(t, "gno.land/r/test/unknown")

	t.Run("results unavailable", func(t *testing.T) {
		o := newStubOracle(stubRPC{
			block:      blockWith(tx),
			resultsErr: errStub,
		})
		require.Error(t, o.processBlock(context.Background(), 1))
		require.Empty(t, o.candidates, "nothing may be queued from a block whose outcomes are unknown")
	})

	t.Run("results do not pair with transactions", func(t *testing.T) {
		o := newStubOracle(stubRPC{
			block:   blockWith(tx),
			results: resultsWith(), // zero results, one tx
		})
		require.Error(t, o.processBlock(context.Background(), 1))
		require.Empty(t, o.candidates)
	})
}

// TestChildEnvExcludesCredentials pins the invariant this daemon's comments
// claim: the process that compiles untrusted code does not hold the approver's
// key material. Before the allow-list the child inherited the parent's whole
// environment, so the claim was false as written even though nothing in the
// child read those variables.
func TestChildEnvExcludesCredentials(t *testing.T) {
	t.Setenv("GPAO_MNEMONIC", "some twelve word phrase that must not leak")
	t.Setenv("GPAO_PASSWORD", "hunter2")
	t.Setenv("HOME", "/home/someone")

	env := childEnv()

	joined := strings.Join(env, "\n")
	require.NotContains(t, joined, "GPAO_MNEMONIC",
		"the verifier must never be handed the approver's mnemonic")
	require.NotContains(t, joined, "GPAO_PASSWORD")
	require.NotContains(t, joined, "hunter2")
	require.Contains(t, joined, "HOME=/home/someone",
		"but it still needs enough to resolve stdlib paths")
}

// TestSpinEnvIsAllowListed keeps the budget test working: the hook that puts a
// child into its heartbeat loop reaches it only if it is on the allow-list, and
// a rename on one side would silently turn that test into a no-op.
func TestSpinEnvIsAllowListed(t *testing.T) {
	require.Contains(t, childEnvAllowed, spinEnv,
		"the budget test's heartbeat hook must reach the child, or it proves nothing")
}

// TestBoundedBufferTruncates pins that attacker-influenced stderr volume cannot
// be buffered without limit, and that a short write is never reported to exec's
// copier -- which would be misread as the child failing.
func TestBoundedBufferTruncates(t *testing.T) {
	b := &boundedBuffer{limit: 8}

	n, err := b.Write([]byte("0123456789"))
	require.NoError(t, err)
	require.Equal(t, 10, n, "must report the full length, not the retained length")

	n, err = b.Write([]byte("more"))
	require.NoError(t, err)
	require.Equal(t, 4, n)

	require.Equal(t, "01234567\n[truncated]", b.String())
}

// TestWouldExceedSpend pins the bound on what one run will pay for approvals.
//
// Every approval costs the full gas fee whether or not it succeeds, and the
// daemon decides on its own when to send one -- so without a bound, anything
// that makes approvals fail repeatedly drains the approver key.
func TestWouldExceedSpend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spent    int64
		fee      int64
		maxSpend int64
		want     bool
	}{
		{"no bound set", 1 << 40, 10, 0, false},
		{"first approval fits", 0, 10, 100, false},
		{"exactly at the bound still fits", 90, 10, 100, false},
		{"one past the bound", 91, 10, 100, true},
		{"already spent it all", 100, 10, 100, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := &oracle{spent: tt.spent, enableFee: tt.fee, maxSpend: tt.maxSpend}
			assert.Equal(t, tt.want, o.wouldExceedSpend())
		})
	}
}

// TestNewOracleRejectsUnusableSpendBound pins the startup checks. A bound that
// cannot pay for a single approval would leave the daemon running and silently
// approving nothing, which is worse than refusing to start.
func TestNewOracleRejectsUnusableSpendBound(t *testing.T) {
	base := func() config {
		return config{
			remote:    "http://127.0.0.1:26657",
			chainID:   "test",
			mnemonic:  testMnemonic,
			gnoRoot:   gnoenv.RootDir(),
			gasFee:    "1000000ugnot",
			gasWanted: defaultGasWanted,
		}
	}
	tio := commands.NewTestIO()
	tio.SetOut(commands.WriteNopCloser(io.Discard))
	tio.SetErr(commands.WriteNopCloser(io.Discard))

	t.Run("below one approval", func(t *testing.T) {
		cfg := base()
		cfg.maxSpend = "999999ugnot"
		_, err := newOracle(cfg, tio)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "below the cost of a single approval")
	})

	t.Run("wrong denom", func(t *testing.T) {
		cfg := base()
		cfg.maxSpend = "100foocoin"
		_, err := newOracle(cfg, tio)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "gas fees are paid in")
	})

	t.Run("a usable bound is accepted", func(t *testing.T) {
		cfg := base()
		cfg.maxSpend = defaultMaxSpend
		o, err := newOracle(cfg, tio)
		require.NoError(t, err)
		assert.Equal(t, int64(1000000), o.enableFee)
		assert.Positive(t, o.maxSpend)
	})
}

// stubGetter answers from a fixed map and records what it was asked for.
type stubGetter struct {
	pkgs  map[string]*std.MemPackage
	asked []string
}

func (g *stubGetter) GetMemPackage(pkgPath string) *std.MemPackage {
	g.asked = append(g.asked, pkgPath)
	return g.pkgs[pkgPath]
}

func onePkg(pkgPath, body string) *std.MemPackage {
	return &std.MemPackage{
		Name:  "x",
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "x.gno", Body: body}},
	}
}

// TestHybridGetterAsksTheChainForUserPackages pins where each kind of import is
// resolved from.
//
// The daemon exists to say what the validator will see, and the validator
// resolves imports from chain state. Disk used to be tried first, so a package
// importing something present in the operator's examples/ but absent from the
// chain verified clean and was approved -- then failed its own type-check at
// enable time, burning a fee and blaming the code for the operator's local tree.
func TestHybridGetterAsksTheChainForUserPackages(t *testing.T) {
	t.Parallel()

	const userPath = "gno.land/p/demo/thing"
	const stdPath = "strings"

	t.Run("a user package comes from the chain even when disk has it", func(t *testing.T) {
		t.Parallel()

		disk := &stubGetter{pkgs: map[string]*std.MemPackage{
			userPath: onePkg(userPath, "package x // from disk\n"),
		}}
		chain := &rpcGetter{
			cache: map[string]*std.MemPackage{
				userPath: onePkg(userPath, "package x // from chain\n"),
			},
		}
		h := hybridGetter{disk: disk, rpc: chain}

		got := h.GetMemPackage(userPath)
		require.NotNil(t, got)
		assert.Contains(t, got.Files[0].Body, "from chain",
			"the chain's copy is the one the validator will use")
		assert.NotContains(t, disk.asked, userPath,
			"disk must not even be consulted for a user package")
	})

	t.Run("a user package the chain lacks stays unresolved", func(t *testing.T) {
		t.Parallel()

		// The failure mode this whole change is about: disk has it, chain does
		// not. The answer must be "unresolved", so the typechecker reports the
		// import and the package is not approved.
		disk := &stubGetter{pkgs: map[string]*std.MemPackage{
			userPath: onePkg(userPath, "package x\n"),
		}}
		// qfile always fails, i.e. the chain does not have it.
		chain := &rpcGetter{
			qfile: func(string) ([]byte, error) { return nil, errStub },
			cache: map[string]*std.MemPackage{},
		}
		h := hybridGetter{disk: disk, rpc: chain}

		assert.Nil(t, h.GetMemPackage(userPath),
			"a package the chain does not have must not be resolved from disk")
	})

	t.Run("a stdlib comes from disk", func(t *testing.T) {
		t.Parallel()

		disk := &stubGetter{pkgs: map[string]*std.MemPackage{
			stdPath: onePkg(stdPath, "package strings\n"),
		}}
		// Would fail if consulted, so a stdlib routed to the chain shows up.
		chain := &rpcGetter{
			qfile: func(string) ([]byte, error) { return nil, errStub },
			cache: map[string]*std.MemPackage{},
		}
		h := hybridGetter{disk: disk, rpc: chain}

		require.NotNil(t, h.GetMemPackage(stdPath),
			"stdlibs ship with the binary; the chain cannot serve them")
		assert.Contains(t, disk.asked, stdPath)
	})

	t.Run("with no remote, disk answers everything", func(t *testing.T) {
		t.Parallel()

		disk := &stubGetter{pkgs: map[string]*std.MemPackage{
			userPath: onePkg(userPath, "package x\n"),
		}}
		h := hybridGetter{disk: disk, rpc: nil}

		assert.NotNil(t, h.GetMemPackage(userPath),
			"development mode: there is nothing to ask, so disk is all there is")
	})
}
