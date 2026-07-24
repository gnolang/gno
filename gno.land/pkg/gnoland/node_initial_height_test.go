package gnoland

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
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

	// Record the height carried by the first NewBlock event. n.Ready() only
	// reports that some block arrived, and the node keeps producing blocks
	// after that, so reading the block store once Ready() fires races with the
	// next commit.
	firstHeight := make(chan int64, 1)
	var once sync.Once
	n.EventSwitch().AddListener("first_block_height", func(ev events.Event) {
		if nb, ok := ev.(bft.EventNewBlock); ok {
			once.Do(func() { firstHeight <- nb.Block.Height })
		}
	})

	require.NoError(t, n.Start())
	t.Cleanup(func() { require.NoError(t, n.Stop()) })

	var height int64
	select {
	case height = <-firstHeight:
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for node to produce first block")
	}

	require.Equal(t, initialHeight, height,
		"first committed block should be at InitialHeight (%d), got %d", initialHeight, height)
}
