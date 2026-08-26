package gnoland

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bftCfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
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
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
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

// TestInitChainer_GenesisValidatorPubKeyType verifies that the initial
// validator set seeded from genesis is gated by the consensus-params pubkey
// allow-list, mirroring the EndBlocker gate for runtime valset updates. A
// genesis validator whose key type is not allowed (e.g. secp256k1, which is
// disallowed for validators) must abort InitChain rather than seed a
// non-compliant validator into the active set.
func TestInitChainer_GenesisValidatorPubKeyType(t *testing.T) {
	t.Parallel()

	ed25519Type := amino.GetTypeURL(ed25519.PubKeyEd25519{})

	newInitReq := func(pubKey crypto.PubKey) abci.RequestInitChain {
		return abci.RequestInitChain{
			ChainID: "dev",
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
				// ed25519 is the only allowed validator key type.
				Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{ed25519Type}},
			},
			Validators: []abci.ValidatorUpdate{
				{Address: pubKey.Address(), PubKey: pubKey, Power: 1},
			},
			AppState: DefaultGenState(),
		}
	}

	t.Run("ed25519 genesis validator accepted", func(t *testing.T) {
		t.Parallel()

		app, err := NewApp(t.TempDir(), NewTestGenesisAppConfig(), config.DefaultAppConfig(), events.NewEventSwitch(), log.NewNoopLogger(), 0)
		require.NoError(t, err)

		resp := app.InitChain(newInitReq(ed25519.GenPrivKey().PubKey()))
		assert.True(t, resp.IsOK(), "resp is not OK: %v", resp)
	})

	t.Run("secp256k1 genesis validator aborts boot", func(t *testing.T) {
		t.Parallel()

		app, err := NewApp(t.TempDir(), NewTestGenesisAppConfig(), config.DefaultAppConfig(), events.NewEventSwitch(), log.NewNoopLogger(), 0)
		require.NoError(t, err)

		// getDummyKey produces a secp256k1 key, which is not in the allow-list.
		secpKey := getDummyKey(t).PubKey()
		wantErr := fmt.Sprintf(
			"genesis validator %s has disallowed pubkey type %s (allowed: %v)",
			secpKey.Address().String(), amino.GetTypeURL(secpKey), []string{ed25519Type},
		)
		assert.PanicsWithError(t, wantErr, func() {
			app.InitChain(newInitReq(secpKey))
		})
	})
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
	ms.MountStoreWithDB(iavlCapKey, storebptree.FastStoreConstructor, db)
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

// TestInitChainer_PanicsOnValoperCoverageFailure pins that an
// assertGenesisValopersConsistent failure aborts boot via a Go-level
// panic, not via the ResponseInitChain.Error field that tm2's
// consensus/replay.go:339-342 silently discards. Without this guarantee
// a hardfork chain can boot in a state where genesis validators have no
// v3 operator-keyed management plane — the safety net would fire but
// not actually stop the boot.
func TestInitChainer_PanicsOnValoperCoverageFailure(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")
	ms.MountStoreWithDB(baseCapKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlCapKey, storebptree.FastStoreConstructor, db)
	ms.LoadLatestVersion()
	testCtx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(),
		&bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())

	// vmk.Call is what assertGenesisValopersConsistent invokes; returning
	// an error from it is the realistic shape of an assertion failure
	// (uncovered genesis validator → v3 panics → vmk.Call returns the
	// wrapped error).
	mock := &mockVMKeeper{
		callFn: func(_ sdk.Context, _ vm.MsgCall) (string, error) {
			return "", fmt.Errorf("synthetic v3 assertion: uncovered validator")
		},
	}

	cfg := InitChainerConfig{
		vmk:   mock,
		acck:  &mockAuthKeeper{},
		bankk: &mockBankKeeper{},
		prmk:  &mockParamsKeeper{},
		gpk:   &mockGasPriceKeeper{},
	}

	// PastChainIDs + Validators → shouldAssertValoperCoverage returns
	// true → the assertion runs against the mocked vmk and fails.
	req := abci.RequestInitChain{
		Validators: generateValidatorUpdates(t, 1),
		AppState:   GnoGenesisState{PastChainIDs: []string{"old-chain"}},
	}

	assert.PanicsWithError(t,
		"genesis valoper coverage assertion failed: synthetic v3 assertion: uncovered validator",
		func() { cfg.InitChainer(testCtx, req) },
		"InitChainer must panic on valoper coverage failure so tm2's handshake aborts; ResponseInitChain.Error is discarded by consensus/replay.go",
	)
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

	for range count {
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

		// Corrupted current panics rather than being silently
		// recovered. Only chain code writes valset:current (via
		// ctx-sentinel), so corruption is by definition a chain-
		// internal bug or store damage and warrants a panic.
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

// TestGasPriceMinimumIsCheckTxOnly verifies that the production ante handler
// enforces the block gas price minimum in CheckTx, but not in DeliverTx.
func TestGasPriceMinimumIsCheckTxOnly(t *testing.T) {
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	baseApp := app.(*sdk.BaseApp)

	initialPrice := std.GasPrice{
		Gas:   1000,
		Price: std.Coin{Amount: 100, Denom: "ugnot"},
	}

	gen := gnoGenesisState(t)
	gen.Auth.Params.InitialGasPrice = initialPrice

	privKey := getDummyKey(t)
	addr := privKey.PubKey().Address()

	gen.Balances = []Balance{
		{
			Address: addr,
			Amount:  []std.Coin{{Amount: 10_000_000_000, Denom: "ugnot"}},
		},
	}

	initRes := baseApp.InitChain(abci.RequestInitChain{
		AppState: gen,
		ChainID:  "test-chain",
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{MaxGas: 10_000_000},
		},
	})
	require.True(
		t,
		initRes.ResponseBase.IsOK(),
		"InitChain failed: %v",
		initRes.ResponseBase.Error,
	)
	baseApp.Commit()

	query := abci.RequestQuery{
		Path: ".store/main/key",
		Data: []byte(auth.GasPriceKey),
	}

	var storedPrice std.GasPrice
	qr := baseApp.Query(query)
	require.True(t, qr.IsOK(), "query failed: %v", qr.ResponseBase.Error)
	require.NoError(t, amino.Unmarshal(qr.Value, &storedPrice))
	require.Equal(t, initialPrice, storedPrice)

	msgs := []std.Msg{
		bank.NewMsgSend(
			addr,
			crypto.AddressFromPreimage([]byte("test-account")),
			std.Coins{{Denom: "ugnot", Amount: 100}},
		),
	}

	tx := createAndSignTx(t, msgs, "test-chain", privKey)
	tx.Fee = std.Fee{
		GasWanted: 10_000_000,
		GasFee:    std.Coin{Amount: 9, Denom: "ugnot"},
	}

	// The fee is part of the sign bytes, so the transaction must be re-signed.
	signBytes, err := tx.GetSignBytes("test-chain", 0, 0)
	require.NoError(t, err)

	sig, err := privKey.Sign(signBytes)
	require.NoError(t, err)

	tx.Signatures = []std.Signature{
		{
			PubKey:    privKey.PubKey(),
			Signature: sig,
		},
	}

	txBytes, err := amino.Marshal(tx)
	require.NoError(t, err)

	checkRes := baseApp.CheckTx(abci.RequestCheckTx{Tx: txBytes})
	require.False(
		t,
		checkRes.IsOK(),
		"CheckTx should reject a transaction below the block gas price: %v",
		checkRes,
	)
	require.Contains(
		t,
		checkRes.Log,
		"as block gas price",
		"unexpected CheckTx error: %s",
		checkRes.Log,
	)

	baseApp.BeginBlock(abci.RequestBeginBlock{
		Header: &bft.Header{
			ChainID: "test-chain",
			Height:  2,
		},
	})

	// This documents the current bug: DeliverTx does not enforce the minimum.
	deliverRes := baseApp.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(
		t,
		deliverRes.IsOK(),
		"DeliverTx failed: %+v",
		deliverRes,
	)
}

func TestGasPriceEmptyBlockUpdate(t *testing.T) {
	startingPrice := std.GasPrice{
		Gas:   1000,
		Price: std.Coin{Amount: 10, Denom: "ugnot"},
	}
	floorPrice := std.GasPrice{
		Gas:   1000,
		Price: std.Coin{Amount: 1, Denom: "ugnot"},
	}

	run := func() []byte {
		app := newGasPriceTestApp(t, startingPrice)
		gen := gnoGenesisState(t)
		gen.Auth.Params.InitialGasPrice = floorPrice
		res := app.InitChain(abci.RequestInitChain{
			AppState: gen,
			ChainID:  "test-chain",
			ConsensusParams: &abci.ConsensusParams{
				Block: &abci.BlockParams{MaxGas: 3_000_000_000},
			},
		})
		require.True(t, res.ResponseBase.IsOK(), "%v", res.ResponseBase.Error)

		genesisHash := app.Commit().Data
		query := abci.RequestQuery{Path: ".store/main/key", Data: []byte(auth.GasPriceKey)}
		var gasPrice std.GasPrice
		qr := app.Query(query)
		require.True(t, qr.IsOK(), "%v", qr.ResponseBase.Error)
		require.NoError(t, amino.Unmarshal(qr.Value, &gasPrice))
		require.Equal(t, startingPrice, gasPrice)

		tx := newCounterTx(1)
		tx.Fee = std.Fee{
			GasWanted: 1000,
			GasFee:    std.Coin{Amount: 9, Denom: "ugnot"},
		}
		txBytes, err := amino.Marshal(tx)
		require.NoError(t, err)
		require.False(t, app.CheckTx(abci.RequestCheckTx{Tx: txBytes}).IsOK())

		app.BeginBlock(abci.RequestBeginBlock{
			Header: &bft.Header{ChainID: "test-chain", Height: 2},
		})
		app.EndBlock(abci.RequestEndBlock{Height: 2})
		emptyBlockHash := app.Commit().Data

		qr = app.Query(query)
		require.True(t, qr.IsOK(), "%v", qr.ResponseBase.Error)
		require.NoError(t, amino.Unmarshal(qr.Value, &gasPrice))
		require.Equal(t, int64(9), gasPrice.Price.Amount)
		require.Equal(t, startingPrice.Gas, gasPrice.Gas)
		require.Equal(t, startingPrice.Price.Denom, gasPrice.Price.Denom)
		require.True(t, app.CheckTx(abci.RequestCheckTx{Tx: txBytes}).IsOK())
		require.NotEqual(t, genesisHash, emptyBlockHash)

		return emptyBlockHash
	}

	require.Equal(t, run(), run(), "identical empty blocks must produce identical state transitions")
}

func newGasPriceTestApp(t *testing.T, storedGasPrice ...std.GasPrice) abci.Application {
	t.Helper()
	cfg := TestAppOptions(memdb.NewMemDB())
	cfg.EventSwitch = events.NewEventSwitch()

	// Capabilities keys.
	mainKey := store.NewStoreKey("main")
	baseKey := store.NewStoreKey("base")

	baseApp := sdk.NewBaseApp("gnoland", cfg.Logger, cfg.DB, baseKey, mainKey)
	baseApp.SetAppVersion("test")

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, cfg.DB)
	baseApp.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, cfg.DB)

	// Construct keepers.
	prmk := params.NewParamsKeeper(mainKey)
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, std.ProtoBaseSessionAccount)
	gpk := auth.NewGasPriceKeeper(mainKey)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName), mainKey, []string{ugnot.Denom})
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
	baseApp.SetInitChainer(func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		res := icc.InitChainer(ctx, req)
		if len(storedGasPrice) > 0 && res.ResponseBase.IsOK() {
			gpk.SetGasPrice(ctx, storedGasPrice[0])
		}
		return res
	})

	// Set AnteHandler
	baseApp.SetAnteHandler(
		// Override default AnteHandler with custom logic.
		func(ctx sdk.Context, tx std.Tx, simulate bool) (
			newCtx sdk.Context, res sdk.Result, abort bool,
		) {
			// Keeper bookkeeping reads run unmetered: the test's price math
			// asserts exact values, so block gas consumed must equal the
			// counter values exactly — calibrated store-read costs would
			// otherwise blow the small block budget and skew the economics.
			readCtx := ctx.WithGasMeter(store.NewInfiniteGasMeter())

			// Add last gas price in the context
			ctx = ctx.WithValue(auth.GasPriceContextKey{}, gpk.LastGasPrice(readCtx))

			// Override auth params.
			ctx = ctx.WithValue(auth.AuthParamsContextKey{}, acck.GetParams(readCtx))
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
	cms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	cms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)

	// Make sure loading a past version doesn't fail
	assert.NoError(t, cms.LoadVersion(1))

	// Release the multistore (its query snapshot) before closing the DB —
	// PebbleDB reports open snapshots as a leak at Close time.
	if closer, ok := cms.(io.Closer); ok {
		require.NoError(t, closer.Close())
	}

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

	// A historical MsgRun must replay even when run_submitters excludes its signer.
	//
	// This is the carve-out in checkCodePolicy, and it is load-bearing for every
	// hardfork: deliverGenesisTx replays history through the same ante with
	// BlockHeight > 0, AFTER InitGenesis installs the NEW params. So without the
	// exemption a fork refuses to replay its own past the moment a historical
	// signer is absent from the new allowlist -- and a fork that rotates
	// operators will routinely have historical signers who are not on the new
	// list. With StrictReplay the node would not boot at all.
	//
	// Left untested, a regression here would not show up in any suite; it would
	// show up the next time somebody forks the chain. The genesis state below
	// populates run_submitters with an address that is NOT the signer, so the
	// gate is armed and the tx is deliverable only because it is recognised as
	// replay. An EMPTY list would prove nothing here — empty means the gate is
	// off, so the tx would pass with or without the carve-out.
	t.Run("historical MsgRun replays under a run_submitters list it is not on", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		// Confirm the premise rather than assuming it: the params this genesis
		// installs really do refuse MsgRun for this signer.
		vmGen := vm.DefaultGenesisState()
		vmGen.Params.RunSubmitters = []crypto.Address{
			crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		}
		require.NotContains(t, vmGen.Params.RunSubmitters, key.PubKey().Address(),
			"premise: the signer must be off the allowlist for this test to mean anything")

		// MsgRun.ValidateBasic forces this exact path, so anything else is
		// refused before the ante is reached and the test would prove nothing.
		runPath := "gno.land/e/" + key.PubKey().Address().String() + "/run"
		runMsg := vm.MsgRun{
			Caller: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "main",
				Path: runPath,
				Files: []*std.MemFile{
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(runPath)},
					{Name: "main.gno", Body: "package main\n\nfunc main() {\n\tprintln(\"replayed\")\n}\n"},
				},
			},
		}
		tx := createAndSignTx(t, []std.Msg{runMsg}, "old-chain", key)

		// PanicOnFailingTxResultHandler (from TestAppOptions) panics if a
		// genesis tx fails, so a refused replay surfaces as a panic here.
		require.NotPanics(t, func() {
			app.InitChain(abci.RequestInitChain{
				ChainID:       chainID,
				Time:          time.Now(),
				InitialHeight: 100,
				ConsensusParams: &abci.ConsensusParams{
					Block:     defaultBlockParams(),
					Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
				},
				AppState: GnoGenesisState{
					Txs: []TxWithMetadata{{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 42,
							ChainID:     "old-chain",
						},
					}},
					Balances: []Balance{{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					}},
					Auth:          auth.DefaultGenesisState(),
					Bank:          bank.DefaultGenesisState(),
					VM:            vmGen,
					PastChainIDs:  []string{"old-chain"},
					InitialHeight: 100,
				},
			})
		}, "a historical MsgRun must replay even though run_submitters excludes its signer")

		// And the gate is still armed for live traffic: the same signer sending
		// the same message NOT as replay must be refused. Without this the test
		// would also pass if the carve-out were replaced by no gate at all.
		liveTx := createAndSignTxWithAccSeq(t, []std.Msg{runMsg}, chainID, key, 0, 1)
		marshalled, err := amino.Marshal(liveTx)
		require.NoError(t, err)
		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: marshalled})
		require.False(t, resp.IsOK(),
			"a live MsgRun must still be refused; only replay is exempt")
		assert.Contains(t, resp.Log, "run_submitters")
	})

	// The same carve-out has to cover a FRESH chain's own genesis txs, not just
	// replayed history.
	//
	// gnoland marks every tx delivered during InitChain, metadata or not, and
	// production bootstrap depends on that: a chain seeds its first GovDAO
	// members with a genesis MsgRun, before any allowlist could name them. The
	// key's doc used to describe it as hardfork-only, so a reasonable-looking
	// tightening -- requiring metadata, or BlockHeight > 0 -- would break every
	// fresh launch that ships a non-empty run_submitters, and no test said so.
	//
	// The genesis below has no Metadata and no PastChainIDs, which is what makes
	// it a fresh launch rather than a replay.
	t.Run("fresh-launch genesis MsgRun is exempt from run_submitters", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "fresh-chain"
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		vmGen := vm.DefaultGenesisState()
		vmGen.Params.RunSubmitters = []crypto.Address{
			crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		}
		require.NotContains(t, vmGen.Params.RunSubmitters, key.PubKey().Address(),
			"premise: the signer must be off the allowlist for this test to mean anything")

		runPath := "gno.land/e/" + key.PubKey().Address().String() + "/run"
		runMsg := vm.MsgRun{
			Caller: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "main",
				Path: runPath,
				Files: []*std.MemFile{
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(runPath)},
					{Name: "main.gno", Body: "package main\n\nfunc main() {\n\tprintln(\"bootstrapped\")\n}\n"},
				},
			},
		}
		tx := createAndSignTx(t, []std.Msg{runMsg}, chainID, key)

		require.NotPanics(t, func() {
			app.InitChain(abci.RequestInitChain{
				ChainID:       chainID,
				Time:          time.Now(),
				InitialHeight: 1,
				ConsensusParams: &abci.ConsensusParams{
					Block:     defaultBlockParams(),
					Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
				},
				AppState: GnoGenesisState{
					// No Metadata: this is a fresh chain's own genesis tx.
					Txs: []TxWithMetadata{{Tx: tx}},
					Balances: []Balance{{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					}},
					Auth: auth.DefaultGenesisState(),
					Bank: bank.DefaultGenesisState(),
					VM:   vmGen,
				},
			})
		}, "a fresh chain's genesis MsgRun must be exempt, or bootstrap cannot seed governance")
	})

	// The carve-out guards two gates, and the two tests above pin only one.
	//
	// checkCodePolicy refuses MsgRun by run_submitters AND MsgAddPackage by
	// code_submitters, and the same exemption covers both. A tightening that
	// kept the MsgRun half working while breaking the MsgAddPackage half -- for
	// instance restricting the carve-out to transactions with no add-package
	// signers -- left the whole package green before this test existed.
	//
	// The shape here is a hardfork replaying an MsgAddPackage under a
	// "permissioned" policy whose code_submitters no longer lists the historical
	// signer, which is what rotating deployers after a fork looks like.
	t.Run("historical MsgAddPackage replays under a code_submitters list it is not on", func(t *testing.T) {
		t.Parallel()

		var (
			db      = memdb.NewMemDB()
			key     = getDummyKey(t)
			chainID = "new-chain"
		)

		app, err := NewAppWithOptions(TestAppOptions(db))
		require.NoError(t, err)

		vmGen := vm.DefaultGenesisState()
		vmGen.Params.CodeSubmissionPolicy = vm.CodeSubmissionPolicyPermissioned
		vmGen.Params.CodeSubmitters = []crypto.Address{
			crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		}
		require.NotContains(t, vmGen.Params.CodeSubmitters, key.PubKey().Address(),
			"premise: the signer must be off the allowlist for this test to mean anything")

		const pkgPath = "gno.land/r/test/replayedpkg"
		addMsg := vm.MsgAddPackage{
			Creator: key.PubKey().Address(),
			Package: &std.MemPackage{
				Name: "replayedpkg",
				Path: pkgPath,
				Files: []*std.MemFile{
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
					{Name: "replayedpkg.gno", Body: "package replayedpkg\n\nfunc Hello() string { return \"hi\" }\n"},
				},
			},
		}
		tx := createAndSignTx(t, []std.Msg{addMsg}, "old-chain", key)

		require.NotPanics(t, func() {
			app.InitChain(abci.RequestInitChain{
				ChainID:       chainID,
				Time:          time.Now(),
				InitialHeight: 100,
				ConsensusParams: &abci.ConsensusParams{
					Block:     defaultBlockParams(),
					Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
				},
				AppState: GnoGenesisState{
					Txs: []TxWithMetadata{{
						Tx: tx,
						Metadata: &GnoTxMetadata{
							Timestamp:   time.Now().Unix(),
							BlockHeight: 42,
							ChainID:     "old-chain",
						},
					}},
					Balances: []Balance{{
						Address: key.PubKey().Address(),
						Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
					}},
					Auth:          auth.DefaultGenesisState(),
					Bank:          bank.DefaultGenesisState(),
					VM:            vmGen,
					PastChainIDs:  []string{"old-chain"},
					InitialHeight: 100,
				},
			})
		}, "a historical MsgAddPackage must replay even though code_submitters excludes its signer")

		// And the gate is still armed for live traffic, so this cannot pass by
		// the policy simply not being enforced.
		liveTx := createAndSignTxWithAccSeq(t, []std.Msg{addMsg}, chainID, key, 0, 1)
		marshalled, err := amino.Marshal(liveTx)
		require.NoError(t, err)
		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: marshalled})
		require.False(t, resp.IsOK(),
			"a live MsgAddPackage must still be refused; only replay is exempt")
		assert.Contains(t, resp.Log, "code_submitters")
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

		// PANICS rather than returning an error, and the distinction is the
		// whole point of the guard. An error here reaches
		// ResponseInitChain.Error and stops: localClient.InitChainSync returns
		// a nil Go error regardless and the handshake reads only that, so the
		// field was populated and never inspected. Asserting on resp.Error --
		// which this test used to do -- passed while the node booted anyway.
		require.PanicsWithError(t,
			"strict replay: 1 genesis tx(s) failed; chain refusing to boot "+
				"(inspect the per-failure 'Genesis replay failure' log lines for details)",
			func() {
				app.InitChain(abci.RequestInitChain{
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
			},
			"StrictReplay must abort the boot, not report and continue")
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
	cms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
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

// TestInitChainer_StreamingAppState exercises the streaming code path: an
// AppState delivered as *GenesisStateRef (an on-disk-backed handle) must be
// applied by loadAppState the same way an in-memory GnoGenesisState would
// be. This is the goal-test for the type switch in loadAppState — without
// it InitChain rejects the AppState as "invalid AppState of type
// *gnoland.GenesisStateRef".
func TestInitChainer_StreamingAppState(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("streaming-init-chain"))
	state := DefaultGenState()
	state.Balances = []Balance{
		{Address: addr, Amount: []std.Coin{{Amount: 1e15, Denom: "ugnot"}}},
	}
	state.Txs = nil // exercise type switch only; tx-delivery path covered separately

	src := writeMinimalGenesisFile(t, "dev", state)

	doc, err := LoadStreamingGenesisDoc(src, t.TempDir(), nil)
	require.NoError(t, err)
	ref, ok := doc.AppState.(*GenesisStateRef)
	require.True(t, ok, "fixture loader must produce *GenesisStateRef")
	require.Equal(t, 1, ref.BalanceCount())
	require.Equal(t, 0, ref.TxCount())

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)

	resp := bapp.InitChain(abci.RequestInitChain{
		Time:    time.Now(),
		ChainID: "dev",
		ConsensusParams: &abci.ConsensusParams{
			Block: defaultBlockParams(),
		},
		Validators: []abci.ValidatorUpdate{},
		AppState:   ref,
	})
	require.True(t, resp.IsOK(), "InitChain response: %v", resp)
	assert.Empty(t, resp.TxResponses, "no genesis txs in this fixture")
}

// TestInitChainer_StreamingAppState_TxParity asserts the streaming
// code path delivers genesis txs and surfaces responses identically to
// the in-memory path. Same AppState fed through both paths must produce
// the same number of tx responses with matching OK/error status.
func TestInitChainer_StreamingAppState_TxParity(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("streaming-tx-parity"))
	state := DefaultGenState()
	state.Balances = []Balance{
		{Address: addr, Amount: []std.Coin{{Amount: 1e15, Denom: "ugnot"}}},
	}
	state.Txs = []TxWithMetadata{
		{
			Tx: std.Tx{
				Msgs: []std.Msg{vm.NewMsgAddPackage(addr, "gno.land/r/demo", []*std.MemFile{
					{Name: "demo.gno", Body: "package demo; func Hello(cur realm) string { return `hello`; }"},
					{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/demo")},
				})},
				Fee:        std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
				Signatures: []std.Signature{{}},
			},
		},
	}

	initReq := func(appState any) abci.RequestInitChain {
		return abci.RequestInitChain{
			Time:    time.Now(),
			ChainID: "dev",
			ConsensusParams: &abci.ConsensusParams{
				Block: defaultBlockParams(),
			},
			Validators: []abci.ValidatorUpdate{},
			AppState:   appState,
		}
	}

	// In-memory path.
	memApp, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	memResp := memApp.(*sdk.BaseApp).InitChain(initReq(state))
	require.True(t, memResp.IsOK(), "in-memory InitChain: %v", memResp)
	require.Len(t, memResp.TxResponses, 1)

	// Streaming path: same state, marshalled to disk, loaded as ref.
	src := writeMinimalGenesisFile(t, "dev", state)
	doc, err := LoadStreamingGenesisDoc(src, t.TempDir(), nil)
	require.NoError(t, err)
	ref := doc.AppState.(*GenesisStateRef)

	streamApp, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	streamResp := streamApp.(*sdk.BaseApp).InitChain(initReq(ref))
	require.True(t, streamResp.IsOK(), "streaming InitChain: %v", streamResp)
	require.Len(t, streamResp.TxResponses, 1)

	// The two paths must agree on tx outcome.
	assert.Equal(t, memResp.TxResponses[0].Error, streamResp.TxResponses[0].Error,
		"in-memory vs streaming tx outcomes must match")
}

func TestInitChainer_VestingAccount(t *testing.T) {
	t.Parallel()

	key := getDummyKey(t)
	addr := key.PubKey().Address()
	chainID := "test"

	vestingAmount := std.NewCoins(std.NewCoin("ugnot", 500_000))
	totalBalance := std.NewCoins(std.NewCoin("ugnot", 1_000_000))

	tests := []struct {
		name      string
		vesting   *std.VestingSchedule
		isVesting bool
	}{
		{
			"continuous vesting",
			&std.VestingSchedule{
				OriginalVesting: vestingAmount,
				StartTime:       100,
				EndTime:         200,
			},
			true,
		},
		{
			"delayed vesting",
			&std.VestingSchedule{
				OriginalVesting: vestingAmount,
				StartTime:       0,
				EndTime:         200,
				Type:            std.VestingDelayed,
			},
			true,
		},
		{
			"no vesting",
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDb := memdb.NewMemDB()
			testApp, err := NewAppWithOptions(TestAppOptions(testDb))
			require.NoError(t, err)

			state := DefaultGenState()
			state.Balances = []Balance{
				{
					Address: addr,
					Amount:  totalBalance,
					Vesting: tt.vesting,
				},
			}

			resp := testApp.InitChain(abci.RequestInitChain{
				ChainID: chainID,
				Time:    time.Unix(150, 0), // halfway through vesting
				ConsensusParams: &abci.ConsensusParams{
					Block: defaultBlockParams(),
					Validator: &abci.ValidatorParams{
						PubKeyTypeURLs: []string{},
					},
				},
				AppState: state,
			})
			require.True(t, resp.IsOK(), "InitChain response: %v", resp)

			// Commit to persist the genesis state before querying.
			cres := testApp.Commit()
			require.NotNil(t, cres)

			// Query the account to verify it exists.
			qres := testApp.Query(abci.RequestQuery{
				Path: fmt.Sprintf("auth/accounts/%s", addr),
			})
			require.True(t, qres.IsOK(), "account query response: %v", qres)

			if tt.isVesting {
				// An ordinary account carrying a schedule, so what identifies it
				// is the field rather than a type name.
				assert.Contains(t, string(qres.Data), `"vesting"`)
				assert.Contains(t, string(qres.Data), vestingAmount.String())
				// The account number must be present.
				assert.Contains(t, string(qres.Data), "account_number")
			} else {
				// An account with no schedule must not carry the key at all: that
				// is what keeps the field out of every existing account's bytes.
				assert.NotContains(t, string(qres.Data), `"vesting"`)
			}

			// Verify the coins are set correctly.
			qresBank := testApp.Query(abci.RequestQuery{
				Path: fmt.Sprintf("bank/balances/%s", addr),
			})
			require.True(t, qresBank.IsOK(), "bank query response: %v", qresBank)
			assert.Contains(t, string(qresBank.Data), "ugnot")
		})
	}
}

// writeMinimalGenesisFile emits a tm2.GenesisDoc-shaped JSON file under
// t.TempDir() that wraps the given GnoGenesisState as `app_state`. Uses
// the same SaveAs serialization the production gnogenesis CLI uses, so
// the fixture's wire shape matches what LoadStreamingGenesisDoc sees in
// production.
func writeMinimalGenesisFile(t *testing.T, chainID string, state GnoGenesisState) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "genesis.json")
	doc := &bft.GenesisDoc{
		ChainID:  chainID,
		AppState: state,
	}
	require.NoError(t, doc.SaveAs(dst))
	return dst
}

// Both genesis paths must call RecomputeSupply. Pinned through the mock rather than
// by calling RecomputeSupply directly: a test that invokes it itself proves the
// function works, not that InitChainer uses it, and would pass with every call site
// deleted. Table-driven over both AppState shapes because the streaming path is a
// separate call site, and covering only the in-memory one left it free to be dropped.
func TestInitChainerSeedsSupply(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		appState func(t *testing.T) any
	}{
		{"in-memory", func(*testing.T) any {
			def := DefaultGenState()
			def.Balances = []Balance{{
				Address: crypto.AddressFromPreimage([]byte("streamed-holder")),
				Amount:  std.Coins{{Denom: ugnot.Denom, Amount: 10}},
			}}
			return def
		}},
		{"streaming", func(t *testing.T) any {
			t.Helper()
			// Marshalled from the same defaults the in-memory case uses, so the two
			// cases differ only in how genesis is delivered.
			def := DefaultGenState()
			small := map[string]string{}
			for k, v := range map[string]any{"auth": def.Auth, "bank": def.Bank, "vm": def.VM} {
				bz, err := amino.MarshalJSON(v)
				require.NoError(t, err)
				small[k] = string(bz)
			}
			holder := crypto.AddressFromPreimage([]byte("streamed-holder"))
			bal := fmt.Sprintf("%q", holder.String()+"=10ugnot")
			ref, err := OpenGenesisStateRef(writeTestCache(t, []string{bal}, nil, small))
			require.NoError(t, err)
			return ref
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seedsSupply(t, tc.appState(t))
		})
	}
}

func seedsSupply(t *testing.T, appState any) {
	t.Helper()

	db := memdb.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	baseKey := store.NewStoreKey("baseKey")
	mainKey := store.NewStoreKey("mainKey")
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(),
		&bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())

	bankk := &mockBankKeeper{}
	cfg := InitChainerConfig{
		vmk:   &mockVMKeeper{},
		acck:  &mockAuthKeeper{},
		bankk: bankk,
		prmk:  &mockParamsKeeper{},
		gpk:   &mockGasPriceKeeper{},
	}
	res := cfg.InitChainer(ctx, abci.RequestInitChain{AppState: appState})
	require.Nil(t, res.Error, "InitChainer must succeed for this to mean anything")
	require.Equal(t, 1, bankk.recomputeSupplyCalls,
		"InitChainer must seed the supply counter from the genesis balances")
	require.Equal(t, 1, bankk.setCoinsAtRecompute,
		"the counter must be recomputed after every balance is applied: SetCoins does "+
			"not maintain it, so recomputing first leaves balances with no supply record")
}

// The values RecomputeSupply seeds, and that an unseeded chain is actually reported.
// Whether InitChainer calls it at all is TestInitChainerSeedsSupply's job: this one
// calls it directly, so it has no opinion on the call sites.
func TestGenesisSeedsTheSupplyCounter(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseKey := store.NewStoreKey("baseKey")
	mainKey := store.NewStoreKey("mainKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(), &bft.Header{ChainID: "test"}, log.NewNoopLogger())

	prmk := params.NewParamsKeeper(mainKey)
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, std.ProtoBaseSessionAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName), mainKey, []string{ugnot.Denom})
	prmk.Register(auth.ModuleName, acck)
	prmk.Register(bank.ModuleName, bankk)
	cfg := InitChainerConfig{acck: acck, bankk: bankk}

	// Both tiers, and a vesting entry — the shape where a delta hook would compute
	// zero, which is why seeding is a sweep.
	realm := "/gno.land/r/x/tok:c"
	vesting := std.Coins{{Denom: ugnot.Denom, Amount: 400}}
	for _, bal := range []Balance{
		{Address: crypto.AddressFromPreimage([]byte("g1")), Amount: std.Coins{{Denom: ugnot.Denom, Amount: 100}}},
		{Address: crypto.AddressFromPreimage([]byte("g2")), Amount: std.Coins{{Denom: realm, Amount: 7}, {Denom: ugnot.Denom, Amount: 250}}},
		{Address: crypto.AddressFromPreimage([]byte("g3")), Amount: vesting, Vesting: &std.VestingSchedule{
			OriginalVesting: vesting, StartTime: 100, EndTime: 1_000_000,
		}},
	} {
		cfg.applyBalance(ctx, bal)
	}

	// Before seeding the counter is empty and the invariant says so.
	require.Zero(t, bankk.TotalSupply(ctx, ugnot.Denom))
	_, broken := bank.SupplyInvariant(bankk.ViewKeeper)(ctx)
	require.True(t, broken, "unseeded supply must be reported, or the seed is untested")

	cfg.bankk.RecomputeSupply(ctx)

	require.Equal(t, int64(100+250+400), bankk.TotalSupply(ctx, ugnot.Denom))
	require.Equal(t, int64(7), bankk.TotalSupply(ctx, realm))
	msg, broken := bank.AllInvariants(bankk.ViewKeeper)(ctx)
	require.False(t, broken, "seeded genesis state must be clean:\n%s", msg)
}

// A genesis balance list may name one address twice — they are assembled from several
// sources, and the integration harness appends to a loaded default set. Pins what that
// actually does, since it is easy to assume it either accumulates or errors: one
// account survives, the last entry's amount wins, and supply seeding agrees.
// A rejected genesis balance aborts InitChain, and the causes include a denom past
// the length bound. The panic has to name the entry: std.ErrInvalidCoins carries its
// detail only under %+v and a panic renders its value with %v, so a bare panic(err)
// would say nothing but "invalid coins error" out of a genesis holding thousands.
func TestGenesisBalanceRejectionNamesTheEntry(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseKey := store.NewStoreKey("baseKey")
	mainKey := store.NewStoreKey("mainKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(),
		&bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())

	prmk := params.NewParamsKeeper(mainKey)
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, ProtoGnoSessionAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName), mainKey, []string{ugnot.Denom})
	cfg := InitChainerConfig{acck: acck, bankk: bankk}

	addr := crypto.AddressFromPreimage([]byte("bad-genesis-entry"))
	overlong := "/gno.land/r/x/" + strings.Repeat("z", 300) + ":c"

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		cfg.applyBalance(ctx, Balance{Address: addr, Amount: std.Coins{{Denom: overlong, Amount: 5}}})
	}()
	require.NotNil(t, recovered, "a denom past the length bound must abort genesis")
	msg := fmt.Sprintf("%v", recovered)
	require.Contains(t, msg, addr.String(), "the panic must name the address to fix")
	require.Contains(t, msg, overlong, "and the amount that was rejected")
}

func TestApplyBalanceWithARepeatedAddress(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseKey := store.NewStoreKey("baseKey")
	mainKey := store.NewStoreKey("mainKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(), &bft.Header{ChainID: "test"}, log.NewNoopLogger())

	prmk := params.NewParamsKeeper(mainKey)
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, std.ProtoBaseSessionAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName), mainKey, []string{ugnot.Denom})
	prmk.Register(auth.ModuleName, acck)
	prmk.Register(bank.ModuleName, bankk)
	cfg := InitChainerConfig{acck: acck, bankk: bankk}

	addr := crypto.AddressFromPreimage([]byte("dup"))
	cfg.applyBalance(ctx, Balance{Address: addr, Amount: std.Coins{{Denom: ugnot.Denom, Amount: 100}}})
	cfg.applyBalance(ctx, Balance{Address: addr, Amount: std.Coins{{Denom: ugnot.Denom, Amount: 250}}})

	acc := acck.GetAccount(ctx, addr)
	require.NotNil(t, acc, "one account must survive")
	require.Equal(t, int64(250), bankk.GetCoin(ctx, addr, ugnot.Denom),
		"the later entry's amount wins; the balance is not accumulated")

	// A plain entry after a vesting one must clear the schedule, or the funds stay
	// locked until EndTime. This is why the account is recreated rather than reused.
	vester := crypto.AddressFromPreimage([]byte("vester"))
	amount := std.Coins{{Denom: ugnot.Denom, Amount: 1000}}
	cfg.applyBalance(ctx, Balance{Address: vester, Amount: amount, Vesting: &std.VestingSchedule{
		OriginalVesting: amount, StartTime: 100, EndTime: 1_000_000,
	}})
	require.IsType(t, &GnoAccount{}, acck.GetAccount(ctx, vester),
		"a vesting balance is an ordinary account carrying a schedule")
	require.False(t, acck.GetAccount(ctx, vester).GetVesting().IsZero(),
		"the schedule must have been set")
	cfg.applyBalance(ctx, Balance{Address: vester, Amount: std.Coins{{Denom: ugnot.Denom, Amount: 500}}})
	require.True(t, acck.GetAccount(ctx, vester).GetVesting().IsZero(),
		"a plain entry must clear an earlier vesting schedule")
	require.NoError(t, bankk.SubtractCoins(ctx, vester, std.Coins{{Denom: ugnot.Denom, Amount: 1}}),
		"and the funds must be spendable")

	// Supply seeding sums what is actually held, so a repeat cannot double-count.
	bankk.RecomputeSupply(ctx)
	require.Equal(t, int64(250+499), bankk.TotalSupply(ctx, ugnot.Denom))
	msg, broken := bank.AllInvariants(bankk.ViewKeeper)(ctx)
	require.False(t, broken, "invariants must be clean after a repeated entry:\n%s", msg)
}

// TestGenesisSignerMintIsAccounted covers the one place genesis creates coins
// outside the balances file: a genesis transaction whose signer has no account is
// funded so it can pay for itself. That has to mint rather than credit. The supply
// counter is seeded from the balances before genesis txs replay, so a
// counter-blind credit would leave the chain holding coins it has no record of,
// and SupplyInvariant broken from block one.
func TestGenesisSignerMintIsAccounted(t *testing.T) {
	t.Parallel()

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)

	const funding int64 = 1e15
	deployer := crypto.AddressFromPreimage([]byte("test1"))
	unfunded := crypto.AddressFromPreimage([]byte("genesis-signer-with-no-balance"))

	appState := DefaultGenState()
	appState.Balances = []Balance{
		{Address: deployer, Amount: []std.Coin{{Amount: funding, Denom: "ugnot"}}},
	}
	appState.Txs = []TxWithMetadata{
		{Tx: std.Tx{
			Msgs: []std.Msg{vm.NewMsgAddPackage(deployer, "gno.land/r/demo", []*std.MemFile{
				{Name: "demo.gno", Body: "package demo; func Hello(cur realm) string { return `hello`; }"},
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/demo")},
			})},
			Fee:        std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
			Signatures: []std.Signature{{}},
		}},
		// Signed by an address with no balances entry, which is what triggers the mint.
		{Tx: std.Tx{
			Msgs:       []std.Msg{vm.NewMsgCall(unfunded, nil, "gno.land/r/demo", "Hello", nil)},
			Fee:        std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
			Signatures: []std.Signature{{}},
		}},
	}

	resp := bapp.InitChain(abci.RequestInitChain{
		Time:            time.Now(),
		ChainID:         "dev",
		ConsensusParams: &abci.ConsensusParams{Block: defaultBlockParams()},
		Validators:      []abci.ValidatorUpdate{},
		AppState:        appState,
	})
	require.True(t, resp.IsOK(), "InitChain response: %v", resp)
	require.NotNil(t, bapp.Commit())

	// The mint must be reflected in the counter. Fees only move coins to the
	// collector, so the total is the funded balance plus exactly what was minted.
	qres := bapp.Query(abci.RequestQuery{Path: "bank/supply/ugnot"})
	require.True(t, qres.IsOK(), "supply query: %v", qres)
	require.Equal(t, strconv.Quote(strconv.FormatInt(funding+genesisSignerFunding, 10)),
		string(qres.Data),
		"genesis auto-funding must mint, or the counter under-records what the chain holds")
}

// txCarriesCode selects which txs must have their signatures verified even on
// the simulate path. It must match both code-bearing messages and nothing else:
// too wide breaks keyless gas estimation for ordinary messages, too narrow
// leaves `.app/simulate` able to compile and run caller-supplied source under
// an address the caller does not control.
func TestTxCarriesCode(t *testing.T) {
	t.Parallel()

	addr := crypto.Address{}
	tests := []struct {
		name string
		msgs []std.Msg
		want bool
	}{
		{"no messages", nil, false},
		{"run alone", []std.Msg{vm.MsgRun{Caller: addr}}, true},
		{"add_package alone", []std.Msg{vm.MsgAddPackage{Creator: addr}}, true},
		{"call alone", []std.Msg{vm.MsgCall{Caller: addr}}, false},
		// A run hidden behind another message must still be found: the
		// predicate is asked once for the whole tx.
		{"run after a call", []std.Msg{vm.MsgCall{Caller: addr}, vm.MsgRun{Caller: addr}}, true},
		{"run before a call", []std.Msg{vm.MsgRun{Caller: addr}, vm.MsgCall{Caller: addr}}, true},
		// enable_package is authorized from msg.Approver -- a caller-supplied
		// field -- and type-checks and init()s a parked package, so an
		// unverified simulate lets anyone name the real approver and have that
		// work done for free. Its omission here was a live hole, reproduced end
		// to end against a running app.
		{"enable_package alone", []std.Msg{vm.MsgEnablePackage{Approver: addr}}, true},
		{"disable_package alone", []std.Msg{vm.MsgDisablePackage{Approver: addr}}, true},
		{"enable_package behind a call", []std.Msg{vm.MsgCall{Caller: addr}, vm.MsgEnablePackage{Approver: addr}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, txCarriesCode(std.Tx{Msgs: tt.msgs}))
		})
	}
}

// txCodeMsgSigners is the single scan both code gates key off, so a message
// kind it fails to report is a gate that silently does not run for that kind.
// This extends TestTxHasMsgRun above to the add_package axis; that test now
// covers the MsgRun projection of the same function.
//
// It asserts the SIGNERS rather than mere presence, because the granularity is
// the part that is easy to get wrong: authorization belongs to the signer of
// each code message, not to every signer of the transaction.
func TestTxCodeMsgSigners(t *testing.T) {
	t.Parallel()

	alice := crypto.AddressFromPreimage([]byte("alice"))
	bob := crypto.AddressFromPreimage([]byte("bob"))

	tests := []struct {
		name           string
		msgs           []std.Msg
		wantAddPkg     []crypto.Address
		wantRunSigners []crypto.Address
	}{
		{"no messages", nil, nil, nil},
		{
			"add_package alone",
			[]std.Msg{vm.MsgAddPackage{Creator: alice}},
			[]crypto.Address{alice}, nil,
		},
		{
			"run alone",
			[]std.Msg{vm.MsgRun{Caller: alice}},
			nil, []crypto.Address{alice},
		},
		{
			// Both in one tx is the case that separates the two rules: under
			// "inert" add_package is open while run stays gated, so conflating
			// them would let a bundled MsgRun through.
			"both, different signers",
			[]std.Msg{
				vm.MsgAddPackage{Creator: alice},
				vm.MsgRun{Caller: bob},
			},
			[]crypto.Address{alice}, []crypto.Address{bob},
		},
		{
			// MsgCall names a package but carries no source, so it must not be
			// classified as code-bearing.
			"call alone",
			[]std.Msg{vm.MsgCall{Caller: alice}},
			nil, nil,
		},
		{
			// A bystander on a non-code message must not appear at all: that is
			// what keeps them from needing code-submission rights.
			"bank send bystander is not collected",
			[]std.Msg{
				bank.MsgSend{FromAddress: bob, ToAddress: alice},
				vm.MsgRun{Caller: alice},
			},
			nil, []crypto.Address{alice},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			addPkg, run := txCodeMsgSigners(std.Tx{Msgs: tt.msgs})
			assert.Equal(t, tt.wantAddPkg, addPkg, "add_package signers")
			assert.Equal(t, tt.wantRunSigners, run, "run signers")
		})
	}
}

// TestSimulateVerifiesSignaturesOnCodeBearingTx pins that gno.land actually
// WIRES txCarriesCode into the ante, not merely that the predicate and the ante
// mechanism each work on their own.
//
// Both halves were already covered and the connection between them was not.
// TestTxCarriesCode checks the predicate; the auth package checks that the ante
// honours RequireSigForSimulate. Deleting `RequireSigForSimulate: txCarriesCode`
// from NewAppWithOptions broke neither -- verified by mutation.
//
// What that line buys: `.app/simulate` is a public query that EXECUTES the
// messages it is given. Skipping signature verification there is safe only for
// messages whose authorization does not depend on who signed. MsgRun and
// MsgAddPackage are gated on their signer, so without verification an
// unauthenticated caller could name a listed address, attach arbitrary bytes as
// a signature, and drive a full type-check and init() per query. The soundness
// argument in checkCodePolicy rests on this line.
func TestSimulateVerifiesSignaturesOnCodeBearingTx(t *testing.T) {
	t.Parallel()

	const chainID = "test-chain"
	key := getDummyKeys(t, 1)[0]
	addr := key.PubKey().Address()

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	app.InitChain(abci.RequestInitChain{
		ChainID: chainID,
		Time:    time.Now(),
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
		AppState: GnoGenesisState{
			Balances: []Balance{{
				Address: addr,
				Amount:  std.NewCoins(std.NewCoin("ugnot", 100_000_000)),
			}},
			Auth: auth.DefaultGenesisState(),
			Bank: bank.DefaultGenesisState(),
			VM:   vm.DefaultGenesisState(),
		},
	})
	app.Commit()

	// Advance one block before simulating, and do not remove this.
	//
	// The ante treats ctx.BlockHeight() == 0 as genesis, and at genesis with
	// SkipGenesisSigVerification -- which TestAppOptions sets -- it skips
	// signature checks entirely, before RequireSigForSimulate is ever consulted.
	// Simulating straight after InitChain therefore accepts any signature, so
	// this test would pass against a build with the wiring deleted.
	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{
		ChainID: chainID, Height: 1, Time: time.Now(),
	}})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// A correctly shaped transaction whose signature is garbage. Shape matters:
	// a missing or miscounted signature is refused in every mode by
	// tx.ValidateBasic, which would make this pass for the wrong reason.
	// The refusal lands in the transaction result, which .app/simulate returns
	// amino-marshalled inside the query response Value. ResponseQuery.Error is
	// only for a failure of the query itself, so asserting on it would pass for
	// the wrong reason -- it is nil here even when the ante refuses.
	simulate := func(t *testing.T, msgs []std.Msg) abci.ResponseDeliverTx {
		t.Helper()
		tx := createAndSignTxWithAccSeq(t, msgs, chainID, key, 0, 0)
		require.Len(t, tx.Signatures, 1)
		tx.Signatures[0].Signature = []byte("not a real signature at all")
		raw, err := amino.Marshal(tx)
		require.NoError(t, err)
		qres := app.Query(abci.RequestQuery{Path: ".app/simulate", Data: raw})
		require.Nil(t, qres.Error, "the query itself must succeed: %v", qres.Error)
		var dres abci.ResponseDeliverTx
		require.NoError(t, amino.Unmarshal(qres.Value, &dres))
		return dres
	}

	t.Run("a code-bearing message is refused", func(t *testing.T) {
		t.Parallel()

		res := simulate(t, []std.Msg{vm.MsgRun{
			Caller: addr,
			Package: &std.MemPackage{
				Name: "main",
				Path: "gno.land/e/" + addr.String() + "/run",
				Files: []*std.MemFile{
					{Name: "main.gno", Body: "package main\n\nfunc main() {}\n"},
				},
			},
		}})
		require.NotNil(t, res.Error,
			"simulate must verify signatures for a message authorized by its signer")
		assert.Contains(t, fmt.Sprintf("%v %s", res.Error, res.Log), "signature",
			"and must refuse on the signature, not on something incidental")
	})

	t.Run("an ordinary message still simulates without a valid signature", func(t *testing.T) {
		t.Parallel()

		// The other half of the predicate: too wide and keyless gas estimation
		// breaks for every ordinary message. MsgCall is not signer-authorized,
		// so it must still be estimable by a caller holding no key.
		res := simulate(t, []std.Msg{vm.MsgCall{
			Caller: addr, PkgPath: "gno.land/r/demo/absent", Func: "Nope",
		}})
		if res.Error != nil {
			assert.NotContains(t, fmt.Sprintf("%v %s", res.Error, res.Log), "signature",
				"an ordinary message must not be refused for its signature")
		}
	})
}

// A vesting account must also be whitelistable against the token lock. The two
// are meant to be combined -- see docs/CONSTITUTION.md, "Whitelisted funds
// remain subject to the vesting schedule" -- and the combination used to abort
// genesis, because applyBalance built a bare std.BaseAccount subtype while
// applyUnrestrictedAddrs asserts every genesis account to *GnoAccount.
func TestInitChainer_VestingAccountCanBeWhitelisted(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	mainKey := store.NewStoreKey("mainKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(), &bft.Header{ChainID: "test"}, log.NewNoopLogger())

	prmk := params.NewParamsKeeper(mainKey)
	acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, std.ProtoBaseSessionAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName), mainKey, []string{ugnot.Denom})
	prmk.Register(auth.ModuleName, acck)
	prmk.Register(bank.ModuleName, bankk)
	cfg := InitChainerConfig{acck: acck, bankk: bankk}

	addr := crypto.AddressFromPreimage([]byte("vested-investor"))
	amount := std.Coins{{Denom: ugnot.Denom, Amount: 1000}}
	cfg.applyBalance(ctx, Balance{Address: addr, Amount: amount, Vesting: &std.VestingSchedule{
		OriginalVesting: amount, StartTime: 100, EndTime: 200,
	}})

	// The panic this guards against happened here.
	require.NotPanics(t, func() {
		cfg.applyUnrestrictedAddrs(ctx, []crypto.Address{addr})
	}, "a vesting account must be whitelistable")

	acc := acck.GetAccount(ctx, addr)
	unrestricter, ok := acc.(std.AccountUnrestricter)
	require.True(t, ok, "a vesting account must carry the token-lock attributes")
	assert.True(t, unrestricter.IsTokenLockWhitelisted(), "the whitelist bit must have been set")

	// Both rules still apply: whitelisted against the token lock, and still
	// locked by the schedule. Whitelisting must not have cleared it.
	assert.False(t, acc.GetVesting().IsZero(), "whitelisting must not clear the schedule")
	locked := acc.LockedCoins(time.Unix(150, 0))
	assert.Equal(t, int64(500), locked.AmountOf(ugnot.Denom),
		"halfway through, half must still be locked")
}

// applyBalance is the runtime path that builds accounts, and it holds the only
// guards on a genesis schedule. Balance.Verify checks the same two things, but it
// runs only under `gnogenesis verify`, which an operator can skip; these must
// abort the boot on their own.
func TestInitChainer_RejectsABadVestingSchedule(t *testing.T) {
	t.Parallel()

	amount := std.Coins{{Denom: ugnot.Denom, Amount: 1000}}

	tests := []struct {
		name    string
		vesting *std.VestingSchedule
		want    string
	}{
		{
			name: "vesting more than the balance",
			vesting: &std.VestingSchedule{
				OriginalVesting: std.Coins{{Denom: ugnot.Denom, Amount: 5000}},
				StartTime:       100,
				EndTime:         200,
			},
			want: "exceeds the balance",
		},
		{
			name: "vesting a denom the balance does not hold",
			vesting: &std.VestingSchedule{
				OriginalVesting: std.Coins{{Denom: "atom", Amount: 1}},
				StartTime:       100,
				EndTime:         200,
			},
			want: "exceeds the balance",
		},
		{
			name: "linear vesting that starts after it ends",
			vesting: &std.VestingSchedule{
				OriginalVesting: amount,
				StartTime:       300,
				EndTime:         100,
			},
			want: "invalid vesting schedule",
		},
		{
			name: "an end time of zero",
			vesting: &std.VestingSchedule{
				OriginalVesting: amount,
				Type:            std.VestingDelayed,
			},
			want: "invalid vesting schedule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := memdb.NewMemDB()
			mainKey := store.NewStoreKey("mainKey")
			ms := store.NewCommitMultiStore(db)
			ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
			require.NoError(t, ms.LoadLatestVersion())
			ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms.MultiCacheWrap(), &bft.Header{ChainID: "test"}, log.NewNoopLogger())

			prmk := params.NewParamsKeeper(mainKey)
			acck := auth.NewAccountKeeper(mainKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount, std.ProtoBaseSessionAccount)
			bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName), mainKey, []string{ugnot.Denom})
			prmk.Register(auth.ModuleName, acck)
			prmk.Register(bank.ModuleName, bankk)
			cfg := InitChainerConfig{acck: acck, bankk: bankk}

			addr := crypto.AddressFromPreimage([]byte("bad-vester"))
			bal := Balance{Address: addr, Amount: amount, Vesting: tt.vesting}

			defer func() {
				r := recover()
				require.NotNil(t, r, "a bad schedule must abort genesis")
				assert.Contains(t, fmt.Sprint(r), tt.want)
				// Nothing may be left behind by a boot that aborted.
				assert.Nil(t, acck.GetAccount(ctx, addr),
					"the account must not have been written")
			}()
			cfg.applyBalance(ctx, bal)
		})
	}
}
