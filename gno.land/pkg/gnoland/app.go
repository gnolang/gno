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

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
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
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
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
// genesisSignerFunding is what a genesis transaction's signer is minted when it has
// no account yet, so the transaction can pay for itself. Consensus-relevant: it
// lands in the supply counter, so it is named rather than inline to keep the mint
// and the test that checks the counter from drifting apart.
const genesisSignerFunding int64 = 10_000_000_000

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
	// B+32 store with the fast index (see storebptree.FastStoreConstructor).
	// Not state-compatible with IAVL databases: fresh chains and
	// export/import forks only.
	baseApp.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, cfg.DB)
	baseApp.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, cfg.DB)

	// Construct keepers.

	prmk := params.NewParamsKeeper(mainKey)
	// Denoms held inside the account object rather than in their own keys. Only
	// the gas denom qualifies: the account object is written on every transaction
	// anyway, so a balance there is free, and that is the whole justification.
	// Everything else is split out. See tm2/pkg/sdk/bank/balance.go.
	//
	// This list decides which physical keyspace a balance lives in, so it is
	// consensus-critical and must stay compiled in. Do NOT make it a flag, a
	// config field, or a governance param: two nodes with different lists route
	// the same denom to different keys and produce different app hashes, which is
	// a silent fork rather than a failed startup. Changing it is a coordinated
	// upgrade, and existing balances must be regenerated by replay or they strand
	// in the tier they were written to — see "Changing the allowlist" in
	// tm2/adr/pr6034_realm_denom_balance_keys.md.
	accountTierDenoms := []string{ugnot.Denom}

	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, ProtoGnoSessionAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName), mainKey, accountTierDenoms)
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
		// MsgAddPackage and MsgRun both compile caller-supplied Gno source,
		// and who is allowed to do that is decided from the signer.
		// `.app/simulate` is a public query that RUNS the messages, so
		// without this an unauthenticated caller could name any address and
		// have that authorization accepted. Restricted to the two
		// code-bearing messages, so gas estimation for every other message
		// type keeps working without a key.
		RequireSigForSimulate: txCarriesCode,
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
			// "Meters" means it does not nil the meter the way
			// acck.GetParams does; the read still costs nothing here,
			// because the ctx is on the infinite meter until
			// auth.SetGasMeter runs. See checkCodePolicy.
			// Kept, not discarded: checkCodePolicy below needs the same
			// struct, and re-reading it there would repeat the decode --
			// GetParams amino-unmarshals every field, and the three
			// allowlists bech32-decode every element, up to
			// maxAddressListLen each. The store reads would be cache hits
			// and cost no gas, but the decode is real work on every
			// code-bearing transaction. Nothing writes vm params between
			// here and there, so the value is identical.
			gasCfg := store.DefaultGasConfig()
			vmParams := vmk.GetParams(ctx)
			vmParams.ApplyToGasConfig(&gasCfg)
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
						// A genuine mint: guarded by the account not existing, so
						// there is no prior balance and this creates the coins.
						// Runs during genesis tx delivery, i.e. after the seed above.
						err := bankk.MintCoins(ctx, signer, std.Coins{std.NewCoin("ugnot", genesisSignerFunding)})
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
			// Code-submission authorization (gno.land layer). Same
			// res-preservation rule as above.
			if codeRes, codeAbort := checkCodePolicy(newCtx, tx, vmParams); codeAbort {
				return newCtx, codeRes, true
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

	// StrictReplay refuses to boot the chain if any non-skipped genesis tx
	// fails replay. Hardfork operators should enable this so a corrupted
	// genesis aborts InitChain loudly instead of producing a chain whose
	// AppHash silently diverges from the source.
	//
	// Skipped txs (those carrying metadata.Failed = true, which were
	// intentionally non-applied on the source chain) do not count as
	// failures.
	StrictReplay bool

	// SkipValoperCoverageAssertion turns off the hardfork-mode
	// AssertGenesisValopersConsistent auto-call. Useful for paths that
	// boot a chain with PastChainIDs set but a synthetic req.Validators
	// that won't match any seeded valoper profile — e.g. gnogenesis
	// fork test replaces genDoc.Validators with a fresh MockPV whose
	// signing addr is never registered, so the assertion would fire
	// spuriously. Production hardfork boots leave this false.
	SkipValoperCoverageAssertion bool

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
	// Gate the initial validator set against the pubkey-type allow-list (like the
	// EndBlocker does for runtime updates); panic to abort boot on a disallowed type.
	var allowedKeyTypes []string
	if req.ConsensusParams != nil && req.ConsensusParams.Validator != nil {
		allowedKeyTypes = req.ConsensusParams.Validator.PubKeyTypeURLs
	}
	if len(allowedKeyTypes) > 0 {
		for _, v := range req.Validators {
			if v.Power == 0 {
				continue // non-voting entry, never enters the active set
			}
			if keyType := amino.GetTypeURL(v.PubKey); !slices.Contains(allowedKeyTypes, keyType) {
				panic(fmt.Errorf(
					"genesis validator %s has disallowed pubkey type %s (allowed: %v)",
					v.Address.String(), keyType, allowedKeyTypes,
				))
			}
		}
	}

	// Note on sentinel scope: internalWriteCtxKey is set on the LOCAL
	// ictx variable, NOT on app.deliverState.ctx. baseapp.Deliver pulls
	// a fresh ctx via getContextForTx (tm2/pkg/sdk/baseapp.go:606-611),
	// so this sentinel does NOT propagate into genesis-tx execution —
	// a malicious genesis tx cannot manufacture a sentinel-bearing ctx
	// and write valset:current directly.
	ictx := ctx.WithValue(internalWriteCtxKey{}, true)
	cfg.prmk.SetStrings(ictx, valsetCurrentPath, abci.EncodeValidatorUpdates(abci.ValidatorUpdates(req.Validators)))
	// Mirror the allow-list into the params store for realms to read (genesis-immutable).
	cfg.prmk.SetStrings(ictx, valsetPubKeyTypesPath, allowedKeyTypes)

	// load app state. AppState may be nil mostly in some minimal testing setups;
	// so log a warning when that happens.
	txResponses, err := cfg.loadAppState(ctx, req.AppState, req.InitialHeight)
	if err != nil {
		// Surface loadAppState errors on the logger before returning. The
		// error is also propagated via ResponseInitChain.Error, but
		// tendermint's handshake does not surface that field — operators
		// otherwise see "Completed ABCI Handshake" with an empty appHash
		// and no indication that genesis replay never happened.
		ctx.Logger().Error("InitChainer: loadAppState failed", "error", err)
		return abci.ResponseInitChain{
			ResponseBase: abci.ResponseBase{
				Error: abci.StringError(err.Error()),
			},
		}
	}

	ctx.Logger().Debug("InitChainer: genesis transactions loaded",
		"elapsed", time.Since(start))

	// Hardfork-mode invariant: every signing addr in valset:current must
	// have a corresponding valoper profile in r/sys/validators/v3's
	// valoperCache. valoper-seed migration .jsonls produce these profiles;
	// the chain refuses to boot if any genesis validator is uncovered.
	//
	// Gated on (a) the hardfork signal (non-empty GnoGenesisState.PastChainIDs)
	// and (b) non-empty req.Validators. Fresh chains and dev/lazy-init/txtar
	// setups have empty PastChainIDs and trivially skip; hardfork tests
	// that set PastChainIDs without seeding validators also skip — there's
	// nothing to cover and the realm may not be loaded.
	//
	// Failure here is unconditionally fatal — independent of StrictReplay
	// — because a hardfork that boots with uncovered genesis validators
	// has lost the operator-keyed management plane for those validators.
	if cfg.shouldRunValoperCoverageAssertion(req) {
		if err := assertGenesisValopersConsistent(ctx, cfg.vmk, req); err != nil {
			// ResponseInitChain.Error is silently discarded by tm2:
			// consensus/replay.go:339-342 only inspects the Go-level
			// err from InitChainSync, and the call chain has no
			// recover() that would convert the proto Error field into
			// one — baseapp.InitChain (baseapp.go:320 + 359-361
			// short-circuit), localClient.InitChainSync
			// (local_client.go:192), and consensus.InitChainSync
			// (app_conn.go:65) are all pass-through. A panic
			// propagates up the boot goroutine (NewNode →
			// Handshaker.ReplayBlocks → InitChainSync) and crashes the
			// process — the only way to abort handshake on uncovered
			// genesis.
			panic(fmt.Errorf("genesis valoper coverage assertion failed: %w", err))
		}
	}

	// Done!
	return abci.ResponseInitChain{
		Validators:  req.Validators,
		TxResponses: txResponses,
	}
}

// shouldRunValoperCoverageAssertion combines the cfg override with the
// request-level gate. See SkipValoperCoverageAssertion for why the
// override exists.
func (cfg InitChainerConfig) shouldRunValoperCoverageAssertion(req abci.RequestInitChain) bool {
	return !cfg.SkipValoperCoverageAssertion && shouldAssertValoperCoverage(req)
}

// shouldAssertValoperCoverage gates the hardfork-mode v3 invariant
// check. Requires (1) non-empty PastChainIDs (authoritative hardfork
// signal — InitialHeight alone isn't, since dev/testnets use
// InitialHeight > 1 for non-hardfork scenarios) and (2) non-empty
// req.Validators (otherwise the check is trivial and would needlessly
// require v3 to be loaded).
func shouldAssertValoperCoverage(req abci.RequestInitChain) bool {
	if len(req.Validators) == 0 {
		return false
	}
	state, ok := req.AppState.(GnoGenesisState)
	if !ok {
		return false
	}
	return len(state.PastChainIDs) > 0
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

func (cfg InitChainerConfig) loadAppState(ctx sdk.Context, appState any, reqInitialHeight int64) ([]abci.ResponseDeliverTx, error) {
	switch state := appState.(type) {
	case GnoGenesisState:
		return cfg.applyInMemoryAppState(ctx, state, reqInitialHeight)
	case *GenesisStateRef:
		return cfg.applyStreamingAppState(ctx, state)
	default:
		return nil, fmt.Errorf("invalid AppState of type %T", appState)
	}
}

func (cfg InitChainerConfig) applyInMemoryAppState(ctx sdk.Context, state GnoGenesisState, reqInitialHeight int64) ([]abci.ResponseDeliverTx, error) {
	// If GnoGenesisState.InitialHeight is set, it must match the authoritative
	// GenesisDoc.InitialHeight (which comes in via req.InitialHeight). These
	// fields are duplicated so tooling can read the app-level one; if they
	// diverge, the genesis file is malformed.
	if state.InitialHeight != 0 && state.InitialHeight != reqInitialHeight {
		return nil, fmt.Errorf(
			"InitialHeight mismatch: GnoGenesisState.InitialHeight=%d, GenesisDoc.InitialHeight=%d",
			state.InitialHeight, reqInitialHeight,
		)
	}

	if err := validateGasReplayMode(state.GasReplayMode); err != nil {
		return nil, err
	}

	// Preflight: every (account-number, address) pair claimed by SignerInfo
	// must be unique, and must not collide with a balance-init account at a
	// different address. NewAccountWithUncheckedNumber does NOT verify this
	// at write-time; a duplicate accNum used with a different address would
	// silently zero the original account's balance. Failing here surfaces a
	// malformed genesis loudly before any state is mutated.
	if err := validateSignerInfo(state); err != nil {
		return nil, err
	}

	if len(state.PastChainIDs) > 0 {
		ctx.Logger().Info("Chain upgrade genesis replay",
			"past_chain_ids", state.PastChainIDs,
			"initial_height", reqInitialHeight,
		)
	}

	cfg.bankk.InitGenesis(ctx, state.Bank)
	for _, bal := range state.Balances {
		cfg.applyBalance(ctx, bal)
	}
	cfg.seedSupply(ctx)
	// The account keeper's initial genesis state must be set after genesis
	// accounts are created in account keeeper with genesis balances
	cfg.acck.InitGenesis(ctx, state.Auth)
	cfg.applyUnrestrictedAddrs(ctx, state.Auth.Params.UnrestrictedAddrs)
	cfg.vmk.InitGenesis(ctx, state.VM)

	ctx = cfg.installAuthParams(ctx)

	// Replay genesis txs.
	txResponses := make([]abci.ResponseDeliverTx, 0, len(state.Txs))
	report := newReplayReport(state.GasReplayMode)

	for txIdx, tx := range state.Txs {
		resp, _ := cfg.deliverGenesisTx(ctx, txIdx, tx, state.PastChainIDs, state.GasReplayMode, report)
		txResponses = append(txResponses, resp)
	}

	if reqInitialHeight > 1 {
		ctx.Logger().Info("Genesis replay complete, chain will start from initial height",
			"initial_height", reqInitialHeight,
		)
	}

	report.emit(ctx.Logger())
	cfg.enforceStrictReplay(report)

	return txResponses, nil
}

// enforceStrictReplay aborts the boot when a genesis transaction failed and the
// operator asked for strictness.
//
// It panics rather than returning an error, for the reason spelled out at the
// valoper coverage assertion above: an error returned from here reaches
// ResponseInitChain.Error and stops there. localClient.InitChainSync returns a
// nil Go error regardless, and the handshake inspects only that — so the field
// is populated and never read. The node would boot anyway, and a fork could come
// up missing packages while logging that it had them.
//
// Panicking propagates up the boot goroutine (NewNode -> Handshaker.ReplayBlocks
// -> InitChainSync) and crashes the process, which is what "refusing to boot"
// has to mean. Opt-in via cfg.StrictReplay, so nothing that boots today starts
// crashing.
func (cfg InitChainerConfig) enforceStrictReplay(report *replayReport) {
	if !cfg.StrictReplay {
		return
	}
	if n := report.FailedCount(); n > 0 {
		panic(fmt.Errorf(
			"strict replay: %d genesis tx(s) failed; chain refusing to boot "+
				"(inspect the per-failure 'Genesis replay failure' log lines for details)",
			n,
		))
	}
}

// applyStreamingAppState mirrors applyInMemoryAppState but pulls each
// genesis element from on-disk JSONL via the ref's iterators, keeping peak
// heap bounded to a single element regardless of total size. Small sibling
// fields (auth, bank, vm) are eagerly amino-decoded out of the envelope.
func (cfg InitChainerConfig) applyStreamingAppState(ctx sdk.Context, ref *GenesisStateRef) ([]abci.ResponseDeliverTx, error) {
	var bankState bank.GenesisState
	if err := decodeSmallField(ref, appStateBankKey, &bankState); err != nil {
		return nil, err
	}
	var authState auth.GenesisState
	if err := decodeSmallField(ref, appStateAuthKey, &authState); err != nil {
		return nil, err
	}
	var vmState vm.GenesisState
	if err := decodeSmallField(ref, appStateVMKey, &vmState); err != nil {
		return nil, err
	}

	cfg.bankk.InitGenesis(ctx, bankState)
	for line, err := range ref.IterBalances(ctx.Context()) {
		if err != nil {
			return nil, fmt.Errorf("iter balances: %w", err)
		}
		var bal Balance
		if err := amino.UnmarshalJSON(line, &bal); err != nil {
			return nil, fmt.Errorf("decode balance: %w", err)
		}
		cfg.applyBalance(ctx, bal)
	}
	cfg.seedSupply(ctx)
	cfg.acck.InitGenesis(ctx, authState)
	cfg.applyUnrestrictedAddrs(ctx, authState.Params.UnrestrictedAddrs)
	cfg.vmk.InitGenesis(ctx, vmState)

	ctx = cfg.installAuthParams(ctx)

	// Decode hardfork replay parameters from the small-field envelope.
	var pastChainIDs []string
	if raw, ok := ref.SmallField("past_chain_ids"); ok {
		if err := amino.UnmarshalJSON(raw, &pastChainIDs); err != nil {
			return nil, fmt.Errorf("decode past_chain_ids: %w", err)
		}
	}
	var gasReplayMode string
	if raw, ok := ref.SmallField("gas_replay_mode"); ok {
		if err := amino.UnmarshalJSON(raw, &gasReplayMode); err != nil {
			return nil, fmt.Errorf("decode gas_replay_mode: %w", err)
		}
	}

	if err := validateGasReplayMode(gasReplayMode); err != nil {
		return nil, err
	}

	report := newReplayReport(gasReplayMode)
	txResponses := make([]abci.ResponseDeliverTx, 0, ref.TxCount())
	txIdx := 0
	for line, err := range ref.IterTxs(ctx.Context()) {
		if err != nil {
			return nil, fmt.Errorf("iter txs: %w", err)
		}
		var tx TxWithMetadata
		if err := amino.UnmarshalJSON(line, &tx); err != nil {
			return nil, fmt.Errorf("decode tx: %w", err)
		}
		resp, _ := cfg.deliverGenesisTx(ctx, txIdx, tx, pastChainIDs, gasReplayMode, report)
		txResponses = append(txResponses, resp)
		txIdx++
	}

	report.emit(ctx.Logger())
	cfg.enforceStrictReplay(report)

	return txResponses, nil
}

func decodeSmallField(ref *GenesisStateRef, key string, into any) error {
	raw, ok := ref.SmallField(key)
	if !ok {
		return fmt.Errorf("missing app_state.%s in genesis cache", key)
	}
	if err := amino.UnmarshalJSON(raw, into); err != nil {
		return fmt.Errorf("decode app_state.%s: %w", key, err)
	}
	return nil
}

// applyBalance creates the genesis account for bal and writes its balance.
//
// A balance list may name one address twice — they are assembled from several
// sources, and the integration harness appends to a loaded default set. That is
// handled by being idempotent in effect rather than in mechanism: the account is
// recreated (which it must be, so a plain entry after a vesting one clears the
// schedule) and SetCoins is replace-all, so the last entry wins. The only trace a
// repeat leaves is a gap in account numbering, which is harmless — numbers must be
// unique and below the global counter, not dense.
//
// Reusing the account to close that gap was tried and reverted: it shifts every
// later account number, which changes the genesis state of any chain whose balance
// file repeats an address, for no correctness gain.
// seedSupply rebuilds the supply counter from the balances genesis just wrote.
//
// A sweep rather than an incremental hook: applyBalance writes the account object with
// the full pre-split amount before calling SetCoins, so SetCoins reads old == new and
// any delta would be zero for every vesting account — and that pre-write cannot be
// removed, since the vesting constructors validate OriginalVesting against it.
//
// Both genesis paths call this. It is a method rather than two inline calls so that a
// step required by one path cannot be added to the other alone; that asymmetry is
// exactly what left the streaming path's seeding untested until it was noticed.
func (cfg InitChainerConfig) seedSupply(ctx sdk.Context) {
	cfg.bankk.RecomputeSupply(ctx)
}

func (cfg InitChainerConfig) applyBalance(ctx sdk.Context, bal Balance) {
	if bal.IsVesting() {
		baseAcc := std.BaseAccount{
			Address:       bal.Address,
			Coins:         bal.Amount,
			AccountNumber: cfg.acck.GetNextAccountNumber(ctx),
		}
		var acc std.Account
		var err error
		switch bal.Vesting.Type {
		case std.VestingDelayed:
			acc, err = std.NewDelayedVestingAccount(&baseAcc, *bal.Vesting)
		default: // VestingContinuous (empty string) — linear vesting
			acc, err = std.NewContinuousVestingAccount(&baseAcc, *bal.Vesting)
		}
		if err != nil {
			panic(fmt.Errorf("invalid vesting account for %s: %w", bal.Address, err))
		}
		cfg.acck.SetAccount(ctx, acc)
	} else {
		acc := cfg.acck.NewAccountWithAddress(ctx, bal.Address)
		cfg.acck.SetAccount(ctx, acc)
	}
	if err := cfg.bankk.SetCoins(ctx, bal.Address, bal.Amount); err != nil {
		// Name the address and the amount. This aborts genesis, and the causes
		// include a denom that is too long or malformed — so an operator forking a
		// chain needs to know which entry to fix. std.ErrInvalidCoins carries the
		// coins only under %+v, and a panic renders its value with %v, so a bare
		// panic(err) here says nothing but "invalid coins error".
		panic(fmt.Errorf("invalid genesis balance for %s (%s): %w", bal.Address, bal.Amount, err))
	}
}

// applyUnrestrictedAddrs flips the token-lock whitelist bit on each
// unrestricted address. Each address must already exist as a genesis
// account (i.e. must have appeared in balances), otherwise the verifier
// can't verify the chain's unrestricted set.
func (cfg InitChainerConfig) applyUnrestrictedAddrs(ctx sdk.Context, addrs []crypto.Address) {
	for _, addr := range addrs {
		acc := cfg.acck.GetAccount(ctx, addr)
		if acc == nil {
			panic(fmt.Errorf("unrestricted address must be one of the genesis accounts: invalid account %q", addr))
		}
		accr := acc.(*GnoAccount)
		accr.SetTokenLockWhitelisted(true)
		cfg.acck.SetAccount(ctx, acc)
	}
}

func (cfg InitChainerConfig) installAuthParams(ctx sdk.Context) sdk.Context {
	params := cfg.acck.GetParams(ctx)
	ctx = ctx.WithValue(auth.AuthParamsContextKey{}, params)
	auth.InitChainer(ctx, cfg.gpk, params.InitialGasPrice)
	return ctx
}

// deliverGenesisTx applies all hardfork-aware context overrides and delivers a
// single genesis tx. Returns the response and a skip flag (true when the tx was
// a known-failed historical tx that must not be re-executed).
func (cfg InitChainerConfig) deliverGenesisTx(
	ctx sdk.Context,
	txIdx int,
	tx TxWithMetadata,
	pastChainIDs []string,
	gasReplayMode string,
	report *replayReport,
) (resp abci.ResponseDeliverTx, skip bool) {
	stdTx := tx.Tx
	metadata := tx.Metadata

	var ctxFn sdk.ContextFn

	// Check if there is metadata associated with the tx
	if metadata != nil {
		ctxFn = func(ctx sdk.Context) sdk.Context {
			header := ctx.BlockHeader().(*bft.Header).Copy()
			if metadata.Timestamp != 0 {
				header.Time = time.Unix(metadata.Timestamp, 0)
			}
			if metadata.BlockHeight > 0 {
				header.Height = metadata.BlockHeight
			}

			ctx = ctx.WithBlockHeader(header)

			// For historical txs (BlockHeight > 0), override the chain ID
			// for signature verification using the per-tx ChainID, provided
			// it is in the genesis allowlist. This allows replaying txs from
			// multiple past chains during a hard fork.
			if metadata.BlockHeight > 0 && metadata.ChainID != "" && isPastChainID(pastChainIDs, metadata.ChainID) {
				ctx = ctx.WithChainID(metadata.ChainID)
			}

			// GasReplayMode="source": bypass the new VM's gas meter for
			// historical txs so outcomes match the source chain even when
			// gas metering changed.
			if gasReplayMode == "source" && metadata.BlockHeight > 0 {
				ctx = ctx.WithValue(auth.SkipGasMeteringKey{}, true)
			}

			return ctx
		}
	}

	// Genesis-mode txs (no metadata) were signed with the original chain
	// ID. During a hardfork (PastChainIDs is set), verify their
	// signatures against the original chain ID. Migration txs
	// (metadata != nil with BlockHeight == 0) carry their own per-tx
	// settings via metadata and are handled in the first branch above;
	// excluding them here prevents the previous overwrite bug where
	// this assignment stomped the metadata-driven Timestamp override.
	//
	// Compose with any prior ctxFn so future broadening of the
	// predicate cannot silently regress.
	if metadata == nil && len(pastChainIDs) > 0 {
		originalChainID := pastChainIDs[0]
		prev := ctxFn
		ctxFn = func(ctx sdk.Context) sdk.Context {
			if prev != nil {
				ctx = prev(ctx)
			}
			return ctx.WithChainID(originalChainID)
		}
	}

	// For historical txs with signer metadata, force-set account state
	// so signature verification succeeds even if prior txs diverged.
	// Uses pre-tx sequence — the value the signature was signed with.
	//
	// Invariant: SignerInfo is only populated by the export tool for historical
	// txs (BlockHeight > 0). Genesis-mode txs (BlockHeight == 0) must never
	// carry SignerInfo — if they did, the force-set would corrupt fresh account
	// state. The BlockHeight > 0 guard enforces this.
	if metadata != nil && metadata.BlockHeight > 0 && len(metadata.SignerInfo) > 0 {
		for _, si := range metadata.SignerInfo {
			acc := cfg.acck.GetAccount(ctx, si.Address)
			if acc == nil {
				// Account doesn't exist yet, create with specific account
				// number, bypassing the auto-increment counter. Uniqueness
				// of (Address, AccountNum) is enforced by the
				// validateSignerInfo preflight above; the keeper does not
				// re-check.
				acc = cfg.acck.NewAccountWithUncheckedNumber(ctx, si.Address, si.AccountNum)
			} else {
				acc.SetAccountNumber(si.AccountNum)
			}
			acc.SetSequence(si.Sequence)
			cfg.acck.SetAccount(ctx, acc)
		}
	}

	// Failed txs: pre-tx sequence already set above. Skip execution —
	// re-executing failed txs could cause double spends or unexpected
	// behavior if the VM fix makes them succeed. The next tx's force-set
	// will handle the correct sequence state.
	// Response carries an explicit error so downstream consumers
	// (indexers, explorers) don't mistake a skipped failed tx for a
	// successful one.
	if metadata != nil && metadata.Failed {
		report.record(txIdx, metadata, 0, 0, replayCategorySkippedFailed, nil)
		return abci.ResponseDeliverTx{
			ResponseBase: abci.ResponseBase{
				Error: abci.StringError("replay skipped: tx failed on source chain"),
				Log:   "genesis replay: skipped failed tx from source chain",
			},
		}, true
	}

	// Every tx delivered during InitChain replay is a genesis replay tx.
	// Mark the ctx so the ante's genesis signature-skip also covers the
	// historical/patched txs whose BlockHeight is overridden above (see
	// auth.GenesisReplayKey). Composes with any ctxFn built above, and
	// only affects sig verification when the node ran with
	// --skip-genesis-sig-verification.
	prev := ctxFn
	ctxFn = func(ctx sdk.Context) sdk.Context {
		if prev != nil {
			ctx = prev(ctx)
		}
		return ctx.WithValue(auth.GenesisReplayKey{}, true)
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

	report.recordDeliverResult(txIdx, metadata, res)
	cfg.GenesisTxResultHandler(ctx, stdTx, res)
	return abci.ResponseDeliverTx{
		ResponseBase: res.ResponseBase,
		GasWanted:    res.GasWanted,
		GasUsed:      res.GasUsed,
	}, false
}

// validatorsV3PkgPath is the realm whose AssertGenesisValopersConsistent
// invariant gates hardfork-mode boot.
const (
	validatorsV3PkgPath       = "gno.land/r/sys/validators/v3"
	assertGenesisValopersFunc = "AssertGenesisValopersConsistent"
	missingV3PkgPanicSubstr   = "unexpected node with location " + validatorsV3PkgPath
)

// assertGenesisValopersConsistent invokes the v3 assertion via the VM
// keeper directly (no tx pipeline, no AnteHandler, no fee accounting).
//
// Caller is the first genesis validator's address; the call sends zero
// coins so no account need exist for it.
//
// If v3 isn't deployed, the underlying gnostore lookup panics outside
// vmk.Call's recover. The defer below catches that case and skips with
// a warning — production hardforks always deploy v3, and if they
// don't, the valoper-seed Register migration txs panic loudly anyway.
func assertGenesisValopersConsistent(ctx sdk.Context, vmk vm.VMKeeperI, req abci.RequestInitChain) (err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprint(r)
			if strings.Contains(msg, missingV3PkgPanicSubstr) {
				ctx.Logger().Warn(
					"valoper coverage assertion skipped: v3 not deployed in genesis",
					"detail", msg,
				)
				err = nil
				return
			}
			err = fmt.Errorf("%s", msg)
		}
	}()
	msg := vm.MsgCall{
		Caller:  req.Validators[0].Address,
		PkgPath: validatorsV3PkgPath,
		Func:    assertGenesisValopersFunc,
	}
	vmCtx := vmk.MakeGnoTransactionStore(ctx)
	if _, e := vmk.Call(vmCtx, msg); e != nil {
		return e
	}
	vmk.CommitGnoTransactionStore(vmCtx)
	return nil
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

// isPastChainID reports whether chainID is present in the pastChainIDs allowlist.
func isPastChainID(pastChainIDs []string, chainID string) bool {
	return slices.Contains(pastChainIDs, chainID)
}

// validateSignerInfo scans every SignerInfo entry across all txs and
// rejects the genesis if two different addresses claim the same account
// number, OR if a SignerInfo claims an account number already reserved by a
// balance-init account at a different address. NewAccountWithUncheckedNumber
// (the keeper primitive replay uses) does not perform this check at
// write-time, so the invariant is enforced here, before any state mutates.
//
// genesis-mode txs (BlockHeight == 0) carry no SignerInfo by invariant of
// the export tool, but we still skip them defensively.
func validateSignerInfo(state GnoGenesisState) error {
	// Map: account number -> address that reserves it.
	numToAddr := map[uint64]crypto.Address{}

	// Treat balance-init accounts as reserving accNum=N, where N is assigned
	// by the auto-increment counter in the order they appear in
	// state.Balances. After all balances are processed, the counter is
	// len(state.Balances). Any SignerInfo with accNum < len(state.Balances)
	// must therefore reference one of those addresses (or it would collide
	// with a different balance-init address).
	for i, bal := range state.Balances {
		numToAddr[uint64(i)] = bal.Address
	}

	for txIdx, tx := range state.Txs {
		if tx.Metadata == nil {
			continue
		}
		for siIdx, si := range tx.Metadata.SignerInfo {
			existing, seen := numToAddr[si.AccountNum]
			if seen && existing != si.Address {
				return fmt.Errorf(
					"genesis SignerInfo collision at txs[%d].SignerInfo[%d]: "+
						"account number %d already assigned to %s, cannot reassign to %s",
					txIdx, siIdx, si.AccountNum, existing, si.Address,
				)
			}
			numToAddr[si.AccountNum] = si.Address
		}
	}
	return nil
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

// txCodeMsgSigners collects the signers of tx's code-bearing messages, split by
// kind. MsgAddPackage and MsgRun are the only two messages that hand the chain
// Gno source to compile, and both gates below key off them, so this
// classification lives in one place rather than being re-walked per caller.
//
// Per MESSAGE, not per transaction, and that distinction is load-bearing. A
// MsgRun's signer is the address whose code executes; a MsgAddPackage's signer
// is the address that deploys. Someone who co-signs an unrelated message in the
// same tx — a bank send, say — is not submitting code, so requiring them to
// hold code-submission rights would refuse legitimate bundles and, worse,
// report the bystander as "not authorized to send MsgRun" when they sent no
// such thing. Using tx.GetSigners() here reads naturally and is wrong.
func txCodeMsgSigners(tx std.Tx) (addPkgSigners, runSigners []crypto.Address) {
	for _, msg := range tx.GetMsgs() {
		switch msg.(type) {
		case vm.MsgAddPackage:
			addPkgSigners = append(addPkgSigners, msg.GetSigners()...)
		case vm.MsgRun:
			runSigners = append(runSigners, msg.GetSigners()...)
		}
	}
	return addPkgSigners, runSigners
}

// txCarriesCode reports whether tx carries any message that hands the chain Gno
// source to compile.
//
// Wired into auth.AnteOptions.RequireSigForSimulate. Both code-bearing messages
// are authorized from their signer and both compile and execute the source they
// carry, so on the simulate path — a public query that runs the messages — an
// unverified signature means the gate reads an attacker-chosen address.
//
// It covers MsgAddPackage as well as MsgRun, and did not always. An earlier
// revision matched MsgRun alone, from when that was the only gated message:
// code_submission_policy existed but nothing enforced it. Adding that
// enforcement made MsgAddPackage gated too and left this predicate behind — so
// under "permissioned", the policy whose entire purpose is keeping strangers
// off the type checker, anyone could name a listed submitter, attach arbitrary
// bytes as a signature, and drive a full type-check plus init() per query, free.
//
// The cost: keyless gas estimation no longer works for either message. gnokey
// signs a second transaction for simulation and is unaffected; other clients
// that estimate before signing must supply a real signature.
// MsgEnablePackage and MsgDisablePackage are covered too, and are NOT part of
// txCodeMsgSigners: that function feeds the code_submitters/run_submitters
// allowlists, which have no authority over enabling. Their gate is
// params.PkgApprovers, checked in the keeper against msg.Approver -- a
// caller-supplied field, exactly like the signers above. So the same reasoning
// applies: on an unverified simulate, anyone may name the real approver, attach
// arbitrary bytes as a signature, and have the chain type-check and init() an
// already-parked package for free. That the bytes are already stored makes it
// worse rather than better, since under "inert" anyone may park them.
//
// MsgRejectPackage is deliberately NOT covered. It is authorized from its own
// payload like the two above, but it executes nothing: it reads the parked
// blob, parses its gnomod.toml and deletes it. The harm the others invite --
// driving a free type-check and init() per query -- has no analogue, and the
// blob decode it does do is already reachable anonymously through
// vm/qpkgmeta_json. Requiring a signature to simulate it would cost keyless
// estimation for no gain.
//
// Enumerated by type rather than derived from the allowlist scan, so that adding
// a message which is authorized from its own payload and executes code is a
// deliberate decision here rather than a silent omission.
func txCarriesCode(tx std.Tx) bool {
	for _, msg := range tx.GetMsgs() {
		switch msg.(type) {
		case vm.MsgAddPackage, vm.MsgRun, vm.MsgEnablePackage, vm.MsgDisablePackage:
			return true
		}
	}
	return false
}

// checkCodePolicy enforces who may submit code, from the tx signers, before any
// message reaches the VM. It implements this (policy x message type) matrix:
//
//	policy          | vm/add_package          | vm/run
//	----------------+-------------------------+---------------------------------
//	permissionless  | allow                   | run_submitters, if non-empty
//	permissioned    | require code_submitters | run_submitters, if non-empty
//	inert           | allow (stored inert)    | run_submitters, if non-empty
//
// "if non-empty" is the whole of run_submitters' opt-in: an unset list is the
// gate switched off, not a chain with MsgRun disabled. See checkRunSubmitters.
//
// Two sibling rules, evaluated independently below (checkRunSubmitters and
// checkCodeSubmissionPolicy). They share only inputs — one params read, one
// signer scan, one replay carve-out — never control flow.
//
// Keeping them separate is the point. Gating both behind one
// `policy != permissionless` test reads as equivalent and is not: it makes
// add_package answer to code_submitters under "inert" too, which contradicts
// what inert is for. Anyone may submit; approval happens later.
//
// MsgRun's column does not vary with policy. "inert" defers MsgAddPackage's
// type-check, but MsgRun still type-checks and executes immediately under every
// policy value, so no policy makes it safe. It is also the only code-bearing
// message with no other gate: MsgAddPackage clears a namespace check and a CLA
// check, while MsgRun's path is forced to /e/<caller>/run and so has no
// namespace to check against. Hence a separate, always-on list.
//
// Enforced here rather than in the keeper, deliberately. This is the only layer
// that can refuse a tx during CheckTx and keep it out of the mempool, and it is
// where gno.land's other signer-derived policy already lives
// (checkSessionRestrictions, directly below).
//
// The keeper does consult IsGenesisReplay, for a different question: this
// decides whether a tx is ADMITTED, the keeper decides what a message DOES.
// Replay must not re-authorize a historical signer, and separately must
// reproduce what the source chain's policy made the message do. Neither answer
// substitutes for the other.
//
// Authorization is read from the signers, which is sound on the simulate path
// only because RequireSigForSimulate (see txCarriesCode) makes the auth ante
// verify MsgRun signatures there too. Without that, `.app/simulate` would let
// an unauthenticated caller name any address and have it accepted here.
//
// Session txs are authorized through their MASTER, not the session key.
// MsgRun.GetSigners() returns msg.Caller — the master address — and that is
// what txCodeMsgSigners collects. The session key lives in
// Signature.SessionAddr, a separate field this check never reads (see the
// contract on std.SessionAccountsContextKey, which spells out that the map is
// keyed on "the master account address returned by msg.GetSigners(), NOT the
// session pubkey address").
//
// So authorization here is a property of the master account, and a session key
// permitted to send vm/run by its AllowPaths inherits it. That is a two-layer
// grant, not a fail-closed one: AllowPaths decides whether the session may
// carry a MsgRun at all, and run_submitters decides whether its master may run
// code. Listing a session address here would have no effect.
func checkCodePolicy(ctx sdk.Context, tx std.Tx, params vm.Params) (sdk.Result, bool) {
	addPkgSigners, runSigners := txCodeMsgSigners(tx)
	if len(addPkgSigners) == 0 && len(runSigners) == 0 {
		// No code-bearing message: skip the params read entirely, so no other
		// message type pays for this check.
		return sdk.Result{}, false
	}

	// Every tx delivered during InitChain is exempt, not only replayed history.
	//
	// A fresh chain seeds its first GovDAO members with a genesis MsgRun, before
	// any allowlist could name them, so narrowing this to the hardfork case
	// (requiring metadata, or BlockHeight > 0) breaks bootstrap on any chain
	// shipping a non-empty run_submitters. TestChainUpgradeGenesisReplay pins
	// both shapes.
	//
	// The hardfork case is the other half: deliverGenesisTx replays historical
	// txs through this same ante with BlockHeight > 0, after InitGenesis has
	// installed the NEW params — so without this carve-out a hardfork would refuse to replay
	// its own history the moment either list fails to contain a historical
	// signer, and would come up missing every package those signers deployed.
	// — or, under StrictReplay, refuse to boot at all. Keyed on the context
	// value rather than BlockHeight, which replay deliberately does not hold
	// at 0.
	if auth.IsGenesisReplay(ctx) {
		return sdk.Result{}, false
	}

	// params is passed in rather than read here. The ante already reads the
	// whole struct near the top of the closure, to build the gas config, on the
	// throwaway meter installed before auth.SetGasMeter replaces it. Reading it
	// again would cost no gas -- baseapp keeps one cache wrap across the ante
	// and the handlers, and cacheStore.Get charges nothing for a hit -- but it
	// would repeat the decode, and GetParams bech32-decodes every entry of all
	// three allowlists. Threading the value keeps that at one decode per tx and
	// removes the ordering constraint the old comment here had to explain.
	return codePolicyResult(addPkgSigners, runSigners, params)
}

// codePolicyResult is the decision half of checkCodePolicy, split from the
// context and keeper reads so the whole matrix above is testable without
// standing up a VM keeper.
func codePolicyResult(addPkgSigners, runSigners []crypto.Address, params vm.Params) (sdk.Result, bool) {
	// The two rules are siblings, evaluated independently. Neither is nested in
	// the other, which is the shape that matters: phase 1 gated add_package and
	// run together behind one policy test, so adding `inert` to the enum
	// silently changed what happened to both. They share only the inputs —
	// params read once by the caller, signers scanned once — not control flow.
	if res, abort := checkRunSubmitters(runSigners, params); abort {
		return res, abort
	}
	return checkCodeSubmissionPolicy(addPkgSigners, params)
}

// checkRunSubmitters enforces the run_submitters allowlist on MsgRun signers.
//
// Policy-independent: it takes no policy argument because no policy value makes
// MsgRun safe. "inert" defers MsgAddPackage's type-check but leaves MsgRun
// type-checking and executing immediately, so the policy has no bearing on the
// hazard. Not accepting a policy parameter is deliberate — it keeps that
// independence structural rather than a branch someone can later "simplify".
//
// An EMPTY list means the allowlist is off and anyone may send MsgRun, which is
// the pre-existing behaviour and therefore the zero value's meaning. This is the
// opposite of code_submitters, and the asymmetry is not an oversight:
// code_submitters is only consulted once code_submission_policy has been
// explicitly moved to "permissioned", so its empty state is a half-configured
// opt-in and refusing is right. run_submitters has no such switch — it is read
// on every MsgRun from the moment the field exists — so a fail-closed empty
// value would silently disable MsgRun on every chain that upgrades without
// touching genesis, including the existing ones. GovDAO proposal CREATION is
// MsgRun-only (a ProposalRequest carries a func value, which MsgCall cannot
// marshal), so that failure mode takes governance with it and cannot be
// repaired in band.
//
// Turning the gate ON is therefore an explicit act: list at least one address.
// The cost is that a chain cannot express "nobody may MsgRun" through this
// param, which is not a configuration anyone has asked for.
func checkRunSubmitters(signers []crypto.Address, params vm.Params) (sdk.Result, bool) {
	if len(signers) == 0 || len(params.RunSubmitters) == 0 {
		return sdk.Result{}, false
	}
	return requireListed(signers, params.RunSubmitters,
		"send MsgRun", "run_submitters")
}

// checkCodeSubmissionPolicy enforces code_submission_policy on MsgAddPackage
// signers. Named after #5885's function, which this replaces, so the two are
// easy to diff.
func checkCodeSubmissionPolicy(signers []crypto.Address, params vm.Params) (sdk.Result, bool) {
	if len(signers) == 0 {
		return sdk.Result{}, false
	}
	// A switch with an explicit default, not `== permissioned`, so an
	// unrecognised policy string refuses rather than allows. Params.Validate
	// makes that unreachable today, but this file promises fail-closed
	// everywhere else and an equality test quietly promises the opposite the
	// moment a new value is added to the enum without being handled here.
	switch params.CodeSubmissionPolicy {
	case vm.CodeSubmissionPolicyPermissionless,
		// "inert" accepts from anyone by design: nothing is compiled or run at
		// submit, and approval is a separate, gated step.
		vm.CodeSubmissionPolicyInert:
		return sdk.Result{}, false
	case vm.CodeSubmissionPolicyPermissioned:
		return requireListed(signers, params.CodeSubmitters,
			"submit packages", "code_submitters")
	default:
		return sdk.ABCIResultFromError(std.ErrUnauthorized(fmt.Sprintf(
			"unknown vm code_submission_policy %q; refusing to submit code",
			params.CodeSubmissionPolicy))), true
	}
}

// requireListed aborts unless every signer appears in allowed. An empty allowed
// list refuses everyone, so callers that treat "unset" as "off" must test for
// emptiness themselves before calling — checkRunSubmitters does, and documents
// why; checkCodeSubmissionPolicy deliberately does not, because it is only
// reached once the policy has been explicitly moved to "permissioned".
func requireListed(signers []crypto.Address, allowed []crypto.Address, action, param string) (sdk.Result, bool) {
	for _, signer := range signers {
		if !slices.Contains(allowed, signer) {
			return sdk.ABCIResultFromError(std.ErrUnauthorized(fmt.Sprintf(
				"%s is not authorized to %s; see the vm %s param",
				signer.String(), action, param))), true
		}
	}
	return sdk.Result{}, false
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
	if msg.Route() == "vm" {
		switch msg.Type() {
		case "add_package":
			return true
		case "enable_package", "disable_package", "reject_package":
			// Approver authority, and it cannot be scoped down. A session's
			// AllowPaths are matched via GetPkgPath(), which only MsgCall
			// implements -- so no path-scoped entry can ever match these, and
			// the only way to grant them would be the "*" wildcard, handing a
			// session its master's full approver power. Deny outright.
			//
			// reject_package is here even though its gate is approver OR
			// creator: a creator withdrawing their own submission is
			// legitimate, but there is no way to grant a session that half
			// alone, and the wildcard that would grant it also hands over the
			// approver half -- which can empty the whole approval queue.
			return true
		}
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
