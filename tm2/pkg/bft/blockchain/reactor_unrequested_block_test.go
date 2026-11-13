package blockchain

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/mock"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/require"
)

// stopRecorder is a Switch that records the peers stopped for an error. The
// reactor is never started, so only the test's own Receive calls reach it.
type stopRecorder struct {
	p2p.Switch

	stopped []p2pTypes.ID
}

func (s *stopRecorder) StopPeerForError(peer p2p.PeerConn, _ error) {
	s.stopped = append(s.stopped, peer.ID())
}

// reported drains the pool's reports, leaving out the timeouts of peers that
// never answer.
func reported(errorsCh chan peerError) (ids []p2pTypes.ID) {
	for {
		select {
		case e := <-errorsCh:
			if e.err.Error() != "peer did not send us anything" {
				ids = append(ids, e.peerID)
			}
		default:
			return ids
		}
	}
}

// A peer sending a block the pool did not ask it for must cost the node
// neither the decode of that block nor a hash of its body, whose work grows
// with the number of txs the peer packed into it. Receive stops the peer of a
// message that fails to decode or to validate, while the pool reports a peer
// through its errors channel.
func TestReceiveSkipsUnrequestedBlock(t *testing.T) {
	t.Parallel()

	requestsCh := make(chan BlockRequest, maxTotalRequesters)
	errorsCh := make(chan peerError, 1000)
	bcR := &BlockchainReactor{pool: NewBlockPool(1, requestsCh, errorsCh)}
	bcR.BaseReactor = *p2p.NewBaseReactor("BlockchainReactor", bcR)
	sw := &stopRecorder{}
	bcR.SetSwitch(sw)
	require.NoError(t, bcR.pool.Start())

	peers := mock.GeneratePeers(t, 4)
	first, second, idle, stranger := peers[0], peers[1], peers[2], peers[3]

	// The pool asks the first peer for heights 1 to 3 and the second for 4 to
	// 6, the only heights it reported above the first's. The idle peer is in
	// the pool but too low to be asked anything.
	bcR.pool.SetPeerHeight(first.ID(), 3)
	require.Eventually(t, func() bool { return len(requestsCh) == 3 }, 5*time.Second, 10*time.Millisecond)
	bcR.pool.SetPeerHeight(second.ID(), 6)
	require.Eventually(t, func() bool { return len(requestsCh) == 6 }, 5*time.Second, 10*time.Millisecond)
	bcR.pool.SetPeerHeight(idle.ID(), 0)

	// One byte over the 64-byte gate, written out so that raising the gate
	// fails the test.
	garbage := bytes.Repeat([]byte{0xff}, 65)
	invalidBlock := func(height int64) []byte {
		return amino.MustMarshalAny(&bcBlockResponseMessage{Block: &types.Block{Header: types.Header{Height: height}}})
	}
	// A block whose header passes ValidateHeaderBasic.
	validHeader := func(height int64) []byte {
		return amino.MustMarshalAny(&bcBlockResponseMessage{Block: types.MakeBlock(height, nil, types.NewCommit(types.BlockID{}, nil))})
	}

	bcR.Receive(BlockchainChannel, stranger, garbage)
	bcR.Receive(BlockchainChannel, idle, garbage)
	require.Empty(t, sw.stopped, "a large message from a peer asked nothing was decoded")
	require.Empty(t, reported(errorsCh))

	// A block for a height the pool asked another peer for, asked nobody for,
	// or already holds, is reported once its header checks out.
	bcR.pool.AddBlock(second.ID(), &types.Block{Header: types.Header{Height: 5}}, 1)
	bcR.Receive(BlockchainChannel, first, validHeader(4))
	bcR.Receive(BlockchainChannel, first, validHeader(50))
	bcR.Receive(BlockchainChannel, second, validHeader(5))
	bcR.Receive(BlockchainChannel, second, validHeader(1))
	require.Empty(t, sw.stopped)
	require.Equal(t, []p2pTypes.ID{first.ID(), first.ID(), second.ID(), second.ID()}, reported(errorsCh))

	// AddBlock reports such a block from a peer the pool waits on for nothing
	// too.
	bcR.pool.AddBlock(idle.ID(), &types.Block{Header: types.Header{Height: 50}}, 1)
	require.Equal(t, []p2pTypes.ID{idle.ID()}, reported(errorsCh))

	// A block the pool asked for is checked against its header alone, so a
	// body its DataHash does not describe is left for the pool routine.
	hash := bytes.Repeat([]byte{1}, tmhash.Size)
	lastID := types.BlockID{Hash: hash, PartsHeader: types.PartSetHeader{Total: 1, Hash: hash}}
	unhashed := types.MakeBlock(3, []types.Tx{types.Tx("tx")}, types.NewCommit(lastID, []*types.CommitSig{nil}))
	unhashed.LastBlockID = lastID
	unhashed.DataHash = hash
	bcR.Receive(BlockchainChannel, first, amino.MustMarshalAny(&bcBlockResponseMessage{Block: unhashed}))
	require.Empty(t, sw.stopped, "a requested block's body was checked against its DataHash before its header was signed")
	require.Empty(t, reported(errorsCh))

	// Every decoded block is checked against its header, asked for or not, and
	// the peer of one that fails is stopped rather than reported.
	bcR.Receive(BlockchainChannel, first, invalidBlock(1))
	bcR.Receive(BlockchainChannel, first, invalidBlock(50))
	bcR.Receive(BlockchainChannel, second, invalidBlock(0))
	bcR.Receive(BlockchainChannel, second, invalidBlock(-1))
	require.Equal(t, []p2pTypes.ID{first.ID(), first.ID(), second.ID(), second.ID()}, sw.stopped)
	require.Empty(t, reported(errorsCh))
	sw.stopped = nil

	// A response carrying no block is small enough to decode from anyone and
	// still fails validation.
	bcR.Receive(BlockchainChannel, stranger, amino.MustMarshalAny(&bcBlockResponseMessage{}))
	require.Equal(t, []p2pTypes.ID{stranger.ID()}, sw.stopped)
	sw.stopped = nil

	// Once the pool stops, as it does when the node switches to consensus,
	// nothing large is decoded.
	bcR.pool.Stop()
	bcR.Receive(BlockchainChannel, second, garbage)
	require.Empty(t, sw.stopped)
}

// A requester is assigned in the pick that counts it against its peer, so no
// removePeer can run between the two and leave it waiting on a peer the pool
// no longer waits on, whose answer Receive would then drop undecoded.
func TestPickAssignsRequester(t *testing.T) {
	t.Parallel()

	pool := NewBlockPool(1, make(chan BlockRequest, 1), make(chan peerError, 1))
	pool.SetPeerHeight("peer", 3)
	bpr := newBPRequester(pool, 1)
	pool.requesters[1] = bpr

	peer := pool.pickIncrAvailablePeer(bpr)
	require.NotNil(t, peer)
	require.Equal(t, peer.id, bpr.getPeerID())
	require.Equal(t, int32(1), peer.numPending)
}

// A requester popped while its pick waited for the pool's lock is not counted
// against a peer, since no pop is left to give that count back.
func TestPickSkipsPoppedRequester(t *testing.T) {
	t.Parallel()

	pool := NewBlockPool(1, make(chan BlockRequest, 1), make(chan peerError, 1))
	pool.SetPeerHeight("peer", 3)

	require.Nil(t, pool.pickIncrAvailablePeer(newBPRequester(pool, 1)))
	require.Zero(t, pool.peers["peer"].numPending)
}

// A peer re-asked for a block whose first deliverer dropped while the pool
// verified it answers after the pop, for a height below pool.height. It sent
// what it was asked for, so it is not reported.
func TestAnswerAfterPopIsNotReported(t *testing.T) {
	t.Parallel()

	requestsCh := make(chan BlockRequest, maxTotalRequesters)
	errorsCh := make(chan peerError, 1000)
	pool := NewBlockPool(1, requestsCh, errorsCh)
	require.NoError(t, pool.Start())
	t.Cleanup(func() { pool.Stop() })

	const p, q = p2pTypes.ID("p"), p2pTypes.ID("q")
	pool.SetPeerHeight(p, 2)
	require.Eventually(t, func() bool { return len(requestsCh) == 2 }, 5*time.Second, time.Millisecond)
	pool.AddBlock(p, &types.Block{Header: types.Header{Height: 1}}, 1)
	pool.AddBlock(p, &types.Block{Header: types.Header{Height: 2}}, 1)
	pool.SetPeerHeight(q, 2)

	// p drops between the peek and the pop, so q is re-asked for height 1.
	first, second := pool.PeekTwoBlocks()
	require.NotNil(t, first)
	require.NotNil(t, second)
	pool.RemovePeer(p)
	require.Eventually(t, func() bool {
		pool.mtx.Lock()
		defer pool.mtx.Unlock()
		return pool.requesters[1].getPeerID() == q && pool.requesters[2].getPeerID() == q
	}, 5*time.Second, time.Millisecond)
	pending := func() int32 {
		pool.mtx.Lock()
		defer pool.mtx.Unlock()
		return pool.peers[q].numPending
	}
	before := pending()
	_, poolBefore, _ := pool.GetStatus()
	pool.PopRequest()
	require.Equal(t, before-1, pending(), "the popped request still counts against q, which then times out")
	_, poolAfter, _ := pool.GetStatus()
	require.Equal(t, poolBefore-1, poolAfter, "the popped request still counts against the pool")

	// Only q was re-asked for height 1, and only once.
	const r = p2pTypes.ID("r")
	pool.SetPeerHeight(r, 0)
	pool.AddBlock(r, &types.Block{Header: types.Header{Height: 1}}, 1)
	require.Equal(t, []p2pTypes.ID{r}, reported(errorsCh))
	pool.AddBlock(q, &types.Block{Header: types.Header{Height: 1}}, 1)
	require.Empty(t, reported(errorsCh))
	pool.AddBlock(q, &types.Block{Header: types.Header{Height: 1}}, 1)
	require.Equal(t, []p2pTypes.ID{q}, reported(errorsCh))
}

// A requester popped after a redo cleared its peer and before a new pick was
// re-asked to nobody, so its pop marks no peer.
func TestPopWithoutPeer(t *testing.T) {
	t.Parallel()

	pool := NewBlockPool(1, make(chan BlockRequest, 1), make(chan peerError, 1))
	pool.requesters[1] = newBPRequester(pool, 1)
	require.NotPanics(t, pool.PopRequest)
}

// A peer the pool waits on is reported for a block below the pool's height
// that it was not re-asked for, so it cannot keep sending stuffed ones there,
// each decoded, until it times out.
func TestBlockBelowPoolHeightIsReported(t *testing.T) {
	t.Parallel()

	requestsCh := make(chan BlockRequest, maxTotalRequesters)
	errorsCh := make(chan peerError, 1000)
	pool := NewBlockPool(1, requestsCh, errorsCh)
	require.NoError(t, pool.Start())
	t.Cleanup(func() { pool.Stop() })

	const p = p2pTypes.ID("p")
	pool.SetPeerHeight(p, 3)
	require.Eventually(t, func() bool { return len(requestsCh) == 3 }, 5*time.Second, time.Millisecond)
	pool.AddBlock(p, &types.Block{Header: types.Header{Height: 1}}, 1)
	pool.AddBlock(p, &types.Block{Header: types.Header{Height: 2}}, 1)
	pool.PopRequest()
	require.True(t, pool.IsWaitingOn(p))

	pool.AddBlock(p, &types.Block{Header: types.Header{Height: 1}}, 1)
	require.Equal(t, []p2pTypes.ID{p}, reported(errorsCh))
}

// Every report is sent under pool.mtx, so a full errors channel must not block
// it: the channel's consumer needs that lock to remove a peer.
func TestReportDoesNotBlockOnFullErrors(t *testing.T) {
	t.Parallel()

	errorsCh := make(chan peerError, 1)
	pool := NewBlockPool(1, make(chan BlockRequest, 1), errorsCh)
	require.NoError(t, pool.Start())
	t.Cleanup(func() { pool.Stop() })
	errorsCh <- peerError{}

	done := make(chan struct{})
	go func() {
		pool.AddBlock("stranger", &types.Block{Header: types.Header{Height: 1000}}, 1)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("a report blocked on the full errors channel")
	}
}

// A peer removed and re-added under the same ID gets a fresh entry that never
// counted the requests made to the old one, so popping such a request gives
// nothing back to it.
func TestPopSkipsReaddedPeer(t *testing.T) {
	t.Parallel()

	pool := NewBlockPool(1, make(chan BlockRequest, 1), make(chan peerError, 1))
	pool.SetPeerHeight("q", 2)
	for h := int64(1); h <= 2; h++ {
		pool.requesters[h] = newBPRequester(pool, h)
		require.NotNil(t, pool.pickIncrAvailablePeer(pool.requesters[h]))
	}
	old := pool.peers["q"]

	pool.RemovePeer("q")
	pool.SetPeerHeight("q", 2)
	require.NotPanics(t, pool.PopRequest)
	require.Zero(t, pool.peers["q"].numPending)
	require.Equal(t, int32(2), old.numPending, "the removed entry was given back a count it no longer answers for")
}

// TestAddBlockSkipsReaddedPeer pins that a delivered block gives its count back
// only to the entry that counted the request, not to one re-added under its ID.
func TestAddBlockSkipsReaddedPeer(t *testing.T) {
	t.Parallel()

	pool := NewBlockPool(1, make(chan BlockRequest, 1), make(chan peerError, 1))
	pool.SetPeerHeight("q", 2)
	pool.requesters[1] = newBPRequester(pool, 1)
	require.NotNil(t, pool.pickIncrAvailablePeer(pool.requesters[1]))
	old := pool.peers["q"]

	pool.RemovePeer("q")
	pool.SetPeerHeight("q", 2)
	pool.requesters[2] = newBPRequester(pool, 2)
	require.NotNil(t, pool.pickIncrAvailablePeer(pool.requesters[2]))

	pool.AddBlock("q", &types.Block{Header: types.Header{Height: 1}}, 1)
	require.NotNil(t, pool.requesters[1].getBlock())
	require.Equal(t, int32(1), pool.peers["q"].numPending, "the re-added entry was given back a count it never held")
	require.Equal(t, int32(1), old.numPending)
}

// peerSwitch holds the given peers, removes one it stops as the switch does,
// and sends each stop's error on a channel. Keeping a stopped peer would have
// poolRoutine stop it again on every pass until its requesters reset.
type peerSwitch struct {
	p2p.Switch
	p2p.PeerSet

	mtx     sync.Mutex
	peers   map[p2pTypes.ID]p2p.PeerConn
	stopped chan error
}

func (s *peerSwitch) Peers() p2p.PeerSet { return s }

func (s *peerSwitch) Broadcast(byte, []byte) {}

func (s *peerSwitch) Get(id p2pTypes.ID) p2p.PeerConn {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.peers[id]
}

func (s *peerSwitch) StopPeerForError(peer p2p.PeerConn, err error) {
	s.mtx.Lock()
	delete(s.peers, peer.ID())
	s.mtx.Unlock()
	s.stopped <- err
}

// The pool routine refuses a body its signed header does not describe.
func TestPoolRoutineChecksBodyAfterCommit(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		signedTamper  bool
		expectedError string
	}{
		{"body changed under a signed header", false, "wrong block id"},
		{"header signed over an invalid body", true, "wrong Header.DataHash"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			val, privVal := types.RandValidator(false, 30)
			genDoc := &types.GenesisDoc{
				GenesisTime: tmtime.Now(),
				ChainID:     "invalid-body",
				Validators:  []types.GenesisValidator{{PubKey: val.PubKey, Power: val.VotingPower}},
			}
			pair := newBlockchainReactor(log.NewNoopLogger(), genDoc, []types.PrivValidator{privVal}, 0)
			defer pair.app.Stop()
			bcR, state := pair.reactor, pair.reactor.initialState

			first := makeBlock(1, state, types.NewCommit(types.BlockID{}, nil))
			tamper := func() { first.DataHash = bytes.Repeat([]byte{1}, len(first.DataHash)) }
			if tc.signedTamper {
				tamper()
			}
			firstID := types.BlockID{Hash: first.Hash(), PartsHeader: first.MakePartSet(types.BlockPartSizeBytes).Header()}
			if !tc.signedTamper {
				tamper()
			}
			vote, err := types.MakeVote(1, firstID, state.Validators, privVal, genDoc.ChainID)
			require.NoError(t, err)
			// Only the second block's commit is read, to verify the first.
			second := types.MakeBlock(2, nil, types.NewCommit(firstID, []*types.CommitSig{vote.CommitSig()}))

			peer := mock.GeneratePeers(t, 1)[0]
			sw := &peerSwitch{peers: map[p2pTypes.ID]p2p.PeerConn{peer.ID(): peer}, stopped: make(chan error, 10)}
			bcR.SetSwitch(sw)
			require.NoError(t, bcR.Start())
			defer bcR.Stop()

			assigned := func(height int64) bool {
				bcR.pool.mtx.Lock()
				defer bcR.pool.mtx.Unlock()
				r := bcR.pool.requesters[height]
				return r != nil && r.getPeerID() == peer.ID()
			}
			// The pool counts itself caught up at the highest peer's height minus
			// one, so a peer two above it keeps the consensus tick from stopping it.
			bcR.pool.SetPeerHeight(peer.ID(), 3)
			require.Eventually(t, func() bool { return assigned(1) && assigned(2) }, 5*time.Second, 10*time.Millisecond)
			bcR.pool.AddBlock(peer.ID(), first, 1)
			bcR.pool.AddBlock(peer.ID(), second, 1)

			select {
			case err := <-sw.stopped:
				require.ErrorContains(t, err, tc.expectedError)
			case <-time.After(5 * time.Second):
				t.Fatal("the sender of the invalid block was not stopped")
			}
			require.Equal(t, int64(0), bcR.store.Height())
		})
	}
}
