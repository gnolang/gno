package gnoland

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gnostd "github.com/gnolang/gno/gnovm/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests that NewAppWithOptions works even when only providing a simple DB.
func TestNewAppWithOptions(t *testing.T) {
	t.Parallel()

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)
	assert.Equal(t, "dev", bapp.AppVersion())
	assert.Equal(t, "gnoland", bapp.Name())

	addr := crypto.AddressFromPreimage([]byte("test1"))

	appState := DefaultGenState()
	appState.Balances = []Balance{
		{
			Address: addr,
			Amount:  []std.Coin{{Amount: 1e15, Denom: "ugnot"}},
		},
	}
	appState.Txs = []std.Tx{
		{
			Msgs: []std.Msg{vm.NewMsgAddPackage(addr, "gno.land/r/demo", []*std.MemFile{
				{
					Name: "demo.gno",
					Body: "package demo; func Hello() string { return `hello`; }",
				},
			})},
			Fee:        std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
			Signatures: []std.Signature{{}}, // one empty signature
		},
	}

	resp := bapp.InitChain(abci.RequestInitChain{
		Time:    time.Now(),
		ChainID: "dev",
		ConsensusParams: &abci.ConsensusParams{
			Block: defaultBlockParams(),
		},
		Validators: []abci.ValidatorUpdate{},
		AppState:   appState,
	})
	require.True(t, resp.IsOK(), "InitChain response: %v", resp)

	tx := amino.MustMarshal(std.Tx{
		Msgs: []std.Msg{vm.NewMsgCall(addr, nil, "gno.land/r/demo", "Hello", nil)},
		Fee: std.Fee{
			GasWanted: 100_000,
			GasFee: std.Coin{
				Denom:  "ugnot",
				Amount: 1_000_000,
			},
		},
		Signatures: []std.Signature{{}}, // one empty signature
		Memo:       "",
	})
	dtxResp := bapp.DeliverTx(abci.RequestDeliverTx{
		RequestBase: abci.RequestBase{},
		Tx:          tx,
	})
	require.True(t, dtxResp.IsOK(), "DeliverTx response: %v", dtxResp)
}

func TestNewAppWithOptions_ErrNoDB(t *testing.T) {
	t.Parallel()

	_, err := NewAppWithOptions(&AppOptions{})
	assert.ErrorContains(t, err, "no db provided")
}

func TestNewApp(t *testing.T) {
	// NewApp should have good defaults and manage to run InitChain.
	td := t.TempDir()

	app, err := NewApp(td, true, events.NewEventSwitch(), log.NewNoopLogger())
	require.NoError(t, err, "NewApp should be successful")

	resp := app.InitChain(abci.RequestInitChain{
		RequestBase: abci.RequestBase{},
		Time:        time.Time{},
		ChainID:     "dev",
		ConsensusParams: &abci.ConsensusParams{
			Block: defaultBlockParams(),
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{},
			},
		},
		Validators: []abci.ValidatorUpdate{},
		AppState:   DefaultGenState(),
	})
	assert.True(t, resp.IsOK(), "resp is not OK: %v", resp)
}

// Test whether InitChainer calls to load the stdlibs correctly.
func TestInitChainer_LoadStdlib(t *testing.T) {
	t.Parallel()

	t.Run("cached", func(t *testing.T) { testInitChainerLoadStdlib(t, true) })
	t.Run("uncached", func(t *testing.T) { testInitChainerLoadStdlib(t, false) })
}

func testInitChainerLoadStdlib(t *testing.T, cached bool) { //nolint:thelper
	t.Parallel()

	type gsContextType string
	const (
		stdlibDir                   = "test-stdlib-dir"
		gnoStoreKey   gsContextType = "gno-store-key"
		gnoStoreValue gsContextType = "gno-store-value"
	)
	db := memdb.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")

	ms.MountStoreWithDB(baseCapKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()
	testCtx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(), &bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())

	// mock set-up
	var (
		makeCalls             int
		commitCalls           int
		loadStdlibCalls       int
		loadStdlibCachedCalls int
	)
	containsGnoStore := func(ctx sdk.Context) bool {
		return ctx.Context().Value(gnoStoreKey) == gnoStoreValue
	}
	// ptr is pointer to either loadStdlibCalls or loadStdlibCachedCalls
	loadStdlib := func(ptr *int) func(ctx sdk.Context, dir string) {
		return func(ctx sdk.Context, dir string) {
			assert.Equal(t, stdlibDir, dir, "stdlibDir should match provided dir")
			assert.True(t, containsGnoStore(ctx), "should contain gno store")
			*ptr++
		}
	}
	mock := &mockVMKeeper{
		makeGnoTransactionStoreFn: func(ctx sdk.Context) sdk.Context {
			makeCalls++
			assert.False(t, containsGnoStore(ctx), "should not already contain gno store")
			return ctx.WithContext(context.WithValue(ctx.Context(), gnoStoreKey, gnoStoreValue))
		},
		commitGnoTransactionStoreFn: func(ctx sdk.Context) {
			commitCalls++
			assert.True(t, containsGnoStore(ctx), "should contain gno store")
		},
		loadStdlibFn:       loadStdlib(&loadStdlibCalls),
		loadStdlibCachedFn: loadStdlib(&loadStdlibCachedCalls),
	}

	// call initchainer
	cfg := InitChainerConfig{
		StdlibDir:       stdlibDir,
		vmKpr:           mock,
		CacheStdlibLoad: cached,
	}
	// Construct keepers.
	paramsKpr := params.NewKeeper(iavlCapKey, nil)
	cfg.acctKpr = auth.NewAccountKeeper(iavlCapKey, paramsKpr, ProtoGnoAccount)
	cfg.gpKpr = auth.NewGasPriceKeeper(iavlCapKey)
	cfg.InitChainer(testCtx, abci.RequestInitChain{
		AppState: DefaultGenState(),
	})

	// assert number of calls
	assert.Equal(t, 1, makeCalls, "should call MakeGnoTransactionStore once")
	assert.Equal(t, 1, commitCalls, "should call CommitGnoTransactionStore once")
	if cached {
		assert.Equal(t, 0, loadStdlibCalls, "should call LoadStdlib never")
		assert.Equal(t, 1, loadStdlibCachedCalls, "should call LoadStdlibCached once")
	} else {
		assert.Equal(t, 1, loadStdlibCalls, "should call LoadStdlib once")
		assert.Equal(t, 0, loadStdlibCachedCalls, "should call LoadStdlibCached never")
	}
}

// generateValidatorUpdates generates dummy validator updates
func generateValidatorUpdates(t *testing.T, count int) []abci.ValidatorUpdate {
	t.Helper()

	validators := make([]abci.ValidatorUpdate, 0, count)

	for i := 0; i < count; i++ {
		// Generate a random private key
		key := getDummyKey(t)

		validator := abci.ValidatorUpdate{
			Address: key.Address(),
			PubKey:  key,
			Power:   1,
		}

		validators = append(validators, validator)
	}

	return validators
}

func TestEndBlocker(t *testing.T) {
	t.Parallel()

	constructVMResponse := func(updates []abci.ValidatorUpdate) string {
		var builder strings.Builder

		builder.WriteString("(slice[")

		for i, update := range updates {
			builder.WriteString(
				fmt.Sprintf(
					"(struct{(%q std.Address),(%q string),(%d uint64)} gno.land/p/sys/validators.Validator)",
					update.Address,
					update.PubKey,
					update.Power,
				),
			)

			if i < len(updates)-1 {
				builder.WriteString(",")
			}
		}

		builder.WriteString("] []gno.land/p/sys/validators.Validator)")

		return builder.String()
	}

	newCommonEvSwitch := func() *mockEventSwitch {
		var cb events.EventCallback

		return &mockEventSwitch{
			addListenerFn: func(_ string, callback events.EventCallback) {
				cb = callback
			},
			fireEventFn: func(event events.Event) {
				cb(event)
			},
		}
	}

	t.Run("no collector events", func(t *testing.T) {
		t.Parallel()

		noFilter := func(_ events.Event) []validatorUpdate {
			return []validatorUpdate{}
		}

		// Create the collector
		c := newCollector[validatorUpdate](&mockEventSwitch{}, noFilter)

		// Create the EndBlocker
		eb := EndBlocker(c, nil, nil, nil, &mockEndBlockerApp{})

		// Run the EndBlocker
		res := eb(sdk.Context{}, abci.RequestEndBlock{})

		// Verify the response was empty
		assert.Equal(t, abci.ResponseEndBlock{}, res)
	})

	t.Run("invalid VM call", func(t *testing.T) {
		t.Parallel()

		var (
			noFilter = func(_ events.Event) []validatorUpdate {
				return make([]validatorUpdate, 1) // 1 update
			}

			vmCalled bool

			mockEventSwitch = newCommonEvSwitch()

			mockVMKeeper = &mockVMKeeper{
				queryFn: func(_ sdk.Context, pkgPath, expr string) (string, error) {
					vmCalled = true

					require.Equal(t, valRealm, pkgPath)
					require.NotEmpty(t, expr)

					return "", errors.New("random call error")
				},
			}
		)

		// Create the collector
		c := newCollector[validatorUpdate](mockEventSwitch, noFilter)

		// Fire a GnoVM event
		mockEventSwitch.FireEvent(gnostd.GnoEvent{})

		// Create the EndBlocker
		eb := EndBlocker(c, nil, nil, mockVMKeeper, &mockEndBlockerApp{})

		// Run the EndBlocker
		res := eb(sdk.Context{}, abci.RequestEndBlock{})

		// Verify the response was empty
		assert.Equal(t, abci.ResponseEndBlock{}, res)

		// Make sure the VM was called
		assert.True(t, vmCalled)
	})

	t.Run("empty VM response", func(t *testing.T) {
		t.Parallel()

		var (
			noFilter = func(_ events.Event) []validatorUpdate {
				return make([]validatorUpdate, 1) // 1 update
			}

			vmCalled bool

			mockEventSwitch = newCommonEvSwitch()

			mockVMKeeper = &mockVMKeeper{
				queryFn: func(_ sdk.Context, pkgPath, expr string) (string, error) {
					vmCalled = true

					require.Equal(t, valRealm, pkgPath)
					require.NotEmpty(t, expr)

					return constructVMResponse([]abci.ValidatorUpdate{}), nil
				},
			}
		)

		// Create the collector
		c := newCollector[validatorUpdate](mockEventSwitch, noFilter)

		// Fire a GnoVM event
		mockEventSwitch.FireEvent(gnostd.GnoEvent{})

		// Create the EndBlocker
		eb := EndBlocker(c, nil, nil, mockVMKeeper, &mockEndBlockerApp{})

		// Run the EndBlocker
		res := eb(sdk.Context{}, abci.RequestEndBlock{})

		// Verify the response was empty
		assert.Equal(t, abci.ResponseEndBlock{}, res)

		// Make sure the VM was called
		assert.True(t, vmCalled)
	})

	t.Run("multiple valset updates", func(t *testing.T) {
		t.Parallel()

		var (
			changes = generateValidatorUpdates(t, 100)

			mockEventSwitch = newCommonEvSwitch()

			mockVMKeeper = &mockVMKeeper{
				queryFn: func(_ sdk.Context, pkgPath, expr string) (string, error) {
					require.Equal(t, valRealm, pkgPath)
					require.NotEmpty(t, expr)

					return constructVMResponse(changes), nil
				},
			}
		)

		// Create the collector
		c := newCollector[validatorUpdate](mockEventSwitch, validatorEventFilter)

		// Construct the GnoVM events
		vmEvents := make([]abci.Event, 0, len(changes))
		for index := range changes {
			event := gnostd.GnoEvent{
				Type:    validatorAddedEvent,
				PkgPath: valRealm,
			}

			// Make half the changes validator removes
			if index%2 == 0 {
				changes[index].Power = 0

				event = gnostd.GnoEvent{
					Type:    validatorRemovedEvent,
					PkgPath: valRealm,
				}
			}

			vmEvents = append(vmEvents, event)
		}

		// Fire the tx result event
		txEvent := bft.EventTx{
			Result: bft.TxResult{
				Response: abci.ResponseDeliverTx{
					ResponseBase: abci.ResponseBase{
						Events: vmEvents,
					},
				},
			},
		}

		mockEventSwitch.FireEvent(txEvent)

		// Create the EndBlocker
		eb := EndBlocker(c, nil, nil, mockVMKeeper, &mockEndBlockerApp{})

		// Run the EndBlocker
		res := eb(sdk.Context{}, abci.RequestEndBlock{})

		// Verify the response was not empty
		require.Len(t, res.ValidatorUpdates, len(changes))

		for index, update := range res.ValidatorUpdates {
			assert.Equal(t, changes[index].Address, update.Address)
			assert.True(t, changes[index].PubKey.Equals(update.PubKey))
			assert.Equal(t, changes[index].Power, update.Power)
		}
	})
}

func TestGasPriceUpdate(t *testing.T) {
	app, err := newGasPriceTestApp()
	require.NoError(t, err)
	// with default initial gas price 0.1 ugnot per gas
	gnoGen, err := gnoGenesisState()
	require.NoError(t, err)

	// abci inintChain
	app.InitChain(abci.RequestInitChain{
		AppState: gnoGen,
		ChainID:  "test-chain",
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxGas: 10000,
			},
		},
	})
	baseApp := app.(*sdk.BaseApp)
	require.Equal(t, int64(0), baseApp.LastBlockHeight())
	// Case 1
	// CheckTx failed because the GasFee is less than the initial gas price.

	tx := newCounterTx(100)
	tx.Fee = std.Fee{
		GasWanted: 100,
		GasFee: sdk.Coin{
			Amount: 9,
			Denom:  "ugnot",
		},
	}
	txBytes, err := amino.Marshal(tx)
	require.NoError(t, err)
	r := app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
	assert.False(t, r.IsOK(), fmt.Sprintf("%v", r))

	// Case 2:
	// A previously successful CheckTx failed after the block gas price increased.
	// Check Tx Ok
	tx2 := newCounterTx(100)
	tx2.Fee = std.Fee{
		GasWanted: 1000,
		GasFee: sdk.Coin{
			Amount: 100,
			Denom:  "ugnot",
		},
	}
	txBytes2, err := amino.Marshal(tx2)
	require.NoError(t, err)
	r = app.CheckTx(abci.RequestCheckTx{Tx: txBytes2})
	assert.True(t, r.IsOK(), fmt.Sprintf("%v", r))

	// After replaying a block, the gas price increased.
	header := &bft.Header{ChainID: "test-chain", Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	// Delvier Tx consumes more than that target block gas 6000.

	tx6001 := newCounterTx(6001)
	tx6001.Fee = std.Fee{
		GasWanted: 20000,
		GasFee: sdk.Coin{
			Amount: 200,
			Denom:  "ugnot",
		},
	}
	txBytes6001, err := amino.Marshal(tx6001)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes6001})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// CheckTx failed because gas price increased
	r = app.CheckTx(abci.RequestCheckTx{Tx: txBytes2})
	assert.False(t, r.IsOK(), fmt.Sprintf("%v", r))

	// Case 3:
	// A previously failed CheckTx successed after block gas price reduced.

	// CheckTx Failed
	r = app.CheckTx(abci.RequestCheckTx{Tx: txBytes2})
	assert.False(t, r.IsOK(), fmt.Sprintf("%v", r))
	// Replayed a Block, the gas price decrease
	header = &bft.Header{ChainID: "test-chain", Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	// Delvier Tx consumes less than that target block gas 6000.

	tx200 := newCounterTx(200)
	tx200.Fee = std.Fee{
		GasWanted: 20000,
		GasFee: sdk.Coin{
			Amount: 200,
			Denom:  "ugnot",
		},
	}
	txBytes200, err := amino.Marshal(tx200)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes200})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// CheckTx earlier failed tx, now is OK
	r = app.CheckTx(abci.RequestCheckTx{Tx: txBytes2})
	assert.True(t, r.IsOK(), fmt.Sprintf("%v", r))

	// Case 4
	// require matching expected GasPrice after three blocks ( increase case)
	replayBlock(t, baseApp, 8000, 3)
	replayBlock(t, baseApp, 8000, 4)
	replayBlock(t, baseApp, 6000, 5)

	key := []byte("gasPrice")
	query := abci.RequestQuery{
		Path: ".store/main/key",
		Data: key,
	}
	qr := app.Query(query)
	var gp std.GasPrice
	err = amino.Unmarshal(qr.Value, &gp)
	require.NoError(t, err)
	require.Equal(t, "108ugnot", gp.Price.String())

	// Case 5,
	// require matching expected GasPrice after low gas blocks ( decrease below initial gas price case)

	replayBlock(t, baseApp, 5000, 6)
	replayBlock(t, baseApp, 5000, 7)
	replayBlock(t, baseApp, 5000, 8)

	qr = app.Query(query)
	err = amino.Unmarshal(qr.Value, &gp)
	require.NoError(t, err)
	require.Equal(t, "102ugnot", gp.Price.String())

	replayBlock(t, baseApp, 5000, 9)

	qr = app.Query(query)
	err = amino.Unmarshal(qr.Value, &gp)
	require.NoError(t, err)
	require.Equal(t, "100ugnot", gp.Price.String())
}

func newGasPriceTestApp() (abci.Application, error) {
	cfg := TestAppOptions(memdb.NewMemDB())
	cfg.EventSwitch = events.NewEventSwitch()

	// Capabilities keys.
	mainKey := store.NewStoreKey("main")
	baseKey := store.NewStoreKey("base")

	baseApp := sdk.NewBaseApp("gnoland", cfg.Logger, cfg.DB, baseKey, mainKey)
	baseApp.SetAppVersion("test")

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoreWithDB(mainKey, iavl.StoreConstructor, cfg.DB)
	baseApp.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, cfg.DB)

	// Construct keepers.
	paramsKpr := params.NewKeeper(mainKey, nil)
	acctKpr := auth.NewAccountKeeper(mainKey, paramsKpr, ProtoGnoAccount)
	gpKpr := auth.NewGasPriceKeeper(mainKey)
	bankKpr := bank.NewBankKeeper(acctKpr)
	vmk := vm.NewVMKeeper(baseKey, mainKey, acctKpr, bankKpr, cfg.MaxCycles)

	// Set InitChainer
	icc := cfg.InitChainerConfig
	icc.baseApp = baseApp
	icc.acctKpr, icc.bankKpr, icc.vmKpr, icc.gpKpr = acctKpr, bankKpr, vmk, gpKpr
	baseApp.SetInitChainer(icc.InitChainer)

	// Set AnteHandler
	baseApp.SetAnteHandler(
		// Override default AnteHandler with custom logic.
		func(ctx sdk.Context, tx std.Tx, simulate bool) (
			newCtx sdk.Context, res sdk.Result, abort bool,
		) {
			// Add last gas price in the context
			ctx = ctx.WithValue(auth.GasPriceContextKey{}, gpKpr.LastGasPrice(ctx))

			// Override auth params.
			ctx = ctx.WithValue(auth.AuthParamsContextKey{}, acctKpr.GetParams(ctx))
			// Continue on with default auth ante handler.
			if ctx.IsCheckTx() {
				res := auth.EnsureSufficientMempoolFees(ctx, tx.Fee)
				if !res.IsOK() {
					return ctx, res, true
				}
			}

			newCtx = auth.SetGasMeter(false, ctx, tx.Fee.GasWanted)

			count := getTotalCount(tx)

			newCtx.GasMeter().ConsumeGas(count, "counter-ante")
			res = sdk.Result{
				GasWanted: getTotalCount(tx),
			}
			return
		},
	)

	// Set up the event collector
	c := newCollector[validatorUpdate](
		cfg.EventSwitch,      // global event switch filled by the node
		validatorEventFilter, // filter fn that keeps the collector valid
	)

	// Set EndBlocker
	baseApp.SetEndBlocker(
		EndBlocker(
			c,
			acctKpr,
			gpKpr,
			nil,
			baseApp,
		),
	)

	// Set a handler Route.
	baseApp.Router().AddRoute("auth", auth.NewHandler(acctKpr))
	baseApp.Router().AddRoute("bank", bank.NewHandler(bankKpr))
	baseApp.Router().AddRoute(
		testutils.RouteMsgCounter,
		newTestHandler(
			func(ctx sdk.Context, msg sdk.Msg) sdk.Result { return sdk.Result{} },
		),
	)

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

// newTx constructs a tx with multiple counter messages.
// we can use the counter as the gas used for the message.

func newCounterTx(counters ...int64) sdk.Tx {
	msgs := make([]sdk.Msg, len(counters))

	for i, c := range counters {
		msgs[i] = testutils.MsgCounter{Counter: c}
	}
	tx := sdk.Tx{Msgs: msgs}
	return tx
}

func getTotalCount(tx sdk.Tx) int64 {
	var c int64
	for _, m := range tx.Msgs {
		c = +m.(testutils.MsgCounter).Counter
	}
	return c
}

func gnoGenesisState() (GnoGenesisState, error) {
	gen := GnoGenesisState{}
	genBytes := []byte(`{
    "@type": "/gno.GenesisState",
    "auth": {
      "params": {
        "gas_price_change_compressor": "8",
        "initial_gasprice": {
          "gas": "1000",
          "price": "100ugnot"
        },
        "max_memo_bytes": "65536",
        "sig_verify_cost_ed25519": "590",
        "sig_verify_cost_secp256k1": "1000",
        "target_gas_ratio": "60",
        "tx_sig_limit": "7",
        "tx_size_cost_per_byte": "10"
      }
    }
  }`)
	err := amino.UnmarshalJSON(genBytes, &gen)

	return gen, err
}

func replayBlock(t *testing.T, app *sdk.BaseApp, gas int64, hight int64) {
	t.Helper()
	tx := newCounterTx(gas)
	tx.Fee = std.Fee{
		GasWanted: 20000,
		GasFee: sdk.Coin{
			Amount: 1000,
			Denom:  "ugnot",
		},
	}
	txBytes, err := amino.Marshal(tx)
	require.NoError(t, err)

	header := &bft.Header{ChainID: "test-chain", Height: hight}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	// consume gas in the block
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

type testHandler struct {
	process func(sdk.Context, sdk.Msg) sdk.Result
	query   func(sdk.Context, abci.RequestQuery) abci.ResponseQuery
}

func (th testHandler) Process(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	return th.process(ctx, msg)
}

func (th testHandler) Query(ctx sdk.Context, req abci.RequestQuery) abci.ResponseQuery {
	return th.query(ctx, req)
}

func newTestHandler(proc func(sdk.Context, sdk.Msg) sdk.Result) sdk.Handler {
	return testHandler{
		process: proc,
	}
}
