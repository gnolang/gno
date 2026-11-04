package consensus

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/stretchr/testify/require"
)

// gossipLogWatcher counts records by message and reports the first want on seen.
type gossipLogWatcher struct {
	want  string
	seen  chan struct{}
	mtx   sync.Mutex
	count map[string]int
}

func newGossipLogWatcher(want string) *gossipLogWatcher {
	return &gossipLogWatcher{
		want:  want,
		seen:  make(chan struct{}, 1),
		count: make(map[string]int),
	}
}

func (w *gossipLogWatcher) counted(message string) int {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.count[message]
}

func (w *gossipLogWatcher) Enabled(context.Context, slog.Level) bool { return true }

func (w *gossipLogWatcher) Handle(_ context.Context, r slog.Record) error {
	w.mtx.Lock()
	w.count[r.Message]++
	w.mtx.Unlock()

	if r.Message == w.want {
		select {
		case w.seen <- struct{}{}:
		default:
		}
	}

	return nil
}

func (w *gossipLogWatcher) WithAttrs([]slog.Attr) slog.Handler { return w }

func (w *gossipLogWatcher) WithGroup(string) slog.Handler { return w }

// TestGossipRoutinesHeightBelowInitialHeight drives the per-peer gossip routines
// against a peer claiming a height the blockstore of a chain starting above 1
// never held, and holds each to one pass per PeerGossipSleepDuration.
func TestGossipRoutinesHeightBelowInitialHeight(t *testing.T) {
	t.Parallel()

	const (
		initialHeight = 100
		peerHeight    = 50
		// Every branch that reports the missing height sleeps for
		// PeerGossipSleepDuration, 5ms under TestConsensusConfig, so this window
		// admits about 40 records per branch. A branch that loops without
		// sleeping produced 1587544 in a second.
		rateWindow = 200 * time.Millisecond
		rateBound  = 1000
	)

	genDoc, privVals := randGenesisDoc(1, false, 30)
	genDoc.InitialHeight = initialHeight

	state, err := sm.MakeGenesisState(genDoc)
	require.NoError(t, err)

	thisConfig, _ := ResetConfig("consensus_gossip_data_test")
	defer os.RemoveAll(thisConfig.RootDir)

	cs := newConsensusStateWithConfig(thisConfig, state, privVals[0], kvstore.NewKVStoreApplication())

	require.Equal(t, int64(initialHeight), cs.GetRoundState().Height)
	require.Equal(t, int64(0), cs.blockStore.Height())

	const (
		missingMeta   = "Failed to load block meta for peer block parts"
		missingCommit = "Failed to load block commit for catchup"
	)

	watcher := newGossipLogWatcher(missingMeta)

	// fastSync leaves the consensus state stopped, so the gossip routines AddPeer
	// spawns are the only thing running.
	reactor := NewConsensusReactor(cs, true)
	reactor.SetLogger(slog.New(watcher))
	reactor.SetEventSwitch(cs.evsw)
	require.NoError(t, reactor.Start())
	defer reactor.Stop()

	peer := p2pTesting.NewPeer(t)
	defer peer.Stop()

	reactor.InitPeer(peer)

	// StateChannel is handled above the FastSync gate, and ValidateBasic weighs
	// the height against nothing local, so the peer state takes 50 as given.
	reactor.Receive(StateChannel, peer, amino.MustMarshalAny(&NewRoundStepMessage{
		Height:          peerHeight,
		Round:           0,
		Step:            cstypes.RoundStepNewHeight,
		LastCommitRound: -1,
	}))

	ps, ok := peer.Get(types.PeerStateKey).(*PeerState)
	require.True(t, ok)
	require.Equal(t, int64(peerHeight), ps.GetRoundState().Height)

	reactor.AddPeer(peer)

	select {
	case <-watcher.seen:
	case <-time.After(ensureTimeout):
		t.Fatalf("gossipDataRoutine reported no %q", missingMeta)
	}

	time.Sleep(rateWindow)
	require.Less(t, watcher.counted(missingMeta), rateBound)
	require.Less(t, watcher.counted(missingCommit), rateBound)
}
