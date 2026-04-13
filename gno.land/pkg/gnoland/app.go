// Package gnoland contains the bootstrapping code to launch a gno.land node.
package gnoland

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"strconv"
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
		PruneStrategy:              types.PruneNothingStrategy,
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
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount)
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
			// Override auth params.
			ctx = ctx.WithValue(auth.AuthParamsContextKey{}, acck.GetParams(ctx))

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

	// load app state. AppState may be nil mostly in some minimal testing setups;
	// so log a warning when that happens.
	txResponses, err := cfg.loadAppState(ctx, req.AppState, req.InitialHeight)
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
}

func (cfg InitChainerConfig) loadAppState(ctx sdk.Context, appState any, reqInitialHeight int64) ([]abci.ResponseDeliverTx, error) {
	state, ok := appState.(GnoGenesisState)
	if !ok {
		return nil, fmt.Errorf("invalid AppState of type %T", appState)
	}

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

	if len(state.PastChainIDs) > 0 {
		ctx.Logger().Info("Chain upgrade genesis replay",
			"past_chain_ids", state.PastChainIDs,
			"initial_height", reqInitialHeight,
		)
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
	report := newReplayReport(state.GasReplayMode)

	// Run genesis txs
	for txIdx, tx := range state.Txs {
		var (
			stdTx    = tx.Tx
			metadata = tx.Metadata

			ctxFn sdk.ContextFn
		)

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
				if metadata.BlockHeight > 0 && metadata.ChainID != "" && isPastChainID(state.PastChainIDs, metadata.ChainID) {
					ctx = ctx.WithChainID(metadata.ChainID)
				}

				// GasReplayMode="source": bypass the new VM's gas meter for
				// historical txs so outcomes match the source chain even when
				// gas metering changed.
				if state.GasReplayMode == "source" && metadata.BlockHeight > 0 {
					ctx = ctx.WithValue(auth.SkipGasMeteringKey{}, true)
				}

				return ctx
			}
		}

		// Genesis-mode txs (no metadata or BlockHeight == 0) were signed with
		// the original chain ID. During a hardfork (PastChainIDs is set), we
		// need to verify their signatures against the original chain ID, not
		// the new one. Use the first PastChainID as the signing context.
		if (metadata == nil || metadata.BlockHeight == 0) && len(state.PastChainIDs) > 0 {
			originalChainID := state.PastChainIDs[0]
			ctxFn = func(ctx sdk.Context) sdk.Context {
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
					// Account doesn't exist yet — create with specific account
					// number, bypassing the auto-increment counter.
					acc = cfg.acck.NewAccountWithNumber(ctx, si.Address, si.AccountNum)
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
			txResponses = append(txResponses, abci.ResponseDeliverTx{
				ResponseBase: abci.ResponseBase{
					Error: abci.StringError("replay skipped: tx failed on source chain"),
					Log:   "genesis replay: skipped failed tx from source chain",
				},
			})
			report.record(txIdx, metadata, 0, 0, replayCategorySkippedFailed, nil)
			continue
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
		report.recordDeliverResult(txIdx, metadata, res)

		cfg.GenesisTxResultHandler(ctx, stdTx, res)
	}

	if reqInitialHeight > 1 {
		ctx.Logger().Info("Genesis replay complete, chain will start from initial height",
			"initial_height", reqInitialHeight,
		)
	}

	report.emit(ctx.Logger())

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

// isPastChainID reports whether chainID is present in the pastChainIDs allowlist.
func isPastChainID(pastChainIDs []string, chainID string) bool {
	return slices.Contains(pastChainIDs, chainID)
}

// Keep in sync with examples/gno.land/r/sys/validators/v3/poc.gno
const (
	vmModulePrefix = "vm"

	// newUpdatesAvailableKey is a flag indicating the chain valset should be updated.
	// Set by the contract, but reset by the chain (EndBlocker).
	newUpdatesAvailableKey = "new_updates_available"

	// valsetNewKey is the param that holds the new proposed valset. Set by the contract,
	// and read (but never modified) by the chain.
	valsetNewKey = "valset_new"

	// valsetPrevKey is the param that holds the latest applied valset. Initially set by
	// the contract (init), but later only written by the chain (EndBlocker).
	valsetPrevKey = "valset_prev"
)

// valsetParamPath constructs the full param key for a valset-realm-scoped param:
//
//	vm:<valset-realm-path>:<valset-param-key>
func valsetParamPath(valsetRealm, key string) string {
	return fmt.Sprintf("%s:%s:%s", vmModulePrefix, valsetRealm, key)
}

// EndBlocker defines the logic executed after every block.
// It reads valset changes from the VM params keeper, checks for a
// governance-requested chain halt, and propagates updates to consensus.
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

		// Determine which realm is responsible for valset management.
		valsetRealm := vm.ValsetRealmDefault
		prmk.GetString(ctx, vm.ValsetRealmParamPath, &valsetRealm)

		// Check if there are any pending valset changes.
		updatesAvailable := false
		prmk.GetBool(ctx, valsetParamPath(valsetRealm, newUpdatesAvailableKey), &updatesAvailable)

		if !updatesAvailable {
			return abci.ResponseEndBlock{}
		}

		var (
			prevValset     []string
			proposedValset []string

			prevValsetPath     = valsetParamPath(valsetRealm, valsetPrevKey)
			proposedValsetPath = valsetParamPath(valsetRealm, valsetNewKey)
		)

		prmk.GetStrings(ctx, prevValsetPath, &prevValset)
		prmk.GetStrings(ctx, proposedValsetPath, &proposedValset)

		// Parse the previous set.
		prevSet, err := extractUpdatesFromParams(prevValset)
		if err != nil {
			app.Logger().Error(
				"unable to parse prev valset in EndBlocker",
				"err", err,
			)

			return abci.ResponseEndBlock{}
		}

		// Parse the proposed set.
		proposedSet, err := extractUpdatesFromParams(proposedValset)
		if err != nil {
			app.Logger().Error(
				"unable to parse proposed valset in EndBlocker",
				"err", err,
			)

			return abci.ResponseEndBlock{}
		}

		// Compute the diff between prev and proposed.
		updates := prevSet.UpdatesFrom(proposedSet)

		app.Logger().Info(
			"valset changes to be applied",
			"count", len(updates),
		)

		// Advance prevValset to match proposedValset.
		prmk.SetStrings(ctx, prevValsetPath, proposedValset)

		// Clear the pending-updates flag.
		prmk.SetBool(ctx, valsetParamPath(valsetRealm, newUpdatesAvailableKey), false)

		allowedKeyTypes := ctx.ConsensusParams().Validator.PubKeyTypeURLs

		// Filter out updates that fail consensus-level validation.
		updates = slices.DeleteFunc(updates, func(u abci.ValidatorUpdate) bool {
			// Power == 0 means removal; skip further validation for removals.
			if u.Power == 0 {
				return false
			}

			// Make sure the public key is an allowed consensus key type.
			if !slices.Contains(allowedKeyTypes, amino.GetTypeURL(u.PubKey)) {
				app.Logger().Error(
					"valset update invalid; unsupported pubkey type",
					"address", u.Address.String(),
					"pubkey_type", amino.GetTypeURL(u.PubKey),
				)

				return true // delete it
			}

			return false // keep it
		})

		return abci.ResponseEndBlock{
			ValidatorUpdates: updates,
		}
	}
}

// extractUpdatesFromParams parses serialized validator updates from the params keeper.
// Each entry is expected to be in the form:
//
//	<address>:<pub-key>:<voting-power>
//
// A voting power of 0 indicates a validator removal.
func extractUpdatesFromParams(changes []string) (abci.ValidatorUpdates, error) {
	updates := make(abci.ValidatorUpdates, 0, len(changes))

	for _, change := range changes {
		parts := strings.Split(change, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf(
				"valset update is not in the format <address>:<pub-key>:<voting-power>, but %q",
				change,
			)
		}

		address, err := crypto.AddressFromBech32(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid validator address: %w", err)
		}

		pubKey, err := crypto.PubKeyFromBech32(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid validator pubkey: %w", err)
		}

		votingPower, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid voting power: %w", err)
		}

		updates = append(updates, abci.ValidatorUpdate{
			Address: address,
			PubKey:  pubKey,
			Power:   votingPower,
		})
	}

	return updates, nil
}
