// Package gnoland contains the bootstrapping code to launch a gno.land node.
package gnoland

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"

	// Only goleveldb is supported for now.
	_ "github.com/gnolang/gno/tm2/pkg/db/_tags"
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb"
)

// AppOptions contains the options to create the gno.land ABCI application.
type AppOptions struct {
	DB                      dbm.DB             // required
	Logger                  *slog.Logger       // required
	EventSwitch             events.EventSwitch // required
	VMOutput                io.Writer          // optional
	SkipGenesisVerification bool               // default to verify genesis transactions
	InitChainerConfig                          // options related to InitChainer
	MinGasPrices            string             // optional
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
		SkipGenesisVerification: true,
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

	// Set InitChainer
	icc := cfg.InitChainerConfig
	icc.baseApp = baseApp
	icc.acck, icc.bankk, icc.vmk, icc.prmk, icc.gpk = acck, bankk, vmk, prmk, gpk
	baseApp.SetInitChainer(icc.InitChainer)

	// Set AnteHandler
	authOptions := auth.AnteOptions{
		VerifyGenesisSignatures: !cfg.SkipGenesisVerification,
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
	evsw events.EventSwitch,
	logger *slog.Logger,
	minGasPrices string,
) (abci.Application, error) {
	var err error

	cfg := &AppOptions{
		Logger:      logger,
		EventSwitch: evsw,
		InitChainerConfig: InitChainerConfig{
			GenesisTxResultHandler: PanicOnFailingTxResultHandler,
			StdlibDir:              filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs"),
		},
		MinGasPrices:            minGasPrices,
		SkipGenesisVerification: genesisCfg.SkipSigVerification,
	}
	if genesisCfg.SkipFailingTxs {
		cfg.GenesisTxResultHandler = NoopGenesisTxResultHandler
	}

	// Get main DB.
	cfg.DB, err = dbm.NewDB("gnolang", dbm.GoLevelDBBackend, filepath.Join(dataRootDir, config.DefaultDBDir))
	if err != nil {
		return nil, fmt.Errorf("error initializing database %q using path %q: %w", dbm.GoLevelDBBackend, dataRootDir, err)
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
		accr := acc.(*GnoAccount)
		accr.SetUnrestricted()
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
}

// Keep in sync with r/sys/validators/v3/poc.gno
const (
	vmModulePrefix = "vm"

	// newUpdatesAvailableKey is a flag indicating the chain valset should be updated.
	// Set by the contract, but reset by the chain (EndBlocker)
	newUpdatesAvailableKey = "new_updates_available"

	// valsetNewKey is the param that holds the new proposed valset. Set by the contract,
	// and read (but never modified) by the chain
	valsetNewKey = "valset_new"

	// valsetPrevKey is the param that holds the latest valset. Initially set by the contract (init),
	// but later only written by the chain (EndBlocker).
	valsetPrevKey = "valset_prev"
)

// valsetParamPath constructs the valset update
// params path
//
// vm:<valset-realm-path>:<valset-param-key>
func valsetParamPath(valsetRealm string, key string) string {
	return fmt.Sprintf(
		"%s:%s:%s",
		vmModulePrefix,
		valsetRealm,
		key,
	)
}

// EndBlocker defines the logic executed after every block.
// Currently, it parses events that happened during execution to calculate
// validator set changes
func EndBlocker(
	prmk params.ParamsKeeperI,
	acck auth.AccountKeeperI,
	gpk auth.GasPriceKeeperI,
	app endBlockerApp,
) func(
	ctx sdk.Context,
	req abci.RequestEndBlock,
) abci.ResponseEndBlock {
	return func(ctx sdk.Context, _ abci.RequestEndBlock) abci.ResponseEndBlock {
		// set the auth params value in the ctx.  The EndBlocker will use InitialGasPrice in
		// the params to calculate the updated gas price.
		if acck != nil {
			ctx = ctx.WithValue(auth.AuthParamsContextKey{}, acck.GetParams(ctx))
		}
		if acck != nil && gpk != nil {
			auth.EndBlocker(ctx, gpk)
		}

		// Fetch the changes for the current height and valset realm,
		// if they're stored in the params
		valsetRealm := vm.ValsetRealmDefault
		prmk.GetString(ctx, vm.ValsetRealmParamPath, &valsetRealm)

		// Check if there need to be any changes applied
		updatesAvailable := false
		prmk.GetBool(ctx, valsetParamPath(valsetRealm, newUpdatesAvailableKey), &updatesAvailable)

		// Check if there need to be changes applied
		if !updatesAvailable {
			// No updates to apply
			return abci.ResponseEndBlock{}
		}

		var (
			prevValset     = make([]string, 0)
			proposedValset = make([]string, 0)

			prevValsetPath     = valsetParamPath(valsetRealm, valsetPrevKey)
			proposedValsetPath = valsetParamPath(valsetRealm, valsetNewKey)
		)

		prmk.GetStrings(ctx, prevValsetPath, &prevValset)
		prmk.GetStrings(ctx, proposedValsetPath, &proposedValset)

		// Parse the previous set
		prevSet, err := extractUpdatesFromParams(prevValset)
		if err != nil {
			app.Logger().Error(
				"unable to parse prev valset updates in EndBlocker",
				"err", err,
			)

			return abci.ResponseEndBlock{}
		}

		// Parse the proposed set
		proposedSet, err := extractUpdatesFromParams(proposedValset)
		if err != nil {
			app.Logger().Error(
				"unable to parse proposed valset updates in EndBlocker",
				"err", err,
			)

			return abci.ResponseEndBlock{}
		}

		// Check for changes
		updates := prevSet.UpdatesFrom(proposedSet)

		app.Logger().Info(
			"valset changes to be applied",
			"count", len(updates),
		)

		// Update the latest valset in the param
		prmk.SetStrings(ctx, prevValsetPath, proposedValset)

		// Clear the new updates params flag
		prmk.SetBool(ctx, valsetParamPath(valsetRealm, newUpdatesAvailableKey), false)

		return abci.ResponseEndBlock{
			ValidatorUpdates: updates,
		}
	}
}

// extractUpdatesFromParams extracts the validator set updates
// from the params keeper value.
// Valset changes are saved in the form:
//
// <address>:<pub-key>:<voting-power>
// voting power == 0 => validator removal
// voting power != 0 => validator power update / validator addition
func extractUpdatesFromParams(changes []string) (abci.ValidatorUpdates, error) {
	updates := make([]abci.ValidatorUpdate, 0, len(changes))

	for _, change := range changes {
		changeParts := strings.Split(change, ":")
		if len(changeParts) != 3 {
			return nil, fmt.Errorf(
				"valset update is not in the format <address>:<pub-key>:<voting-power>, but %s",
				change,
			)
		}

		// Grab the address
		address, err := crypto.AddressFromBech32(changeParts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid validator address: %w", err)
		}

		// Grab the public key
		pubKey, err := crypto.PubKeyFromBech32(changeParts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid validator pubkey: %w", err)
		}

		// Make sure the address matches the public key
		if pubKey.Address().Compare(address) != 0 {
			return nil, fmt.Errorf(
				"address (%s) does not match public key address (%s)",
				address,
				pubKey.Address(),
			)
		}

		// Validate the voting power
		votingPower, err := strconv.ParseInt(changeParts[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid voting power: %w", err)
		}

		// Save the update
		update := abci.ValidatorUpdate{
			Address: address,
			PubKey:  pubKey,
			Power:   votingPower,
		}
		updates = append(updates, update)
	}

	return updates, nil
}
