package sdk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

var (
	baseKey = store.NewStoreKey("base") // in all test apps
	mainKey = store.NewStoreKey("main") // in all test apps
)

type (
	msgCounter  = testutils.MsgCounter
	msgCounter2 = testutils.MsgCounter2
	msgNoRoute  = testutils.MsgNoRoute
)

const (
	routeMsgCounter  = testutils.RouteMsgCounter
	routeMsgCounter2 = testutils.RouteMsgCounter2
)

// txInt: used as counter in incrementing counter tests,
// or as how much gas will be consumed in antehandler
// (depending on anteHandler used in tests)
func newTxCounter(txInt int64, msgInts ...int64) std.Tx {
	msgs := make([]std.Msg, len(msgInts))

	for i, msgInt := range msgInts {
		msgs[i] = msgCounter{Counter: msgInt, FailOnHandler: false}
	}

	tx := std.Tx{Msgs: msgs}
	setCounter(&tx, txInt)
	setFailOnHandler(&tx, false)
	return tx
}

func defaultLogger() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	return logger.With("module", "sdk/app")
}

func newBaseApp(name string, db dbm.DB, options ...func(*BaseApp)) *BaseApp {
	logger := defaultLogger()
	app := NewBaseApp(name, logger, db, baseKey, mainKey, options...)
	app.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, nil)
	app.MountStoreWithDB(mainKey, iavl.StoreConstructor, nil)
	return app
}

// simple one store baseapp
func setupBaseApp(t *testing.T, options ...func(*BaseApp)) *BaseApp {
	t.Helper()

	db := memdb.NewMemDB()
	app := newBaseApp(t.Name(), db, options...)
	require.Equal(t, t.Name(), app.Name())
	err := app.LoadLatestVersion()
	require.Nil(t, err)
	return app
}

func TestMountStores(t *testing.T) {
	t.Parallel()

	app := setupBaseApp(t)

	// check both stores
	store1 := app.cms.GetCommitStore(baseKey)
	require.NotNil(t, store1)
	store2 := app.cms.GetCommitStore(mainKey)
	require.NotNil(t, store2)
}

// Test that we can make commits and then reload old versions.
// Test that LoadLatestVersion actually does.
func TestLoadVersion(t *testing.T) {
	t.Parallel()

	pruningOpt := SetPruningOptions(store.PruneSyncable)
	name := t.Name()
	db := memdb.NewMemDB()
	app := newBaseApp(name, db, pruningOpt)

	// make a cap key and mount the store
	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := store.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	// execute a block, collect commit ID
	header := &bft.Header{ChainID: "test-chain", Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := store.CommitID{Version: 1, Hash: res.Data}

	// execute a block, collect commit ID
	header = &bft.Header{ChainID: "test-chain", Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res = app.Commit()
	commitID2 := store.CommitID{Version: 2, Hash: res.Data}

	// reload with LoadLatestVersion
	app = newBaseApp(name, db, pruningOpt)
	err = app.LoadLatestVersion()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(2), commitID2)

	// reload with LoadVersion, see if you can commit the same block and get
	// the same result
	app = newBaseApp(name, db, pruningOpt)
	err = app.LoadVersion(1)
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(1), commitID1)
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()
	testLoadVersionHelper(t, app, int64(2), commitID2)
}

func TestAppVersionSetterGetter(t *testing.T) {
	t.Parallel()

	pruningOpt := SetPruningOptions(store.PruneSyncable)
	name := t.Name()
	db := memdb.NewMemDB()
	app := newBaseApp(name, db, pruningOpt)

	require.Equal(t, "", app.AppVersion())
	res := app.Query(abci.RequestQuery{Path: ".app/version"})
	require.True(t, res.IsOK())
	require.Equal(t, "", string(res.Value))

	versionString := "1.0.0"
	app.SetAppVersion(versionString)
	require.Equal(t, versionString, app.AppVersion())
	res = app.Query(abci.RequestQuery{Path: ".app/version"})
	require.True(t, res.IsOK())
	require.Equal(t, versionString, string(res.Value))
}

func TestLoadVersionInvalid(t *testing.T) {
	t.Parallel()

	pruningOpt := SetPruningOptions(store.PruneSyncable)
	name := t.Name()
	db := memdb.NewMemDB()
	app := newBaseApp(name, db, pruningOpt)

	err := app.LoadLatestVersion()
	require.Nil(t, err)

	// require error when loading an invalid version
	err = app.LoadVersion(-1)
	require.Error(t, err)

	header := &bft.Header{ChainID: "test-chain", Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := store.CommitID{Version: 1, Hash: res.Data}

	// create a new app with the stores mounted under the same cap key
	app = newBaseApp(name, db, pruningOpt)

	// require we can load the latest version
	err = app.LoadVersion(1)
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(1), commitID1)

	// require error when loading an invalid version
	err = app.LoadVersion(2)
	require.Error(t, err)
}

func TestOptionSetters(t *testing.T) {
	t.Parallel()

	tt := []struct {
		// Calling BaseApp.[method]([value]) should change BaseApp.[fieldName] to [value].
		method    string
		fieldName string
		value     any
	}{
		{"SetName", "name", "hello"},
		{"SetAppVersion", "appVersion", "12345"},
		{"SetDB", "db", memdb.NewMemDB()},
		{"SetCMS", "cms", store.NewCommitMultiStore(memdb.NewMemDB())},
		{"SetInitChainer", "initChainer", func(Context, abci.RequestInitChain) abci.ResponseInitChain { panic("not implemented") }},
		{"SetBeginBlocker", "beginBlocker", func(Context, abci.RequestBeginBlock) abci.ResponseBeginBlock { panic("not implemented") }},
		{"SetEndBlocker", "endBlocker", func(Context, abci.RequestEndBlock) abci.ResponseEndBlock { panic("not implemented") }},
		{"SetAnteHandler", "anteHandler", func(Context, Tx, bool) (Context, Result, bool) { panic("not implemented") }},
		{"SetBeginTxHook", "beginTxHook", func(Context) Context { panic("not implemented") }},
		{"SetEndTxHook", "endTxHook", func(Context, Result) { panic("not implemented") }},
	}

	for _, tc := range tt {
		t.Run(tc.method, func(t *testing.T) {
			t.Parallel()

			var ba BaseApp
			rv := reflect.ValueOf(&ba)

			rv.MethodByName(tc.method).Call([]reflect.Value{reflect.ValueOf(tc.value)})
			changed := rv.Elem().FieldByName(tc.fieldName)

			if reflect.TypeOf(tc.value).Kind() == reflect.Func {
				assert.Equal(t, reflect.ValueOf(tc.value).Pointer(), changed.Pointer(), "%s(%#v): function value should have changed", tc.method, tc.value)
			} else {
				assert.True(t, reflect.ValueOf(tc.value).Equal(changed), "%s(%#v): wanted %v got %v", tc.method, tc.value, tc.value, changed)
			}
			assert.False(t, changed.IsZero(), "%s(%#v): field's new value should not be zero value", tc.method, tc.value)
		})
	}
}

func testLoadVersionHelper(t *testing.T, app *BaseApp, expectedHeight int64, expectedID store.CommitID) {
	t.Helper()

	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func TestOptionFunction(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	bap := newBaseApp("starting name", db, testChangeNameHelper("new name"))
	require.Equal(t, bap.name, "new name", "BaseApp should have had name changed via option function")
}

func testChangeNameHelper(name string) func(*BaseApp) {
	return func(bap *BaseApp) {
		bap.name = name
	}
}

// Test that Info returns the latest committed state.
func TestInfo(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	app := newBaseApp(t.Name(), db)

	// ----- test an empty response -------
	reqInfo := abci.RequestInfo{}
	res := app.Info(reqInfo)

	// should be empty
	assert.Equal(t, "", res.AppVersion)
	assert.Equal(t, t.Name(), string(res.Data))
	assert.Equal(t, int64(0), res.LastBlockHeight)
	require.Equal(t, []uint8(nil), res.LastBlockAppHash)

	// ----- test a proper response -------
	// TODO
}

func TestBaseAppOptionSeal(t *testing.T) {
	t.Parallel()

	app := setupBaseApp(t)

	require.Panics(t, func() {
		app.SetName("")
	})
	require.Panics(t, func() {
		app.SetAppVersion("")
	})
	require.Panics(t, func() {
		app.SetDB(nil)
	})
	require.Panics(t, func() {
		app.SetCMS(nil)
	})
	require.Panics(t, func() {
		app.SetInitChainer(nil)
	})
	require.Panics(t, func() {
		app.SetBeginBlocker(nil)
	})
	require.Panics(t, func() {
		app.SetEndBlocker(nil)
	})
	require.Panics(t, func() {
		app.SetAnteHandler(nil)
	})
	require.Panics(t, func() {
		app.SetBeginTxHook(nil)
	})
	require.Panics(t, func() {
		app.SetEndTxHook(nil)
	})
}

func TestSetMinGasPrices(t *testing.T) {
	t.Parallel()

	minGasPrices, err := ParseGasPrices("5000stake/10gas")
	require.Nil(t, err)
	db := memdb.NewMemDB()
	app := newBaseApp(t.Name(), db, SetMinGasPrices("5000stake/10gas"))
	require.Equal(t, minGasPrices, app.minGasPrices)
}

func TestInitChainer(t *testing.T) {
	t.Parallel()

	name := t.Name()
	// keep the db and logger ourselves so
	// we can reload the same  app later
	db := memdb.NewMemDB()
	app := newBaseApp(name, db)

	// set a value in the store on init chain
	key, value := []byte("hello"), []byte("goodbye")
	var initChainer InitChainer = func(ctx Context, req abci.RequestInitChain) abci.ResponseInitChain {
		store := ctx.Store(mainKey)
		store.Set(key, value)
		return abci.ResponseInitChain{}
	}

	query := abci.RequestQuery{
		Path: ".store/main/key",
		Data: key,
	}

	// set initChainer and try again - should see the value
	app.SetInitChainer(initChainer)

	// stores are mounted and private members are set - sealing baseapp
	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)
	require.Equal(t, int64(0), app.LastBlockHeight())

	// initChainer is nil - nothing happens
	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})
	res := app.Query(query)
	require.Equal(t, 0, len(res.Value))

	app.InitChain(abci.RequestInitChain{AppState: nil, ChainID: "test-chain-id"}) // must have valid JSON genesis file, even if empty

	// assert that chainID is set correctly in InitChain
	chainID := app.deliverState.ctx.ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = app.checkState.ctx.ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in checkState not set correctly in InitChain")

	app.Commit()
	res = app.Query(query)
	require.Equal(t, int64(1), app.LastBlockHeight())
	require.Equal(t, value, res.Value)

	// reload app
	app = newBaseApp(name, db)
	app.SetInitChainer(initChainer)
	err = app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)
	require.Equal(t, int64(1), app.LastBlockHeight())

	// ensure we can still query after reloading
	res = app.Query(query)
	require.Equal(t, value, res.Value)

	// commit and ensure we can still query
	header := &bft.Header{ChainID: "test-chain", Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()

	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

type testTxData struct {
	FailOnAnte bool
	Counter    int64
}

func getFailOnAnte(tx Tx) bool {
	var testdata testTxData
	amino.MustUnmarshalJSON([]byte(tx.Memo), &testdata)
	return testdata.FailOnAnte
}

func setFailOnAnte(tx *Tx, fail bool) {
	var testdata testTxData
	if tx.Memo == "" {
		tx.Memo = "{}"
	}
	amino.MustUnmarshalJSON([]byte(tx.Memo), &testdata)
	testdata.FailOnAnte = fail
	tx.Memo = string(amino.MustMarshalJSON(testdata))
}

func getCounter(tx Tx) int64 {
	var testdata testTxData
	amino.MustUnmarshalJSON([]byte(tx.Memo), &testdata)
	return testdata.Counter
}

func setCounter(tx *Tx, counter int64) {
	var testdata testTxData
	if tx.Memo == "" {
		tx.Memo = "{}"
	}
	amino.MustUnmarshalJSON([]byte(tx.Memo), &testdata)
	testdata.Counter = counter
	tx.Memo = string(amino.MustMarshalJSON(testdata))
}

func setFailOnHandler(tx *Tx, fail bool) {
	for i, msg := range tx.Msgs {
		tx.Msgs[i] = msgCounter{Counter: msg.(msgCounter).Counter, FailOnHandler: fail}
	}
}

func anteHandlerTxTest(t *testing.T, capKey store.StoreKey, storeKey []byte) AnteHandler {
	t.Helper()

	return func(ctx Context, tx std.Tx, simulate bool) (newCtx Context, res Result, abort bool) {
		store := ctx.GasStore(capKey)
		if getFailOnAnte(tx) {
			res.Error = ABCIError(std.ErrInternal("ante handler failure"))
			return newCtx, res, true
		}

		res = incrementingCounter(t, store, storeKey, getCounter(tx))
		newCtx = ctx
		return
	}
}

type testHandler struct {
	process func(Context, Msg) Result
	query   func(Context, abci.RequestQuery) abci.ResponseQuery
}

func (th testHandler) Process(ctx Context, msg Msg) Result {
	return th.process(ctx, msg)
}

func (th testHandler) Query(ctx Context, req abci.RequestQuery) abci.ResponseQuery {
	return th.query(ctx, req)
}

func newTestHandler(proc func(Context, Msg) Result) Handler {
	return testHandler{
		process: proc,
	}
}

type msgCounterHandler struct {
	t          *testing.T
	capKey     store.StoreKey
	deliverKey []byte
}

func newMsgCounterHandler(t *testing.T, capKey store.StoreKey, deliverKey []byte) Handler {
	t.Helper()

	return msgCounterHandler{t, capKey, deliverKey}
}

func (mch msgCounterHandler) Process(ctx Context, msg Msg) (res Result) {
	store := ctx.Store(mch.capKey)
	var msgCount int64
	switch m := msg.(type) {
	case msgCounter:
		if m.FailOnHandler {
			res.Error = ABCIError(std.ErrInternal("message handler failure"))
			return
		}
		msgCount = m.Counter
	case msgCounter2:
		msgCount = m.Counter
	default:
		panic(fmt.Sprint("unexpected msg type", reflect.TypeOf(msg)))
	}
	return incrementingCounter(mch.t, store, mch.deliverKey, msgCount)
}

func (mch msgCounterHandler) Query(ctx Context, req abci.RequestQuery) abci.ResponseQuery {
	panic("should not happen")
}

func getIntFromStore(store store.Store, key []byte) int64 {
	bz := store.Get(key)
	if len(bz) == 0 {
		return 0
	}
	i, err := binary.ReadVarint(bytes.NewBuffer(bz))
	if err != nil {
		panic(err)
	}
	return i
}

func setIntOnStore(store store.Store, key []byte, i int64) {
	bz := make([]byte, 8)
	n := binary.PutVarint(bz, i)
	store.Set(key, bz[:n])
}

// check counter matches what's in store.
// increment and store
func incrementingCounter(t *testing.T, store store.Store, counterKey []byte, counter int64) (res Result) {
	t.Helper()

	storedCounter := getIntFromStore(store, counterKey)
	require.Equal(t, storedCounter, counter)
	setIntOnStore(store, counterKey, counter+1)
	return
}

// ---------------------------------------------------------------------
// Tx processing - CheckTx, DeliverTx, SimulateTx.
// These tests use the serialized tx as input, while most others will use the
// Check(), Deliver(), Simulate() methods directly.
// Ensure that Check/Deliver/Simulate work as expected with the store.

// Test that successive CheckTx can see each others' effects
// on the store within a block, and that the CheckTx state
// gets reset to the latest committed state during Commit
func TestCheckTx(t *testing.T) {
	t.Parallel()

	// This ante handler reads the key and checks that the value matches the current counter.
	// This ensures changes to the kvstore persist across successive CheckTx.
	counterKey := []byte("counter-key")

	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, mainKey, counterKey)) }
	routerOpt := func(bapp *BaseApp) {
		// TODO: can remove this once CheckTx doesn't process msgs.
		bapp.Router().AddRoute(routeMsgCounter, newTestHandler(func(ctx Context, msg Msg) Result { return Result{} }))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	nTxs := int64(5)
	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})

	for i := int64(0); i < nTxs; i++ {
		tx := newTxCounter(i, 0)
		txBytes, err := amino.Marshal(tx)
		require.NoError(t, err)
		r := app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
		assert.True(t, r.IsOK(), fmt.Sprintf("%v", r))
	}

	checkStateStore := app.checkState.ctx.Store(mainKey)
	storedCounter := getIntFromStore(checkStateStore, counterKey)

	// Ensure AnteHandler ran
	require.Equal(t, nTxs, storedCounter)

	// If a block is committed, CheckTx state should be reset.
	header := &bft.Header{ChainID: "test-chain", Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	checkStateStore = app.checkState.ctx.Store(mainKey)
	storedBytes := checkStateStore.Get(counterKey)
	require.Nil(t, storedBytes)
}

// Test that successive DeliverTx can see each others' effects
// on the store, both within and across blocks.
func TestDeliverTx(t *testing.T) {
	t.Parallel()

	// test increments in the ante
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, mainKey, anteKey)) }

	// test increments in the handler
	deliverKey := []byte("deliver-key")
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newMsgCounterHandler(t, mainKey, deliverKey))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})

	nBlocks := 3
	txPerHeight := 5

	for blockN := range nBlocks {
		header := &bft.Header{ChainID: "test-chain", Height: int64(blockN) + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		for i := range txPerHeight {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(counter, counter)

			txBytes, err := amino.Marshal(tx)
			require.NoError(t, err)

			res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
		}

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

// Test that the gas used between Simulate and DeliverTx is the same.
func TestGasUsedBetweenSimulateAndDeliver(t *testing.T) {
	t.Parallel()

	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, mainKey, anteKey)) }

	deliverKey := []byte("deliver-key")
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newMsgCounterHandler(t, mainKey, deliverKey))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})

	header := &bft.Header{ChainID: "test-chain", Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	tx := newTxCounter(0, 0)
	txBytes, err := amino.Marshal(tx)
	require.Nil(t, err)

	simulateRes := app.Simulate(txBytes, tx)
	require.True(t, simulateRes.IsOK(), fmt.Sprintf("%v", simulateRes))
	require.Greater(t, simulateRes.GasUsed, int64(0)) // gas used should be greater than 0

	deliverRes := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, deliverRes.IsOK(), fmt.Sprintf("%v", deliverRes))

	require.Equal(t, simulateRes.GasUsed, deliverRes.GasUsed) // gas used should be the same from simulate and deliver
}

// One call to DeliverTx should process all the messages, in order.
func TestMultiMsgDeliverTx(t *testing.T) {
	t.Parallel()

	// increment the tx counter
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, mainKey, anteKey)) }

	// increment the msg counter
	deliverKey := []byte("deliver-key")
	deliverKey2 := []byte("deliver-key2")
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newMsgCounterHandler(t, mainKey, deliverKey))
		bapp.Router().AddRoute(routeMsgCounter2, newMsgCounterHandler(t, mainKey, deliverKey2))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	// run a multi-msg tx
	// with all msgs the same route

	header := &bft.Header{ChainID: "test-chain", Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	tx := newTxCounter(0, 0, 1, 2)
	txBytes, err := amino.Marshal(tx)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	store := app.deliverState.ctx.Store(mainKey)

	// tx counter only incremented once
	txCounter := getIntFromStore(store, anteKey)
	require.Equal(t, int64(1), txCounter)

	// msg counter incremented three times
	msgCounter := getIntFromStore(store, deliverKey)
	require.Equal(t, int64(3), msgCounter)

	// replace the second message with a msgCounter2

	tx = newTxCounter(1, 3)
	tx.Msgs = append(tx.Msgs, msgCounter2{Counter: 0})
	tx.Msgs = append(tx.Msgs, msgCounter2{Counter: 1})
	txBytes, err = amino.Marshal(tx)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	store = app.deliverState.ctx.Store(mainKey)

	// tx counter only incremented once
	txCounter = getIntFromStore(store, anteKey)
	require.Equal(t, int64(2), txCounter)

	// original counter increments by one
	// new counter increments by two
	msgCounter = getIntFromStore(store, deliverKey)
	require.Equal(t, int64(4), msgCounter)
	msgCounter2 := getIntFromStore(store, deliverKey2)
	require.Equal(t, int64(2), msgCounter2)
}

// Simulate a transaction that uses gas to compute the gas.
// Simulate() and Query(".app/simulate", txBytes) should give
// the same results.
func TestSimulateTx(t *testing.T) {
	t.Parallel()

	gasConsumed := int64(5)

	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (newCtx Context, res Result, abort bool) {
			limit := gasConsumed
			newCtx = ctx.WithGasMeter(store.NewGasMeter(limit))
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newTestHandler(func(ctx Context, msg Msg) Result {
			ctx.GasMeter().ConsumeGas(gasConsumed, "test")
			return Result{GasUsed: ctx.GasMeter().GasConsumed()}
		}))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})

	nBlocks := 3
	for blockN := range nBlocks {
		count := int64(blockN + 1)
		header := &bft.Header{ChainID: "test-chain", Height: count}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		tx := newTxCounter(count, count)
		txBytes, err := amino.Marshal(tx)
		require.Nil(t, err)

		// simulate a message, check gas reported
		result := app.Simulate(txBytes, tx)
		require.True(t, result.IsOK(), result.Log)
		require.Equal(t, gasConsumed, result.GasUsed)

		// simulate again, same result
		result = app.Simulate(txBytes, tx)
		require.True(t, result.IsOK(), result.Log)
		require.Equal(t, gasConsumed, result.GasUsed)

		// simulate by calling Query with encoded tx
		query := abci.RequestQuery{
			Path: ".app/simulate",
			Data: txBytes,
		}
		queryResult := app.Query(query)
		require.True(t, queryResult.IsOK(), queryResult.Log)

		var res Result
		require.NoError(t, amino.Unmarshal(queryResult.Value, &res))
		require.Nil(t, err, "Result unmarshalling failed")
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, gasConsumed, res.GasUsed, res.Log)
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

func TestRunInvalidTransaction(t *testing.T) {
	t.Parallel()

	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (newCtx Context, res Result, abort bool) {
			newCtx = ctx
			return
		})
	}
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newTestHandler(func(ctx Context, msg Msg) (res Result) { return }))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	header := &bft.Header{ChainID: "test-chain", Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Transaction with no messages
	{
		emptyTx := std.Tx{}
		err := app.Deliver(emptyTx)
		_, ok := err.Error.(std.UnknownRequestError)
		require.True(t, ok)
	}

	// Transaction where ValidateBasic fails
	{
		testCases := []struct {
			tx   std.Tx
			fail bool
		}{
			{newTxCounter(0, 0), false},
			{newTxCounter(-1, 0), false},
			{newTxCounter(100, 100), false},
			{newTxCounter(100, 5, 4, 3, 2, 1), false},

			{newTxCounter(0, -1), true},
			{newTxCounter(0, 1, -2), true},
			{newTxCounter(0, 1, 2, -10, 5), true},
		}

		for _, testCase := range testCases {
			tx := testCase.tx
			res := app.Deliver(tx)
			if testCase.fail {
				_, ok := res.Error.(std.InvalidSequenceError)
				require.True(t, ok)
			} else {
				require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
			}
		}
	}

	// Transaction with no known route
	{
		unknownRouteTx := std.Tx{Msgs: []Msg{msgNoRoute{}}}
		err := app.Deliver(unknownRouteTx)
		_, ok := err.Error.(std.UnknownRequestError)
		require.True(t, ok)

		unknownRouteTx = std.Tx{Msgs: []Msg{msgCounter{}, msgNoRoute{}}}
		err = app.Deliver(unknownRouteTx)
		_, ok = err.Error.(std.UnknownRequestError)
		require.True(t, ok)
	}

	// Transaction with an unregistered message
	{
		txBytes := []byte{0xFF, 0xFF, 0xFF}
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		_, ok := res.Error.(std.TxDecodeError)
		require.True(t, ok)
	}
}

// Test that transactions exceeding gas limits fail
func TestTxGasLimits(t *testing.T) {
	t.Parallel()

	gasGranted := int64(10)
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (newCtx Context, res Result, abort bool) {
			gmeter := store.NewPassthroughGasMeter(
				ctx.GasMeter(),
				gasGranted,
			)
			newCtx = ctx.WithGasMeter(gmeter)

			count := getCounter(tx)
			newCtx.GasMeter().ConsumeGas(count, "counter-ante")
			res = Result{
				GasWanted: gasGranted,
			}
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newTestHandler(func(ctx Context, msg Msg) Result {
			count := msg.(msgCounter).Counter
			ctx.GasMeter().ConsumeGas(count, "counter-handler")
			return Result{}
		}))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	header := &bft.Header{ChainID: "test-chain", Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	testCases := []struct {
		tx      std.Tx
		gasUsed int64
		fail    bool
	}{
		{newTxCounter(0, 0), 0, false},
		{newTxCounter(1, 1), 2, false},
		{newTxCounter(9, 1), 10, false},
		{newTxCounter(1, 9), 10, false},
		{newTxCounter(10, 0), 10, false},
		{newTxCounter(0, 10), 10, false},
		{newTxCounter(0, 8, 2), 10, false},
		{newTxCounter(0, 5, 1, 1, 1, 1, 1), 10, false},
		{newTxCounter(0, 5, 1, 1, 1, 1), 9, false},

		{newTxCounter(9, 2), 11, true},
		{newTxCounter(2, 9), 11, true},
		{newTxCounter(9, 1, 1), 11, true},
		{newTxCounter(1, 8, 1, 1), 11, true},
		{newTxCounter(11, 0), 11, true},
		{newTxCounter(0, 11), 11, true},
		{newTxCounter(0, 5, 11), 16, true},
	}

	for i, tc := range testCases {
		tx := tc.tx
		res := app.Deliver(tx)

		// check gas used and wanted
		require.Equal(t, tc.gasUsed, res.GasUsed, fmt.Sprintf("%d: %v, %v", i, tc, res))

		// check for out of gas
		if !tc.fail {
			require.True(t, res.IsOK(), fmt.Sprintf("%d: %v, %v", i, tc, res))
		} else {
			_, ok := res.Error.(std.OutOfGasError)
			require.True(t, ok, fmt.Sprintf("%d: %v, %v", i, tc, res))
		}
	}
}

func TestConsensusMaxGasMentionedInOutOfGasLog(t *testing.T) {
	t.Parallel()

	maxGas := int64(50)
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (newCtx Context, res Result, abort bool) {
			res.GasWanted = 100
			return ctx, res, false
		})
	}
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newTestHandler(func(ctx Context, msg Msg) Result {
			ctx.GasMeter().ConsumeGas(store.Gas(maxGas+10), "burn beyond maxGas")
			return Result{}
		}))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.setConsensusParams(&abci.ConsensusParams{
		Block: &abci.BlockParams{MaxGas: maxGas},
	})

	header := &bft.Header{ChainID: "test-chain", Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	txBytes, err := amino.Marshal(newTxCounter(0, 1))
	require.NoError(t, err)

	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})

	require.True(t, res.IsErr())
	_, ok := res.Error.(std.OutOfGasError)
	require.True(t, ok)
	assert.Contains(t, res.Log, "hit consensus maxGas")
}

// Test that transactions exceeding gas limits fail
func TestMaxBlockGasLimits(t *testing.T) {
	t.Parallel()

	gasGranted := int64(10)
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (newCtx Context, res Result, abort bool) {
			gmeter := store.NewPassthroughGasMeter(
				ctx.GasMeter(),
				gasGranted,
			)
			newCtx = ctx.WithGasMeter(gmeter)

			count := getCounter(tx)
			newCtx.GasMeter().ConsumeGas(count, "counter-ante")
			res = Result{
				GasWanted: gasGranted,
			}
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newTestHandler(func(ctx Context, msg Msg) Result {
			count := msg.(msgCounter).Counter
			ctx.GasMeter().ConsumeGas(count, "counter-handler")
			return Result{}
		}))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.InitChain(abci.RequestInitChain{
		ChainID: "test-chain",
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxGas: 100,
			},
		},
	})

	testCases := []struct {
		tx                std.Tx
		numDelivers       int
		gasUsedPerDeliver int64
		fail              bool
		failAfterDeliver  int
	}{
		{newTxCounter(0, 0), 0, 0, false, 0},
		{newTxCounter(9, 1), 2, 10, false, 0},
		{newTxCounter(10, 0), 3, 10, false, 0},
		{newTxCounter(10, 0), 10, 10, false, 0},
		{newTxCounter(2, 7), 11, 9, false, 0},
		{newTxCounter(10, 0), 10, 10, false, 0}, // hit the limit but pass

		{newTxCounter(10, 0), 11, 10, true, 10},
		{newTxCounter(10, 0), 15, 10, true, 10},
		{newTxCounter(9, 0), 12, 9, true, 11}, // fly past the limit
	}

	for i, tc := range testCases {
		tx := tc.tx

		// reset the block gas
		header := &bft.Header{ChainID: "test-chain", Height: app.LastBlockHeight() + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		// execute the transaction multiple times
		for j := range tc.numDelivers {
			res := app.Deliver(tx)

			ctx := app.getState(RunTxModeDeliver).ctx
			blockGasUsed := ctx.BlockGasMeter().GasConsumed()

			// check for failed transactions
			if tc.fail && (j+1) > tc.failAfterDeliver {
				_, ok := res.Error.(std.OutOfGasError)
				require.True(t, ok, fmt.Sprintf("%d: %v, %v", i, tc, res))
				require.True(t, ctx.BlockGasMeter().IsOutOfGas())
			} else {
				// check gas used and wanted
				expBlockGasUsed := tc.gasUsedPerDeliver * int64(j+1)
				require.Equal(t, expBlockGasUsed, blockGasUsed,
					fmt.Sprintf("%d,%d: %v, %v, %v, %v", i, j, tc, expBlockGasUsed, blockGasUsed, res))

				require.True(t, res.IsOK(), fmt.Sprintf("%d,%d: %v, %v", i, j, tc, res))
				require.False(t, ctx.BlockGasMeter().IsPastLimit())
			}
		}
	}
}

func TestBaseAppAnteHandler(t *testing.T) {
	t.Parallel()

	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, mainKey, anteKey))
	}

	deliverKey := []byte("deliver-key")
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newMsgCounterHandler(t, mainKey, deliverKey))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})

	header := &bft.Header{ChainID: "test-chain", Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// execute a tx that will fail ante handler execution
	//
	// NOTE: State should not be mutated here. This will be implicitly checked by
	// the next txs ante handler execution (anteHandlerTxTest).
	tx := newTxCounter(0, 0)
	setFailOnAnte(&tx, true)
	txBytes, err := amino.Marshal(tx)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx := app.getState(RunTxModeDeliver).ctx
	store := ctx.Store(mainKey)
	require.Equal(t, int64(0), getIntFromStore(store, anteKey))

	// execute at tx that will pass the ante handler (the checkTx state should
	// mutate) but will fail the message handler
	tx = newTxCounter(0, 0)
	setFailOnHandler(&tx, true)

	txBytes, err = amino.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx = app.getState(RunTxModeDeliver).ctx
	store = ctx.Store(mainKey)
	require.Equal(t, int64(1), getIntFromStore(store, anteKey))
	require.Equal(t, int64(0), getIntFromStore(store, deliverKey))

	// execute a successful ante handler and message execution where state is
	// implicitly checked by previous tx executions
	tx = newTxCounter(1, 0)

	txBytes, err = amino.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx = app.getState(RunTxModeDeliver).ctx
	store = ctx.Store(mainKey)
	require.Equal(t, int64(2), getIntFromStore(store, anteKey))
	require.Equal(t, int64(1), getIntFromStore(store, deliverKey))

	// commit
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func TestGasConsumptionBadTx(t *testing.T) {
	t.Parallel()

	gasWanted := int64(5)
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (newCtx Context, res Result, abort bool) {
			gmeter := store.NewPassthroughGasMeter(
				ctx.GasMeter(),
				gasWanted,
			)
			newCtx = ctx.WithGasMeter(gmeter)

			newCtx.GasMeter().ConsumeGas(getCounter(tx), "counter-ante")
			if getFailOnAnte(tx) {
				res.Error = ABCIError(std.ErrInternal("ante handler failure"))
				return newCtx, res, true
			}

			res = Result{
				GasWanted: gasWanted,
			}
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newTestHandler(func(ctx Context, msg Msg) Result {
			count := msg.(msgCounter).Counter
			ctx.GasMeter().ConsumeGas(count, "counter-handler")
			return Result{}
		}))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.InitChain(abci.RequestInitChain{
		ChainID: "test-chain",
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxGas: 9,
			},
		},
	})
	// app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})

	header := &bft.Header{ChainID: "test-chain", Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	tx := newTxCounter(5, 0)
	setFailOnAnte(&tx, true)
	txBytes, err := amino.Marshal(tx)
	require.NoError(t, err)

	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	// require next tx to fail due to black gas limit
	tx = newTxCounter(5, 0)
	txBytes, err = amino.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))
}

// Test that we can only query from the latest committed state.
func TestQuery(t *testing.T) {
	t.Parallel()

	key, value := []byte("hello"), []byte("goodbye")
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (newCtx Context, res Result, abort bool) {
			newCtx = ctx
			store := ctx.Store(mainKey)
			store.Set(key, value)
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, newTestHandler(func(ctx Context, msg Msg) Result {
			store := ctx.Store(mainKey)
			store.Set(key, value)
			return Result{}
		}))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})

	// NOTE: "/store/main" tells us Store
	// and the final "/key" says to use the data as the
	// key in the given Store ...
	query := abci.RequestQuery{
		Path: ".store/main/key",
		Data: key,
	}
	tx := newTxCounter(0, 0)

	// query is empty before we do anything
	res := app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a CheckTx
	resTx := app.Check(tx)
	require.True(t, resTx.IsOK(), fmt.Sprintf("%v", resTx))
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a DeliverTx before we commit
	header := &bft.Header{ChainID: "test-chain", Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	resTx = app.Deliver(tx)
	require.True(t, resTx.IsOK(), fmt.Sprintf("%v", resTx))
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query returns correct value after Commit
	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

func TestGetMaximumBlockGas(t *testing.T) {
	app := setupBaseApp(t)

	app.setConsensusParams(&abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 0}})
	require.Equal(t, int64(0), app.getMaximumBlockGas())

	app.setConsensusParams(&abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: -1}})
	require.Equal(t, int64(0), app.getMaximumBlockGas())

	app.setConsensusParams(&abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 5000000}})
	require.Equal(t, int64(5000000), app.getMaximumBlockGas())

	app.setConsensusParams(&abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: -5000000}})
	require.Panics(t, func() { app.getMaximumBlockGas() })
}
