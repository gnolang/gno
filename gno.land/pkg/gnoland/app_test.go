package gnoland

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
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

// endBlockerParamsMock is a ParamsKeeperI mock with optional per-method
// hooks, scoped to TestEndBlocker. Unset hooks are no-ops, matching the
// minimal-by-default behavior of mockParamsKeeper but adding per-key
// observation/injection where each subtest needs it.
type endBlockerParamsMock struct {
	getStringFn  func(sdk.Context, string, *string)
	getInt64Fn   func(sdk.Context, string, *int64)
	getBoolFn    func(sdk.Context, string, *bool)
	getStringsFn func(sdk.Context, string, *[]string)
	setBoolFn    func(sdk.Context, string, bool)
	setStringsFn func(sdk.Context, string, []string)
}

func (m *endBlockerParamsMock) GetString(ctx sdk.Context, key string, ptr *string) {
	if m.getStringFn != nil {
		m.getStringFn(ctx, key, ptr)
	}
}

func (m *endBlockerParamsMock) GetInt64(ctx sdk.Context, key string, ptr *int64) {
	if m.getInt64Fn != nil {
		m.getInt64Fn(ctx, key, ptr)
	}
}

func (m *endBlockerParamsMock) GetBool(ctx sdk.Context, key string, ptr *bool) {
	if m.getBoolFn != nil {
		m.getBoolFn(ctx, key, ptr)
	}
}

func (m *endBlockerParamsMock) GetStrings(ctx sdk.Context, key string, ptr *[]string) {
	if m.getStringsFn != nil {
		m.getStringsFn(ctx, key, ptr)
	}
}

func (m *endBlockerParamsMock) SetBool(ctx sdk.Context, key string, value bool) {
	if m.setBoolFn != nil {
		m.setBoolFn(ctx, key, value)
	}
}

func (m *endBlockerParamsMock) SetStrings(ctx sdk.Context, key string, value []string) {
	if m.setStringsFn != nil {
		m.setStringsFn(ctx, key, value)
	}
}

// Remaining ParamsKeeperI methods are not exercised by EndBlocker.
func (m *endBlockerParamsMock) GetUint64(sdk.Context, string, *uint64)         {}
func (m *endBlockerParamsMock) GetBytes(sdk.Context, string, *[]byte)          {}
func (m *endBlockerParamsMock) SetString(sdk.Context, string, string)          {}
func (m *endBlockerParamsMock) SetInt64(sdk.Context, string, int64)            {}
func (m *endBlockerParamsMock) SetUint64(sdk.Context, string, uint64)          {}
func (m *endBlockerParamsMock) SetBytes(sdk.Context, string, []byte)           {}
func (m *endBlockerParamsMock) Has(sdk.Context, string) bool                   { return false }
func (m *endBlockerParamsMock) GetStruct(sdk.Context, string, any)             {}
func (m *endBlockerParamsMock) SetStruct(sdk.Context, string, any)             {}
func (m *endBlockerParamsMock) GetAny(sdk.Context, string) any                 { return nil }
func (m *endBlockerParamsMock) SetAny(sdk.Context, string, any)                {}

func TestEndBlocker(t *testing.T) {
	t.Parallel()

	t.Run("no valset changes", func(t *testing.T) {
		t.Parallel()

		var (
			mockParamsKeeper = &endBlockerParamsMock{
				getStringFn: func(_ sdk.Context, key string, ptr *string) {
					// valset realm lookup - return default
				},
				getBoolFn: func(_ sdk.Context, key string, ptr *bool) {
					// updatesAvailable stays false (default)
				},
			}

			mockApp = &mockEndBlockerApp{}
		)

		eb := EndBlocker(mockParamsKeeper, nil, nil, mockApp)

		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		assert.Equal(t, abci.ResponseEndBlock{}, res)
	})

	t.Run("invalid valset changes in prev recovers", func(t *testing.T) {
		t.Parallel()

		// Recovery contract: a corrupted prev valset must not wedge consensus.
		// EndBlocker logs loudly, advances prev to proposed, and clears the
		// pending-updates flag so subsequent proposals can land.
		proposed := []string{} // empty proposed set is fine

		var (
			updateFlag   = true
			paramUpdates []string

			mockParamsKeeper = &endBlockerParamsMock{
				getStringFn: func(_ sdk.Context, key string, ptr *string) {},
				getStringsFn: func(_ sdk.Context, key string, ptr *[]string) {
					switch key {
					case valsetParamPath(vm.ValsetRealmDefault, valsetPrevKey):
						*ptr = []string{"totally invalid format"}
					case valsetParamPath(vm.ValsetRealmDefault, valsetNewKey):
						*ptr = proposed
					}
				},
				getBoolFn: func(_ sdk.Context, key string, ptr *bool) {
					if key == valsetParamPath(vm.ValsetRealmDefault, valsetDirtyKey) {
						*ptr = updateFlag
					}
				},
				setBoolFn: func(_ sdk.Context, key string, value bool) {
					updateFlag = value
				},
				setStringsFn: func(_ sdk.Context, key string, value []string) {
					paramUpdates = value
				},
			}

			mockApp = &mockEndBlockerApp{}
		)

		eb := EndBlocker(mockParamsKeeper, nil, nil, mockApp)

		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		assert.Equal(t, abci.ResponseEndBlock{}, res)
		// Flag cleared, prev advanced to proposed (recovery applied).
		assert.False(t, updateFlag, "flag must be cleared so future proposals land")
		assert.Equal(t, proposed, paramUpdates, "prev must advance to proposed")
	})

	t.Run("invalid valset changes in proposed recovers", func(t *testing.T) {
		t.Parallel()

		// Recovery contract for proposed parse failure: clear the flag so a
		// future re-propose can land. We do NOT touch prev (it's still good).
		var (
			updateFlag      = true
			prevSetCalls    int
			updatesProposed = []string{"totally invalid format"}

			mockParamsKeeper = &endBlockerParamsMock{
				getStringFn: func(_ sdk.Context, key string, ptr *string) {},
				getStringsFn: func(_ sdk.Context, key string, ptr *[]string) {
					switch key {
					case valsetParamPath(vm.ValsetRealmDefault, valsetPrevKey):
						*ptr = []string{}
					case valsetParamPath(vm.ValsetRealmDefault, valsetNewKey):
						*ptr = updatesProposed
					}
				},
				getBoolFn: func(_ sdk.Context, key string, ptr *bool) {
					if key == valsetParamPath(vm.ValsetRealmDefault, valsetDirtyKey) {
						*ptr = updateFlag
					}
				},
				setBoolFn: func(_ sdk.Context, key string, value bool) {
					updateFlag = value
				},
				setStringsFn: func(_ sdk.Context, key string, value []string) {
					prevSetCalls++
				},
			}

			mockApp = &mockEndBlockerApp{}
		)

		eb := EndBlocker(mockParamsKeeper, nil, nil, mockApp)
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		assert.Equal(t, abci.ResponseEndBlock{}, res)
		assert.False(t, updateFlag, "flag must be cleared so future proposals land")
		assert.Zero(t, prevSetCalls, "prev must NOT be touched when proposed is bad")
	})

	t.Run("valid valset changes", func(t *testing.T) {
		t.Parallel()

		updates := generateValidatorUpdates(t, 10)

		serializeUpdate := func(u abci.ValidatorUpdate) string {
			return fmt.Sprintf("%s:%s:%d", u.Address.String(), u.PubKey, u.Power)
		}

		var (
			updateFlag   = true
			paramUpdates []string

			mockParamsKeeper = &endBlockerParamsMock{
				getStringFn: func(_ sdk.Context, key string, ptr *string) {},
				getStringsFn: func(_ sdk.Context, key string, ptr *[]string) {
					switch key {
					case valsetParamPath(vm.ValsetRealmDefault, valsetPrevKey):
						*ptr = []string{} // empty prev set
					case valsetParamPath(vm.ValsetRealmDefault, valsetNewKey):
						serialized := make([]string, 0, len(updates))
						for _, u := range updates {
							serialized = append(serialized, serializeUpdate(u))
						}
						*ptr = serialized
					}
				},
				getBoolFn: func(_ sdk.Context, key string, ptr *bool) {
					if key == valsetParamPath(vm.ValsetRealmDefault, valsetDirtyKey) {
						*ptr = updateFlag
					}
				},
				setBoolFn: func(_ sdk.Context, key string, value bool) {
					updateFlag = value
				},
				setStringsFn: func(_ sdk.Context, key string, value []string) {
					paramUpdates = value
				},
			}

			mockApp = &mockEndBlockerApp{}
		)

		eb := EndBlocker(mockParamsKeeper, nil, nil, mockApp)

		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		require.Len(t, res.ValidatorUpdates, len(updates))

		// Sort both for comparison
		sort.Slice(updates, func(i, j int) bool {
			return updates[i].Address.Compare(updates[j].Address) < 0
		})
		sort.Slice(res.ValidatorUpdates, func(i, j int) bool {
			return res.ValidatorUpdates[i].Address.Compare(res.ValidatorUpdates[j].Address) < 0
		})

		for i, u := range updates {
			assert.Equal(t, u.Address.String(), res.ValidatorUpdates[i].Address.String())
			assert.True(t, u.PubKey.Equals(res.ValidatorUpdates[i].PubKey))
			assert.Equal(t, u.Power, res.ValidatorUpdates[i].Power)
		}

		// Flag cleared, prev updated
		assert.False(t, updateFlag)
		assert.NotEmpty(t, paramUpdates)
	})

	t.Run("wrong pubkey type filtered out", func(t *testing.T) {
		t.Parallel()

		updates := generateValidatorUpdates(t, 1)

		serializeUpdate := func(u abci.ValidatorUpdate) string {
			return fmt.Sprintf("%s:%s:%d", u.Address.String(), u.PubKey, u.Power)
		}

		var (
			updateFlag = true

			mockParamsKeeper = &endBlockerParamsMock{
				getStringFn: func(_ sdk.Context, key string, ptr *string) {},
				getStringsFn: func(_ sdk.Context, key string, ptr *[]string) {
					switch key {
					case valsetParamPath(vm.ValsetRealmDefault, valsetPrevKey):
						*ptr = []string{}
					case valsetParamPath(vm.ValsetRealmDefault, valsetNewKey):
						*ptr = []string{serializeUpdate(updates[0])}
					}
				},
				getBoolFn: func(_ sdk.Context, key string, ptr *bool) {
					if key == valsetParamPath(vm.ValsetRealmDefault, valsetDirtyKey) {
						*ptr = updateFlag
					}
				},
				setBoolFn:    func(_ sdk.Context, _ string, value bool) { updateFlag = value },
				setStringsFn: func(_ sdk.Context, _ string, _ []string) {},
			}

			mockApp = &mockEndBlockerApp{}
		)

		eb := EndBlocker(mockParamsKeeper, nil, nil, mockApp)

		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeyEd25519"}, // wrong type
			},
		}), abci.RequestEndBlock{})

		// The update is filtered out due to wrong pubkey type
		assert.Empty(t, res.ValidatorUpdates)
	})

	t.Run("diff applied: kept + power-change + new + removed", func(t *testing.T) {
		t.Parallel()

		// Build prev = [v1@10, v2@20, v3@30]
		// proposed = [v1@10 (kept), v2@99 (power change), v4@40 (new)]
		// expected updates: v2@99, v3@0 (removal), v4@40
		prevUpdates := generateValidatorUpdates(t, 3)
		newUpdate := generateValidatorUpdates(t, 1)[0]

		serialize := func(u abci.ValidatorUpdate) string {
			return fmt.Sprintf("%s:%s:%d", u.Address.String(), u.PubKey, u.Power)
		}
		prevUpdates[0].Power = 10
		prevUpdates[1].Power = 20
		prevUpdates[2].Power = 30
		newUpdate.Power = 40

		prevSerialized := []string{
			serialize(prevUpdates[0]),
			serialize(prevUpdates[1]),
			serialize(prevUpdates[2]),
		}
		v2Changed := prevUpdates[1]
		v2Changed.Power = 99
		proposedSerialized := []string{
			serialize(prevUpdates[0]), // unchanged
			serialize(v2Changed),      // power change
			serialize(newUpdate),      // new
			// prevUpdates[2] dropped → removal
		}

		var (
			updateFlag = true
			prevWrites [][]string

			mockParamsKeeper = &endBlockerParamsMock{
				getStringFn: func(_ sdk.Context, key string, ptr *string) {},
				getStringsFn: func(_ sdk.Context, key string, ptr *[]string) {
					switch key {
					case valsetParamPath(vm.ValsetRealmDefault, valsetPrevKey):
						*ptr = prevSerialized
					case valsetParamPath(vm.ValsetRealmDefault, valsetNewKey):
						*ptr = proposedSerialized
					}
				},
				getBoolFn: func(_ sdk.Context, key string, ptr *bool) {
					if key == valsetParamPath(vm.ValsetRealmDefault, valsetDirtyKey) {
						*ptr = updateFlag
					}
				},
				setBoolFn:    func(_ sdk.Context, _ string, value bool) { updateFlag = value },
				setStringsFn: func(_ sdk.Context, _ string, value []string) { prevWrites = append(prevWrites, value) },
			}
			mockApp = &mockEndBlockerApp{}
		)

		eb := EndBlocker(mockParamsKeeper, nil, nil, mockApp)
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		require.Len(t, res.ValidatorUpdates, 3, "expect: 1 power change, 1 removal, 1 new")

		// Build by-address map for assertion (UpdatesFrom output is sorted but
		// we don't want to depend on the sort key here).
		byAddr := map[string]abci.ValidatorUpdate{}
		for _, u := range res.ValidatorUpdates {
			byAddr[u.Address.String()] = u
		}

		assert.Equal(t, int64(99), byAddr[prevUpdates[1].Address.String()].Power, "v2 power must be 99")
		assert.Equal(t, int64(0), byAddr[prevUpdates[2].Address.String()].Power, "v3 must be removed (Power=0)")
		assert.Equal(t, int64(40), byAddr[newUpdate.Address.String()].Power, "v4 must be added")
		_, kept := byAddr[prevUpdates[0].Address.String()]
		assert.False(t, kept, "v1 (unchanged) must NOT appear in updates")

		assert.False(t, updateFlag)
		require.Len(t, prevWrites, 1)
		assert.Equal(t, proposedSerialized, prevWrites[0], "prev advances to proposed")
	})

	t.Run("wipe valset: prev=[v1,v2] proposed=[] -> 2 removals", func(t *testing.T) {
		t.Parallel()

		prev := generateValidatorUpdates(t, 2)
		serialize := func(u abci.ValidatorUpdate) string {
			return fmt.Sprintf("%s:%s:%d", u.Address.String(), u.PubKey, u.Power)
		}
		prevSerialized := []string{serialize(prev[0]), serialize(prev[1])}

		updateFlag := true
		mockParamsKeeper := &endBlockerParamsMock{
			getStringFn: func(_ sdk.Context, key string, ptr *string) {},
			getStringsFn: func(_ sdk.Context, key string, ptr *[]string) {
				switch key {
				case valsetParamPath(vm.ValsetRealmDefault, valsetPrevKey):
					*ptr = prevSerialized
				case valsetParamPath(vm.ValsetRealmDefault, valsetNewKey):
					*ptr = []string{}
				}
			},
			getBoolFn: func(_ sdk.Context, key string, ptr *bool) {
				if key == valsetParamPath(vm.ValsetRealmDefault, valsetDirtyKey) {
					*ptr = updateFlag
				}
			},
			setBoolFn:    func(_ sdk.Context, _ string, value bool) { updateFlag = value },
			setStringsFn: func(_ sdk.Context, _ string, _ []string) {},
		}

		eb := EndBlocker(mockParamsKeeper, nil, nil, &mockEndBlockerApp{})
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		require.Len(t, res.ValidatorUpdates, 2, "wiping the set must surface 2 removals")
		for _, u := range res.ValidatorUpdates {
			assert.Equal(t, int64(0), u.Power)
		}
	})

	t.Run("custom valset_realm_path override is honored", func(t *testing.T) {
		t.Parallel()

		const customRealm = "gno.land/r/test/custom_valset"

		updates := generateValidatorUpdates(t, 1)
		serialize := func(u abci.ValidatorUpdate) string {
			return fmt.Sprintf("%s:%s:%d", u.Address.String(), u.PubKey, u.Power)
		}

		var (
			seenRealmInPaths []string
			updateFlag       = true
		)
		mockParamsKeeper := &endBlockerParamsMock{
			getStringFn: func(_ sdk.Context, key string, ptr *string) {
				if key == vm.ValsetRealmParamPath {
					*ptr = customRealm
				}
			},
			getStringsFn: func(_ sdk.Context, key string, ptr *[]string) {
				seenRealmInPaths = append(seenRealmInPaths, key)
				switch key {
				case valsetParamPath(customRealm, valsetPrevKey):
					*ptr = []string{}
				case valsetParamPath(customRealm, valsetNewKey):
					*ptr = []string{serialize(updates[0])}
				}
			},
			getBoolFn: func(_ sdk.Context, key string, ptr *bool) {
				if key == valsetParamPath(customRealm, valsetDirtyKey) {
					*ptr = updateFlag
				}
			},
			setBoolFn:    func(_ sdk.Context, _ string, value bool) { updateFlag = value },
			setStringsFn: func(_ sdk.Context, _ string, _ []string) {},
		}

		eb := EndBlocker(mockParamsKeeper, nil, nil, &mockEndBlockerApp{})
		res := eb(sdk.Context{}.WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypeURLs: []string{"/tm.PubKeySecp256k1"},
			},
		}), abci.RequestEndBlock{})

		require.Len(t, res.ValidatorUpdates, 1)
		// Verify EndBlocker actually queried the custom realm's keyspace.
		require.NotEmpty(t, seenRealmInPaths)
		for _, p := range seenRealmInPaths {
			assert.Contains(t, p, customRealm, "EndBlocker must read from custom realm's keyspace")
		}
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

	t.Run("halts at exact height", func(t *testing.T) {
		t.Parallel()

		var haltSet uint64
		mockApp := &mockEndBlockerApp{
			setHaltHeightFn: func(h uint64) { haltSet = h },
		}
		mockPrmk := &mockConfigurableParamsKeeper{
			int64s: map[string]int64{nodeParamHaltHeight: 100},
		}

		eb := EndBlocker(mockPrmk, nil, nil, mockApp)
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

		eb := EndBlocker(mockPrmk, nil, nil, mockApp)
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

		eb := EndBlocker(mockPrmk, nil, nil, mockApp)
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

		eb := EndBlocker(mockPrmk, nil, nil, mockApp)
		eb(sdk.Context{}, abci.RequestEndBlock{Height: 100})

		assert.Equal(t, uint64(0), haltSet, "SetHaltHeight should NOT be called when halt_height=0 (cancelled)")
	})
}
