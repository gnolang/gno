package sdk

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"syscall"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Key to store the consensus params in the main store.
var (
	mainConsensusParamsKey = []byte("consensus_params")
	mainLastHeaderKey      = []byte("last_header")
)

// BaseApp reflects the ABCI application implementation.
type BaseApp struct {
	// initialized on creation
	logger *slog.Logger
	name   string                 // application name from abci.Info
	db     dbm.DB                 // common DB backend
	cms    store.CommitMultiStore // Main (uncached) state
	router Router                 // handle any kind of message

	// set upon LoadVersion or LoadLatestVersion.
	baseKey store.StoreKey // Base Store in cms (raw db, not hashed)
	mainKey store.StoreKey // Main Store in cms (e.g. iavl, merkle-ized)

	anteHandler  AnteHandler  // ante handler for fee and auth
	initChainer  InitChainer  // initialize state with validators and state blob
	beginBlocker BeginBlocker // logic to run before any txs
	endBlocker   EndBlocker   // logic to run after all txs, and to determine valset changes

	// --------------------
	// Volatile state
	// checkState is set on initialization and reset on Commit.
	// deliverState is set in InitChain and BeginBlock and cleared on Commit.
	// See methods setCheckState and setDeliverState.
	checkState   *state          // for CheckTx
	deliverState *state          // for DeliverTx
	voteInfos    []abci.VoteInfo // absent validators from begin block

	// consensus params
	// TODO: Move this in the future to baseapp param store on main store.
	consensusParams *abci.ConsensusParams

	// The minimum gas prices a validator is willing to accept for processing a
	// transaction. This is mainly used for DoS and spam prevention.
	minGasPrices []GasPrice

	// flag for sealing options and parameters to a BaseApp
	sealed bool // TODO: needed?

	// block height at which to halt the chain and gracefully shutdown
	haltHeight uint64

	// minimum block time (in Unix seconds) at which to halt the chain and gracefully shutdown
	haltTime uint64

	// application's version string
	appVersion string
}

var _ abci.Application = (*BaseApp)(nil)

// NewBaseApp returns a reference to an initialized BaseApp. It accepts a
// variadic number of option functions, which act on the BaseApp to set
// configuration choices.
//
// NOTE: The db is used to store the version number for now.
func NewBaseApp(
	name string,
	logger *slog.Logger,
	db dbm.DB,
	baseKey store.StoreKey,
	mainKey store.StoreKey,
	options ...func(*BaseApp),
) *BaseApp {
	app := &BaseApp{
		logger:  logger,
		name:    name,
		db:      db,
		cms:     store.NewCommitMultiStore(db),
		router:  NewRouter(),
		baseKey: baseKey,
		mainKey: mainKey,
	}
	for _, option := range options {
		option(app)
	}

	return app
}

// Name returns the name of the BaseApp.
func (app *BaseApp) Name() string {
	return app.name
}

// AppVersion returns the application's version string.
func (app *BaseApp) AppVersion() string {
	return app.appVersion
}

// Logger returns the logger of the BaseApp.
func (app *BaseApp) Logger() *slog.Logger {
	return app.logger
}

// MountStoreWithDB mounts a store to the provided key in the BaseApp
// multistore, using a specified DB.
func (app *BaseApp) MountStoreWithDB(key store.StoreKey, cons store.CommitStoreConstructor, db dbm.DB) {
	app.cms.MountStoreWithDB(key, cons, db)
}

// MountStore mounts a store to the provided key in the BaseApp multistore,
// using the default DB.
func (app *BaseApp) MountStore(key store.StoreKey, cons store.CommitStoreConstructor) {
	app.cms.MountStoreWithDB(key, cons, nil)
}

// LoadLatestVersion loads the latest application version. It will panic if
// called more than once on a running BaseApp.
// This, or LoadVersion() MUST be called even after first init.
func (app *BaseApp) LoadLatestVersion() error {
	err := app.cms.LoadLatestVersion()
	if err != nil {
		return err
	}
	return app.initFromMainStore()
}

// LoadVersion loads the BaseApp application version. It will panic if called
// more than once on a running baseapp.
// This, or LoadLatestVersion() MUST be called even after first init.
func (app *BaseApp) LoadVersion(version int64) error {
	err := app.cms.LoadVersion(version)
	if err != nil {
		return err
	}
	return app.initFromMainStore()
}

// LastCommitID returns the last CommitID of the multistore.
func (app *BaseApp) LastCommitID() store.CommitID {
	return app.cms.LastCommitID()
}

// LastBlockHeight returns the last committed block height.
func (app *BaseApp) LastBlockHeight() int64 {
	return app.cms.LastCommitID().Version
}

// initializes the app from app.cms after loading.
func (app *BaseApp) initFromMainStore() error {
	baseStore := app.cms.GetStore(app.baseKey)
	if baseStore == nil {
		return errors.New("baseapp expects MultiStore with 'base' Store")
	}
	mainStore := app.cms.GetStore(app.mainKey)
	if mainStore == nil {
		return errors.New("baseapp expects MultiStore with 'main' Store")
	}

	// Load the consensus params from the main store. If the consensus params are
	// nil, it will be saved later during InitChain.
	//
	// TODO: assert that InitChain hasn't yet been called.
	consensusParamsBz := mainStore.Get(mainConsensusParamsKey)
	if consensusParamsBz != nil {
		consensusParams := &abci.ConsensusParams{}
		err := amino.Unmarshal(consensusParamsBz, consensusParams)
		if err != nil {
			panic(err)
		}

		app.setConsensusParams(consensusParams)
	}

	// Load the consensus header from the main store.
	// This is needed to setCheckState with the right chainID etc.
	lastHeaderBz := baseStore.Get(mainLastHeaderKey)
	if lastHeaderBz != nil {
		lastHeader := &bft.Header{}
		err := amino.Unmarshal(lastHeaderBz, lastHeader)
		if err != nil {
			panic(err)
		}
		app.setCheckState(lastHeader)
	}
	// Done.
	app.Seal()

	return nil
}

func (app *BaseApp) setMinGasPrices(gasPrices []GasPrice) {
	app.minGasPrices = gasPrices
}

func (app *BaseApp) setHaltHeight(haltHeight uint64) {
	app.haltHeight = haltHeight
}

func (app *BaseApp) setHaltTime(haltTime uint64) {
	app.haltTime = haltTime
}

// Returns a read-only (cache) MultiStore.
// This may be used by keepers for initialization upon restart.
func (app *BaseApp) GetCacheMultiStore() store.MultiStore {
	return app.cms.MultiCacheWrap()
}

// Router returns the router of the BaseApp.
func (app *BaseApp) Router() Router {
	if app.sealed {
		// We cannot return a router when the app is sealed because we can't have
		// any routes modified which would cause unexpected routing behavior.
		panic("Router() on sealed BaseApp")
	}
	return app.router
}

// Seal seals a BaseApp. It prohibits any further modifications to a BaseApp.
func (app *BaseApp) Seal() { app.sealed = true }

// IsSealed returns true if the BaseApp is sealed and false otherwise.
func (app *BaseApp) IsSealed() bool { return app.sealed }

// setCheckState sets checkState with the cached multistore and
// the context wrapping it.
// It is called by InitChain() and Commit()
func (app *BaseApp) setCheckState(header abci.Header) {
	ms := app.cms.MultiCacheWrap()
	app.checkState = &state{
		ms:  ms,
		ctx: NewContext(RunTxModeCheck, ms, header, app.logger).WithMinGasPrices(app.minGasPrices),
	}
}

// setDeliverState sets deliverState with the cached multistore and
// the context wrapping it.
// It is called by InitChain() and BeginBlock(),
// and deliverState is set nil on Commit().
func (app *BaseApp) setDeliverState(header abci.Header) {
	ms := app.cms.MultiCacheWrap()
	app.deliverState = &state{
		ms:  ms,
		ctx: NewContext(RunTxModeDeliver, ms, header, app.logger),
	}
}

// setConsensusParams memoizes the consensus params.
func (app *BaseApp) setConsensusParams(consensusParams *abci.ConsensusParams) {
	app.consensusParams = consensusParams
}

// setConsensusParams stores the consensus params to the main store.
func (app *BaseApp) storeConsensusParams(consensusParams *abci.ConsensusParams) {
	consensusParamsBz, err := amino.Marshal(consensusParams)
	if err != nil {
		panic(err)
	}
	mainStore := app.cms.GetStore(app.mainKey)
	mainStore.Set(mainConsensusParamsKey, consensusParamsBz)
}

// getMaximumBlockGas gets the maximum gas from the consensus params. It panics
// if maximum block gas is less than negative one and returns zero if negative
// one.
func (app *BaseApp) getMaximumBlockGas() int64 {
	if app.consensusParams == nil || app.consensusParams.Block == nil {
		return 0
	}

	maxGas := app.consensusParams.Block.MaxGas
	switch {
	case maxGas < -1:
		panic(fmt.Sprintf("invalid maximum block gas: %d", maxGas))

	case maxGas == -1:
		return 0

	default:
		return maxGas
	}
}

// ----------------------------------------------------------------------------
// ABCI

// Info implements the ABCI interface.
func (app *BaseApp) Info(req abci.RequestInfo) (res abci.ResponseInfo) {
	lastCommitID := app.cms.LastCommitID()

	// return res
	res.Data = []byte(app.Name())
	res.LastBlockHeight = lastCommitID.Version
	res.LastBlockAppHash = lastCommitID.Hash
	return
}

// SetOption implements the ABCI interface.
func (app *BaseApp) SetOption(req abci.RequestSetOption) (res abci.ResponseSetOption) {
	// TODO: Implement!
	return
}

// InitChain implements the ABCI interface. It runs the initialization logic
// directly on the CommitMultiStore.
func (app *BaseApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	// stash the consensus params in the cms main store and memoize
	if req.ConsensusParams != nil {
		app.setConsensusParams(req.ConsensusParams)
		app.storeConsensusParams(req.ConsensusParams)
	}

	initHeader := &bft.Header{ChainID: req.ChainID, Time: req.Time}

	// initialize the deliver state and check state with a correct header
	app.setDeliverState(initHeader)
	app.setCheckState(initHeader)

	if app.initChainer == nil {
		return
	}

	// add block gas meter for any genesis transactions (allow infinite gas)
	app.deliverState.ctx = app.deliverState.ctx.
		WithBlockGasMeter(store.NewInfiniteGasMeter())

	res = app.initChainer(app.deliverState.ctx, req)

	// sanity check
	if len(req.Validators) > 0 {
		if len(req.Validators) != len(res.Validators) {
			panic(fmt.Errorf(
				"len(RequestInitChain.Validators) != len(validators) (%d != %d)",
				len(req.Validators), len(res.Validators)))
		}
		sort.Sort(abci.ValidatorUpdates(req.Validators))
		sort.Sort(abci.ValidatorUpdates(res.Validators))
		for i, val := range res.Validators {
			if !val.Equals(req.Validators[i]) {
				panic(fmt.Errorf("validators[%d] != req.Validators[%d] ", i, i))
			}
		}
	}

	// NOTE: We don't commit, but BeginBlock for block 1 starts from this
	// deliverState.
	return
}

// Splits a string path using the delimiter '/'.
// e.g. "this/is/funny" becomes []string{"this", "is", "funny"}
func splitPath(requestPath string) (path []string) {
	path = strings.Split(requestPath, "/")
	// first element is empty string
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}
	return path
}

// Query implements the ABCI interface. It delegates to CommitMultiStore if it
// implements Queryable.
func (app *BaseApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	path := splitPath(req.Path)
	if len(path) == 0 {
		msg := "no query path provided"
		res.Error = ABCIError(std.ErrUnknownRequest(msg))
		return
	}

	switch path[0] {
	// "/.app", "/.store" prefix for special application queries
	case ".app":
		return handleQueryApp(app, path, req)

	case ".store":
		return handleQueryStore(app, path, req)

	// default router queries
	default:
		return handleQueryCustom(app, path, req)
	}

	msg := "unknown query path " + req.Path
	res.Error = ABCIError(std.ErrUnknownRequest(msg))
	return
}

func handleQueryApp(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(path) >= 2 {
		var result Result

		switch path[1] {
		case "simulate":
			txBytes := req.Data
			var tx Tx
			err := amino.Unmarshal(txBytes, &tx)
			if err != nil {
				res.Error = ABCIError(std.ErrTxDecode(err.Error()))
			} else {
				result = app.Simulate(txBytes, tx)
			}
			res.Height = req.Height
			res.Value = amino.MustMarshal(result)
			return res
		case "version":
			res.Height = req.Height
			res.Value = []byte(app.appVersion)
			return res
		default:
			res.Error = ABCIError(std.ErrUnknownRequest(fmt.Sprintf("Unknown query: %s", path)))
			return
		}
	} else {
		res.Error = ABCIError(std.ErrUnknownRequest(fmt.Sprintf("Unknown query: %s", path)))
		return
	}
}

func handleQueryStore(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	// "/store" prefix for store queries
	queryable, ok := app.cms.(store.Queryable)
	if !ok {
		msg := "multistore doesn't support queries"
		res.Error = ABCIError(std.ErrUnknownRequest(msg))
		return
	}

	req.Path = "/" + strings.Join(path[1:], "/")

	// when a client did not provide a query height, manually inject the latest
	if req.Height == 0 {
		req.Height = app.LastBlockHeight()
	}

	if req.Height <= 1 && req.Prove {
		res.Error = ABCIError(std.ErrInternal("cannot query with proof when height <= 1; please provide a valid height"))
		return
	}

	resp := queryable.Query(req)
	resp.Height = req.Height
	return resp
}

func handleQueryCustom(app *BaseApp, path []string, req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(path) < 1 || path[0] == "" {
		res.Error = ABCIError(std.ErrUnknownRequest("No route for custom query specified"))
		return
	}

	handler := app.router.Route(path[0])
	if handler == nil {
		res.Error = ABCIError(std.ErrUnknownRequest(fmt.Sprintf("no custom handler found for route %s", path[0])))
		return
	}

	// when a client did not provide a query height, manually inject the latest
	if req.Height == 0 {
		req.Height = app.LastBlockHeight()
	}

	if req.Height <= 1 && req.Prove {
		res.Error = ABCIError(std.ErrInternal("cannot query with proof when height <= 1; please provide a valid height"))
		return
	}

	cacheMS, err := app.cms.MultiImmutableCacheWrapWithVersion(req.Height)
	if err != nil {
		res.Error = ABCIError(std.ErrInternal(
			fmt.Sprintf(
				"failed to load state at height %d; %s (latest height: %d)",
				req.Height, err, app.LastBlockHeight(),
			),
		))
		return
	}

	// cache wrap the commit-multistore for safety
	// XXX RunTxModeQuery?
	ctx := NewContext(RunTxModeCheck, cacheMS, app.checkState.ctx.BlockHeader(), app.logger).WithMinGasPrices(app.minGasPrices)

	// Passes the query to the handler.
	res = handler.Query(ctx, req)
	return
}

func (app *BaseApp) validateHeight(req abci.RequestBeginBlock) error {
	if req.Header.GetHeight() < 1 {
		return fmt.Errorf("invalid height: %d", req.Header.GetHeight())
	}

	prevHeight := app.LastBlockHeight()
	if req.Header.GetHeight() != prevHeight+1 {
		return fmt.Errorf("invalid height: %d; expected: %d", req.Header.GetHeight(), prevHeight+1)
	}

	return nil
}

// BeginBlock implements the ABCI application interface.
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	if err := app.validateHeight(req); err != nil {
		panic(err)
	}

	// Initialize the DeliverTx state. If this is the first block, it should
	// already be initialized in InitChain. Otherwise app.deliverState will be
	// nil, since it is reset on Commit.
	if app.deliverState == nil {
		app.setDeliverState(req.Header)
	} else {
		// In the first block, app.deliverState.ctx will already be initialized
		// by InitChain. Context is now updated with Header information.
		app.deliverState.ctx = app.deliverState.ctx.
			WithBlockHeader(req.Header)
	}

	// add block gas meter
	var gasMeter store.GasMeter
	if maxGas := app.getMaximumBlockGas(); maxGas > 0 {
		gasMeter = store.NewGasMeter(maxGas)
	} else {
		gasMeter = store.NewInfiniteGasMeter()
	}

	app.deliverState.ctx = app.deliverState.ctx.WithBlockGasMeter(gasMeter)

	if app.beginBlocker != nil {
		res = app.beginBlocker(app.deliverState.ctx, req)
	}

	// set the signed validators for addition to context in deliverTx
	if req.LastCommitInfo != nil {
		app.voteInfos = req.LastCommitInfo.Votes
	}
	return
}

// CheckTx implements the ABCI interface. It runs the "basic checks" to see
// whether or not a transaction can possibly be executed, first decoding and then
// the ante handler (which checks signatures/fees/ValidateBasic).
//
// NOTE:CheckTx does not run the actual Msg handler function(s).
func (app *BaseApp) CheckTx(req abci.RequestCheckTx) (res abci.ResponseCheckTx) {
	var tx Tx
	err := amino.Unmarshal(req.Tx, &tx)
	if err != nil {
		res.Error = ABCIError(std.ErrTxDecode(err.Error()))
		return
	} else {
		result := app.runTx(RunTxModeCheck, req.Tx, tx)
		res.ResponseBase = result.ResponseBase
		res.GasWanted = result.GasWanted
		res.GasUsed = result.GasUsed
		return
	}
}

// DeliverTx implements the ABCI interface.
func (app *BaseApp) DeliverTx(req abci.RequestDeliverTx) (res abci.ResponseDeliverTx) {
	var tx Tx
	err := amino.Unmarshal(req.Tx, &tx)
	if err != nil {
		res.Error = ABCIError(std.ErrTxDecode(err.Error()))
		return
	} else {
		result := app.runTx(RunTxModeDeliver, req.Tx, tx)
		res.ResponseBase = result.ResponseBase
		res.GasWanted = result.GasWanted
		res.GasUsed = result.GasUsed
		return
	}
}

// validateBasicTxMsgs executes basic validator calls for messages.
func validateBasicTxMsgs(msgs []Msg) error {
	if msgs == nil || len(msgs) == 0 {
		return std.ErrUnknownRequest("Tx.GetMsgs() must return at least one message in list")
	}

	for _, msg := range msgs {
		// Validate the Msg.
		err := msg.ValidateBasic()
		if err != nil {
			return err
		}
	}

	return nil
}

// retrieve the context for the tx w/ txBytes and other memoized values.
func (app *BaseApp) getContextForTx(mode RunTxMode, txBytes []byte) (ctx Context) {
	ctx = app.getState(mode).ctx.
		WithMode(mode).
		WithTxBytes(txBytes).
		WithVoteInfos(app.voteInfos).
		WithConsensusParams(app.consensusParams)

	if mode == RunTxModeSimulate {
		ctx, _ = ctx.CacheContext()
	}

	return
}

// / runMsgs iterates through all the messages and executes them.
func (app *BaseApp) runMsgs(ctx Context, msgs []Msg, mode RunTxMode) (result Result) {
	msgLogs := make([]string, 0, len(msgs))

	data := make([]byte, 0, len(msgs))
	err := error(nil)
	events := []Event{}

	// NOTE: GasWanted is determined by ante handler and GasUsed by the GasMeter.
	for i, msg := range msgs {
		// match message route
		msgRoute := msg.Route()
		handler := app.router.Route(msgRoute)
		if handler == nil {
			result.Error = ABCIError(std.ErrUnknownRequest("unrecognized message type: " + msgRoute))
			return
		}

		var msgResult Result
		ctx = ctx.WithEventLogger(NewEventLogger())

		// run the message!
		// skip actual execution for CheckTx mode
		if mode != RunTxModeCheck {
			msgResult = handler.Process(ctx, msg)
		}

		// Each message result's Data must be length prefixed in order to separate
		// each result.
		data = append(data, msgResult.Data...)
		events = append(events, msgResult.Events...)
		defer func() {
			events = append(events, ctx.EventLogger().Events()...)
			result.Events = events
		}()
		// TODO append msgevent from ctx. XXX XXX

		// stop execution and return on first failed message
		if !msgResult.IsOK() {
			msgLogs = append(msgLogs,
				fmt.Sprintf("msg:%d,success:%v,log:%s,events:%v",
					i, false, msgResult.Log, events))
			err = msgResult.Error
			break
		}

		msgLogs = append(msgLogs,
			fmt.Sprintf("msg:%d,success:%v,log:%s,events:%v",
				i, true, msgResult.Log, events))
	}

	result.Error = ABCIError(err)
	result.Data = data
	result.Log = strings.Join(msgLogs, "\n")
	result.GasUsed = ctx.GasMeter().GasConsumed()
	result.Events = events
	return result
}

// Returns the applications's deliverState if app is in RunTxModeDeliver,
// otherwise it returns the application's checkstate.
func (app *BaseApp) getState(mode RunTxMode) *state {
	if mode == RunTxModeCheck || mode == RunTxModeSimulate {
		return app.checkState
	}

	return app.deliverState
}

// cacheTxContext returns a new context based off of the provided context with
// a cache wrapped multi-store.
func (app *BaseApp) cacheTxContext(ctx Context, txBytes []byte) (
	Context, store.MultiStore,
) {
	ms := ctx.MultiStore()
	// TODO: https://github.com/tendermint/classic/sdk/issues/2824
	msCache := ms.MultiCacheWrap()
	return ctx.WithMultiStore(msCache), msCache
}

// runTx processes a transaction. The transactions is processed via an
// anteHandler. The provided txBytes may be nil in some cases, eg. in tests. For
// further details on transaction execution, reference the BaseApp SDK
// documentation.
func (app *BaseApp) runTx(mode RunTxMode, txBytes []byte, tx Tx) (result Result) {
	// NOTE: GasWanted should be returned by the AnteHandler. GasUsed is
	// determined by the GasMeter. We need access to the context to get the gas
	// meter so we initialize upfront.
	var gasWanted int64

	ctx := app.getContextForTx(mode, txBytes)
	ms := ctx.MultiStore()
	if mode == RunTxModeDeliver {
		gasleft := ctx.BlockGasMeter().Remaining()
		ctx = ctx.WithGasMeter(store.NewPassthroughGasMeter(
			ctx.GasMeter(),
			gasleft,
		))
	}

	// only run the tx if there is block gas remaining
	if mode == RunTxModeDeliver && ctx.BlockGasMeter().IsOutOfGas() {
		result.Error = ABCIError(std.ErrOutOfGas("no block gas left to run tx"))
		return
	}

	var startingGas int64
	if mode == RunTxModeDeliver {
		startingGas = ctx.BlockGasMeter().GasConsumed()
	}

	defer func() {
		if r := recover(); r != nil {
			switch ex := r.(type) {
			case store.OutOfGasException:
				log := fmt.Sprintf(
					"out of gas, gasWanted: %d, gasUsed: %d location: %v",
					gasWanted,
					ctx.GasMeter().GasConsumed(),
					ex.Descriptor,
				)
				result.Error = ABCIError(std.ErrOutOfGas(log))
				result.Log = log
				result.GasWanted = gasWanted
				result.GasUsed = ctx.GasMeter().GasConsumed()
				return
			default:
				log := fmt.Sprintf("recovered: %v\nstack:\n%v", r, string(debug.Stack()))
				result.Error = ABCIError(std.ErrInternal(log))
				result.Log = log
				result.GasWanted = gasWanted
				result.GasUsed = ctx.GasMeter().GasConsumed()
				return
			}
		}
		// Whether AnteHandler panics or not.
		result.GasWanted = gasWanted
		result.GasUsed = ctx.GasMeter().GasConsumed()
	}()

	// If BlockGasMeter() panics it will be caught by the above recover and will
	// return an error - in any case BlockGasMeter will consume gas past the limit.
	//
	// NOTE: This must exist in a separate defer function for the above recovery
	// to recover from this one.
	defer func() {
		if mode == RunTxModeDeliver {
			ctx.BlockGasMeter().ConsumeGas(
				ctx.GasMeter().GasConsumedToLimit(),
				"block gas meter",
			)

			if ctx.BlockGasMeter().GasConsumed() < startingGas {
				panic(std.ErrGasOverflow("tx gas summation"))
			}
		}
	}()

	msgs := tx.GetMsgs()
	if err := validateBasicTxMsgs(msgs); err != nil {
		result.Error = ABCIError(err)
		return
	}

	if app.anteHandler != nil {
		var anteCtx Context
		var msCache store.MultiStore

		// Cache wrap context before anteHandler call in case
		// it aborts.  This is required for both CheckTx and
		// DeliverTx.  Ref:
		// https://github.com/tendermint/classic/sdk/issues/2772
		//
		// NOTE: Alternatively, we could require that
		// anteHandler ensures that writes do not happen if
		// aborted/failed.  This may have some performance
		// benefits, but it'll be more difficult to get
		// right.
		anteCtx, msCache = app.cacheTxContext(ctx, txBytes)
		// Call AnteHandler.
		// NOTE: It is the responsibility of the anteHandler
		// to use something like passthroughGasMeter to
		// account for ante handler gas usage, despite
		// OutOfGasExceptions.
		newCtx, result, abort := app.anteHandler(anteCtx, tx, mode == RunTxModeSimulate)
		if newCtx.IsZero() {
			panic("newCtx must not be zero")
		}
		if abort && result.Error == nil {
			panic("result.Error should be set for abort")
		}
		if abort {
			// NOTE: first we must set ctx above,
			// because a previous defer call sets
			// result.GasUsed, regardless of error.
			return result
		} else {
			// Revert cache wrapping of multistore.
			ctx = newCtx.WithMultiStore(ms)
			msCache.MultiWrite()
			gasWanted = result.GasWanted
		}
	}

	// Create a new context based off of the existing context with a cache wrapped
	// multi-store in case message processing fails.
	runMsgCtx, msCache := app.cacheTxContext(ctx, txBytes)
	result = app.runMsgs(runMsgCtx, msgs, mode)
	result.GasWanted = gasWanted

	// Safety check: don't write the cache state unless we're in DeliverTx.
	if mode != RunTxModeDeliver {
		return result
	}

	// only update state if all messages pass
	if result.IsOK() {
		msCache.MultiWrite()
	}

	return result
}

// EndBlock implements the ABCI interface.
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	if app.endBlocker != nil {
		res = app.endBlocker(app.deliverState.ctx, req)
	}

	return
}

// Commit implements the ABCI interface. It will commit all state that exists in
// the deliver state's multi-store and includes the resulting commit ID in the
// returned abci.ResponseCommit. Commit will set the check state based on the
// latest header and reset the deliver state. Also, if a non-zero halt height is
// defined in config, Commit will execute a deferred function call to check
// against that height and gracefully halt if it matches the latest committed
// height.
func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	header := app.deliverState.ctx.BlockHeader()

	var halt bool

	switch {
	case app.haltHeight > 0 && uint64(header.GetHeight()) >= app.haltHeight:
		halt = true

	case app.haltTime > 0 && header.GetTime().Unix() >= int64(app.haltTime):
		halt = true
	}

	if halt {
		app.halt()

		// Note: State is not actually committed when halted. Logs from Tendermint
		// can be ignored.
		return abci.ResponseCommit{}
	}

	// Write the DeliverTx state which is cache-wrapped and commit the MultiStore.
	// The write to the DeliverTx state writes all state transitions to the root
	// MultiStore (app.cms) so when Commit() is called is persists those values.
	app.deliverState.ms.MultiWrite()
	commitID := app.cms.Commit()
	app.logger.Debug("Commit synced", "commit", fmt.Sprintf("%X", commitID))

	// Save this header.
	baseStore := app.cms.GetStore(app.baseKey)
	if baseStore == nil {
		res.Error = ABCIError(errors.New("baseapp expects MultiStore with 'base' Store"))
		return
	}
	headerBz := amino.MustMarshal(header)
	baseStore.Set(mainLastHeaderKey, headerBz)

	// Reset the Check state to the latest committed.
	//
	// NOTE: This is safe because Tendermint holds a lock on the mempool for
	// Commit. Use the header from this latest block.
	app.setCheckState(header)

	// empty/reset the deliver state
	app.deliverState = nil

	// return.
	res.Data = commitID.Hash
	return
}

// halt attempts to gracefully shutdown the node via SIGINT and SIGTERM falling
// back on os.Exit if both fail.
func (app *BaseApp) halt() {
	app.logger.Info("halting node per configuration", "height", app.haltHeight, "time", app.haltTime)

	p, err := os.FindProcess(os.Getpid())
	if err == nil {
		// attempt cascading signals in case SIGINT fails (os dependent)
		sigIntErr := p.Signal(syscall.SIGINT)
		sigTermErr := p.Signal(syscall.SIGTERM)

		if sigIntErr == nil || sigTermErr == nil {
			return
		}
	}

	// Resort to exiting immediately if the process could not be found or killed
	// via SIGINT/SIGTERM signals.
	app.logger.Info("failed to send SIGINT/SIGTERM; exiting...")
	os.Exit(0)
}

// TODO implement cleanup
func (app *BaseApp) Close() error {
	return nil // XXX
}

// ----------------------------------------------------------------------------
// State

type state struct {
	ms  store.MultiStore
	ctx Context
}

func (st *state) MultiCacheWrap() store.MultiStore {
	return st.ms.MultiCacheWrap()
}

func (st *state) Context() Context {
	return st.ctx
}
