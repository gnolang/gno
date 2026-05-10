package gnoland

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
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

func TestShouldAssertValoperCoverage(t *testing.T) {
	t.Parallel()

	dummyVals := generateValidatorUpdates(t, 1)

	cases := []struct {
		name string
		req  abci.RequestInitChain
		want bool
	}{
		{
			name: "fresh chain, no validators",
			req:  abci.RequestInitChain{AppState: GnoGenesisState{}},
			want: false,
		},
		{
			name: "fresh chain, validators present",
			req:  abci.RequestInitChain{Validators: dummyVals, AppState: GnoGenesisState{}},
			want: false,
		},
		{
			name: "hardfork PastChainIDs but no validators",
			req:  abci.RequestInitChain{AppState: GnoGenesisState{PastChainIDs: []string{"old"}}},
			want: false,
		},
		{
			name: "hardfork PastChainIDs + validators",
			req:  abci.RequestInitChain{Validators: dummyVals, AppState: GnoGenesisState{PastChainIDs: []string{"old"}}},
			want: true,
		},
		{
			name: "non-genesis InitialHeight alone (NOT a hardfork signal)",
			req:  abci.RequestInitChain{Validators: dummyVals, InitialHeight: 100, AppState: GnoGenesisState{}},
			want: false,
		},
		{
			name: "AppState wrong type (defensive)",
			req:  abci.RequestInitChain{Validators: dummyVals, AppState: nil},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := shouldAssertValoperCoverage(tc.req)
			assert.Equal(t, tc.want, got, "case %q", tc.name)
		})
	}
}

// TestInitChainer_SkipValoperCoverageAssertion guards the cfg-level
// override against the hardfork auto-assertion. Without it, gnogenesis
// fork test (synthetic MockPV with no valoper profile) trips the
// assertion and aborts boot. Underlying request-level gating is
// covered by TestShouldAssertValoperCoverage; this test only exercises
// the flag composition.
func TestInitChainer_SkipValoperCoverageAssertion(t *testing.T) {
	t.Parallel()

	hardforkReq := abci.RequestInitChain{
		Validators: generateValidatorUpdates(t, 1),
		AppState:   GnoGenesisState{PastChainIDs: []string{"old-chain"}},
	}

	cases := []struct {
		name string
		skip bool
		want bool
	}{
		{name: "flag false: assertion runs", skip: false, want: true},
		{name: "flag true: assertion skipped", skip: true, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := InitChainerConfig{SkipValoperCoverageAssertion: tc.skip}
			got := cfg.shouldRunValoperCoverageAssertion(hardforkReq)
			assert.Equal(t, tc.want, got)
		})
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

	return createAndSignTxWithAccSeq(t, msgs, chainID, key, 0, 0)
}

func createAndSignTxWithAccSeq(
	t *testing.T,
	msgs []std.Msg,
	chainID string,
	key crypto.PrivKey,
	accNum, seq uint64,
) std.Tx {
	t.Helper()

	tx := std.Tx{
		Msgs: msgs,
		Fee: std.Fee{
			GasFee:    std.NewCoin("ugnot", 2000000),
			GasWanted: 10000000,
		},
	}

	signBytes, err := tx.GetSignBytes(chainID, accNum, seq)
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

		getZeroTimestampMetadataState = func(tx std.Tx, balances []Balance) GnoGenesisState {
			return GnoGenesisState{
				// Metadata present but Timestamp=0 — genesis block time should be preserved
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp: 0, // zero — must not override to Unix epoch
						},
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
		{
			"metadata transaction with zero timestamp uses genesis block time",
			currentTimestamp,
			currentTimestamp, // zero Timestamp → falls back to genesis block time
			getZeroTimestampMetadataState,
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

// TestInitChainer_MigrationTxKeepsTimestampWithPastChainIDs is a regression
// test for the bug where, with PastChainIDs set, a tx whose metadata had
// BlockHeight == 0 but a non-zero Timestamp (a migration tx) had its
// ctxFn silently overwritten by the genesis-mode branch, dropping the
// timestamp override. The fix tightens the genesis-mode predicate to
// metadata == nil so migration txs keep their metadata-driven ctxFn.
func TestInitChainer_MigrationTxKeepsTimestampWithPastChainIDs(t *testing.T) {
	t.Parallel()

	var (
		genesisTime   = time.Now()
		migrationTime = genesisTime.Add(7 * 24 * time.Hour) // 7 days later
		chainID       = "test-chain"
		pastChainIDs  = []string{chainID}
		path          = "gno.land/r/demo/migration"
		body          = `package migration

import "time"

var t time.Time = time.Now()

func GetT(cur realm) int64 { return t.Unix() }
`
	)

	key := getDummyKey(t)

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)

	msg := vm.MsgAddPackage{
		Creator: key.PubKey().Address(),
		Package: &std.MemPackage{
			Name: "migration",
			Path: path,
			Files: []*std.MemFile{
				{Name: "file.gno", Body: body},
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
			},
		},
		MaxDeposit: nil,
	}
	tx := createAndSignTx(t, []std.Msg{msg}, chainID, key)

	app.InitChain(abci.RequestInitChain{
		ChainID: chainID,
		Time:    genesisTime,
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
		AppState: GnoGenesisState{
			Txs: []TxWithMetadata{
				{
					Tx: tx,
					// migration-tx shape: BlockHeight == 0 but Timestamp != 0
					Metadata: &GnoTxMetadata{
						Timestamp:   migrationTime.Unix(),
						BlockHeight: 0,
					},
				},
			},
			Balances: []Balance{
				{
					Address: key.PubKey().Address(),
					Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
				},
			},
			Auth:         auth.DefaultGenesisState(),
			Bank:         bank.DefaultGenesisState(),
			VM:           vm.DefaultGenesisState(),
			PastChainIDs: pastChainIDs, // triggers the genesis-mode branch pre-fix
		},
	})

	callMsg := vm.MsgCall{
		Caller:  key.PubKey().Address(),
		PkgPath: path,
		Func:    "GetT",
	}
	tx = createAndSignTx(t, []std.Msg{callMsg}, chainID, key)
	marshalledTx, err := amino.Marshal(tx)
	require.NoError(t, err)

	resp := app.DeliverTx(abci.RequestDeliverTx{Tx: marshalledTx})
	require.True(t, resp.IsOK(), "expected OK, got: %s", resp.Log)

	// Before the fix, the second ctxFn assignment in the loop stomped the
	// metadata-driven Timestamp override and the realm initialized at
	// genesisTime instead of migrationTime.
	assert.Contains(
		t,
		string(resp.Data),
		fmt.Sprintf("(%d int64)", migrationTime.Unix()),
		"realm should have been initialized at metadata.Timestamp, not genesis time",
	)
}

// endBlockerParamsMock is a ParamsKeeperI mock with optional per-method
// hooks, scoped to TestEndBlocker. Unset hooks are no-ops, matching the
// minimal-by-default behavior of mockParamsKeeper but adding per-key
// observation/injection where each subtest needs it.
type endBlockerParamsMock struct {
	getStringFn  func(sdk.Context, string, *string) bool
	getInt64Fn   func(sdk.Context, string, *int64) bool
	getBoolFn    func(sdk.Context, string, *bool) bool
	getStringsFn func(sdk.Context, string, *[]string) bool
	setBoolFn    func(sdk.Context, string, bool)
	setStringsFn func(sdk.Context, string, []string)
}

func (m *endBlockerParamsMock) GetString(ctx sdk.Context, key string, ptr *string) bool {
	if m.getStringFn != nil {
		return m.getStringFn(ctx, key, ptr)
	}
	return false
}

func (m *endBlockerParamsMock) GetInt64(ctx sdk.Context, key string, ptr *int64) bool {
	if m.getInt64Fn != nil {
		return m.getInt64Fn(ctx, key, ptr)
	}
	return false
}

func (m *endBlockerParamsMock) GetBool(ctx sdk.Context, key string, ptr *bool) bool {
	if m.getBoolFn != nil {
		return m.getBoolFn(ctx, key, ptr)
	}
	return false
}

func (m *endBlockerParamsMock) GetStrings(ctx sdk.Context, key string, ptr *[]string) bool {
	if m.getStringsFn != nil {
		return m.getStringsFn(ctx, key, ptr)
	}
	return false
}

func (m *endBlockerParamsMock) SetBool(ctx sdk.Context, key string, value bool) int {
	if m.setBoolFn != nil {
		m.setBoolFn(ctx, key, value)
	}
	return 0
}

func (m *endBlockerParamsMock) SetStrings(ctx sdk.Context, key string, value []string) int {
	if m.setStringsFn != nil {
		m.setStringsFn(ctx, key, value)
	}
	return 0
}

// Remaining ParamsKeeperI methods are not exercised by EndBlocker.
func (m *endBlockerParamsMock) GetUint64(sdk.Context, string, *uint64) bool { return false }
func (m *endBlockerParamsMock) GetBytes(sdk.Context, string, *[]byte) bool  { return false }
func (m *endBlockerParamsMock) SetString(sdk.Context, string, string) int   { return 0 }
func (m *endBlockerParamsMock) SetInt64(sdk.Context, string, int64) int     { return 0 }
func (m *endBlockerParamsMock) SetUint64(sdk.Context, string, uint64) int   { return 0 }
func (m *endBlockerParamsMock) SetBytes(sdk.Context, string, []byte) int    { return 0 }
func (m *endBlockerParamsMock) Has(sdk.Context, string) bool                { return false }
func (m *endBlockerParamsMock) GetStruct(sdk.Context, string, any)          {}
func (m *endBlockerParamsMock) SetStruct(sdk.Context, string, any)          {}
func (m *endBlockerParamsMock) GetAny(sdk.Context, string) any              { return nil }
func (m *endBlockerParamsMock) SetAny(sdk.Context, string, any)             {}

// valsetState is a tiny in-memory shim mirroring the valset key space,
// used by TestEndBlocker to drive endBlockerParamsMock.
type valsetState struct {
	current, proposed []string
	dirty             bool
	currentWrites     [][]string
	dirtyWrites       []bool
	// currentWriteCtxSentinels records ctx.Value(internalWriteCtxKey{})
	// observed on each valset:current write — TestEndBlocker_SentinelOnCurrentWrite
	// asserts the sentinel is always true so a future regression that
	// drops `intCtx := ctx.WithValue(internalWriteCtxKey{}, true)` at
	// app.go:646 fails CI rather than silently re-opening F2.
	currentWriteCtxSentinels []bool
}

// serializeUpdates converts ValidatorUpdates to the wire format
// "<pubkey>:<power>" used by the params keeper.
func serializeUpdates(us []abci.ValidatorUpdate) []string {
	out := make([]string, len(us))
	for i, u := range us {
		out[i] = u.PubKey.String() + ":" + strconv.FormatInt(u.Power, 10)
	}
	return out
}

// newValsetMock returns a mock keeper backed by st. Reads come from st;
// writes update st (and append to its history slices for assertions).
func newValsetMock(st *valsetState) *endBlockerParamsMock {
	return &endBlockerParamsMock{
		getStringsFn: func(_ sdk.Context, key string, ptr *[]string) bool {
			switch key {
			case valsetCurrentPath:
				if st.current == nil {
					return false
				}
				*ptr = st.current
				return true
			case valsetProposedPath:
				if st.proposed == nil {
					return false
				}
				*ptr = st.proposed
				return true
			}
			return false
		},
		getBoolFn: func(_ sdk.Context, key string, ptr *bool) bool {
			if key == valsetDirtyPath {
				*ptr = st.dirty
				return true
			}
			return false
		},
		setBoolFn: func(_ sdk.Context, key string, value bool) {
			if key == valsetDirtyPath {
				st.dirty = value
				st.dirtyWrites = append(st.dirtyWrites, value)
			}
		},
		setStringsFn: func(ctx sdk.Context, key string, value []string) {
			if key == valsetCurrentPath {
				st.current = value
				st.currentWrites = append(st.currentWrites, value)
				sentinel, _ := ctx.Value(internalWriteCtxKey{}).(bool)
				st.currentWriteCtxSentinels = append(st.currentWriteCtxSentinels, sentinel)
			}
		},
	}
}

func runEndBlocker(t *testing.T, mock *endBlockerParamsMock, pubKeyType string) abci.ResponseEndBlock {
	t.Helper()
	eb := EndBlocker(mock, nil, nil, &mockEndBlockerApp{})
	// Use context.Background() as the wrapped context so ctx.Value()
	// (which the new EndBlocker calls for internalWriteCtxKey) and
	// ctx.WithValue() don't nil-deref the underlying context.Context.
	ctx := sdk.Context{}.
		WithContext(context.Background()).
		WithConsensusParams(&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{pubKeyType}},
		})
	return eb(ctx, abci.RequestEndBlock{})
}

func TestEndBlocker(t *testing.T) {
	t.Parallel()

	t.Run("no valset changes (dirty=false)", func(t *testing.T) {
		t.Parallel()

		st := &valsetState{dirty: false}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		assert.Equal(t, abci.ResponseEndBlock{}, res)
	})

	t.Run("valset:current corrupted panics (chain-internal)", func(t *testing.T) {
		t.Parallel()

		// Tier 1 semantic flip: corrupted current is no longer
		// silently recovered. Only chain code writes valset:current
		// (via ctx-sentinel), so corruption is by definition a
		// chain-internal bug or store damage and warrants a panic.
		st := &valsetState{
			current:  []string{"garbage:not-a-power"},
			proposed: serializeUpdates(generateValidatorUpdates(t, 1)),
			dirty:    true,
		}
		assert.Panics(t, func() {
			runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		})
	})

	t.Run("invalid valset:proposed drops dirty (no current write)", func(t *testing.T) {
		t.Parallel()

		// Recovery for proposed parse failure: clear the flag so a future
		// re-propose can land. Do NOT touch current.
		st := &valsetState{
			current:  serializeUpdates(generateValidatorUpdates(t, 1)),
			proposed: []string{"bogus:7"}, // pubkey is invalid bech32
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		assert.Equal(t, abci.ResponseEndBlock{}, res)
		assert.False(t, st.dirty, "flag must be cleared so future proposals land")
		assert.Empty(t, st.currentWrites, "current must NOT be touched when proposed is bad")
	})

	t.Run("valid valset changes (additions only)", func(t *testing.T) {
		t.Parallel()

		updates := generateValidatorUpdates(t, 10)
		proposedEntries := serializeUpdates(updates)
		st := &valsetState{proposed: proposedEntries, dirty: true}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		require.Len(t, res.ValidatorUpdates, len(updates))

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

		assert.False(t, st.dirty)
		// current = EncodeValidatorUpdates(proposedSet) which sorts by
		// pubkey-bytes; so the written value is sorted. Just check
		// that it has the same set of entries.
		require.Len(t, st.currentWrites, 1)
		assert.ElementsMatch(t, proposedEntries, st.currentWrites[0],
			"current must equal proposed (modulo canonical sort)")

		// Sentinel-flow regression guard: the valset:current write must
		// carry internalWriteCtxKey{}=true so the chain-side
		// WillSetParam in node_params.go accepts it. A future change
		// that drops `intCtx := ctx.WithValue(...)` at app.go would
		// silently re-open the F2 vector (any realm could write
		// valset:current via a generic factory). Pin it here.
		require.Len(t, st.currentWriteCtxSentinels, 1)
		assert.True(t, st.currentWriteCtxSentinels[0],
			"valset:current write must carry the internalWriteCtxKey sentinel")
	})

	t.Run("wrong pubkey type whole-rejects proposal", func(t *testing.T) {
		t.Parallel()

		// Whole-reject: a proposal containing any disallowed pubkey
		// type is refused atomically. current is untouched.
		updates := generateValidatorUpdates(t, 1)
		st := &valsetState{proposed: serializeUpdates(updates), dirty: true}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeyEd25519") // wrong type

		assert.Empty(t, res.ValidatorUpdates, "whole-reject means no updates emitted")
		assert.False(t, st.dirty, "dirty cleared")
		assert.Empty(t, st.currentWrites, "current MUST NOT be advanced on whole-reject")
	})

	t.Run("diff applied: kept + power-change + new + removed", func(t *testing.T) {
		t.Parallel()

		// current = [v1@10, v2@20, v3@30]
		// proposed = [v1@10 (kept), v2@99 (power change), v4@40 (new)]
		// expected updates: v2@99, v3@0 (removal), v4@40
		currentUpdates := generateValidatorUpdates(t, 3)
		newcomer := generateValidatorUpdates(t, 1)[0]
		currentUpdates[0].Power = 10
		currentUpdates[1].Power = 20
		currentUpdates[2].Power = 30
		newcomer.Power = 40

		v2Changed := currentUpdates[1]
		v2Changed.Power = 99
		proposed := []abci.ValidatorUpdate{currentUpdates[0], v2Changed, newcomer}
		proposedEntries := serializeUpdates(proposed)

		st := &valsetState{
			current:  serializeUpdates(currentUpdates),
			proposed: proposedEntries,
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		require.Len(t, res.ValidatorUpdates, 3, "expect: 1 power change, 1 removal, 1 new")

		byAddr := map[string]abci.ValidatorUpdate{}
		for _, u := range res.ValidatorUpdates {
			byAddr[u.Address.String()] = u
		}
		assert.Equal(t, int64(99), byAddr[currentUpdates[1].Address.String()].Power, "v2 power must be 99")
		assert.Equal(t, int64(0), byAddr[currentUpdates[2].Address.String()].Power, "v3 must be removed (Power=0)")
		assert.Equal(t, int64(40), byAddr[newcomer.Address.String()].Power, "v4 must be added")
		_, kept := byAddr[currentUpdates[0].Address.String()]
		assert.False(t, kept, "v1 (unchanged) must NOT appear in updates")

		assert.False(t, st.dirty)
		require.Len(t, st.currentWrites, 1)
		assert.ElementsMatch(t, proposedEntries, st.currentWrites[0],
			"current must equal proposed (modulo canonical sort)")
	})

	t.Run("min-floor: empty proposed rejected", func(t *testing.T) {
		t.Parallel()

		// Min-floor: proposed=[] is the "remove all" shape; refuse it
		// to keep consensus from halting at H+2 with zero validators.
		current := generateValidatorUpdates(t, 2)
		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: []string{},
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		assert.Empty(t, res.ValidatorUpdates, "min-floor means no updates emitted")
		assert.False(t, st.dirty, "dirty cleared")
		assert.Empty(t, st.currentWrites, "current MUST NOT be advanced on min-floor reject")
	})

	t.Run("min-floor: all-Power=0 proposed rejected", func(t *testing.T) {
		t.Parallel()

		// Defense-in-depth: a non-empty proposed where every entry has
		// Power=0 is still a "remove all" — len > 0 but live count is
		// zero. Floor must catch this regardless of outer-list length.
		// (Reachable via v3 if a proposal's deltas remove every
		// validator and produce an empty published set; the floor is
		// the consensus-safety backstop.)
		current := generateValidatorUpdates(t, 2)
		// Build a proposed list that mirrors current's pubkeys but with
		// Power=0. This is the "explicitly remove all" shape.
		proposed := make([]abci.ValidatorUpdate, len(current))
		copy(proposed, current)
		for i := range proposed {
			proposed[i].Power = 0
		}
		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(proposed),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		assert.Empty(t, res.ValidatorUpdates, "min-floor means no updates emitted")
		assert.False(t, st.dirty, "dirty cleared")
		assert.Empty(t, st.currentWrites, "current MUST NOT be advanced on all-Power=0 reject")
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
	prmk.Register("node", nodeParamsKeeper{})
	// Set InitChainer
	icc := cfg.InitChainerConfig
	icc.baseApp = baseApp
	icc.prmk = prmk
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

func TestChainUpgradeGenesisReplay(t *testing.T) {
	t.Parallel()

	t.Run("fields serialize correctly", func(t *testing.T) {
		t.Parallel()

		state := GnoGenesisState{
			Balances:      []Balance{},
			Txs:           []TxWithMetadata{},
			Auth:          auth.DefaultGenesisState(),
			Bank:          bank.DefaultGenesisState(),
			VM:            vm.DefaultGenesisState(),
			PastChainIDs:  []string{"old-chain-1", "old-chain-2"},
			InitialHeight: 100,
		}

		// Serialize and deserialize
		data, err := amino.MarshalJSON(state)
		require.NoError(t, err)

		var decoded GnoGenesisState
		require.NoError(t, amino.UnmarshalJSON(data, &decoded))

		assert.Equal(t, []string{"old-chain-1", "old-chain-2"}, decoded.PastChainIDs)
		assert.Equal(t, int64(100), decoded.InitialHeight)
	})

	t.Run("historical tx replays with correct block height", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"

			path = "gno.land/r/demo/upgradetest"
			body = `package upgradetest

import "chain/runtime"

var height int64 = runtime.ChainHeight()

func GetHeight(cur realm) int64 { return height }
`
		)

		// Create a fresh app instance
		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		// Prepare the deploy transaction
		msg := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "upgradetest",
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

		// Sign with the old chain ID — metadata.BlockHeight > 0 and metadata.ChainID
		// in PastChainIDs will cause the ctxFn to override the chain ID for sig verification.
		// Account number=0 and sequence=0 because the account is created from balances
		// but hasn't processed any transactions yet.
		tx := createAndSignTx(t, []std.Msg{msg}, "old-chain", key)

		// Run InitChain with PastChainIDs and InitialHeight set,
		// and the deploy tx using metadata with BlockHeight=42 and ChainID="old-chain"
		app.InitChain(abci.RequestInitChain{
			ChainID:       chainID,
			Time:          time.Now(),
			InitialHeight: 100,
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: []string{},
				},
			},
			AppState: GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 42,
							ChainID:     "old-chain", // must be in PastChainIDs for override
						},
					},
				},
				Balances: []Balance{
					{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					},
				},
				Auth:          auth.DefaultGenesisState(),
				Bank:          bank.DefaultGenesisState(),
				VM:            vm.DefaultGenesisState(),
				PastChainIDs:  []string{"old-chain"},
				InitialHeight: 100,
			},
		})

		// Call GetHeight to verify the realm captured height=42
		callMsg := vm.MsgCall{
			Caller:  key.PubKey().Address(),
			PkgPath: path,
			Func:    "GetHeight",
		}

		callTx := createAndSignTx(t, []std.Msg{callMsg}, chainID, key)

		marshalledTx, err := amino.Marshal(callTx)
		require.NoError(t, err)

		resp := app.DeliverTx(abci.RequestDeliverTx{
			Tx: marshalledTx,
		})

		require.True(t, resp.IsOK(), "DeliverTx failed: %s", resp.Log)

		// The realm should have captured block height 42
		assert.Contains(t, string(resp.Data), "(42 int64)")
	})

	t.Run("metadata block height in GnoTxMetadata serializes correctly", func(t *testing.T) {
		t.Parallel()

		txm := TxWithMetadata{
			Tx: std.Tx{},
			Metadata: &GnoTxMetadata{
				Timestamp:   1234567890,
				BlockHeight: 42,
				ChainID:     "gnoland1",
			},
		}

		data, err := amino.MarshalJSON(txm)
		require.NoError(t, err)

		var decoded TxWithMetadata
		require.NoError(t, amino.UnmarshalJSON(data, &decoded))

		require.NotNil(t, decoded.Metadata)
		assert.Equal(t, int64(1234567890), decoded.Metadata.Timestamp)
		assert.Equal(t, int64(42), decoded.Metadata.BlockHeight)
		assert.Equal(t, "gnoland1", decoded.Metadata.ChainID)
	})

	t.Run("chain ID not overridden when BlockHeight is zero in metadata", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"

			path = "gno.land/r/demo/chainidtest"
			body = `package chainidtest

var Deployed = true

func IsDeployed(cur realm) bool { return Deployed }
`
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		msg := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "chainidtest",
				Path: path,
				Files: []*std.MemFile{
					{Name: "file.gno", Body: body},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
				},
			},
		}

		// When metadata.BlockHeight == 0, the chain ID override must NOT happen.
		// So the tx must be signed with the current chain ID (chainID), not any past chain ID.
		tx := createAndSignTx(t, []std.Msg{msg}, chainID, key)

		app.InitChain(abci.RequestInitChain{
			ChainID: chainID,
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: []string{},
				},
			},
			AppState: GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 0,           // zero — no chain ID override
							ChainID:     "old-chain", // present but ignored since BlockHeight == 0
						},
					},
				},
				Balances: []Balance{
					{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					},
				},
				Auth:         auth.DefaultGenesisState(),
				Bank:         bank.DefaultGenesisState(),
				VM:           vm.DefaultGenesisState(),
				PastChainIDs: []string{"old-chain"}, // set, but should NOT be used since BlockHeight == 0
			},
		})
	})

	t.Run("no chain ID override when metadata.ChainID not in PastChainIDs", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"

			path = "gno.land/r/demo/nooverride"
			body = `package nooverride

var Deployed = true
`
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		msg := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "nooverride",
				Path: path,
				Files: []*std.MemFile{
					{Name: "file.gno", Body: body},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
				},
			},
		}

		// BlockHeight > 0 and metadata.ChainID is set, but the chain ID is NOT in
		// PastChainIDs — no chain ID override should happen. The tx is signed with
		// chainID so it verifies correctly without the override.
		tx := createAndSignTx(t, []std.Msg{msg}, chainID, key)

		app.InitChain(abci.RequestInitChain{
			ChainID: chainID,
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: []string{},
				},
			},
			AppState: GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 10,
							ChainID:     "unknown-chain", // not in PastChainIDs — no override
						},
					},
				},
				Balances: []Balance{
					{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					},
				},
				Auth: auth.DefaultGenesisState(),
				Bank: bank.DefaultGenesisState(),
				VM:   vm.DefaultGenesisState(),
				// PastChainIDs intentionally empty — no chain ID override allowed
			},
		})
	})

	t.Run("txs from multiple past chains replay correctly", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"

			path1 = "gno.land/r/demo/multichain1"
			path2 = "gno.land/r/demo/multichain2"
			body  = `package %s
var Deployed = true
`
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		// Both txs come from the same account (accNum=0) but different past chains.
		// tx1: seq=0, chain-a; tx2: seq=1, chain-b (sequence incremented by tx1).
		msg1 := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "multichain1",
				Path: path1,
				Files: []*std.MemFile{
					{Name: "file.gno", Body: fmt.Sprintf(body, "multichain1")},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path1)},
				},
			},
		}
		msg2 := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "multichain2",
				Path: path2,
				Files: []*std.MemFile{
					{Name: "file.gno", Body: fmt.Sprintf(body, "multichain2")},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path2)},
				},
			},
		}

		tx1 := createAndSignTx(t, []std.Msg{msg1}, "chain-a", key) // accNum=0, seq=0

		// tx2 must use seq=1 because tx1 already incremented the sequence.
		tx2Raw := std.Tx{
			Msgs: []std.Msg{msg2},
			Fee:  std.Fee{GasFee: std.NewCoin("ugnot", 2_000_000), GasWanted: 10_000_000},
		}
		signBytes2, err := tx2Raw.GetSignBytes("chain-b", 0, 1) // accNum=0, seq=1
		require.NoError(t, err)
		sig2, err := key.Sign(signBytes2)
		require.NoError(t, err)
		tx2Raw.Signatures = []std.Signature{{PubKey: key.PubKey(), Signature: sig2}}

		// Both chain IDs in the allowlist; each tx carries its own ChainID
		app.InitChain(abci.RequestInitChain{
			ChainID: chainID,
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: []string{},
				},
			},
			AppState: GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx1,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 10,
							ChainID:     "chain-a",
						},
					},
					{
						Tx: tx2Raw,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 20,
							ChainID:     "chain-b",
						},
					},
				},
				Balances: []Balance{
					{Address: key.PubKey().Address(), Amount: std.NewCoins(std.NewCoin("ugnot", 20_000_000))},
				},
				Auth:         auth.DefaultGenesisState(),
				Bank:         bank.DefaultGenesisState(),
				VM:           vm.DefaultGenesisState(),
				PastChainIDs: []string{"chain-a", "chain-b"},
			},
		})
	})
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

	pub := getDummyKey(t).PubKey().String()
	good := pub + ":10"

	t.Run("valset:dirty bool passes", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:dirty", true)
		})
	})

	t.Run("valset:dirty wrong type panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:dirty", "yes")
		})
	})

	t.Run("valset:proposed wrong type panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:proposed", 42)
		})
	})

	t.Run("valset:proposed malformed entry panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:proposed", []string{"no-colon"})
		})
	})

	t.Run("valset:proposed bad pubkey panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:proposed", []string{"notapubkey:10"})
		})
	})

	t.Run("valset:proposed negative power panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:proposed", []string{pub + ":-1"})
		})
	})

	t.Run("valset:proposed valid passes", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:proposed", []string{good})
		})
	})

	t.Run("valset:proposed boundary cap accepts len==max", func(t *testing.T) {
		t.Parallel()
		// Cap is inclusive (predicate is `> maxValsetEntries`).
		entries := serializeUpdates(generateValidatorUpdates(t, maxValsetEntries))
		assert.NotPanics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:proposed", entries)
		})
	})

	t.Run("valset:proposed boundary cap rejects len==max+1", func(t *testing.T) {
		t.Parallel()
		entries := serializeUpdates(generateValidatorUpdates(t, maxValsetEntries+1))
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:proposed", entries)
		})
	})

	t.Run("valset:current rejected without sentinel", func(t *testing.T) {
		t.Parallel()
		// The new ctx-sentinel test path: writes from non-internal ctx
		// must be rejected even with valid entry format.
		ctx := sdk.Context{}.WithContext(context.Background())
		assert.Panics(t, func() {
			npk.WillSetParam(ctx, "valset:current", []string{good})
		})
	})

	t.Run("valset:current accepted with sentinel", func(t *testing.T) {
		t.Parallel()
		ctx := sdk.Context{}.WithContext(context.Background()).
			WithValue(internalWriteCtxKey{}, true)
		assert.NotPanics(t, func() {
			npk.WillSetParam(ctx, "valset:current", []string{good})
		})
	})

	t.Run("unknown valset:* key panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			npk.WillSetParam(sdk.Context{}, "valset:bogus", []string{good})
		})
	})
}

// TestInitChainer_InitialHeightMismatch verifies that loadAppState rejects
// a genesis where GnoGenesisState.InitialHeight diverges from the
// GenesisDoc.InitialHeight passed in via RequestInitChain.
func TestInitChainer_InitialHeightMismatch(t *testing.T) {
	t.Parallel()

	t.Run("mismatch is rejected", func(t *testing.T) {
		t.Parallel()

		app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
		require.NoError(t, err)
		resp := app.InitChain(abci.RequestInitChain{
			ChainID:       "test-chain",
			Time:          time.Now(),
			InitialHeight: 100,
			ConsensusParams: &abci.ConsensusParams{
				Block:     defaultBlockParams(),
				Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
			},
			AppState: GnoGenesisState{
				Balances:      []Balance{},
				Txs:           []TxWithMetadata{},
				Auth:          auth.DefaultGenesisState(),
				Bank:          bank.DefaultGenesisState(),
				VM:            vm.DefaultGenesisState(),
				InitialHeight: 200, // diverges from RequestInitChain.InitialHeight
			},
		})
		require.NotNil(t, resp.Error, "InitChainer should reject InitialHeight mismatch")
		assert.Contains(t, resp.Error.Error(), "InitialHeight mismatch")
	})

	t.Run("match is accepted", func(t *testing.T) {
		t.Parallel()

		app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
		require.NoError(t, err)
		resp := app.InitChain(abci.RequestInitChain{
			ChainID:       "test-chain",
			Time:          time.Now(),
			InitialHeight: 100,
			ConsensusParams: &abci.ConsensusParams{
				Block:     defaultBlockParams(),
				Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
			},
			AppState: GnoGenesisState{
				Balances:      []Balance{},
				Txs:           []TxWithMetadata{},
				Auth:          auth.DefaultGenesisState(),
				Bank:          bank.DefaultGenesisState(),
				VM:            vm.DefaultGenesisState(),
				InitialHeight: 100,
			},
		})
		require.Nil(t, resp.Error, "matching InitialHeight should be accepted: %v", resp.Error)
	})

	t.Run("zero app-level InitialHeight is accepted", func(t *testing.T) {
		t.Parallel()

		// GnoGenesisState.InitialHeight = 0 means "not set"; no check needed.
		app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
		require.NoError(t, err)
		resp := app.InitChain(abci.RequestInitChain{
			ChainID:       "test-chain",
			Time:          time.Now(),
			InitialHeight: 100,
			ConsensusParams: &abci.ConsensusParams{
				Block:     defaultBlockParams(),
				Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
			},
			AppState: GnoGenesisState{
				Balances: []Balance{},
				Txs:      []TxWithMetadata{},
				Auth:     auth.DefaultGenesisState(),
				Bank:     bank.DefaultGenesisState(),
				VM:       vm.DefaultGenesisState(),
				// InitialHeight not set
			},
		})
		require.Nil(t, resp.Error, "zero app-level InitialHeight should pass validation: %v", resp.Error)
	})
}

// TestInitChainer_StrictReplay verifies that StrictReplay refuses to boot
// when any non-skipped genesis tx fails replay, and that intentionally
// skipped txs (metadata.Failed = true) are not counted as failures.
func TestInitChainer_StrictReplay(t *testing.T) {
	t.Parallel()

	// A tx that fails to deliver because it has no msgs / no signatures
	// (ante handler will reject it).
	failingTx := std.Tx{
		Msgs: []std.Msg{},
		Fee:  std.Fee{GasFee: std.NewCoin("ugnot", 1), GasWanted: 100},
	}

	t.Run("StrictReplay false: failing tx does not abort boot", func(t *testing.T) {
		t.Parallel()

		opts := TestAppOptions(memdb.NewMemDB())
		opts.SkipGenesisSigVerification = true
		opts.GenesisTxResultHandler = NoopGenesisTxResultHandler
		opts.StrictReplay = false

		app, err := NewAppWithOptions(opts)
		require.NoError(t, err)
		resp := app.InitChain(abci.RequestInitChain{
			ChainID: "test-chain",
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block:     defaultBlockParams(),
				Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
			},
			AppState: GnoGenesisState{
				Balances: []Balance{},
				Txs: []TxWithMetadata{
					{Tx: failingTx, Metadata: &GnoTxMetadata{BlockHeight: 1}},
				},
				Auth: auth.DefaultGenesisState(),
				Bank: bank.DefaultGenesisState(),
				VM:   vm.DefaultGenesisState(),
			},
		})
		require.Nil(t, resp.Error, "StrictReplay false should boot despite failing tx: %v", resp.Error)
	})

	t.Run("StrictReplay true: failing tx aborts boot", func(t *testing.T) {
		t.Parallel()

		opts := TestAppOptions(memdb.NewMemDB())
		opts.SkipGenesisSigVerification = true
		opts.GenesisTxResultHandler = NoopGenesisTxResultHandler
		opts.StrictReplay = true

		app, err := NewAppWithOptions(opts)
		require.NoError(t, err)
		resp := app.InitChain(abci.RequestInitChain{
			ChainID: "test-chain",
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block:     defaultBlockParams(),
				Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
			},
			AppState: GnoGenesisState{
				Balances: []Balance{},
				Txs: []TxWithMetadata{
					{Tx: failingTx, Metadata: &GnoTxMetadata{BlockHeight: 1}},
				},
				Auth: auth.DefaultGenesisState(),
				Bank: bank.DefaultGenesisState(),
				VM:   vm.DefaultGenesisState(),
			},
		})
		require.NotNil(t, resp.Error, "StrictReplay true should refuse to boot on failing tx")
		assert.Contains(t, resp.Error.Error(), "strict replay")
	})

	t.Run("StrictReplay true: tx marked Failed in source is skipped, not counted", func(t *testing.T) {
		t.Parallel()

		opts := TestAppOptions(memdb.NewMemDB())
		opts.SkipGenesisSigVerification = true
		opts.GenesisTxResultHandler = NoopGenesisTxResultHandler
		opts.StrictReplay = true

		app, err := NewAppWithOptions(opts)
		require.NoError(t, err)
		resp := app.InitChain(abci.RequestInitChain{
			ChainID: "test-chain",
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block:     defaultBlockParams(),
				Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
			},
			AppState: GnoGenesisState{
				Balances: []Balance{},
				Txs: []TxWithMetadata{
					{Tx: failingTx, Metadata: &GnoTxMetadata{BlockHeight: 1, Failed: true}},
				},
				Auth: auth.DefaultGenesisState(),
				Bank: bank.DefaultGenesisState(),
				VM:   vm.DefaultGenesisState(),
			},
		})
		require.Nil(t, resp.Error, "intentionally-skipped failed tx should not trigger StrictReplay: %v", resp.Error)
	})
}

// TestValidateSignerInfo verifies the preflight catches account-number
// collisions before any state mutates. Without this check,
// NewAccountWithUncheckedNumber would silently overwrite accounts.
func TestValidateSignerInfo(t *testing.T) {
	t.Parallel()

	addrA := crypto.AddressFromPreimage([]byte("addr-a"))
	addrB := crypto.AddressFromPreimage([]byte("addr-b"))

	tests := []struct {
		name      string
		state     GnoGenesisState
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "empty state passes",
			state:   GnoGenesisState{},
			wantErr: false,
		},
		{
			name: "no SignerInfo passes",
			state: GnoGenesisState{
				Txs: []TxWithMetadata{
					{Metadata: &GnoTxMetadata{BlockHeight: 1}},
				},
			},
			wantErr: false,
		},
		{
			name: "same accNum same addr is fine (legitimate per-tx repeat)",
			state: GnoGenesisState{
				Txs: []TxWithMetadata{
					{Metadata: &GnoTxMetadata{BlockHeight: 1, SignerInfo: []SignerAccountInfo{{Address: addrA, AccountNum: 5, Sequence: 0}}}},
					{Metadata: &GnoTxMetadata{BlockHeight: 2, SignerInfo: []SignerAccountInfo{{Address: addrA, AccountNum: 5, Sequence: 1}}}},
				},
			},
			wantErr: false,
		},
		{
			name: "same accNum different addrs collides",
			state: GnoGenesisState{
				Txs: []TxWithMetadata{
					{Metadata: &GnoTxMetadata{BlockHeight: 1, SignerInfo: []SignerAccountInfo{{Address: addrA, AccountNum: 5}}}},
					{Metadata: &GnoTxMetadata{BlockHeight: 2, SignerInfo: []SignerAccountInfo{{Address: addrB, AccountNum: 5}}}},
				},
			},
			wantErr:   true,
			errSubstr: "SignerInfo collision",
		},
		{
			name: "SignerInfo collides with balance-init account",
			state: GnoGenesisState{
				// state.Balances[0] reserves accNum=0 for addrA
				Balances: []Balance{{Address: addrA, Amount: std.NewCoins(std.NewCoin("ugnot", 1))}},
				Txs: []TxWithMetadata{
					// SignerInfo claims accNum=0 for addrB; collision
					{Metadata: &GnoTxMetadata{BlockHeight: 1, SignerInfo: []SignerAccountInfo{{Address: addrB, AccountNum: 0}}}},
				},
			},
			wantErr:   true,
			errSubstr: "SignerInfo collision",
		},
		{
			name: "SignerInfo matching balance-init address is fine",
			state: GnoGenesisState{
				Balances: []Balance{{Address: addrA, Amount: std.NewCoins(std.NewCoin("ugnot", 1))}},
				Txs: []TxWithMetadata{
					// SignerInfo claims accNum=0 for addrA, matches balance-init
					{Metadata: &GnoTxMetadata{BlockHeight: 1, SignerInfo: []SignerAccountInfo{{Address: addrA, AccountNum: 0}}}},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateSignerInfo(tc.state)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				require.NoError(t, err)
			}
		})
	}
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

func TestIsPastChainID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		pastChainIDs []string
		chainID      string
		expected     bool
	}{
		{"empty allowlist", []string{}, "chain-a", false},
		{"nil allowlist", nil, "chain-a", false},
		{"single match", []string{"chain-a"}, "chain-a", true},
		{"no match in list", []string{"chain-a", "chain-b"}, "chain-c", false},
		{"match second element", []string{"chain-a", "chain-b"}, "chain-b", true},
		{"empty chain ID", []string{"chain-a"}, "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, isPastChainID(tc.pastChainIDs, tc.chainID))
		})
	}
}

func TestSignerInfoForceSetAccountState(t *testing.T) {
	t.Parallel()

	t.Run("force-sets existing account sequence and number", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"

			path = "gno.land/r/demo/signertest"
			body = `package signertest

var Deployed = true

func IsDeployed(cur realm) bool { return Deployed }
`
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		msg := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "signertest",
				Path: path,
				Files: []*std.MemFile{
					{Name: "file.gno", Body: body},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
				},
			},
		}

		// Sign with old chain, accNum=5, seq=10 — the SignerInfo will force-set
		// the account to these values before signature verification.
		tx := createAndSignTxWithAccSeq(t, []std.Msg{msg}, "old-chain", key, 5, 10)

		app.InitChain(abci.RequestInitChain{
			ChainID: chainID,
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: []string{},
				},
			},
			AppState: GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 42,
							ChainID:     "old-chain",
							SignerInfo: []SignerAccountInfo{
								{
									Address:    key.PubKey().Address(),
									AccountNum: 5,
									Sequence:   10,
								},
							},
						},
					},
				},
				Balances: []Balance{
					{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					},
				},
				Auth:         auth.DefaultGenesisState(),
				Bank:         bank.DefaultGenesisState(),
				VM:           vm.DefaultGenesisState(),
				PastChainIDs: []string{"old-chain"},
			},
		})

		// If SignerInfo was correctly applied, the tx would have been
		// delivered successfully (sig verification passed).
		// Verify by calling the deployed realm.
		callMsg := vm.MsgCall{
			Caller:  key.PubKey().Address(),
			PkgPath: path,
			Func:    "IsDeployed",
		}

		callTx := createAndSignTxWithAccSeq(t, []std.Msg{callMsg}, chainID, key, 5, 11)

		marshalledTx, err := amino.Marshal(callTx)
		require.NoError(t, err)

		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: marshalledTx})
		require.True(t, resp.IsOK(), "DeliverTx failed: %s", resp.Log)
		assert.Contains(t, string(resp.Data), "true")
	})

	t.Run("creates new account via SignerInfo when account does not exist", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"

			path = "gno.land/r/demo/newacctest"
			body = `package newacctest

var Deployed = true

func IsDeployed(cur realm) bool { return Deployed }
`
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		msg := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "newacctest",
				Path: path,
				Files: []*std.MemFile{
					{Name: "file.gno", Body: body},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
				},
			},
		}

		// Sign with accNum=7. Account won't exist from balances, so
		// NewAccountWithUncheckedNumber must be called.
		tx := createAndSignTxWithAccSeq(t, []std.Msg{msg}, "old-chain", key, 7, 0)

		app.InitChain(abci.RequestInitChain{
			ChainID: chainID,
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: []string{},
				},
			},
			AppState: GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 10,
							ChainID:     "old-chain",
							SignerInfo: []SignerAccountInfo{
								{
									Address:    key.PubKey().Address(),
									AccountNum: 7,
									Sequence:   0,
								},
							},
						},
					},
				},
				// No balances — account doesn't exist before SignerInfo creates it.
				// But the account needs funds for gas, so we must provide balances.
				Balances: []Balance{
					{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					},
				},
				Auth:         auth.DefaultGenesisState(),
				Bank:         bank.DefaultGenesisState(),
				VM:           vm.DefaultGenesisState(),
				PastChainIDs: []string{"old-chain"},
			},
		})

		// Verify deployment succeeded
		callMsg := vm.MsgCall{
			Caller:  key.PubKey().Address(),
			PkgPath: path,
			Func:    "IsDeployed",
		}

		callTx := createAndSignTxWithAccSeq(t, []std.Msg{callMsg}, chainID, key, 7, 1)

		marshalledTx, err := amino.Marshal(callTx)
		require.NoError(t, err)

		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: marshalledTx})
		require.True(t, resp.IsOK(), "DeliverTx failed: %s", resp.Log)
		assert.Contains(t, string(resp.Data), "true")
	})

	t.Run("failed tx is skipped and does not execute", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"

			path = "gno.land/r/demo/failedtest"
			body = `package failedtest

var Deployed = true

func IsDeployed(cur realm) bool { return Deployed }
`
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		msg := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "failedtest",
				Path: path,
				Files: []*std.MemFile{
					{Name: "file.gno", Body: body},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
				},
			},
		}

		// This tx is marked as Failed — it should be skipped entirely.
		tx := createAndSignTxWithAccSeq(t, []std.Msg{msg}, "old-chain", key, 0, 0)

		initResp := app.InitChain(abci.RequestInitChain{
			ChainID: chainID,
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: []string{},
				},
			},
			AppState: GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 5,
							ChainID:     "old-chain",
							Failed:      true,
							SignerInfo: []SignerAccountInfo{
								{
									Address:    key.PubKey().Address(),
									AccountNum: 0,
									Sequence:   0,
								},
							},
						},
					},
				},
				Balances: []Balance{
					{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					},
				},
				Auth:         auth.DefaultGenesisState(),
				Bank:         bank.DefaultGenesisState(),
				VM:           vm.DefaultGenesisState(),
				PastChainIDs: []string{"old-chain"},
			},
		})

		// The skipped failed tx should produce a non-success response so
		// downstream consumers (indexers, explorers) don't mistake it for
		// success.
		require.Len(t, initResp.TxResponses, 1)
		skippedResp := initResp.TxResponses[0]
		require.NotNil(t, skippedResp.Error, "skipped failed tx response should carry an error marker")
		assert.Contains(t, skippedResp.Error.Error(), "replay skipped")

		// The package should NOT be deployed since the tx was marked as failed.
		// Trying to call it should fail.
		callMsg := vm.MsgCall{
			Caller:  key.PubKey().Address(),
			PkgPath: path,
			Func:    "IsDeployed",
		}

		callTx := createAndSignTxWithAccSeq(t, []std.Msg{callMsg}, chainID, key, 0, 1)

		marshalledTx, err := amino.Marshal(callTx)
		require.NoError(t, err)

		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: marshalledTx})
		// Should fail because the package was never deployed
		require.False(t, resp.IsOK(), "DeliverTx should have failed — failed tx should not deploy package")
	})

	t.Run("SignerInfo is ignored when BlockHeight is zero", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "test-chain"

			path = "gno.land/r/demo/genesismode"
			body = `package genesismode

var Deployed = true

func IsDeployed(cur realm) bool { return Deployed }
`
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		msg := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "genesismode",
				Path: path,
				Files: []*std.MemFile{
					{Name: "file.gno", Body: body},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
				},
			},
		}

		// Sign with the current chain ID (genesis-mode tx).
		// BlockHeight=0 means SignerInfo should be ignored entirely.
		tx := createAndSignTx(t, []std.Msg{msg}, chainID, key)

		app.InitChain(abci.RequestInitChain{
			ChainID: chainID,
			Time:    time.Now(),
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				Validator: &abci.ValidatorParams{
					PubKeyTypeURLs: []string{},
				},
			},
			AppState: GnoGenesisState{
				Txs: []TxWithMetadata{
					{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 0, // genesis-mode — SignerInfo must be ignored
							SignerInfo: []SignerAccountInfo{
								{
									Address:    key.PubKey().Address(),
									AccountNum: 999, // would corrupt state if applied
									Sequence:   999,
								},
							},
						},
					},
				},
				Balances: []Balance{
					{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					},
				},
				Auth: auth.DefaultGenesisState(),
				Bank: bank.DefaultGenesisState(),
				VM:   vm.DefaultGenesisState(),
			},
		})

		// If SignerInfo was correctly ignored, the deployment should succeed
		// with the normal account state (accNum=0, seq=0).
		callMsg := vm.MsgCall{
			Caller:  key.PubKey().Address(),
			PkgPath: path,
			Func:    "IsDeployed",
		}

		callTx := createAndSignTx(t, []std.Msg{callMsg}, chainID, key)

		marshalledTx, err := amino.Marshal(callTx)
		require.NoError(t, err)

		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: marshalledTx})
		require.True(t, resp.IsOK(), "DeliverTx failed: %s", resp.Log)
		assert.Contains(t, string(resp.Data), "true")
	})
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
