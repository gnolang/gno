package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	// Importing the vm package also registers its message types (MsgAddPackage,
	// MsgEnablePackage, ...) with amino so block txs decode correctly.
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
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

	// failedEnable counts how many times a package that PASSED verification
	// failed to enable on-chain. Same shape and same reason as overBudget: the
	// bytes are not the problem, so the path is left retryable until the count
	// says otherwise. Touched only by the verifier goroutine, like seen and
	// overBudget, or they race on a plain map.
	failedEnable map[string]int

	// spent is the gas fees paid for approvals so far this run, and maxSpend the
	// bound from -max-spend. Touched only by the verifier goroutine, which is
	// the only thing that approves.
	spent     int64
	maxSpend  int64
	enableFee int64

	// blockMaxGas is the chain's Block.MaxGas. It bounds both the probe used for
	// estimation and the resulting gas-wanted, because the ante refuses a
	// transaction above it rather than clamping. Set to defaultBlockMaxGas when
	// the chain reports no bound, and until the chain answers at all.
	//
	// Atomic because the two goroutines are on opposite ends of it: the block
	// reader keeps asking for it until the chain answers, and the verifier
	// reads it on every approval.
	blockMaxGas atomic.Int64

	// status is the one piece of oracle state readable from outside the
	// verifier goroutine, and it takes a lock for that reason. See statusBoard.
	status *statusBoard

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

// maxEnableAttempts caps how often a package that passed verification may fail
// to enable before the oracle stops trying.
//
// A failed enable is usually not about the bytes -- the creator cannot cover the
// storage deposit, a dependency is not live yet, a namespace or governance param
// moved, the block ran out of gas, or the gas estimate was measured against
// state that has since changed. Those clear on their own, and the estimate in
// particular is re-measured against fresher state on the next attempt. Marking
// the path seen on the first failure retires those bytes for the rest of the run
// with the reason visible only in this process's stderr, which is how a creator
// pays a submission charge and then never learns why nothing happened.
const maxEnableAttempts = 3

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
		cfg:          cfg,
		io:           io,
		client:       gnoclient.Client{Signer: signer, RPCClient: rpc},
		approver:     info.GetAddress(),
		candidates:   make(chan *std.MemPackage, candidateQueueSize),
		seen:         make(map[string]struct{}),
		overBudget:   make(map[string]int),
		failedEnable: make(map[string]int),
		status:       newStatusBoard(),
		enableFee:    gasFee.Amount,
		maxSpend:     maxSpend.Amount,
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

// serveStatus starts the read-only status API and stops it when ctx is done.
//
// Failing to listen is logged, not fatal: the oracle's job is approving
// packages, and refusing to do it because a reporting port is taken would trade
// the thing that matters for the thing that explains it.
func (o *oracle) serveStatus(ctx context.Context, addr string) {
	srv := &http.Server{
		Handler:           o.status.statusHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		o.errf("gpao: status API not available on %s: %v", addr, err)
		return
	}
	o.logf("gpao: status API on http://%s/status", ln.Addr())

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			o.errf("gpao: status API stopped: %v", err)
		}
	}()
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()
}

// run polls the node for new blocks and processes each one, until ctx is done.
func (o *oracle) run(ctx context.Context) error {
	// The chain's ceiling is asked for with the poll rather than once here,
	// because a node that is not up yet would otherwise settle it for the whole
	// process -- and settling for the fallback is not harmless: on a chain
	// configured BELOW it, every probe is signed above Block.MaxGas, the ante
	// refuses it rather than clamping, and the simulate comes back as a message
	// the node ran and rejected. Nothing would ever be approved, and nothing
	// would say why. The fallback stands until the chain answers.
	o.blockMaxGas.Store(defaultBlockMaxGas)
	ceilingKnown := false

	if o.cfg.statusListen != "" {
		o.serveStatus(ctx, o.cfg.statusListen)
	}

	height := o.cfg.startHeight

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

		if !ceilingKnown {
			if maxGas, ok := o.queryBlockMaxGas(ctx); ok {
				o.blockMaxGas.Store(maxGas)
				ceilingKnown = true
			}
		}

		status, err := o.client.RPCClient.Status(ctx, nil)
		if err != nil {
			o.errf("gpao: status query failed: %v", err)
			continue
		}
		latest := status.SyncInfo.LatestBlockHeight
		if height <= 0 {
			// -start-height 0 means begin at the tip. Resolved on the first
			// poll that answers, so a node that is not up yet costs a poll
			// interval rather than the process.
			height = latest + 1
		}

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
		o.status.record(path, statusPending, "the oracle could not run verification: "+err.Error(), 0)
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
			o.status.record(path, statusGaveUp,
				"verification ran out of time repeatedly; needs a larger -verify-budget", n)
			o.errf(
				"gpao: %q exceeded the verify budget %d times, giving up on it this run "+
					"(needs a human, or a larger -verify-budget): %v", path, n, err)
			return
		}
		o.status.record(path, statusPending, "verification ran out of time; will be retried",
			o.overBudget[key])
		o.errf("gpao: %q exceeded the verify budget, leaving it pending: %v", path, err)
		return
	}
	// A rejection IS a verdict about the bytes, so record it: re-verifying them
	// would reach the same answer, and the submitter has to change something for
	// it to be worth another look -- which produces a different key.
	if err != nil {
		// The one a submitter can actually act on: their code did not pass.
		o.status.record(path, statusRejected, err.Error(), 0)
		o.seen[key] = struct{}{}
		o.logf("gpao: %q rejected, not approving: %v", path, err)
		return
	}

	// Already live? Then there is nothing to enable, and sending the message
	// anyway costs the full fee to be told so. This is the common case when
	// catching up with -start-height over blocks that were already approved.
	// Terminal, so recorded.
	if o.isActive(ctx, path) {
		o.status.record(path, statusApproved, "already active on-chain", 0)
		o.seen[key] = struct{}{}
		o.logf("gpao: %q is already active, nothing to approve", path)
		return
	}

	// Stop before spending past the bound rather than after. The fee is charged
	// whether or not the message succeeds, so the check has to come first.
	//
	// Left UNSEEN. The message tells the operator to raise the bound or restart,
	// and both are no-ops if the package has been retired -- the whole point is
	// that it should be approved once there is budget for it.
	if o.wouldExceedSpend() {
		// Nothing is wrong with the package; the oracle is out of allowance.
		// Saying so is the difference between "your code is bad" and "ask the
		// operator", which the submitter cannot otherwise tell apart.
		o.status.record(path, statusBlocked,
			"the oracle has reached its spending limit for this run", 0)
		o.errf("gpao: not approving %q: it would take this run past its "+
			"-max-spend of %d%s (already spent %d). Raise the bound or restart.",
			path, o.maxSpend, ugnotDenom, o.spent)
		return
	}

	o.logf("gpao: %q passed typecheck, broadcasting approval", path)
	// Counted before the call, not after: the fee is deducted by the ante
	// handler, so a failed approval costs exactly as much as a successful one.
	o.spent += o.enableFee
	if err := o.enable(path, vm.PackageContentHash(mpkg)); err != nil {
		// Left unseen until the count runs out, for the reason at
		// maxEnableAttempts: the package verified, so the failure is about the
		// chain's state rather than the code, and most such causes clear.
		//
		// As with overBudget, this does not re-offer the package by itself --
		// heights only move forward. It is what makes a restart (with
		// -start-height at or below the submitting block) or a resubmission of
		// the same bytes effective instead of a silent no-op.
		if n, giveUp := o.recordEnableFailure(key); giveUp {
			o.status.record(path, statusGaveUp, err.Error(), n)
			o.errf("gpao: failed to approve %q %d times, giving up on it this run "+
				"(needs a human): %v", path, n, err)
			return
		}
		o.status.record(path, statusPending, err.Error(), o.failedEnable[key])
		o.errf("gpao: failed to approve %q, leaving it pending: %v", path, err)
		return
	}
	o.status.record(path, statusApproved, "", 0)
	o.seen[key] = struct{}{}
	o.logf("gpao: %q approved and enabled", path)
}

// queryBlockMaxGas reads the chain's Block.MaxGas, and reports whether the
// chain answered at all.
//
// The two are separate the way classifySimulate separates them: a node that
// could not be reached has said nothing about the ceiling and must be asked
// again, while a chain that answered has settled it -- including when the
// answer is unusable and the fallback stands in. Answered once, it is not
// asked for again: it is a consensus param, so it changes rarely, and a
// per-approval query would add a round trip to every enable to learn something
// that almost never moves.
//
// A chain may set -1, meaning no bound. The fallback is used there too -- an
// unbounded ceiling would let one absurd estimate ask for unbounded gas, and
// nothing gpao approves should need more than a full block's worth anyway.
func (o *oracle) queryBlockMaxGas(ctx context.Context) (maxGas int64, answered bool) {
	res, err := o.client.RPCClient.ConsensusParams(ctx, nil)
	if err != nil {
		o.errf("gpao: block max gas query failed, using %d until it answers: %v",
			defaultBlockMaxGas, err)
		return 0, false
	}
	maxGas = blockMaxGasFrom(res, nil)
	if maxGas == defaultBlockMaxGas {
		o.logf("gpao: using %d for block max gas", defaultBlockMaxGas)
	}
	return maxGas, true
}

// blockMaxGasFrom picks the ceiling from a consensus-params response, falling
// back to defaultBlockMaxGas on anything unusable.
//
// Split out from the query so it can be tested without a node. A chain may
// legitimately report -1, meaning no bound; the fallback covers that too,
// because an unbounded ceiling would let one absurd estimate ask for unbounded
// gas.
func blockMaxGasFrom(res *ctypes.ResultConsensusParams, err error) int64 {
	if err != nil || res == nil || res.ConsensusParams.Block == nil {
		return defaultBlockMaxGas
	}
	if maxGas := res.ConsensusParams.Block.MaxGas; maxGas > 0 {
		return maxGas
	}
	return defaultBlockMaxGas
}

// recordEnableFailure counts a failed enable for this content and reports the
// count, plus whether the oracle should stop retrying it.
//
// Marks the content seen only on the last attempt, which is the whole point:
// until then the path stays eligible, so a restart or a resubmission of the same
// bytes gets another go at a failure that was never about the bytes.
func (o *oracle) recordEnableFailure(key string) (n int, giveUp bool) {
	o.failedEnable[key]++
	n = o.failedEnable[key]
	if n >= maxEnableAttempts {
		o.seen[key] = struct{}{}
		return n, true
	}
	return n, false
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

// errVerifyUnavailable reports that verification could not produce evidence --
// a fork failure, a signal death, a missing binary, or the child running fine
// while the network under its import resolver failed. Distinguished from both
// a rejection and an overrun: it says
// nothing about the package, and unlike an overrun it does not count against
// the per-path allowance, because the operator's box misbehaving is not the
// submitter's doing.
var errVerifyUnavailable = errors.New("verifier unavailable")

// defaultBlockMaxGas is the fallback ceiling on a gas-wanted, used until the
// chain's own Block.MaxGas is known. It matches tm2's MaxBlockMaxGas, which is
// also the default a chain gets if it sets nothing.
//
// The real value matters because the ante REFUSES a transaction whose
// GasWanted exceeds Block.MaxGas rather than clamping it. So on a chain
// configured below this fallback, a probe signed at the fallback is rejected
// and every estimate fails -- which is why the chain is asked for blockMaxGas
// on each poll until it answers, instead of assuming it.
const defaultBlockMaxGas = int64(3_000_000_000)

// gasHeadroomNum/Den add 20% to a measured estimate.
//
// The estimate is a measurement of one execution against the state the
// simulation saw, and the enable lands against later state: the package's
// dependencies may have grown, the names or CLA realms may have been touched,
// and the storage deposit is priced off realm bytes that the simulation
// computed on a different tree. Headroom absorbs that. It costs nothing when
// unused -- fees are flat, so an over-large GasWanted is not over-paid, it just
// has to clear the mempool's minimum.
const (
	gasHeadroomNum = 12
	gasHeadroomDen = 10
)

// simulateVerdict says what a simulation told us about an enable.
type simulateVerdict int

const (
	// verdictReady: the message ran and succeeded. Use the measured gas.
	verdictReady simulateVerdict = iota
	// verdictUnknown: the node did not answer. Nothing was learned about the
	// message.
	verdictUnknown
	// verdictWillFail: the message ran and failed. Broadcasting would pay a
	// gas fee to be told the same thing on-chain.
	verdictWillFail
)

// classifySimulate splits a simulation outcome three ways.
//
// The distinction is the whole point: gnoclient.Simulate reports a transport
// failure, a rejected query and a failed message as one error, and the right
// response to the first two is the opposite of the right response to the third.
// A node that cannot be reached must not stop approvals -- anyone able to
// disturb the query path could then stall the chain's deploys -- while a
// message the node has already run and rejected is not worth paying to repeat.
//
// Split out from the query so it can be tested without a node.
func classifySimulate(res *abci.ResponseDeliverTx, err error) simulateVerdict {
	switch {
	case err != nil || res == nil:
		return verdictUnknown
	case res.Error != nil:
		return verdictWillFail
	default:
		return verdictReady
	}
}

// gasWantedFor sizes an enable's GasWanted from a measured estimate, bounded by
// the chain's Block.MaxGas.
//
// Returns the fallback when the estimate is unusable (zero or negative, i.e.
// the simulation did not produce a number), so a failed estimate degrades to
// the configured -gas-wanted rather than to zero gas. The fallback is bounded
// too: -gas-wanted is operator input and can exceed the chain's Block.MaxGas,
// which the ante refuses outright rather than clamping. An unbounded fallback
// would fail every enable on exactly the chains it exists to keep working on.
func gasWantedFor(estimated, fallback, ceiling int64) int64 {
	if estimated <= 0 {
		if fallback > ceiling {
			return ceiling
		}
		return fallback
	}
	// Overflow guard before the multiply: an absurd estimate would otherwise
	// wrap negative and be refused as a non-positive GasWanted.
	if estimated > ceiling {
		return ceiling
	}
	if want := estimated * gasHeadroomNum / gasHeadroomDen; want < ceiling {
		return want
	}
	return ceiling
}

// enable builds, signs and broadcasts a MsgEnablePackage for pkgPath.
//
// Signed twice, deliberately. The fee is part of the sign bytes, so the
// GasWanted cannot be corrected after signing -- the first signature exists only
// to make the transaction simulatable, and the second carries the real number.
//
// The provisional GasWanted is the block ceiling rather than the configured
// value, because simulate executes under the transaction's own limit: sizing it
// at -gas-wanted would make the simulation run out of gas exactly on the
// packages whose cost we most need to learn, and report that failure instead of
// a measurement.
func (o *oracle) enable(pkgPath, pkgHash string) error {
	gasFee, err := std.ParseCoin(o.cfg.gasFee)
	if err != nil {
		return fmt.Errorf("invalid gas fee %q: %w", o.cfg.gasFee, err)
	}
	// Name the source that was verified. The keeper hashes the parked blob the
	// same way and refuses if they differ, so a creator who replaces the bytes
	// after verification cannot ride this approval.
	msg := vm.MsgEnablePackage{Approver: o.approver, PkgPath: pkgPath, PkgHash: pkgHash}

	// Read once, so the probe and the gas-wanted it sizes cannot straddle the
	// block reader adopting the chain's ceiling.
	ceiling := o.blockMaxGas.Load()

	// accountNumber/sequenceNumber == 0 lets SignTx auto-query the chain.
	probe, err := o.client.SignTx(std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(ceiling, gasFee),
	}, 0, 0)
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}

	gasWanted := gasWantedFor(0, o.cfg.gasWanted, ceiling)
	sim, simErr := o.client.SimulateResult(probe)
	switch classifySimulate(sim, simErr) {
	case verdictWillFail:
		// The node has already run this and rejected it. Broadcasting would
		// spend a gas fee to have it rejected again, in a block, and the
		// creator would still be told nothing they can act on.
		return fmt.Errorf("simulate says the enable would fail: %w", sim.Error)
	case verdictUnknown:
		// Not fatal, and deliberately so: refusing to approve because the node
		// would not answer hands anyone who can disturb the query path a way to
		// stall approvals chain-wide. Fall back to the configured value.
		o.logf("estimate failed for %s, using %d: %v", pkgPath, gasWanted, simErr)
	case verdictReady:
		gasWanted = gasWantedFor(sim.GasUsed, o.cfg.gasWanted, ceiling)
		o.logf("estimated %d gas for %s, sending with %d", sim.GasUsed, pkgPath, gasWanted)
	}

	signed, err := o.client.SignTx(std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(gasWanted, gasFee),
	}, 0, 0)
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	// BroadcastTxCommit returns an error if CheckTx or DeliverTx failed.
	if _, err := o.client.BroadcastTxCommit(signed); err != nil {
		return fmt.Errorf("broadcast: %w", err)
	}
	return nil
}
