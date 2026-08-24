package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	// Importing the vm package also registers its message types (MsgAddPackage,
	// MsgEnablePackage, ...) with amino so block txs decode correctly.
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// oracle watches a chain, typechecks submitted packages off-chain, and
// broadcasts approvals for the ones that pass.
type oracle struct {
	cfg      config
	io       commands.IO
	client   gnoclient.Client
	approver crypto.Address

	// candidates carries submitted packages from the block reader to the
	// verifier goroutine. See runVerifier for why they are separate.
	candidates chan *std.MemPackage

	// seen dedupes packages already processed in this run. Touched ONLY by the
	// verifier goroutine, never by the block reader, or the two race on a plain
	// map.
	seen map[string]struct{}

	// overBudget counts how many times a path has run out of verification time
	// without reaching a verdict.
	overBudget map[string]int

	// spent is the gas fees paid for approvals so far this run, and maxSpend the
	// bound from -max-spend. Touched only by the verifier goroutine, which is
	// the only thing that approves.
	spent     int64
	maxSpend  int64
	enableFee int64

	// logMu serializes writes to io. commands.IOImpl buffers through a
	// bufio.Writer, which is not goroutine-safe, and the block reader and the
	// verifier goroutine both log -- a real race, not a theoretical one
	// (reproduced under -race). Every log in this daemon goes through logf/errf.
	logMu sync.Mutex
}

func (o *oracle) logf(format string, args ...any) {
	o.logMu.Lock()
	defer o.logMu.Unlock()
	o.io.Printfln(format, args...)
}

func (o *oracle) errf(format string, args ...any) {
	o.logMu.Lock()
	defer o.logMu.Unlock()
	o.io.ErrPrintfln(format, args...)
}

func (o *oracle) logln(args ...any) {
	o.logMu.Lock()
	defer o.logMu.Unlock()
	o.io.Println(args...)
}

// maxOverBudgetAttempts caps how often one package path may consume a full
// budget without producing a verdict.
//
// Leaving an overrun unrecorded is what lets a transient one be retried, but
// taken alone it lets a single path consume a full budget per resubmission
// indefinitely, and resubmitting MsgAddPackage is cheap under "inert". After
// this many attempts the path is recorded as seen and left for a human — the
// oracle declines to keep paying, and says so.
//
// This bounds repeat spend on one set of BYTES, which is all it claims. Fresh
// paths get a fresh allowance, and so does fresh content at the same path --
// correct, since different bytes are different work. The per-attempt budget is
// what bounds those.
const maxOverBudgetAttempts = 3

// candidateQueueSize bounds how far the block reader may run ahead of the
// verifier. Generous, because its whole job is absorbing a bursty block; past
// that, blocking the reader is the honest response to a saturated oracle.
const candidateQueueSize = 256

func newOracle(cfg config, io commands.IO) (*oracle, error) {
	signer, err := buildSigner(cfg, io)
	if err != nil {
		return nil, err
	}
	if err := signer.Validate(); err != nil {
		return nil, fmt.Errorf("invalid signer: %w", err)
	}
	info, err := signer.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to read signer info: %w", err)
	}

	rpc, err := rpcclient.NewHTTPClient(cfg.remote)
	if err != nil {
		return nil, fmt.Errorf("failed to build RPC client: %w", err)
	}

	gasFee, err := std.ParseCoin(cfg.gasFee)
	if err != nil {
		return nil, fmt.Errorf("invalid gas fee %q: %w", cfg.gasFee, err)
	}
	var maxSpend std.Coin
	if cfg.maxSpend != "" {
		maxSpend, err = std.ParseCoin(cfg.maxSpend)
		if err != nil {
			return nil, fmt.Errorf("invalid max spend %q: %w", cfg.maxSpend, err)
		}
		if maxSpend.Denom != gasFee.Denom {
			return nil, fmt.Errorf("max spend is in %s but gas fees are paid in %s",
				maxSpend.Denom, gasFee.Denom)
		}
		if maxSpend.Amount < gasFee.Amount {
			return nil, fmt.Errorf("max spend %s is below the cost of a single "+
				"approval (%s), so nothing could ever be approved",
				cfg.maxSpend, cfg.gasFee)
		}
	}

	return &oracle{
		cfg:        cfg,
		io:         io,
		client:     gnoclient.Client{Signer: signer, RPCClient: rpc},
		approver:   info.GetAddress(),
		candidates: make(chan *std.MemPackage, candidateQueueSize),
		seen:       make(map[string]struct{}),
		overBudget: make(map[string]int),
		enableFee:  gasFee.Amount,
		maxSpend:   maxSpend.Amount,
	}, nil
}

// buildSigner constructs the transaction signer. It prefers the local gnokey
// keystore (home + key), which keeps the approver key encrypted on disk and
// only unlocked at startup — the recommended setup. As a dev-only fallback, a
// mnemonic supplied via $GPAO_MNEMONIC is used instead.
//
// Note: consensus KMSes such as tmkms or gnokms are NOT usable here — they sign
// consensus votes over the privval protocol, not application transactions.
func buildSigner(cfg config, io commands.IO) (gnoclient.Signer, error) {
	if cfg.mnemonic != "" {
		signer, err := gnoclient.SignerFromBip39(cfg.mnemonic, cfg.chainID, "", 0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to build signer from mnemonic: %w", err)
		}
		return signer, nil
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.home)
	if err != nil {
		return nil, fmt.Errorf("failed to open keystore %q: %w", cfg.home, err)
	}
	// The password unlocks the key. Prefer $GPAO_PASSWORD for unattended/service
	// deployments; otherwise prompt once at startup.
	password := os.Getenv("GPAO_PASSWORD")
	if password == "" {
		password, err = io.GetPassword(fmt.Sprintf("enter password for key %q: ", cfg.key), true)
		if err != nil {
			return nil, fmt.Errorf("failed to read key password: %w", err)
		}
	}
	return gnoclient.SignerFromKeybase{
		Keybase:  kb,
		Account:  cfg.key,
		Password: password,
		ChainID:  cfg.chainID,
	}, nil
}

// run polls the node for new blocks and processes each one, until ctx is done.
func (o *oracle) run(ctx context.Context) error {
	height := o.cfg.startHeight
	if height <= 0 {
		status, err := o.client.RPCClient.Status(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to query node status: %w", err)
		}
		height = status.SyncInfo.LatestBlockHeight + 1
	}

	// Verification runs on its own goroutine, never on the block reader.
	//
	// The original reason was that verification was unbounded; a child process
	// now bounds it, so this is no longer about runaway work. What remains is
	// burst absorption: a single block can carry many MsgAddPackage, and inline
	// each one costs up to the full budget, so a burst would stall
	// chain-following for the sum of them.
	//
	// The visible cost is that `height` runs ahead of what has actually been
	// verified. That is cheap here because no progress is persisted either way
	// -- height is in-memory and a restart resumes from -start-height -- so
	// running ahead loses nothing a crash would not have lost anyway.
	//
	// One verifier, not a pool. Each verification is an isolated process now,
	// so concurrency would be safe; it stays serial because verification is the
	// expensive thing this daemon does and running several at once would make
	// each slower and the budget harder to interpret.
	go o.runVerifier(ctx)

	ticker := time.NewTicker(o.cfg.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			o.logln("gpao: shutting down")
			return nil
		case <-ticker.C:
		}

		status, err := o.client.RPCClient.Status(ctx, nil)
		if err != nil {
			o.errf("gpao: status query failed: %v", err)
			continue
		}
		latest := status.SyncInfo.LatestBlockHeight

		for ; height <= latest; height++ {
			// Catching up can span many blocks, and enqueue blocks when the
			// verifier is behind. Without this the only ctx check is between
			// ticks, so SIGINT would be ignored for the whole catch-up.
			select {
			case <-ctx.Done():
				o.logln("gpao: shutting down")
				return nil
			default:
			}
			if err := o.processBlock(ctx, height); err != nil {
				// Do NOT advance past it. Heights are read once and only move
				// forward, so skipping drops every MsgAddPackage in that block
				// permanently -- and the usual causes (node restart, timeout)
				// are transient. The outer ticker paces the retry, so this
				// cannot hot-spin.
				o.errf("gpao: block %d processing failed, will retry: %v",
					height, err)
				break
			}
		}
	}
}

// processBlock decodes a block's transactions and handles every MsgAddPackage.
func (o *oracle) processBlock(ctx context.Context, height int64) error {
	res, err := o.client.RPCClient.Block(ctx, &height)
	if err != nil {
		return err
	}
	if res.Block == nil {
		return nil
	}
	txs := res.Block.Data.Txs
	if len(txs) == 0 {
		return nil
	}
	// Only transactions that SUCCEEDED count. A block carries every transaction
	// that was proposed, including those that failed in DeliverTx -- CheckTx
	// admission is not success. Without this a submitter parks slow but valid
	// bytes at a path, then sends a second MsgAddPackage at the SAME path with
	// gas too low to survive the preprocess charge: it fails and changes
	// nothing, yet it sits in the block, so verification times the trivial
	// bytes and approves a path whose stored contents are the slow ones. The
	// enable this daemon signs would then be the vehicle for exactly the
	// unmetered compile it exists to prevent.
	//
	// Fails closed. If results cannot be fetched, or do not pair one-for-one
	// with the transactions, nothing here is queued and the caller retries the
	// height -- acting on an unverifiable block is worse than lagging.
	//
	// This does NOT close the other half: the submitter may still replace parked
	// bytes between verification and enable, because MsgEnablePackage names a
	// path rather than a content hash. That needs a chain-side change.
	results, err := o.client.RPCClient.BlockResults(ctx, &height)
	if err != nil {
		return fmt.Errorf("cannot read results for block %d, refusing to verify "+
			"transactions whose outcome is unknown: %w", height, err)
	}
	if results.Results == nil || len(results.Results.DeliverTxs) != len(txs) {
		return fmt.Errorf("block %d has %d transactions but %d results; "+
			"refusing to pair them by position",
			height, len(txs), deliverTxCount(results))
	}

	for i, raw := range txs {
		if !results.Results.DeliverTxs[i].IsOK() {
			continue
		}
		var tx std.Tx
		if err := amino.Unmarshal(raw, &tx); err != nil {
			o.errf("gpao: skipping undecodable tx at height %d: %v", height, err)
			continue
		}
		for _, msg := range tx.Msgs {
			add, ok := msg.(vm.MsgAddPackage)
			if !ok || add.Package == nil {
				continue
			}
			if err := o.enqueue(ctx, add.Package); err != nil {
				return err
			}
		}
	}
	return nil
}

func deliverTxCount(r *ctypes.ResultBlockResults) int {
	if r == nil || r.Results == nil {
		return 0
	}
	return len(r.Results.DeliverTxs)
}

// runVerifier drains the candidate queue, one package at a time.
func (o *oracle) runVerifier(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case mpkg := <-o.candidates:
			o.handleCandidate(ctx, mpkg)
		}
	}
}

// enqueue hands a candidate to the verifier, blocking if it is behind.
//
// Nothing is dropped. A dropped candidate would never be verified and so never
// approved, and since each block is read exactly once there would be no later
// retry -- it would stay inert with no record of why. Blocking is the honest
// alternative, and saturation is announced rather than left to be inferred from
// the oracle mysteriously lagging.
func (o *oracle) enqueue(ctx context.Context, mpkg *std.MemPackage) error {
	select {
	case o.candidates <- mpkg:
		return nil
	default:
	}
	o.errf(
		"gpao: verify queue full (%d), pausing block reads; the oracle is CPU-saturated",
		cap(o.candidates))
	select {
	case o.candidates <- mpkg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// handleCandidate typechecks a submitted package and, if it passes, broadcasts
// a MsgEnablePackage to activate it on-chain.
func (o *oracle) handleCandidate(ctx context.Context, mpkg *std.MemPackage) {
	path := mpkg.Path
	// Keyed on the bytes, not just the path. A rejection is a verdict about the
	// code, so a submitter who fixes the code and resubmits deserves a fresh
	// look -- keying on the path alone ignored them for the lifetime of the
	// process, which turned "your package does not compile" into "this path is
	// dead until someone restarts the daemon".
	//
	// It also means a package whose bytes changed is re-verified rather than
	// assumed settled, which matters because the submitter may replace parked
	// bytes at any time.
	key := candidateKey(mpkg)
	if _, done := o.seen[key]; done {
		return
	}

	o.logf("gpao: verifying %q (budget %s)", path, o.cfg.verifyBudget)
	err := o.verify(ctx, mpkg)
	if errors.Is(err, errVerifyUnavailable) {
		// Left unseen and uncounted, so a restart or resubmission retries it.
		o.errf("gpao: could not verify %q, leaving it pending: %v", path, err)
		return
	}
	if errors.Is(err, errVerifyBudget) {
		// Not (usually) marked seen: the verdict is "we ran out of time", not
		// "this package is bad", and CPU contention is transient.
		//
		// Note what that does and does not buy. Heights are read once and only
		// move forward, so nothing re-offers this package on its own; the retry
		// is a restart (with -start-height at or below the submitting block) or
		// a resubmission. Leaving it unseen is what makes either effective
		// instead of a no-op, and it keeps a slow package off the rejected list.
		o.overBudget[key]++
		if n := o.overBudget[key]; n >= maxOverBudgetAttempts {
			o.seen[key] = struct{}{}
			o.errf(
				"gpao: %q exceeded the verify budget %d times, giving up on it this run "+
					"(needs a human, or a larger -verify-budget): %v", path, n, err)
			return
		}
		o.errf("gpao: %q exceeded the verify budget, leaving it pending: %v", path, err)
		return
	}
	// Everything below is a verdict about the package itself, so record it.
	o.seen[key] = struct{}{}
	if err != nil {
		o.logf("gpao: %q rejected, not approving: %v", path, err)
		return
	}

	// Already live? Then there is nothing to enable, and sending the message
	// anyway costs the full fee to be told so. This is the common case when
	// catching up with -start-height over blocks that were already approved.
	if o.isActive(ctx, path) {
		o.logf("gpao: %q is already active, nothing to approve", path)
		return
	}

	// Stop before spending past the bound rather than after. The fee is charged
	// whether or not the message succeeds, so the check has to come first.
	if o.wouldExceedSpend() {
		o.errf("gpao: not approving %q: it would take this run past its "+
			"-max-spend of %d%s (already spent %d). Raise the bound or restart.",
			path, o.maxSpend, ugnotDenom, o.spent)
		return
	}

	o.logf("gpao: %q passed typecheck, broadcasting approval", path)
	// Counted before the call, not after: the fee is deducted by the ante
	// handler, so a failed approval costs exactly as much as a successful one.
	o.spent += o.enableFee
	if err := o.enable(path); err != nil {
		o.errf("gpao: failed to approve %q: %v", path, err)
		return
	}
	o.logf("gpao: %q approved and enabled", path)
}

// wouldExceedSpend reports whether paying for one more approval would take this
// run past its bound. Zero means no bound.
//
// Asked BEFORE approving, because the fee is deducted by the ante handler
// whether or not the message succeeds -- so checking afterwards would always be
// one approval too late.
func (o *oracle) wouldExceedSpend() bool {
	return o.maxSpend > 0 && o.spent+o.enableFee > o.maxSpend
}

const ugnotDenom = "ugnot"

// isActive reports whether a package is already deployed at this path.
//
// vm/qfile reads the active store, so a successful answer means the path is
// live. A parked package is invisible to it, which is what makes this a useful
// pre-flight: it distinguishes "waiting to be enabled" from "already enabled".
//
// On a query error this returns false, so a node that cannot answer does not
// silently stop the oracle approving. The spend bound is what limits the damage
// if the node is wrong.
func (o *oracle) isActive(ctx context.Context, pkgPath string) bool {
	res, err := o.client.Query(gnoclient.QueryCfg{
		Path: "vm/qfile",
		Data: []byte(pkgPath),
	})
	if err != nil || res == nil {
		return false
	}
	return res.Response.Error == nil
}

// candidateKey identifies what was verified: the path plus a hash of the
// package. Two submissions at the same path with different code are different
// candidates, because the verdict is about the code.
func candidateKey(mpkg *std.MemPackage) string {
	sum := sha256.Sum256(amino.MustMarshal(mpkg))
	return mpkg.Path + "@" + hex.EncodeToString(sum[:8])
}

// errVerifyBudget reports that verification ran out of time rather than
// finding a problem with the package. Distinguished from a rejection because
// the two deserve opposite treatment: a bad package is settled, a slow one may
// simply have lost a race with whatever else the box was doing.
var errVerifyBudget = errors.New("verify budget exceeded")

// errVerifyUnavailable reports that the verifier itself could not run -- a fork
// failure, a signal death, a missing binary. Distinguished from both a rejection
// and an overrun: it says nothing about the package, and unlike an overrun it
// does not count against the per-path allowance, because the operator's box
// misbehaving is not the submitter's doing.
var errVerifyUnavailable = errors.New("verifier unavailable")

// enable builds, signs and broadcasts a MsgEnablePackage for pkgPath.
func (o *oracle) enable(pkgPath string) error {
	gasFee, err := std.ParseCoin(o.cfg.gasFee)
	if err != nil {
		return fmt.Errorf("invalid gas fee %q: %w", o.cfg.gasFee, err)
	}

	tx := std.Tx{
		Msgs:       []std.Msg{vm.MsgEnablePackage{Approver: o.approver, PkgPath: pkgPath}},
		Fee:        std.NewFee(o.cfg.gasWanted, gasFee),
		Signatures: nil,
	}

	// accountNumber/sequenceNumber == 0 lets SignTx auto-query the chain.
	signed, err := o.client.SignTx(tx, 0, 0)
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	// BroadcastTxCommit returns an error if CheckTx or DeliverTx failed.
	if _, err := o.client.BroadcastTxCommit(signed); err != nil {
		return fmt.Errorf("broadcast: %w", err)
	}
	return nil
}
