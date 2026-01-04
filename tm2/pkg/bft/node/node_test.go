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
