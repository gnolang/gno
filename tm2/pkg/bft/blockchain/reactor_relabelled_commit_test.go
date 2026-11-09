package blockchain

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	cs "github.com/gnolang/gno/tm2/pkg/bft/consensus"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

// relabelledChain is a single-validator chain of genuine blocks 1..n.
type relabelledChain struct {
	genDoc   *types.GenesisDoc
	privVals []types.PrivValidator
	state    sm.State
	stateDB  *memdb.MemDB
	bs       *store.BlockStore
	exec     *sm.BlockExecutor
	blocks   []*types.Block
	parts    []*types.PartSet
}

func buildRelabelledChain(t *testing.T, n int64) *relabelledChain {
	t.Helper()

	genDoc, privVals := randGenesisDoc(1, false, 30)
	proxyApp := appconn.NewAppConns(proxy.NewLocalClientCreator(&testApp{}))
	require.NoError(t, proxyApp.Start())
	t.Cleanup(func() { proxyApp.Stop() })

	state, err := sm.LoadStateFromDBOrGenesisDoc(memdb.NewMemDB(), genDoc)
	require.NoError(t, err)
	db := memdb.NewMemDB()
	sm.SaveState(db, state)
	ch := &relabelledChain{
		genDoc:   genDoc,
		privVals: privVals,
		stateDB:  db,
		bs:       store.NewBlockStore(memdb.NewMemDB()),
		exec:     sm.NewBlockExecutor(db, log.NewNoopLogger(), proxyApp.Consensus(), mock.Mempool{}),
	}
	for h := int64(1); h <= n; h++ {
		lastCommit := types.NewCommit(types.BlockID{}, nil)
		if h > 1 {
			lastCommit = ch.commit(t, state, h-1)
		}
		block := makeBlock(h, state, lastCommit)
		parts := block.MakePartSet(types.BlockPartSizeBytes)
		state, err = ch.exec.ApplyBlock(state, types.BlockID{Hash: block.Hash(), PartsHeader: parts.Header()}, block)
		require.NoError(t, err)
		ch.bs.SaveBlock(block, parts, lastCommit)
		ch.blocks = append(ch.blocks, block)
		ch.parts = append(ch.parts, parts)
	}
	ch.state = state
	return ch
}

// commit returns the genuine commit for the stored block at height h.
func (ch *relabelledChain) commit(t *testing.T, state sm.State, h int64) *types.Commit {
	t.Helper()

	meta := ch.bs.LoadBlockMeta(h)
	vote, err := types.MakeVote(h, meta.BlockID, state.Validators, ch.privVals[0], ch.genDoc.ChainID)
	require.NoError(t, err)
	return types.NewCommit(meta.BlockID, []*types.CommitSig{vote.CommitSig()})
}

// relabelled returns the genuine commit for height h with its precommit's
// unsigned ValidatorAddress rewritten: the signature still verifies.
func (ch *relabelledChain) relabelled(t *testing.T, h int64) *types.Commit {
	t.Helper()

	c := ch.commit(t, ch.state, h)
	c.Precommits[0].ValidatorAddress = crypto.Address(bytes.Repeat([]byte{0xAA}, 20))
	return c
}

// startConsensus runs NewConsensusState, which rebuilds LastCommit from the
// stored seen-commit, and returns what it panicked with, if anything.
func startConsensus(state sm.State, stateDB *memdb.MemDB, bs sm.BlockStore) (panicked any) {
	defer func() { panicked = recover() }()
	exec := sm.NewBlockExecutor(stateDB, log.NewNoopLogger(), nil, mock.Mempool{})
	_ = cs.NewConsensusState(cfg.TestConfig().Consensus, state, exec, bs, mock.Mempool{}, cs.NoOpEvidencePool{})
	return nil
}

// A fast-sync peer serves genuine blocks 1..H and a tip block whose
// LastCommit is block H's commit with a relabelled address. The syncing node
// must reject that commit rather than store it as H's seen-commit, where
// handing off to consensus would panic on it, and so would every restart.
func TestFastSyncRejectsRelabelledTipCommit(t *testing.T) {
	const H = 5
	config, _ = cfg.ResetTestRoot("fastsync_relabelled_commit")
	defer os.RemoveAll(config.RootDir)

	hostile := buildRelabelledChain(t, H)
	tip, tipParts := hostile.state.MakeBlock(H+1, makeTxs(H+1), hostile.relabelled(t, H), hostile.state.Validators.GetProposer().Address)
	hostile.bs.SaveBlock(tip, tipParts, hostile.relabelled(t, H))
	hostileState := hostile.state.Copy()
	hostileState.LastBlockHeight = H + 1
	hostileR := NewBlockchainReactor(hostileState, hostile.exec, hostile.bs, true, nil)
	hostileR.SetLogger(log.NewNoopLogger())

	proxyApp := appconn.NewAppConns(proxy.NewLocalClientCreator(&testApp{}))
	require.NoError(t, proxyApp.Start())
	defer proxyApp.Stop()
	victimStore := store.NewBlockStore(memdb.NewMemDB())
	victimState, err := sm.LoadStateFromDBOrGenesisDoc(memdb.NewMemDB(), hostile.genDoc)
	require.NoError(t, err)
	victimDB := memdb.NewMemDB()
	sm.SaveState(victimDB, victimState)
	victimExec := sm.NewBlockExecutor(victimDB, log.NewNoopLogger(), proxyApp.Consensus(), mock.Mempool{})

	var (
		mu       sync.Mutex
		handoff  bool
		panicked any
	)
	victimR := NewBlockchainReactor(victimState.Copy(), victimExec, victimStore, true, func(s sm.State, _ int) {
		// Stands in for SwitchToConsensus, which runs the same LastCommit
		// reconstruction on the stored seen-commit.
		p := startConsensus(s, victimDB, victimStore)
		mu.Lock()
		handoff, panicked = true, p
		mu.Unlock()
	})
	victimR.SetLogger(log.NewNoopLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	switches, _ := p2pTesting.MakeConnectedPeers(t, ctx, p2pTesting.TestingConfig{
		Count:  2,
		P2PCfg: config.P2P,
		SwitchOptions: map[int][]p2p.SwitchOption{
			0: {p2p.WithReactor("BLOCKCHAIN", hostileR)},
			1: {p2p.WithReactor("BLOCKCHAIN", victimR)},
		},
		Channels: []byte{BlockchainChannel},
	})
	defer hostileR.Stop()
	defer victimR.Stop()

	// Done once the victim hands off or drops the hostile peer.
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return handoff || switches[1].Peers().NumOutbound()+switches[1].Peers().NumInbound() == 0
	}, 8*time.Second, 50*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Nil(t, panicked, "consensus handoff panicked")
	assert.False(t, handoff, "the victim synced past a relabelled commit")
	assert.Equal(t, int64(H-1), victimStore.Height(), "block H stays unverified without a valid commit for it")
	assert.Nil(t, victimStore.LoadSeenCommit(H), "the relabelled commit was stored as a seen-commit")
}

// A node that already stored a relabelled seen-commit, as an unpatched fast
// sync would have, must still start consensus.
func TestConsensusStartsOnStoredRelabelledCommit(t *testing.T) {
	const H = 5
	config, _ = cfg.ResetTestRoot("restart_relabelled_commit")
	defer os.RemoveAll(config.RootDir)

	ch := buildRelabelledChain(t, H)
	bs := store.NewBlockStore(memdb.NewMemDB())
	for h := int64(1); h <= H; h++ {
		seen := ch.relabelled(t, H)
		if h < H {
			seen = ch.blocks[h].LastCommit
		}
		bs.SaveBlock(ch.blocks[h-1], ch.parts[h-1], seen)
	}
	for range 3 {
		assert.Nil(t, startConsensus(ch.state, ch.stateDB, bs))
	}
}
