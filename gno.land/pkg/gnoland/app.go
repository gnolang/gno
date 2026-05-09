// Package gnoland contains the bootstrapping code to launch a gno.land node.
package gnoland

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/_tags"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	sdkCfg "github.com/gnolang/gno/tm2/pkg/sdk/config"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// AppOptions contains the options to create the gno.land ABCI application.
type AppOptions struct {
	DB                         dbm.DB             // required
	Logger                     *slog.Logger       // required
	EventSwitch                events.EventSwitch // required
	VMOutput                   io.Writer          // optional
	SkipGenesisSigVerification bool               // default to verify genesis transactions
	SkipUpgradeHeight          int64              // if set, skip the halt_min_version check at this height
	InitChainerConfig                             // options related to InitChainer
	MinGasPrices               string             // optional
	PruneStrategy              types.PruneStrategy
}

// TestAppOptions provides a "ready" default [AppOptions] for use with
// [NewAppWithOptions], using the provided db.
func TestAppOptions(db dbm.DB) *AppOptions {
	return &AppOptions{
		DB:          db,
		Logger:      log.NewNoopLogger(),
		EventSwitch: events.NewEventSwitch(),
		InitChainerConfig: InitChainerConfig{
			GenesisTxResultHandler: PanicOnFailingTxResultHandler,
			StdlibDir:              filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs"),
			CacheStdlibLoad:        true,
		},
		SkipGenesisSigVerification: true,
		PruneStrategy:              types.PruneSyncableStrategy,
	}
}

func (c AppOptions) validate() error {
	// Required fields
	switch {
	case c.DB == nil:
		return fmt.Errorf("no db provided")
	case c.Logger == nil:
		return fmt.Errorf("no logger provided")
	case c.EventSwitch == nil:
		return fmt.Errorf("no event switch provided")
	}
	return nil
}

// NewAppWithOptions creates the gno.land application with specified options.
func NewAppWithOptions(cfg *AppOptions) (abci.Application, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Capabilities keys.
	mainKey := store.NewStoreKey("main")
	baseKey := store.NewStoreKey("base")

	//  set sdk app options
	var appOpts []func(*sdk.BaseApp)
	if cfg.MinGasPrices != "" {
		appOpts = append(appOpts, sdk.SetMinGasPrices(cfg.MinGasPrices))
	}

	appOpts = append(appOpts, sdk.SetPruningOptions(cfg.PruneStrategy.Options()))

	// Create BaseApp.
	baseApp := sdk.NewBaseApp("gnoland", cfg.Logger, cfg.DB, baseKey, mainKey, appOpts...)
	baseApp.SetAppVersion("dev")

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoreWithDB(mainKey, iavl.StoreConstructor, cfg.DB)
	baseApp.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, cfg.DB)

	// Construct keepers.

	prmk := params.NewParamsKeeper(mainKey)
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, ProtoGnoSessionAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName))
	gpk := auth.NewGasPriceKeeper(mainKey)
	vmk := vm.NewVMKeeper(baseKey, mainKey, acck, bankk, prmk)
	vmk.Output = cfg.VMOutput

	prmk.Register(auth.ModuleName, acck)
	prmk.Register(bank.ModuleName, bankk)
	prmk.Register(vm.ModuleName, vmk)
	prmk.Register("node", nodeParamsKeeper{})

	// Set InitChainer
	icc := cfg.InitChainerConfig
	icc.baseApp = baseApp
	icc.acck, icc.bankk, icc.vmk, icc.prmk, icc.gpk = acck, bankk, vmk, prmk, gpk
	baseApp.SetInitChainer(icc.InitChainer)

	// Set AnteHandler
	authOptions := auth.AnteOptions{
		VerifyGenesisSignatures: !cfg.SkipGenesisSigVerification,
	}
	authAnteHandler := auth.NewAnteHandler(
		acck, bankk, auth.DefaultSigVerificationGasConsumer, authOptions)
	baseApp.SetAnteHandler(
		// Override default AnteHandler with custom logic.
		func(ctx sdk.Context, tx std.Tx, simulate bool) (
			newCtx sdk.Context, res sdk.Result, abort bool,
		) {
			// Add last gas price in the context
			ctx = ctx.WithValue(auth.GasPriceContextKey{}, gpk.LastGasPrice(ctx))
			// Override auth params. acck.GetParams internally bypasses
			// the gas meter (see tm2/pkg/sdk/auth/params.go) so this
			// read costs nothing.
			ctx = ctx.WithValue(auth.AuthParamsContextKey{}, acck.GetParams(ctx))
			// Apply VM gas config so all store operations (account
			// reads/writes in ante, message handlers, etc.) use the
			// governed depth parameters. vmk.GetParams DOES meter (vm
			// params are user-tunable consensus state and we want a
			// real gas signal on changes), so this read uses the ctx's
			// current (default) gasCfg until it's replaced below.
			gasCfg := store.DefaultGasConfig()
			vmk.GetParams(ctx).ApplyToGasConfig(&gasCfg)
			ctx = ctx.WithGasConfig(gasCfg)

			// During genesis (block height 0), automatically create accounts for signers
			// if they don't exist. This allows packages with custom creators to be loaded.
			if ctx.BlockHeight() == 0 {
				for _, signer := range tx.GetSigners() {
					if acck.GetAccount(ctx, signer) == nil {
						// Create a new account for the signer
						acc := acck.NewAccountWithAddress(ctx, signer)
						acck.SetAccount(ctx, acc)
						// Give it enough funds to pay for the transaction
						// This is only for genesis - in normal operation accounts must be funded
						err := bankk.SetCoins(ctx, signer, std.Coins{std.NewCoin("ugnot", 10_000_000_000)})
						if err != nil {
							panic(fmt.Sprintf("failed to set coins for genesis account %s: %v", signer, err))
						}
					}
				}
			}

			// Continue on with default auth ante handler.
			newCtx, res, abort = authAnteHandler(ctx, tx, simulate)
			if abort {
				return
			}

			// Session message restrictions (gno.land layer). Only
			// overwrite res when the check aborts — on success,
			// preserve the ante's res (which carries GasWanted from
			// tx.Fee). checkSessionRestrictions returns sdk.Result{}
			// on success, which would otherwise zero out GasWanted.
			if sessRes, sessAbort := checkSessionRestrictions(newCtx, tx); sessAbort {
				return newCtx, sessRes, true
			}
			return
		},
	)

	// Set begin and end transaction hooks.
	// These are used to create gno transaction stores and commit them when finishing
	// the tx - in other words, data from a failing transaction won't be persisted
	// to the gno store caches.
	baseApp.SetBeginTxHook(func(ctx sdk.Context) sdk.Context {
		// Create Gno transaction store.
		return vmk.MakeGnoTransactionStore(ctx)
	})
	baseApp.SetEndTxHook(func(ctx sdk.Context, result sdk.Result) {
		if result.IsOK() {
			vmk.CommitGnoTransactionStore(ctx)
		}
	})

	// Set EndBlocker
	baseApp.SetEndBlocker(
		EndBlocker(
			prmk,
			acck,
			gpk,
			baseApp,
		),
	)

	// Set a handler Route.
	baseApp.Router().AddRoute("auth", auth.NewHandler(acck, gpk))
	baseApp.Router().AddRoute("bank", bank.NewHandler(bankk))
	baseApp.Router().AddRoute("params", params.NewHandler(prmk))
	baseApp.Router().AddRoute("vm", vm.NewHandler(vmk))

	// Load latest version.
	if err := baseApp.LoadLatestVersion(); err != nil {
		return nil, err
	}

	// Initialize the VMKeeper.
	ms := baseApp.GetCacheMultiStore()
	vmk.Initialize(cfg.Logger, ms)
	ms.MultiWrite() // XXX why was't this needed?

	// Verify node startup constraints set by governance halt proposals.
	if err := checkNodeStartupParams(prmk, baseApp.GetCacheMultiStore(), baseApp.LastBlockHeight(), cfg.SkipUpgradeHeight); err != nil {
		return nil, err
	}

	return baseApp, nil
}

// GenesisAppConfig wraps the most important
// genesis params relating to the App
type GenesisAppConfig struct {
	SkipFailingTxs      bool // does not stop the chain from starting if any tx fails
	SkipSigVerification bool // does not verify the transaction signatures in genesis
}

// NewTestGenesisAppConfig returns a testing genesis app config
func NewTestGenesisAppConfig() GenesisAppConfig {
	return GenesisAppConfig{
		SkipFailingTxs:      true,
		SkipSigVerification: true,
	}
}

// NewApp creates the gno.land application.
func NewApp(
	dataRootDir string,
	genesisCfg GenesisAppConfig,
	appCfg *sdkCfg.AppConfig,
	evsw events.EventSwitch,
	logger *slog.Logger,
	skipUpgradeHeight int64,
) (abci.Application, error) {
	var err error

	cfg := &AppOptions{
		Logger:      logger,
		EventSwitch: evsw,
		InitChainerConfig: InitChainerConfig{
			GenesisTxResultHandler: PanicOnFailingTxResultHandler,
			StdlibDir:              filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs"),
		},
		MinGasPrices:               appCfg.MinGasPrices,
		SkipGenesisSigVerification: genesisCfg.SkipSigVerification,
		SkipUpgradeHeight:          skipUpgradeHeight,
		PruneStrategy:              appCfg.PruneStrategy,
	}
	if genesisCfg.SkipFailingTxs {
		cfg.GenesisTxResultHandler = NoopGenesisTxResultHandler
	}

	// Get main DB.
	cfg.DB, err = dbm.NewDB("gnolang", dbm.PebbleDBBackend, filepath.Join(dataRootDir, config.DefaultDBDir))
	if err != nil {
		return nil, fmt.Errorf("error initializing database %q using path %q: %w", dbm.PebbleDBBackend, dataRootDir, err)
	}

	return NewAppWithOptions(cfg)
}

// GenesisTxResultHandler is called in the InitChainer after a genesis
// transaction is executed.
type GenesisTxResultHandler func(ctx sdk.Context, tx std.Tx, res sdk.Result)

// NoopGenesisTxResultHandler is a no-op GenesisTxResultHandler.
func NoopGenesisTxResultHandler(_ sdk.Context, _ std.Tx, _ sdk.Result) {}

// PanicOnFailingTxResultHandler handles genesis transactions by panicking if
// res.IsErr() returns true.
func PanicOnFailingTxResultHandler(_ sdk.Context, _ std.Tx, res sdk.Result) {
	if res.IsErr() {
		panic(res.Log)
	}
}

// InitChainerConfig keeps the configuration for the InitChainer.
// [NewAppWithOptions] will set [InitChainerConfig.InitChainer] as its InitChainer
// function.
type InitChainerConfig struct {
	// Handles the results of each genesis transaction.
	GenesisTxResultHandler

	// Standard library directory.
	StdlibDir string
	// Whether to keep a record of the DB operations to load standard libraries,
	// so they can be quickly replicated on additional genesis executions.
	// This should be used for integration testing, where InitChainer will be
	// called several times.
	CacheStdlibLoad bool

	// These fields are passed directly by NewAppWithOptions, and should not be
	// configurable by end-users.
	baseApp *sdk.BaseApp
	vmk     vm.VMKeeperI
	acck    auth.AccountKeeperI
	bankk   bank.BankKeeperI
	prmk    params.ParamsKeeperI
	gpk     auth.GasPriceKeeperI
}

// InitChainer is the function that can be used as a [sdk.InitChainer].
func (cfg InitChainerConfig) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	start := time.Now()
	ctx.Logger().Debug("InitChainer: started")

	// load standard libraries; immediately committed to store so that they are
	// available for use when processing the genesis transactions below.
	cfg.loadStdlibs(ctx)
	ctx.Logger().Debug("InitChainer: standard libraries loaded",
		"elapsed", time.Since(start))

	// Seed valset:current from genesis validators BEFORE loadAppState so
	// that any genesis-time realm reads of sysparams.GetValsetEffective /
	// GetValsetEntries see the authoritative set instead of empty.
	//
	// Note on sentinel scope: internalWriteCtxKey is set on the LOCAL
	// ictx variable, NOT on app.deliverState.ctx. baseapp.Deliver pulls
	// a fresh ctx via getContextForTx (tm2/pkg/sdk/baseapp.go:606-611),
	// so this sentinel does NOT propagate into genesis-tx execution —
	// a malicious genesis tx cannot manufacture a sentinel-bearing ctx
	// and write valset:current directly.
	ictx := ctx.WithValue(internalWriteCtxKey{}, true)
	cfg.prmk.SetStrings(ictx, valsetCurrentPath, abci.EncodeValidatorUpdates(abci.ValidatorUpdates(req.Validators)))

	// load app state. AppState may be nil mostly in some minimal testing setups;
	// so log a warning when that happens.
	txResponses, err := cfg.loadAppState(ctx, req.AppState)
	if err != nil {
		return abci.ResponseInitChain{
			ResponseBase: abci.ResponseBase{
				Error: abci.StringError(err.Error()),
			},
		}
	}

	ctx.Logger().Debug("InitChainer: genesis transactions loaded",
		"elapsed", time.Since(start))

	// Done!
	return abci.ResponseInitChain{
		Validators:  req.Validators,
		TxResponses: txResponses,
	}
}

func (cfg InitChainerConfig) loadStdlibs(ctx sdk.Context) {
	// cache-wrapping is necessary for non-validator nodes; in the tm2 BaseApp,
	// this is done using BaseApp.cacheTxContext; so we replicate it here.
	ms := ctx.MultiStore()
	msCache := ms.MultiCacheWrap()

	stdlibCtx := cfg.vmk.MakeGnoTransactionStore(ctx)
	stdlibCtx = stdlibCtx.WithMultiStore(msCache)
	if cfg.CacheStdlibLoad {
		cfg.vmk.LoadStdlibCached(stdlibCtx, cfg.StdlibDir)
	} else {
		cfg.vmk.LoadStdlib(stdlibCtx, cfg.StdlibDir)
	}
	cfg.vmk.CommitGnoTransactionStore(stdlibCtx)

	msCache.MultiWrite()

	// Populate stdlib byte cache for gas-free stdlib reads.
	// Must read from the deliver state's baseStore (where stdlib objects
	// were written), not the persistent gnoStore's baseStore (which is
	// a different cache layer that doesn't have them yet).
	cfg.vmk.PopulateStdlibCacheFrom(ms)
}

func (cfg InitChainerConfig) loadAppState(ctx sdk.Context, appState any) ([]abci.ResponseDeliverTx, error) {
	state, ok := appState.(GnoGenesisState)
	if !ok {
		return nil, fmt.Errorf("invalid AppState of type %T", appState)
	}

	cfg.bankk.InitGenesis(ctx, state.Bank)
	// Apply genesis balances.
	for _, bal := range state.Balances {
		acc := cfg.acck.NewAccountWithAddress(ctx, bal.Address)
		cfg.acck.SetAccount(ctx, acc)
		err := cfg.bankk.SetCoins(ctx, bal.Address, bal.Amount)
		if err != nil {
			panic(err)
		}
	}
	// The account keeper's initial genesis state must be set after genesis
	// accounts are created in account keeeper with genesis balances
	cfg.acck.InitGenesis(ctx, state.Auth)

	// The unrestricted address must have been created as one of the genesis accounts.
	// Otherwise, we cannot verify the unrestricted address in the genesis state.

	for _, addr := range state.Auth.Params.UnrestrictedAddrs {
		acc := cfg.acck.GetAccount(ctx, addr)
		if acc == nil {
			panic(fmt.Errorf("unrestricted address must be one of the genesis accounts: invalid account %q", addr))
		}

		accr := acc.(*GnoAccount)
		accr.SetTokenLockWhitelisted(true)
		cfg.acck.SetAccount(ctx, acc)
	}

	cfg.vmk.InitGenesis(ctx, state.VM)

	params := cfg.acck.GetParams(ctx)
	ctx = ctx.WithValue(auth.AuthParamsContextKey{}, params)
	auth.InitChainer(ctx, cfg.gpk, params.InitialGasPrice)

	// Replay genesis txs.
	txResponses := make([]abci.ResponseDeliverTx, 0, len(state.Txs))

	// Run genesis txs
	for _, tx := range state.Txs {
		var (
			stdTx    = tx.Tx
			metadata = tx.Metadata

			ctxFn sdk.ContextFn
		)

		// Check if there is metadata associated with the tx
		if metadata != nil {
			// Create a custom context modifier
			ctxFn = func(ctx sdk.Context) sdk.Context {
				// Create a copy of the header, in
				// which only the timestamp information is modified
				header := ctx.BlockHeader().(*bft.Header).Copy()
				header.Time = time.Unix(metadata.Timestamp, 0)

				// Save the modified header
				return ctx.WithBlockHeader(header)
			}
		}

		res := cfg.baseApp.Deliver(stdTx, ctxFn)
		if res.IsErr() {
			ctx.Logger().Error(
				"Unable to deliver genesis tx",
				"log", res.Log,
				"error", res.Error,
				"gas-used", res.GasUsed,
			)
		}

		txResponses = append(txResponses, abci.ResponseDeliverTx{
			ResponseBase: res.ResponseBase,
			GasWanted:    res.GasWanted,
			GasUsed:      res.GasUsed,
		})

		cfg.GenesisTxResultHandler(ctx, stdTx, res)
	}
	return txResponses, nil
}

// endBlockerApp is the app abstraction required by any EndBlocker
type endBlockerApp interface {
	// LastBlockHeight returns the latest app height
	LastBlockHeight() int64

	// Logger returns the logger reference
	Logger() *slog.Logger

	// SetHaltHeight sets the block height at which the node will halt.
	SetHaltHeight(uint64)
}

// EndBlocker defines the logic executed after every block.
// It checks for a governance-requested chain halt, then reads valset changes
// from the params keeper and propagates them to consensus.
func EndBlocker(
	prmk params.ParamsKeeperI,
	acck auth.AccountKeeperI,
	gpk auth.GasPriceKeeperI,
	app endBlockerApp,
) func(
	ctx sdk.Context,
	req abci.RequestEndBlock,
) abci.ResponseEndBlock {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		// Set the auth params value in the ctx. The EndBlocker will use InitialGasPrice in
		// the params to calculate the updated gas price.
		if acck != nil {
			ctx = ctx.WithValue(auth.AuthParamsContextKey{}, acck.GetParams(ctx))
		}
		if acck != nil && gpk != nil {
			auth.EndBlocker(ctx, gpk)
		}

		// Check if GovDAO has requested a halt at this height.
		// Use == (not >=) so we only trigger once: at the exact halt height.
		// SetHaltHeight causes BeginBlock of the *next* block to panic, ensuring
		// this block is fully committed before the node stops.
		// On restart, req.Height > halt_height, so == never re-fires — no infinite loop.
		if prmk != nil {
			var haltHeight int64
			prmk.GetInt64(ctx, nodeParamHaltHeight, &haltHeight)
			if haltHeight > 0 && req.Height == haltHeight {
				app.Logger().Info(
					"GovDAO halt height reached, will halt after this block",
					"height", req.Height,
					"halt_height", haltHeight,
				)
				app.SetHaltHeight(uint64(haltHeight))
			}
		}

		// Check if there are any pending valset changes.
		dirty := false
		prmk.GetBool(ctx, valsetDirtyPath, &dirty)
		if !dirty {
			return abci.ResponseEndBlock{}
		}

		var currentEntries, proposedEntries []string
		prmk.GetStrings(ctx, valsetCurrentPath, &currentEntries)
		prmk.GetStrings(ctx, valsetProposedPath, &proposedEntries)

		// Parse proposed first; on parse failure, drop the proposal by
		// clearing dirty without writing anything else. WillSetParam
		// guards realm-side writes, but a direct chain-internal write
		// could still seed bad data; either way the recovery is just
		// "drop the bad proposal."
		proposedSet, err := abci.ParseValidatorUpdates(proposedEntries)
		if err != nil {
			app.Logger().Error("valset:proposed corrupted; dropping proposal", "err", err)
			prmk.SetBool(ctx, valsetDirtyPath, false)
			return abci.ResponseEndBlock{}
		}

		// Parse current; corruption here is chain-internal (only chain
		// code writes valset:current via ctx-sentinel + WillSetParam
		// validates on every write) so panic.
		//
		// Why not "recover by applying the proposal as adds-only when
		// current is corrupt but proposed parses"? Because ABCI's
		// ResponseEndBlock.ValidatorUpdates is a DELTA, not a snapshot.
		// tm2 applies it on top of state.NextValidators (the prior set)
		// and commits the result; there's no "replace whole set"
		// primitive at the ABCI boundary. To produce a delta that
		// yields consensus == proposed, we must know what's currently
		// in consensus so we can emit removals for the validators that
		// proposed drops. valset:current is the chain's record of that
		// set. If it's unparseable we can't compute the right delta —
		// we could only emit "add everything in proposed", which leaves
		// the real (now-untracked) prior validators in consensus and
		// produces permanent divergence between valset:current and the
		// actual signing set (the v1 prev-vs-actual bug we redesigned
		// to fix). Silent recovery would be a wrong proposal applied
		// while pretending it was the right one.
		//
		// In practice this branch is unreachable in normal operation:
		// store damage, partial commit, or a chain-code bug that wrote
		// past WillSetParam are the only ways to get here. Panic is
		// the right "this shouldn't happen, investigate" signal.
		currentSet, err := abci.ParseValidatorUpdates(currentEntries)
		if err != nil {
			panic(fmt.Sprintf("valset:current corrupted (chain-internal): %v", err))
		}

		// Min-validator floor: refuse to empty consensus.
		// proposed is the full target set, so the post-apply set has
		// exactly the entries with Power > 0. v3's normal flow emits
		// the effective set as positive-power entries — but the
		// callback also accepts an all-removes proposal that
		// publishes entries=[]string{}, so all-Power=0 is reachable
		// at the v3 boundary; this floor is the consensus-safety
		// backstop.
		liveCount := 0
		for _, u := range proposedSet {
			if u.Power > 0 {
				liveCount++
			}
		}
		if liveCount == 0 {
			app.Logger().Error("valset proposal would empty consensus; rejecting",
				"proposed_len", len(proposedSet),
				"live_count", liveCount)
			prmk.SetBool(ctx, valsetDirtyPath, false)
			return abci.ResponseEndBlock{}
		}

		// Compute diff. Whole-reject if any add/update has a disallowed
		// pubkey type — atomic accept-or-reject avoids partial-application
		// ambiguity (no filter losses, so valset:current = proposed exactly).
		diff := currentSet.UpdatesFrom(proposedSet)
		var allowedKeyTypes []string
		if cp := ctx.ConsensusParams(); cp != nil && cp.Validator != nil {
			allowedKeyTypes = cp.Validator.PubKeyTypeURLs
		}
		for _, u := range diff {
			if u.Power == 0 {
				continue // removals always allowed
			}
			if len(allowedKeyTypes) == 0 {
				continue // no allow-list configured -> accept all
			}
			if !slices.Contains(allowedKeyTypes, amino.GetTypeURL(u.PubKey)) {
				app.Logger().Error(
					"valset proposal contains disallowed pubkey type; rejecting whole proposal",
					"address", u.Address.String(),
					"pubkey_type", amino.GetTypeURL(u.PubKey),
				)
				prmk.SetBool(ctx, valsetDirtyPath, false)
				return abci.ResponseEndBlock{}
			}
		}

		app.Logger().Info("valset changes to be applied", "count", len(diff))

		// Whole-apply: advance valset:current = proposed (no filter losses
		// possible since the disallowed-pubkey scan above whole-rejects).
		// At this point valset:current records V_{H+2} — the set that will
		// be active at H+2 once the most recent EndBlock's updates apply
		// (NOT the active-signing set at H+1, which tm2 has already
		// locked in from the prior commit).
		intCtx := ctx.WithValue(internalWriteCtxKey{}, true)
		prmk.SetStrings(intCtx, valsetCurrentPath, abci.EncodeValidatorUpdates(proposedSet))
		// dirty clear uses original (no-sentinel) ctx; valset:dirty is
		// not sentinel-gated since it's bool-typed only and the realm
		// side already enforces single-writer via assertValsetCaller.
		prmk.SetBool(ctx, valsetDirtyPath, false)

		return abci.ResponseEndBlock{
			ValidatorUpdates: diff,
		}
	}
}

// checkSessionRestrictions enforces gno.land session key restrictions.
// Two filters apply, in order:
//
//  1. sessionAlwaysDenied — auth/* and vm/add_package. Hard floor: never
//     permitted, even with "*" entry.
//  2. AllowPaths match — session's per-msg allow-list (validated at
//     create-time by handleMsgCreateSession; the "*" entry matches any).
//
// SpendLimit is enforced separately at the bank keeper layer; see ADR-001.
func checkSessionRestrictions(ctx sdk.Context, tx std.Tx) (sdk.Result, bool) {
	sa := ctx.Value(std.SessionAccountsContextKey{})
	if sa == nil {
		return sdk.Result{}, false
	}
	sessions := sa.(map[crypto.Address]std.DelegatedAccount)
	for _, msg := range tx.GetMsgs() {
		for _, signer := range msg.GetSigners() {
			sess, ok := sessions[signer]
			if !ok {
				continue
			}
			if sessionAlwaysDenied(msg) {
				return sdk.ABCIResultFromError(std.ErrSessionNotAllowed(fmt.Sprintf(
					"msg %s/%s cannot be signed by a session (privilege escalation)",
					msg.Route(), msg.Type(),
				))), true
			}
			entries, err := parseAllowPaths(sessionAllowPathsRaw(sess))
			if err != nil {
				// Handler validates at create-time; fail closed if seen at runtime.
				return sdk.ABCIResultFromError(std.ErrSessionNotAllowed(
					"invalid stored AllowPaths: " + err.Error())), true
			}
			if !anyEntryMatches(entries, msg) {
				return sdk.ABCIResultFromError(std.ErrSessionNotAllowed(fmt.Sprintf(
					"msg %s/%s%s not permitted by session AllowPaths %v",
					msg.Route(), msg.Type(), pkgPathSuffix(msg),
					sessionAllowPathsRaw(sess),
				))), true
			}
		}
	}
	return sdk.Result{}, false
}

// sessionAlwaysDenied reports whether a msg can never be signed by a session,
// regardless of AllowPaths. Auth is denied at the route level (forward-compat
// against new auth msgs); vm/add_package at the type level.
func sessionAlwaysDenied(msg std.Msg) bool {
	if msg.Route() == "auth" {
		return true
	}
	if msg.Route() == "vm" && msg.Type() == "add_package" {
		return true
	}
	return false
}

// sessionAllowPathsRaw extracts the AllowPaths slice via a local interface
// (concrete type is *GnoSessionAccount).
func sessionAllowPathsRaw(sess std.DelegatedAccount) []string {
	type pathRestricted interface{ GetAllowPaths() []string }
	if pr, ok := sess.(pathRestricted); ok {
		return pr.GetAllowPaths()
	}
	return nil
}

// anyEntryMatches reports whether any allow-list entry permits the msg.
func anyEntryMatches(entries []allowPathsEntry, msg std.Msg) bool {
	for _, e := range entries {
		if entryMatchesMsg(e, msg) {
			return true
		}
	}
	return false
}

// entryMatchesMsg reports whether a single entry permits the msg. Path
// matching uses the prefix rule (exact or sub-path guarded by "/") to
// preserve the prefix-attack defense (see TestSessionAllowPathsPrefixAttack).
func entryMatchesMsg(e allowPathsEntry, msg std.Msg) bool {
	if e.Wildcard {
		return true
	}
	if e.Route != msg.Route() || e.Type != msg.Type() {
		return false
	}
	if e.Path == "" {
		return true
	}
	type pkgPather interface{ GetPkgPath() string }
	pp, ok := msg.(pkgPather)
	if !ok {
		return false
	}
	path := pp.GetPkgPath()
	return path == e.Path || strings.HasPrefix(path, e.Path+"/")
}

// pkgPathSuffix renders " (path %q)" for path-bearing msgs and "" otherwise,
// for use in error messages.
func pkgPathSuffix(msg std.Msg) string {
	type pkgPather interface{ GetPkgPath() string }
	if pp, ok := msg.(pkgPather); ok {
		return fmt.Sprintf(" (path %q)", pp.GetPkgPath())
	}
	return ""
}
