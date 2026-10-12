package gnoland

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// TestNodeBootWithInitialHeight boots a full in-memory node whose genesis doc
// has InitialHeight = 100.  It verifies that:
//
//   - The node starts without panicking (exercises all the InitialHeight paths
//     through Handshaker → ConsensusState.reconstructLastCommit →
//     BlockchainReactor.NewBlockchainReactor).
//   - The first committed block is at height 100, not 1.
func TestNodeBootWithInitialHeight(t *testing.T) {
	const initialHeight = int64(100)

	td := t.TempDir()
	tmcfg := NewDefaultTMConfig(td)

	pv := bft.NewMockPV()
	pk := pv.PubKey()

	genesis := &bft.GenesisDoc{
		GenesisTime:   time.Now(),
		ChainID:       tmcfg.ChainID(),
		InitialHeight: initialHeight,
		ConsensusParams: abci.ConsensusParams{
			Block: defaultBlockParams(),
		},
		Validators: []bft.GenesisValidator{
			{
				Address: pk.Address(),
				PubKey:  pk,
				Power:   10,
				Name:    "self",
			},
		},
		AppState: DefaultGenState(),
	}

	cfg := &InMemoryNodeConfig{
		PrivValidator: pv,
		Genesis:       genesis,
		TMConfig:      tmcfg,
		DB:            memdb.NewMemDB(),
		InitChainerConfig: InitChainerConfig{
			GenesisTxResultHandler: PanicOnFailingTxResultHandler,
			StdlibDir:              filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs"),
			CacheStdlibLoad:        true,
		},
	}

	n, err := NewInMemoryNode(log.NewTestingLogger(t), cfg)
	require.NoError(t, err)

	require.NoError(t, n.Start())
	t.Cleanup(func() { require.NoError(t, n.Stop()) })

	select {
	case <-n.Ready():
		// first block committed
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for node to produce first block")
	}

	height := n.BlockStore().Height()
	require.Equal(t, initialHeight, height,
		"first committed block should be at InitialHeight (%d), got %d", initialHeight, height)
}

// A genesis the InitChainer refuses, its InitialHeight corrected and booted
// again on the same DB, must reach the app hash a fresh DB reaches.
func TestNodeRejectedGenesisRetryMatchesFreshBoot(t *testing.T) {
	pv := bft.NewMockPV()
	pk := pv.PubKey()
	genTime := time.Now()

	genesis := func(docInitialHeight int64) *bft.GenesisDoc {
		state := DefaultGenState()
		state.InitialHeight = 999
		return &bft.GenesisDoc{
			GenesisTime:     genTime,
			ChainID:         "retry-test",
			InitialHeight:   docInitialHeight,
			ConsensusParams: abci.ConsensusParams{Block: defaultBlockParams()},
			Validators: []bft.GenesisValidator{
				{Address: pk.Address(), PubKey: pk, Power: 10, Name: "self"},
			},
			AppState: state,
		}
	}

	// boot returns the app hash once block 999 is committed.
	boot := func(db *memdb.MemDB, gen *bft.GenesisDoc) ([]byte, error) {
		tmcfg := NewDefaultTMConfig(t.TempDir())
		tmcfg.Consensus.CreateEmptyBlocks = false
		n, err := NewInMemoryNode(log.NewNoopLogger(), &InMemoryNodeConfig{
			PrivValidator: pv,
			Genesis:       gen,
			TMConfig:      tmcfg,
			DB:            db,
			InitChainerConfig: InitChainerConfig{
				GenesisTxResultHandler: PanicOnFailingTxResultHandler,
				StdlibDir:              filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs"),
				CacheStdlibLoad:        true,
			},
		})
		if err != nil {
			return nil, err
		}
		require.NoError(t, n.Start())
		defer n.Stop()
		var appHash []byte
		require.Eventually(t, func() bool {
			state := sm.LoadState(db)
			appHash = state.AppHash
			return state.LastBlockHeight >= 999
		}, 60*time.Second, 100*time.Millisecond, "block 999 was never committed")
		return appHash, nil
	}

	shared := memdb.NewMemDB()
	_, err := boot(shared, genesis(100))
	require.ErrorContains(t, err, "InitialHeight mismatch")

	retried, err := boot(shared, genesis(999))
	require.NoError(t, err, "the corrected genesis must boot on the same DB")

	fresh, err := boot(memdb.NewMemDB(), genesis(999))
	require.NoError(t, err)
	require.Equal(t, fresh, retried, "the refused boot must leave nothing behind")
}

// AppState must be a GnoGenesisState value: loadAppState rejects a
// *GnoGenesisState as an invalid AppState.
func TestNewDefaultGenesisConfig_AppStateType(t *testing.T) {
	t.Parallel()

	genesis := NewDefaultGenesisConfig("test-chain", "gno.land")

	_, ok := genesis.AppState.(GnoGenesisState)
	require.True(t, ok, "AppState must be a GnoGenesisState value, got %T", genesis.AppState)
}
