package node

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	sserver "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/random"
)

func TestNodeStartStop(t *testing.T) {
	config, genesisFile := cfg.ResetTestRoot("node_node_test")
	defer os.RemoveAll(config.RootDir)

	// create & start node
	n, err := DefaultNewNode(config, genesisFile, events.NewEventSwitch(), log.NewNoopLogger())
	require.NoError(t, err)
	err = n.Start()
	require.NoError(t, err)

	// wait for the node to produce a block
	blocksSub := events.SubscribeToEvent(n.EventSwitch(), "node_test", types.EventNewBlock{})
	require.NoError(t, err)
	select {
	case _, ok := <-blocksSub:
		if !ok {
			t.Fatal("blocksSub was cancelled")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the node to produce a block")
	}

	// stop the node
	go func() {
		n.Stop()
	}()

	select {
	case <-n.Quit():
	case <-time.After(5 * time.Second):
		pid := os.Getpid()
		p, err := os.FindProcess(pid)
		if err != nil {
			panic(err)
		}
		err = p.Signal(syscall.SIGABRT)
		fmt.Println(err)
		t.Fatal("timed out waiting for shutdown")
	}
}

// TestDefaultNewNodeWithGenesisProvider verifies that a custom GenesisDocProvider
// is used in place of the on-disk loader and that its returned doc reaches the
// node. This is the seam gnoland uses to inject the streaming genesis loader.
func TestDefaultNewNodeWithGenesisProvider(t *testing.T) {
	config, genesisFile := cfg.ResetTestRoot("node_genesis_provider_test")
	defer os.RemoveAll(config.RootDir)

	const sentinel = "from-custom-provider"
	calls := 0
	provider := func() (*types.GenesisDoc, error) {
		calls++
		doc, err := types.GenesisDocFromFile(genesisFile)
		if err != nil {
			return nil, err
		}
		doc.AppState = sentinel
		return doc, nil
	}

	n, err := DefaultNewNodeWithGenesisProvider(
		config,
		provider,
		events.NewEventSwitch(),
		log.NewNoopLogger(),
	)
	require.NoError(t, err)
	require.Equal(t, 1, calls, "custom genesis provider must be called exactly once")
	require.Equal(t, sentinel, n.GenesisDoc().AppState,
		"node must carry the AppState produced by the custom provider")
}

// TestSaveGenesisDoc_OmitsAppState verifies that the persisted genesis doc in
// the state DB drops AppState. AppState is only consumed at appBlockHeight==0
// (replay.go ReplayBlocks); the source genesis file is the canonical
// reference, and the DB copy is an audit trail for chainID / validators /
// genesis_time / app_hash. Persisting AppState would (a) panic for
// non-amino-registered types, and (b) bloat the state DB by hundreds of MB
// on real-world genesis files.
func TestSaveGenesisDoc_OmitsAppState(t *testing.T) {
	db := memdb.NewMemDB()

	original := &types.GenesisDoc{
		GenesisTime: tmtime.Now(),
		ChainID:     "test-chain-omit-appstate",
		AppHash:     []byte{0xAB, 0xCD},
		AppState:    types.MockAppState{AccountOwner: "Alice"},
	}

	saveGenesisDoc(db, original)

	loaded, err := loadGenesisDoc(db)
	require.NoError(t, err)
	require.Equal(t, original.ChainID, loaded.ChainID)
	require.Equal(t, original.AppHash, loaded.AppHash)
	require.Nil(t, loaded.AppState, "AppState must be omitted from the persisted doc")
}

// TestLoadStateFromDBOrGenesisDocProvider_ReinvokesProviderForAppState verifies
// that the provider is called again when the DB-loaded doc has nil AppState,
// so streaming providers can re-attach a fresh handle (e.g. *GenesisStateRef)
// on each boot rather than amino-marshaling it into the state DB.
func TestLoadStateFromDBOrGenesisDocProvider_ReinvokesProviderForAppState(t *testing.T) {
	db := memdb.NewMemDB()

	calls := 0
	provider := func() (*types.GenesisDoc, error) {
		calls++
		return &types.GenesisDoc{
			GenesisTime: tmtime.Now(),
			ChainID:     "test-chain-reinvoke",
			Validators: []types.GenesisValidator{{
				PubKey: ed25519.GenPrivKey().PubKey(),
				Power:  10,
			}},
			AppHash:  []byte{},
			AppState: types.MockAppState{AccountOwner: fmt.Sprintf("call-%d", calls)},
		}, nil
	}

	_, doc1, err := LoadStateFromDBOrGenesisDocProvider(db, provider)
	require.NoError(t, err)
	require.Equal(t, 1, calls)
	require.Equal(t, types.MockAppState{AccountOwner: "call-1"}, doc1.AppState)

	_, doc2, err := LoadStateFromDBOrGenesisDocProvider(db, provider)
	require.NoError(t, err)
	require.Equal(t, 2, calls,
		"provider must be re-invoked when the DB-loaded doc has nil AppState")
	require.Equal(t, types.MockAppState{AccountOwner: "call-2"}, doc2.AppState)
	require.Equal(t, doc1.ChainID, doc2.ChainID,
		"non-AppState fields must come from the persisted doc")
}

// TestLoadStateFromDBOrGenesisDocProvider_RejectsChainIDMismatchOnReinvoke
// guards the case where an operator swaps the source genesis.json between
// boots. The DB-persisted doc has chain A's metadata; the provider re-
// invocation reads chain B's genesis. Without a freshness check, the
// streaming AppState (chain B) would be paired with chain A's validators
// and AppHash — silent data corruption at appBlockHeight==0. The fix
// must refuse with a clear error.
func TestLoadStateFromDBOrGenesisDocProvider_RejectsChainIDMismatchOnReinvoke(t *testing.T) {
	db := memdb.NewMemDB()

	originalChain := &types.GenesisDoc{
		GenesisTime: tmtime.Now(),
		ChainID:     "chain-A",
		Validators: []types.GenesisValidator{{
			PubKey: ed25519.GenPrivKey().PubKey(),
			Power:  10,
		}},
		AppHash:  []byte{0xaa},
		AppState: types.MockAppState{AccountOwner: "boot-1"},
	}
	swapped := *originalChain
	swapped.ChainID = "chain-B" // simulates operator pointing the node at a different genesis.json

	calls := 0
	provider := func() (*types.GenesisDoc, error) {
		calls++
		if calls == 1 {
			return originalChain, nil
		}
		clone := swapped
		return &clone, nil
	}

	_, _, err := LoadStateFromDBOrGenesisDocProvider(db, provider)
	require.NoError(t, err)

	_, _, err = LoadStateFromDBOrGenesisDocProvider(db, provider)
	require.Error(t, err, "second boot with mismatched chain id must be rejected")
	require.Contains(t, err.Error(), "chain id",
		"error must explain that chain id changed between boots: %v", err)
}

// TestLoadStateFromDBOrGenesisDocProvider_RejectsAppHashMismatchOnReinvoke
// guards a subtler operator-error case: same chain id, different app_hash.
// Could happen if the genesis.json on disk drifts (e.g., a balance row was
// added). The DB still has the original app_hash; pairing fresh app_state
// with the persisted app_hash would corrupt InitChain. Refuse with a
// clear error.
func TestLoadStateFromDBOrGenesisDocProvider_RejectsAppHashMismatchOnReinvoke(t *testing.T) {
	db := memdb.NewMemDB()

	original := &types.GenesisDoc{
		GenesisTime: tmtime.Now(),
		ChainID:     "stable-chain",
		Validators: []types.GenesisValidator{{
			PubKey: ed25519.GenPrivKey().PubKey(),
			Power:  10,
		}},
		AppHash:  []byte{0xaa, 0xbb},
		AppState: types.MockAppState{AccountOwner: "boot-1"},
	}
	tampered := *original
	tampered.AppHash = []byte{0xcc, 0xdd}

	calls := 0
	provider := func() (*types.GenesisDoc, error) {
		calls++
		if calls == 1 {
			return original, nil
		}
		clone := tampered
		return &clone, nil
	}

	_, _, err := LoadStateFromDBOrGenesisDocProvider(db, provider)
	require.NoError(t, err)

	_, _, err = LoadStateFromDBOrGenesisDocProvider(db, provider)
	require.Error(t, err, "second boot with mismatched app_hash must be rejected")
	require.Contains(t, err.Error(), "app_hash",
		"error must explain that app_hash changed between boots: %v", err)
}

// TestLoadStateFromDBOrGenesisDocProvider_SurfacesReinvokeError verifies
// that a provider failure during re-invocation surfaces cleanly to the
// caller rather than being absorbed (which would leave the node booting
// with a nil AppState — guaranteed to crash at InitChain replay).
func TestLoadStateFromDBOrGenesisDocProvider_SurfacesReinvokeError(t *testing.T) {
	db := memdb.NewMemDB()

	calls := 0
	provider := func() (*types.GenesisDoc, error) {
		calls++
		if calls == 1 {
			return &types.GenesisDoc{
				GenesisTime: tmtime.Now(),
				ChainID:     "err-chain",
				Validators: []types.GenesisValidator{{
					PubKey: ed25519.GenPrivKey().PubKey(),
					Power:  10,
				}},
				AppHash:  []byte{},
				AppState: types.MockAppState{AccountOwner: "boot-1"},
			}, nil
		}
		return nil, fmt.Errorf("synthetic provider failure on boot %d", calls)
	}

	_, _, err := LoadStateFromDBOrGenesisDocProvider(db, provider)
	require.NoError(t, err)

	_, _, err = LoadStateFromDBOrGenesisDocProvider(db, provider)
	require.Error(t, err, "provider failure on re-invoke must surface to caller")
	require.Contains(t, err.Error(), "synthetic provider failure")
}

func TestSplitAndTrimEmpty(t *testing.T) {
	testCases := []struct {
		s        string
		sep      string
		cutset   string
		expected []string
	}{
		{"a,b,c", ",", " ", []string{"a", "b", "c"}},
		{" a , b , c ", ",", " ", []string{"a", "b", "c"}},
		{" a, b, c ", ",", " ", []string{"a", "b", "c"}},
		{" a, ", ",", " ", []string{"a"}},
		{"   ", ",", " ", []string{}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, splitAndTrimEmpty(tc.s, tc.sep, tc.cutset), "%s", tc.s)
	}
}

func TestNodeDelayedStart(t *testing.T) {
	config, genesisFile := cfg.ResetTestRoot("node_delayed_start_test")
	defer os.RemoveAll(config.RootDir)
	now := tmtime.Now()

	// create & start node
	n, err := DefaultNewNode(config, genesisFile, events.NewEventSwitch(), log.NewTestingLogger(t))
	n.GenesisDoc().GenesisTime = now.Add(2 * time.Second)
	require.NoError(t, err)

	err = n.Start()
	require.NoError(t, err)
	defer n.Stop()

	startTime := tmtime.Now()
	assert.Equal(t, true, startTime.After(n.GenesisDoc().GenesisTime))
}

func TestNodeReady(t *testing.T) {
	config, genesisFile := cfg.ResetTestRoot("node_node_test")
	defer os.RemoveAll(config.RootDir)

	// Create & start node
	n, err := DefaultNewNode(config, genesisFile, events.NewEventSwitch(), log.NewTestingLogger(t))
	require.NoError(t, err)

	// Assert that blockstore has zero block before waiting for the first block
	require.Equal(t, int64(0), n.BlockStore().Height())

	// Assert that first block signal is not alreay received by calling Ready
	select {
	case <-n.Ready():
		require.FailNow(t, "first block signal should not be close before starting the node")
	default: // ok
	}

	err = n.Start()
	require.NoError(t, err)
	defer n.Stop()

	// Wait until the node is ready or timeout
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "timeout while waiting for first block signal")
	case <-n.Ready(): // ready
	}

	// Check that blockstore have at last one block
	require.GreaterOrEqual(t, n.BlockStore().Height(), int64(1))
}

func TestNodeSetAppVersion(t *testing.T) {
	config, genesisFile := cfg.ResetTestRoot("node_app_version_test")
	defer os.RemoveAll(config.RootDir)

	// create & start node
	n, err := DefaultNewNode(config, genesisFile, events.NewEventSwitch(), log.NewTestingLogger(t))
	require.NoError(t, err)

	// default config uses the kvstore app
	appVersion := kvstore.AppVersion

	// check version is set in state
	state := sm.LoadState(n.stateDB)
	assert.Equal(t, state.AppVersion, appVersion)

	// check version is set in node info
	appVersion2, ok := n.nodeInfo.VersionSet.Get("app")
	assert.True(t, ok)
	assert.Equal(t, appVersion2.Version, appVersion)
}

func TestNodeSetPrivValTCP(t *testing.T) {
	addr := "tcp://" + testFreeAddr(t)

	config, genesisFile := cfg.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.Consensus.PrivValidator.RemoteSigner.ServerAddress = addr

	signer := types.NewMockSigner()

	pvss, err := sserver.NewRemoteSignerServer(
		signer,
		addr,
		log.NewNoopLogger(),
	)
	require.NoError(t, err)

	go func() {
		err := pvss.Start()
		require.NoError(t, err)
	}()
	defer pvss.Stop()

	n, err := DefaultNewNode(config, genesisFile, events.NewEventSwitch(), log.NewTestingLogger(t))
	require.NotNil(t, n)
	require.NoError(t, err)

	privValPK := n.PrivValidator().PubKey()
	require.NotNil(t, privValPK)

	signerPK := signer.PubKey()
	require.NotNil(t, signerPK)

	require.Equal(t, signerPK, privValPK)
}

func TestNodeSetPrivValIPC(t *testing.T) {
	unixSocket := "unix:///tmp/kms." + random.RandStr(6) + ".sock"
	defer os.Remove(unixSocket) // clean up

	config, genesisFile := cfg.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.Consensus.PrivValidator.RemoteSigner.ServerAddress = unixSocket

	signer := types.NewMockSigner()

	pvss, err := sserver.NewRemoteSignerServer(
		signer,
		unixSocket,
		log.NewNoopLogger(),
	)
	require.NoError(t, err)

	go func() {
		err := pvss.Start()
		require.NoError(t, err)
	}()
	defer pvss.Stop()

	n, err := DefaultNewNode(config, genesisFile, events.NewEventSwitch(), log.NewTestingLogger(t))
	require.NotNil(t, n)
	require.NoError(t, err)

	privValPK := n.PrivValidator().PubKey()
	require.NotNil(t, privValPK)

	signerPK := signer.PubKey()
	require.NotNil(t, signerPK)

	require.Equal(t, signerPK, privValPK)
}

// testFreeAddr claims a free port so we don't block on listener being ready.
func testFreeAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	return fmt.Sprintf("127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)
}

// create a proposal block using real and full
// mempool pool and validate it.
func TestCreateProposalBlock(t *testing.T) {
	config, _ := cfg.ResetTestRoot("node_create_proposal")
	defer os.RemoveAll(config.RootDir)
	cc := proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication())
	proxyApp := appconn.NewAppConns(cc)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	logger := log.NewTestingLogger(t)

	var height int64 = 1
	state, stateDB := state(1, height)
	maxBlockBytes := 16384
	state.ConsensusParams.Block.MaxBlockBytes = int64(maxBlockBytes)
	proposerAddr, _ := state.Validators.GetByIndex(0)

	// Make Mempool
	mempool := mempl.NewCListMempool(
		config.Mempool,
		proxyApp.Mempool(),
		state.LastBlockHeight,
		state.ConsensusParams.Block.MaxTxBytes,
		mempl.WithPreCheck(sm.TxPreCheck(state)),
	)
	mempool.SetLogger(logger)

	// fill the mempool with more txs
	// than can fit in a block
	txLength := 1000
	for range maxBlockBytes / txLength {
		tx := random.RandBytes(txLength)
		err := mempool.CheckTx(tx, nil)
		assert.NoError(t, err)
	}

	blockExec := sm.NewBlockExecutor(
		stateDB,
		logger,
		proxyApp.Consensus(),
		mempool,
	)

	commit := types.NewCommit(types.BlockID{}, nil)
	block, _ := blockExec.CreateProposalBlock(
		height,
		state, commit,
		proposerAddr,
	)

	err = state.ValidateBlock(block)
	assert.NoError(t, err)
}

func state(nVals int, height int64) (sm.State, dbm.DB) {
	vals := make([]types.GenesisValidator, nVals)
	for i := range nVals {
		secret := fmt.Appendf(nil, "test%d", i)
		pk := ed25519.GenPrivKeyFromSecret(secret)
		vals[i] = types.GenesisValidator{
			Address: pk.PubKey().Address(),
			PubKey:  pk.PubKey(),
			Power:   1000,
			Name:    fmt.Sprintf("test%d", i),
		}
	}
	s, _ := sm.MakeGenesisState(&types.GenesisDoc{
		ChainID:    "test-chain",
		Validators: vals,
		AppHash:    nil,
	})

	// save validators to db for 2 heights
	stateDB := memdb.NewMemDB()
	sm.SaveState(stateDB, s)

	for i := 1; i < int(height); i++ {
		s.LastBlockHeight++
		s.LastValidators = s.Validators.Copy()
		sm.SaveState(stateDB, s)
	}
	return s, stateDB
}
