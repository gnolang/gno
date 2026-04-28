package gnoland

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bftCfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/config"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
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
	appState.Txs = []TxWithMetadata{
		{
			Tx: std.Tx{
				Msgs: []std.Msg{vm.NewMsgAddPackage(addr, "gno.land/r/demo", []*std.MemFile{
					{
						Name: "demo.gno",
						Body: "package demo; func Hello(cur realm) string { return `hello`; }",
					},
					{
						Name: "gnomod.toml",
						Body: gnolang.GenGnoModLatest("gno.land/r/demo"),
					},
				})},
				Fee:        std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
				Signatures: []std.Signature{{}}, // one empty signature
			},
		},
	}
	appState.VM.RealmParams = []params.Param{
		params.NewParam("gno.land/r/sys/testrealm:bar_string", "hello"),
		params.NewParam("gno.land/r/sys/testrealm:bar_int64", int64(-42)),
		params.NewParam("gno.land/r/sys/testrealm:bar_uint64", uint64(1337)),
		params.NewParam("gno.land/r/sys/testrealm:bar_bool", true),
		params.NewParam("gno.land/r/sys/testrealm:bar_strings", []string{"some", "strings"}),
		params.NewParam("gno.land/r/sys/testrealm:bar_bytes", []byte{0x48, 0x69, 0x21}),
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

	cres := bapp.Commit()
	require.NotNil(t, cres)

	tcs := []struct {
		path        string
		expectedVal string
	}{
		{"params/vm:gno.land/r/sys/testrealm:bar_string", `"hello"`},
		{"params/vm:gno.land/r/sys/testrealm:bar_int64", `"-42"`},
		{"params/vm:gno.land/r/sys/testrealm:bar_uint64", `"1337"`},
		{"params/vm:gno.land/r/sys/testrealm:bar_bool", `true`},
		{"params/vm:gno.land/r/sys/testrealm:bar_strings", `["some","strings"]`},
		{"params/vm:gno.land/r/sys/testrealm:bar_bytes", string([]byte{0x48, 0x69, 0x21})}, // XXX: make this test more readable
	}

	for _, tc := range tcs {
		qres := bapp.Query(abci.RequestQuery{
			Path: tc.path,
		})
		require.True(t, qres.IsOK())
		assert.Equal(t, qres.Data, []byte(tc.expectedVal))
	}
}

func TestNewAppWithOptions_ErrNoDB(t *testing.T) {
	t.Parallel()

	_, err := NewAppWithOptions(&AppOptions{})
	assert.ErrorContains(t, err, "no db provided")
}

func TestNewAppWithOptions_ErrNoLogger(t *testing.T) {
	t.Parallel()

	opts := TestAppOptions(memdb.NewMemDB())
	opts.Logger = nil
	_, err := NewAppWithOptions(opts)
	assert.ErrorContains(t, err, "no logger provided")
}

func TestNewAppWithOptions_ErrNoEventSwitch(t *testing.T) {
	t.Parallel()

	opts := TestAppOptions(memdb.NewMemDB())
	opts.EventSwitch = nil
	_, err := NewAppWithOptions(opts)
	assert.ErrorContains(t, err, "no event switch provided")
}

func TestNewApp(t *testing.T) {
	// NewApp should have good defaults and manage to run InitChain.
	td := t.TempDir()

	app, err := NewApp(td, NewTestGenesisAppConfig(), config.DefaultAppConfig(), events.NewEventSwitch(), log.NewNoopLogger(), 0)
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
		vmk:             mock,
		acck:            &mockAuthKeeper{},
		bankk:           &mockBankKeeper{},
		prmk:            &mockParamsKeeper{},
		gpk:             &mockGasPriceKeeper{},
		CacheStdlibLoad: cached,
	}

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

	for range count {
		// Generate a random private key
		key := getDummyKey(t).PubKey()

		validator := abci.ValidatorUpdate{
			Address: key.Address(),
			PubKey:  key,
			Power:   1,
		}

		validators = append(validators, validator)
	}

	return validators
}

// generateDummyKeys generates a slice of dummy private keys
func generateDummyKeys(t *testing.T, count int) []crypto.PrivKey {
	t.Helper()

	keys := make([]crypto.PrivKey, 0, count)

	for i := 0; i < count; i++ {
		key := getDummyKey(t)
		keys = append(keys, key)
	}

	return keys
}

func createAndSignTx(
	t *testing.T,
	msgs []std.Msg,
	chainID string,
	key crypto.PrivKey,
) std.Tx {
	t.Helper()

	tx := std.Tx{
		Msgs: msgs,
		Fee: std.Fee{
			GasFee:    std.NewCoin("ugnot", 2000000),
			GasWanted: 10000000,
		},
	}

	signBytes, err := tx.GetSignBytes(chainID, 0, 0)
	require.NoError(t, err)

	// Sign the tx
	signedTx, err := key.Sign(signBytes)
	require.NoError(t, err)

	tx.Signatures = []std.Signature{
		{
			PubKey:    key.PubKey(),
			Signature: signedTx,
		},
	}

	return tx
}

func TestInitChainer_MetadataTxs(t *testing.T) {
	var (
		currentTimestamp = time.Now()
		laterTimestamp   = currentTimestamp.Add(10 * 24 * time.Hour) // 10 days

		getMetadataState = func(tx std.Tx, balances []Balance) GnoGenesisState {
			return GnoGenesisState{
				// Set the package deployment as the genesis tx
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp: laterTimestamp.Unix(),
						},
					},
				},
				// Make sure the deployer account has a balance
				Balances: balances,
				Auth:     auth.DefaultGenesisState(),
				Bank:     bank.DefaultGenesisState(),
				VM:       vm.DefaultGenesisState(),
			}
		}

		getNonMetadataState = func(tx std.Tx, balances []Balance) GnoGenesisState {
			return GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx,
					},
				},
				Balances: balances,
				Auth:     auth.DefaultGenesisState(),
				Bank:     bank.DefaultGenesisState(),
				VM:       vm.DefaultGenesisState(),
			}
		}
	)

	testTable := []struct {
		name         string
		genesisTime  time.Time
		expectedTime time.Time
		stateFn      func(std.Tx, []Balance) GnoGenesisState
	}{
		{
			"non-metadata transaction",
			currentTimestamp,
			currentTimestamp,
			getNonMetadataState,
		},
		{
			"metadata transaction",
			currentTimestamp,
			laterTimestamp,
			getMetadataState,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				db = memdb.NewMemDB()

				key     = getDummyKey(t) // user account, and genesis deployer
				chainID = "test"

				path = "gno.land/r/demo/metadatatx"
				body = `package metadatatx

	import "time"

	// Time is initialized on deployment (genesis)
	var t time.Time = time.Now()

	// GetT returns the time that was saved from genesis
	func GetT(cur realm) int64 { return t.Unix() }
`
			)

			// Create a fresh app instance
			app, err := NewAppWithOptions(TestAppOptions(db))
			require.NoError(t, err)

			// Prepare the deploy transaction
			msg := vm.MsgAddPackage{
				Creator: key.PubKey().Address(),
				Package: &std.MemPackage{
					Name: "metadatatx",
					Path: path,
					Files: []*std.MemFile{
						{
							Name: "file.gno",
							Body: body,
						},
						{
							Name: "gnomod.toml",
							Body: gnolang.GenGnoModLatest(path),
						},
					},
				},
				MaxDeposit: nil,
			}

			// Create the initial genesis tx
			tx := createAndSignTx(t, []std.Msg{msg}, chainID, key)

			// Run the top-level init chain process
			app.InitChain(abci.RequestInitChain{
				ChainID: chainID,
				Time:    testCase.genesisTime,
				ConsensusParams: &abci.ConsensusParams{
					Block: defaultBlockParams(),
					Validator: &abci.ValidatorParams{
						PubKeyTypeURLs: []string{},
					},
				},
				// Set the package deployment as the genesis tx,
				// and make sure the deployer account has a balance
				AppState: testCase.stateFn(tx, []Balance{
					{
						// Make sure the deployer account has a balance
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					},
				}),
			})

			// Prepare the call transaction
			callMsg := vm.MsgCall{
				Caller:  key.PubKey().Address(),
				PkgPath: path,
				Func:    "GetT",
			}

			tx = createAndSignTx(t, []std.Msg{callMsg}, chainID, key)

			// Marshal the transaction to Amino binary
			marshalledTx, err := amino.Marshal(tx)
			require.NoError(t, err)

			// Execute the call to the "GetT" method
			// on the deployed Realm
			resp := app.DeliverTx(abci.RequestDeliverTx{
				Tx: marshalledTx,
			})

			require.True(t, resp.IsOK())

			// Make sure the initialized Realm state is
			// the injected context timestamp from the tx metadata
			assert.Contains(
				t,
				string(resp.Data),
				fmt.Sprintf("(%d int64)", testCase.expectedTime.Unix()),
			)
		})
	}
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
		eb := EndBlocker(c, nil, nil, nil, nil, &mockEndBlockerApp{})

		// Run the EndBlocker
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

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
		mockEventSwitch.FireEvent(chain.Event{})

		// Create the EndBlocker
		eb := EndBlocker(c, nil, nil, mockVMKeeper, nil, &mockEndBlockerApp{})

		// Run the EndBlocker
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

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
		mockEventSwitch.FireEvent(chain.Event{})

		// Create the EndBlocker
		eb := EndBlocker(c, nil, nil, mockVMKeeper, nil, &mockEndBlockerApp{})

		// Run the EndBlocker
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

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
			event := chain.Event{
				Type:    validatorAddedEvent,
				PkgPath: valRealm,
			}

			// Make half the changes validator removes
			if index%2 == 0 {
				changes[index].Power = 0

				event = chain.Event{
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
		eb := EndBlocker(c, nil, nil, mockVMKeeper, nil, &mockEndBlockerApp{})

		// Run the EndBlocker
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		// Verify the response was not empty
		require.Len(t, res.ValidatorUpdates, len(changes))

		for index, update := range res.ValidatorUpdates {
			assert.Equal(t, changes[index].Address, update.Address)
			assert.True(t, changes[index].PubKey.Equals(update.PubKey))
			assert.Equal(t, changes[index].Power, update.Power)
		}
	})

	t.Run("negative power filtered out", func(t *testing.T) {
		t.Parallel()

		var (
			keys = generateDummyKeys(t, 2)

			validUpdate = abci.ValidatorUpdate{
				Address: keys[0].PubKey().Address(),
				PubKey:  keys[0].PubKey(),
				Power:   1,
			}

			invalidUpdate = abci.ValidatorUpdate{
				Address: keys[1].PubKey().Address(),
				PubKey:  keys[1].PubKey(),
				Power:   -1, // Invalid negative power
			}

			updates = []abci.ValidatorUpdate{validUpdate, invalidUpdate}

			mockEventSwitch = newCommonEvSwitch()

			mockVMKeeper = &mockVMKeeper{
				queryFn: func(_ sdk.Context, pkgPath, expr string) (string, error) {
					require.Equal(t, valRealm, pkgPath)
					require.NotEmpty(t, expr)

					return constructVMResponse(updates), nil
				},
			}

			vmEvents = []abci.Event{
				chain.Event{
					Type:    validatorAddedEvent,
					PkgPath: valRealm,
				},
				chain.Event{
					Type:    validatorAddedEvent,
					PkgPath: valRealm,
				},
			}
			txEvent = bft.EventTx{
				Result: bft.TxResult{
					Response: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Events: vmEvents,
						},
					},
				},
			}
		)

		c := newCollector[validatorUpdate](mockEventSwitch, validatorEventFilter)
		mockEventSwitch.FireEvent(txEvent)

		eb := EndBlocker(c, nil, nil, mockVMKeeper, nil, &mockEndBlockerApp{})
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})
		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, validUpdate.Address, res.ValidatorUpdates[0].Address)
		assert.Equal(t, validUpdate.Power, res.ValidatorUpdates[0].Power)
	})

	t.Run("pubkey address mismatch filtered out", func(t *testing.T) {
		t.Parallel()

		var (
			keys = generateDummyKeys(t, 3)

			validUpdate = abci.ValidatorUpdate{
				Address: keys[0].PubKey().Address(),
				PubKey:  keys[0].PubKey(),
				Power:   1,
			}

			invalidUpdate = abci.ValidatorUpdate{
				Address: keys[1].PubKey().Address(), // Address from key1
				PubKey:  keys[2].PubKey(),           // PubKey from key2 (mismatch)
				Power:   1,
			}

			updates = []abci.ValidatorUpdate{validUpdate, invalidUpdate}

			mockEventSwitch = newCommonEvSwitch()

			mockVMKeeper = &mockVMKeeper{
				queryFn: func(_ sdk.Context, pkgPath, expr string) (string, error) {
					require.Equal(t, valRealm, pkgPath)
					require.NotEmpty(t, expr)

					return constructVMResponse(updates), nil
				},
			}

			vmEvents = []abci.Event{
				chain.Event{
					Type:    validatorAddedEvent,
					PkgPath: valRealm,
				},
				chain.Event{
					Type:    validatorAddedEvent,
					PkgPath: valRealm,
				},
			}
			txEvent = bft.EventTx{
				Result: bft.TxResult{
					Response: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Events: vmEvents,
						},
					},
				},
			}
		)

		c := newCollector[validatorUpdate](mockEventSwitch, validatorEventFilter)
		mockEventSwitch.FireEvent(txEvent)
		eb := EndBlocker(c, nil, nil, mockVMKeeper, nil, &mockEndBlockerApp{})
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		// Verify only the valid update is returned
		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, validUpdate.Address, res.ValidatorUpdates[0].Address)
		assert.True(t, validUpdate.PubKey.Equals(res.ValidatorUpdates[0].PubKey))
	})

	t.Run("wrong pubkey type", func(t *testing.T) {
		t.Parallel()

		var (
			key1 = getDummyKey(t)

			updates = []abci.ValidatorUpdate{
				{
					Address: key1.PubKey().Address(),
					PubKey:  key1.PubKey(),
					Power:   1,
				},
			}

			mockEventSwitch = newCommonEvSwitch()

			mockVMKeeper = &mockVMKeeper{
				queryFn: func(_ sdk.Context, pkgPath, expr string) (string, error) {
					require.Equal(t, valRealm, pkgPath)
					require.NotEmpty(t, expr)

					return constructVMResponse(updates), nil
				},
			}
			txEvent = bft.EventTx{
				Result: bft.TxResult{
					Response: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Events: []abci.Event{
								chain.Event{
									Type:    validatorAddedEvent,
									PkgPath: valRealm,
								},
							},
						},
					},
				},
			}
		)

		c := newCollector[validatorUpdate](mockEventSwitch, validatorEventFilter)
		mockEventSwitch.FireEvent(txEvent)
		eb := EndBlocker(c, nil, nil, mockVMKeeper, nil, &mockEndBlockerApp{})
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeyEd25519"},
			},
		}), abci.RequestEndBlock{})

		// Verify only the valid update is returned
		require.Len(t, res.ValidatorUpdates, 0)
	})

	t.Run("extract updates error", func(t *testing.T) {
		t.Parallel()

		var (
			noFilter = func(_ events.Event) []validatorUpdate {
				return make([]validatorUpdate, 1) // 1 update
			}

			mockEventSwitchInner = newCommonEvSwitch()

			mockVMKeeperInner = &mockVMKeeper{
				queryFn: func(_ sdk.Context, pkgPath, expr string) (string, error) {
					require.Equal(t, valRealm, pkgPath)
					// Return a response that matches the regex but has an invalid bech32 address.
					// This causes extractUpdatesFromResponse to return an error.
					return `{("notabech32" std.Address),("notapubkey" string),(1 uint64)}`, nil
				},
			}
		)

		c := newCollector[validatorUpdate](mockEventSwitchInner, noFilter)
		mockEventSwitchInner.FireEvent(chain.Event{})

		eb := EndBlocker(c, nil, nil, mockVMKeeperInner, nil, &mockEndBlockerApp{})
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		// Error from extractUpdatesFromResponse → EndBlocker returns empty response
		assert.Equal(t, abci.ResponseEndBlock{}, res)
	})
}

func TestGasPriceUpdate(t *testing.T) {
	app := newGasPriceTestApp(t)

	// with default initial gas price 0.1 ugnot per gas
	gnoGen := gnoGenesisState(t)

	// abci inintChain
	app.InitChain(abci.RequestInitChain{
		AppState: gnoGen,
		ChainID:  "test-chain",
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxGas: 1_000_000,
			},
		},
	})
	baseApp := app.(*sdk.BaseApp)
	require.Equal(t, int64(0), baseApp.LastBlockHeight())
	// Case 1
	// CheckTx failed because the GasFee is less than the initial gas price.

	tx := newCounterTx(100)
	tx.Fee = std.Fee{
		GasWanted: 10000,
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
		GasWanted: 100000,
		GasFee: sdk.Coin{
			Amount: 10000,
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
	// Delvier Tx consumes more than that target block gas 600000.

	tx6001 := newCounterTx(610000)
	tx6001.Fee = std.Fee{
		GasWanted: 2000000,
		GasFee: sdk.Coin{
			Amount: 200000,
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
	// Delvier Tx consumes less than that target block gas 600000.

	tx200 := newCounterTx(20000)
	tx200.Fee = std.Fee{
		GasWanted: 2000000,
		GasFee: sdk.Coin{
			Amount: 200000,
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
	replayBlock(t, baseApp, 800000, 3)
	replayBlock(t, baseApp, 800000, 4)
	replayBlock(t, baseApp, 600000, 5)

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

	replayBlock(t, baseApp, 500000, 6)
	replayBlock(t, baseApp, 500000, 7)
	replayBlock(t, baseApp, 500000, 8)

	qr = app.Query(query)
	err = amino.Unmarshal(qr.Value, &gp)
	require.NoError(t, err)
	require.Equal(t, "102ugnot", gp.Price.String())

	replayBlock(t, baseApp, 500000, 9)

	qr = app.Query(query)
	err = amino.Unmarshal(qr.Value, &gp)
	require.NoError(t, err)
	require.Equal(t, "100ugnot", gp.Price.String())
}

func newGasPriceTestApp(t *testing.T) abci.Application {
	t.Helper()
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
	prmk := params.NewParamsKeeper(mainKey)
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, std.ProtoBaseSessionAccount)
	gpk := auth.NewGasPriceKeeper(mainKey)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName))
	vmk := vm.NewVMKeeper(baseKey, mainKey, acck, bankk, prmk)
	prmk.Register(auth.ModuleName, acck)
	prmk.Register(bank.ModuleName, bankk)
	prmk.Register(vm.ModuleName, vmk)
	// Set InitChainer
	icc := cfg.InitChainerConfig
	icc.baseApp = baseApp
	icc.acck, icc.bankk, icc.vmk, icc.gpk = acck, bankk, vmk, gpk
	baseApp.SetInitChainer(icc.InitChainer)

	// Set AnteHandler
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
			if ctx.IsCheckTx() {
				res := auth.EnsureSufficientMempoolFees(ctx, tx.Fee)
				if !res.IsOK() {
					return ctx, res, true
				}
			}

			newCtx = auth.SetGasMeter(ctx, tx.Fee.GasWanted)

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
			acck,
			gpk,
			nil,
			nil,
			baseApp,
		),
	)

	// Set a handler Route.
	baseApp.Router().AddRoute("auth", auth.NewHandler(acck, gpk))
	baseApp.Router().AddRoute("bank", bank.NewHandler(bankk))
	baseApp.Router().AddRoute(
		testutils.RouteMsgCounter,
		newTestHandler(
			func(ctx sdk.Context, msg sdk.Msg) sdk.Result { return sdk.Result{} },
		),
	)

	baseApp.Router().AddRoute("vm", vm.NewHandler(vmk))

	// Load latest version.
	if err := baseApp.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load the lastest state: %v", err)
	}

	// Initialize the VMKeeper.
	ms := baseApp.GetCacheMultiStore()
	vmk.Initialize(cfg.Logger, ms)
	ms.MultiWrite() // XXX why was't this needed?

	return baseApp
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

func gnoGenesisState(t *testing.T) GnoGenesisState {
	t.Helper()
	gen := GnoGenesisState{}
	genBytes := []byte(`{
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
        "tx_size_cost_per_byte": "10",
        "fee_collector": "g1najfm5t7dr4f2m38cg55xt6gh2lxsk92tgh0xy"
      }
    }
  }`)
	err := amino.UnmarshalJSON(genBytes, &gen)

	gen.Bank = bank.DefaultGenesisState()
	gen.VM = vm.DefaultGenesisState()

	if err != nil {
		t.Fatalf("failed to create genesis state: %v", err)
	}
	return gen
}

func replayBlock(t *testing.T, app *sdk.BaseApp, gas int64, hight int64) {
	t.Helper()
	tx := newCounterTx(gas)
	tx.Fee = std.Fee{
		GasWanted: 2000000,
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

func TestPruneStrategyNothing(t *testing.T) {
	t.Parallel()

	var (
		chainID = "dev"
		appDir  = t.TempDir()
	)

	appCfg := config.DefaultAppConfig()
	appCfg.PruneStrategy = types.PruneNothingStrategy

	app, err := NewApp(
		appDir,
		NewTestGenesisAppConfig(),
		appCfg,
		events.NewEventSwitch(),
		log.NewNoopLogger(),
		0,
	)
	require.NoError(t, err)

	base := app.(*sdk.BaseApp)

	// Run the genesis initialization, and commit it
	base.InitChain(abci.RequestInitChain{
		ChainID: chainID,
		Time:    time.Now(),
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{MaxGas: 1_000_000},
		},
		AppState: DefaultGenState(),
	})
	base.Commit()

	// Simulate a few empty blocks being committed
	startHeight := base.LastBlockHeight() + 1
	for h := startHeight; h <= startHeight+5; h++ {
		base.BeginBlock(abci.RequestBeginBlock{
			Header: &bft.Header{ChainID: chainID, Height: h},
		})

		base.EndBlock(abci.RequestEndBlock{})

		base.Commit()
	}

	// Close the app, so it releases the DB
	require.NoError(t, base.Close())

	// Reopen the same DB
	db, err := dbm.NewDB(
		"gnolang",
		dbm.PebbleDBBackend,
		filepath.Join(appDir, bftCfg.DefaultDBDir),
	)
	require.NoError(t, err)

	var (
		mainKey = store.NewStoreKey("main")
		baseKey = store.NewStoreKey("base")
	)

	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(mainKey, iavl.StoreConstructor, db)
	cms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)

	// Make sure loading a past version doesn't fail
	assert.NoError(t, cms.LoadVersion(1))

	err = db.Close()
	require.NoError(t, err)
}

func TestNodeParamsKeeperWillSetParam(t *testing.T) {
	t.Parallel()

	npk := nodeParamsKeeper{}

	t.Run("valid halt_height (no block context)", func(t *testing.T) {
		t.Parallel()
		// Without a block header, safeBlockHeight returns 0, so no future check.
		assert.NotPanics(t, func() {
			npk.WillSetParam(sdk.Context{}, "p:halt_height", int64(100))
		})
	})

	t.Run("halt_height zero is allowed (cancel sentinel)", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			npk.WillSetParam(sdk.Context{}, "p:halt_height", int64(0))
		})
	})

	t.Run("halt_height in the future is valid when block height is known", func(t *testing.T) {
		t.Parallel()
		ctx := sdk.Context{}.WithBlockHeader(&bft.Header{Height: 50})
		assert.NotPanics(t, func() {
			npk.WillSetParam(ctx, "p:halt_height", int64(100))
		})
	})

	t.Run("halt_height equal to current block height panics", func(t *testing.T) {
		t.Parallel()
		ctx := sdk.Context{}.WithBlockHeader(&bft.Header{Height: 100})
		assert.Panics(t, func() {
			npk.WillSetParam(ctx, "p:halt_height", int64(100))
		})
	})

	t.Run("halt_height in the past panics", func(t *testing.T) {
		t.Parallel()
		ctx := sdk.Context{}.WithBlockHeader(&bft.Header{Height: 200})
		assert.Panics(t, func() {
			npk.WillSetParam(ctx, "p:halt_height", int64(100))
		})
	})

	t.Run("negative halt_height panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "p:halt_height", int64(-1))
		})
	})

	t.Run("halt_height wrong type panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "p:halt_height", "not-an-int64")
		})
	})

	t.Run("valid halt_min_version", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			npk.WillSetParam(sdk.Context{}, "p:halt_min_version", "chain/gnoland1.1")
		})
	})

	t.Run("empty halt_min_version is allowed", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			npk.WillSetParam(sdk.Context{}, "p:halt_min_version", "")
		})
	})

	t.Run("halt_min_version wrong type panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "p:halt_min_version", int64(1))
		})
	})

	t.Run("unknown p: key panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "p:unknown_key", int64(0))
		})
	})

	t.Run("non-p: key is allowed", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			npk.WillSetParam(sdk.Context{}, "other:key", "value")
		})
	})
}

func TestMeetsMinVersion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		binary string
		minVer string
		want   bool
	}{
		// Empty minVersion always passes
		{"chain/gnoland1.0", "", true},
		{"develop", "", true},

		// Same version passes
		{"chain/gnoland1.0", "chain/gnoland1.0", true},
		{"chain/gnoland1.1", "chain/gnoland1.1", true},

		// Newer binary passes
		{"chain/gnoland1.1", "chain/gnoland1.0", true},
		{"chain/gnoland2.0", "chain/gnoland1.0", true},
		{"chain/gnoland1.2", "chain/gnoland1.1", true},

		// Older binary fails
		{"chain/gnoland1.0", "chain/gnoland1.1", false},
		{"chain/gnoland1.0", "chain/gnoland2.0", false},

		// Non-gnoland format: requires exact match
		{"develop", "chain/gnoland1.1", false},
		{"v1.0.0", "v1.0.0", true},
		{"v1.0.0", "v1.1.0", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.binary+">="+tc.minVer, func(t *testing.T) {
			t.Parallel()
			got := meetsMinVersion(tc.binary, tc.minVer)
			assert.Equal(t, tc.want, got,
				"meetsMinVersion(%q, %q)", tc.binary, tc.minVer)
		})
	}
}

func TestParseGnolandVersion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		major int
		minor int
		ok    bool
	}{
		{"chain/gnoland1.0", 1, 0, true},
		{"chain/gnoland1.1", 1, 1, true},
		{"chain/gnoland2.3", 2, 3, true},
		{"develop", 0, 0, false},
		{"v1.0.0", 0, 0, false},
		{"chain/gnoland", 0, 0, false},
		{"chain/gnolandX.Y", 0, 0, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			major, minor, ok := parseGnolandVersion(tc.input)
			assert.Equal(t, tc.ok, ok)
			if tc.ok {
				assert.Equal(t, tc.major, major)
				assert.Equal(t, tc.minor, minor)
			}
		})
	}
}

// newTestParamsKeeper creates a minimal ParamsKeeper with an in-memory store
// and pre-seeds it with the given halt params.
func newTestParamsKeeper(t *testing.T, haltHeight int64, minVersion string) (params.ParamsKeeper, store.MultiStore) {
	t.Helper()

	db := memdb.NewMemDB()
	mainKey := store.NewStoreKey("main")

	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(mainKey, iavl.StoreConstructor, db)
	require.NoError(t, cms.LoadLatestVersion())

	prmk := params.NewParamsKeeper(mainKey)
	prmk.Register("node", nodeParamsKeeper{})

	ms := cms.MultiCacheWrap()
	ctx := sdk.Context{}.WithMultiStore(ms).WithChainID("_")

	prmk.SetInt64(ctx, nodeParamHaltHeight, haltHeight)
	prmk.SetString(ctx, nodeParamHaltMinVersion, minVersion)
	ms.MultiWrite()
	cms.Commit()

	return prmk, cms.MultiCacheWrap()
}

func TestCheckNodeStartupParams(t *testing.T) {
	t.Parallel()

	t.Run("no halt configured", func(t *testing.T) {
		t.Parallel()
		prmk, ms := newTestParamsKeeper(t, 0, "")
		require.NoError(t, checkNodeStartupParams(prmk, ms, 50, 0))
	})

	t.Run("halt with no version passes", func(t *testing.T) {
		t.Parallel()
		prmk, ms := newTestParamsKeeper(t, 100, "")
		require.NoError(t, checkNodeStartupParams(prmk, ms, 100, 0))
	})

	t.Run("binary meets version after halt", func(t *testing.T) {
		t.Parallel()
		prmk, ms := newTestParamsKeeper(t, 100, "develop")
		// binary "develop" == "develop" -> meetsMinVersion (exact match), lastBlock >= haltHeight
		require.NoError(t, checkNodeStartupParams(prmk, ms, 100, 0))
	})

	t.Run("old binary rejected after halt", func(t *testing.T) {
		t.Parallel()
		prmk, ms := newTestParamsKeeper(t, 100, "chain/gnoland9.9")
		// binary "develop" doesn't meet "chain/gnoland9.9" -> rejected
		err := checkNodeStartupParams(prmk, ms, 100, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not meet the minimum version")
	})

	t.Run("new binary rejected before halt height", func(t *testing.T) {
		t.Parallel()
		prmk, ms := newTestParamsKeeper(t, 100, "develop")
		// binary "develop" == "develop" -> meetsMinVersion, but chain hasn't halted yet
		err := checkNodeStartupParams(prmk, ms, 50, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upgrade intended for halt height")
	})

	t.Run("old binary allowed before halt height", func(t *testing.T) {
		t.Parallel()
		prmk, ms := newTestParamsKeeper(t, 100, "chain/gnoland9.9")
		// binary "develop" doesn't meet "chain/gnoland9.9", chain hasn't halted -> old binary, OK
		require.NoError(t, checkNodeStartupParams(prmk, ms, 50, 0))
	})

	t.Run("skip_upgrade_height bypasses check", func(t *testing.T) {
		t.Parallel()
		prmk, ms := newTestParamsKeeper(t, 100, "develop")
		// Even though binary meets version before halt, skip_upgrade_height=100 bypasses
		require.NoError(t, checkNodeStartupParams(prmk, ms, 50, 100))
	})
}

func TestEndBlockerHalt(t *testing.T) {
	t.Parallel()

	noFilter := func(_ events.Event) []validatorUpdate { return nil }

	t.Run("halts at exact height", func(t *testing.T) {
		t.Parallel()

		var haltSet uint64
		mockApp := &mockEndBlockerApp{
			setHaltHeightFn: func(h uint64) { haltSet = h },
		}
		mockPrmk := &mockConfigurableParamsKeeper{
			int64s: map[string]int64{nodeParamHaltHeight: 100},
		}

		c := newCollector[validatorUpdate](&mockEventSwitch{}, noFilter)
		eb := EndBlocker(c, nil, nil, nil, mockPrmk, mockApp)
		eb(sdk.Context{}, abci.RequestEndBlock{Height: 100})

		assert.Equal(t, uint64(100), haltSet, "SetHaltHeight should be called with halt_height")
	})

	t.Run("does not halt before halt height", func(t *testing.T) {
		t.Parallel()

		var haltSet uint64
		mockApp := &mockEndBlockerApp{
			setHaltHeightFn: func(h uint64) { haltSet = h },
		}
		mockPrmk := &mockConfigurableParamsKeeper{
			int64s: map[string]int64{nodeParamHaltHeight: 100},
		}

		c := newCollector[validatorUpdate](&mockEventSwitch{}, noFilter)
		eb := EndBlocker(c, nil, nil, nil, mockPrmk, mockApp)
		eb(sdk.Context{}, abci.RequestEndBlock{Height: 99})

		assert.Equal(t, uint64(0), haltSet, "SetHaltHeight should NOT be called before halt height")
	})

	t.Run("does not re-halt after halt height (no infinite loop)", func(t *testing.T) {
		t.Parallel()

		var haltSet uint64
		mockApp := &mockEndBlockerApp{
			setHaltHeightFn: func(h uint64) { haltSet = h },
		}
		mockPrmk := &mockConfigurableParamsKeeper{
			int64s: map[string]int64{nodeParamHaltHeight: 100},
		}

		c := newCollector[validatorUpdate](&mockEventSwitch{}, noFilter)
		eb := EndBlocker(c, nil, nil, nil, mockPrmk, mockApp)
		// After restart at height 101, halt_height=100 still in params but == doesn't re-fire
		eb(sdk.Context{}, abci.RequestEndBlock{Height: 101})

		assert.Equal(t, uint64(0), haltSet, "SetHaltHeight must NOT be called after halt height (prevents infinite loop)")
	})

	t.Run("cancel: halt_height zero never halts", func(t *testing.T) {
		t.Parallel()

		var haltSet uint64
		mockApp := &mockEndBlockerApp{
			setHaltHeightFn: func(h uint64) { haltSet = h },
		}
		mockPrmk := &mockConfigurableParamsKeeper{
			int64s: map[string]int64{nodeParamHaltHeight: 0},
		}

		c := newCollector[validatorUpdate](&mockEventSwitch{}, noFilter)
		eb := EndBlocker(c, nil, nil, nil, mockPrmk, mockApp)
		eb(sdk.Context{}, abci.RequestEndBlock{Height: 100})

		assert.Equal(t, uint64(0), haltSet, "SetHaltHeight should NOT be called when halt_height=0 (cancelled)")
	})
}

func TestExtractUpdatesFromResponse(t *testing.T) {
	t.Parallel()

	t.Run("empty response returns nil", func(t *testing.T) {
		t.Parallel()
		updates, err := extractUpdatesFromResponse("")
		require.NoError(t, err)
		assert.Nil(t, updates)
	})

	t.Run("no regex match returns nil", func(t *testing.T) {
		t.Parallel()
		updates, err := extractUpdatesFromResponse("some random string with no validator data")
		require.NoError(t, err)
		assert.Nil(t, updates)
	})

	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()
		// The regex captures any quoted string as the address, so we can inject an invalid bech32.
		response := `{("notabech32" std.Address),("notapubkey" string),(1 uint64)}`
		_, err := extractUpdatesFromResponse(response)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse address")
	})

	t.Run("invalid pubkey", func(t *testing.T) {
		t.Parallel()
		// Valid bech32 address, but invalid pubkey string.
		addr := crypto.AddressFromPreimage([]byte("test"))
		response := fmt.Sprintf(`{(%q std.Address),("notapubkey" string),(1 uint64)}`, addr)
		_, err := extractUpdatesFromResponse(response)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse public key")
	})
}
