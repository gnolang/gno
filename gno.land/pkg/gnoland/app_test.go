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
	resp := bapp.InitChain(abci.RequestInitChain{
		Time:    time.Now(),
		ChainID: "dev",
		ConsensusParams: &abci.ConsensusParams{
			Block: defaultBlockParams(),
		},
		Validators: []abci.ValidatorUpdate{},
		AppState: GnoGenesisState{
			Balances: []Balance{
				{
					Address: addr,
					Amount:  []std.Coin{{Amount: 1e15, Denom: "ugnot"}},
				},
			},
			Txs: []std.Tx{
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
			},
		},
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
		AppState:   GnoGenesisState{},
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
	cfg.InitChainer(testCtx, abci.RequestInitChain{
		AppState: GnoGenesisState{},
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
		eb := EndBlocker(c, nil, &mockEndBlockerApp{})

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
		eb := EndBlocker(c, mockVMKeeper, &mockEndBlockerApp{})

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
		eb := EndBlocker(c, mockVMKeeper, &mockEndBlockerApp{})

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
		eb := EndBlocker(c, mockVMKeeper, &mockEndBlockerApp{})

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
