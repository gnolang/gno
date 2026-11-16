package blockchain

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

const (
	// BlockchainChannel is a channel for blocks and status updates (`BlockStore` height)
	BlockchainChannel = byte(0x40)

	trySyncIntervalMS = 10

	// stop syncing when last block's time is
	// within this much of the system time.
	// stopSyncingDurationMinutes = 10

	// ask for best height every 10s
	statusUpdateIntervalSeconds = 10
	// check if we should switch to consensus reactor
	switchToConsensusIntervalSeconds = 1

	// While deferring its writes (SetDeferredWrites), block sync takes a
	// flush point at least this often, bounding what a crash or a stop costs
	// again, and when a database holds more than flushPointBytes.
	flushPointInterval = time.Second
	flushPointBytes    = 64 << 20 // 64 MiB

	// maxBlockMsgOverhead is the allowance for the bcBlockResponseMessage
	// envelope wrapped around an already length-prefixed block: an amino type
	// prefix, a field key and a length prefix, i.e. tens of bytes. 64KB is
	// generous headroom for that framing.
	//
	// Note it does NOT need to cover the block header or the LastCommit:
	// MaxDataBytes is also the max size ConsensusState.addProposalBlockPart
	// passes to amino when decoding a proposal block, so a block can only ever
	// be committed if its *whole* serialized form — header and commit
	// included — fits in MaxDataBytes.
	maxBlockMsgOverhead = 64 << 10 // 64KB

	// maxMsgSize is the maximum size of a blockchain-reactor message. The
	// largest message is a bcBlockResponseMessage carrying a full block. Since a
	// committed block's serialized size is bounded by the chain's MaxDataBytes,
	// which consensus-param validation caps at MaxBlockDataBytesLimit, deriving
	// the limit from the ceiling covers every block the chain could ever have
	// committed, including blocks committed before MaxDataBytes was lowered.
	//
	// It doubles as the channel's RecvMessageCapacity, so keeping it tight
	// bounds the per-connection recving-buffer exposure; it was previously
	// MaxBlockSizeBytes (100MB), far larger than any real block.
	maxMsgSize = int(types.MaxBlockDataBytesLimit) + maxBlockMsgOverhead

	// maxSmallMsgSize bounds every message of this reactor but a block
	// response: each carries one height, and at math.MaxInt64 the largest is
	// 34 bytes.
	maxSmallMsgSize = 64

	// blockRequestsKey is the peer data key of the heights the peer
	// requested and respondRoutine has yet to answer.
	blockRequestsKey = "BlockchainReactor.blockRequests"

	// sendRetryDelay paces sending again a block request or response that
	// failed to queue. Neither is ever asked for twice, so neither is dropped
	// while its peer is connected.
	sendRetryDelay = 100 * time.Millisecond
)

// SwitchToConsensusFn is a callback method that is meant to
// stop the syncing process as soon as the latest known height is reached,
// and start the consensus process for the validator node
type SwitchToConsensusFn func(sm.State, int)

type peerError struct {
	err    error
	peerID p2pTypes.ID
}

func (e peerError) Error() string {
	return fmt.Sprintf("error with peer %v: %s", e.peerID, e.err.Error())
}

// BlockchainReactor handles long-term catchup syncing.
type BlockchainReactor struct {
	p2p.BaseReactor

	// immutable
	initialState sm.State

	blockExec *sm.BlockExecutor
	store     *store.BlockStore
	pool      *BlockPool
	fastSync  bool

	requestsCh <-chan BlockRequest
	errorsCh   <-chan peerError

	switchToConsensusFn SwitchToConsensusFn

	deferred *DeferredWrites // nil: every block is written as it is applied

	routineDone chan struct{} // closed when poolRoutine returns; nil if it never started
}

// NewBlockchainReactor returns new reactor instance.
func NewBlockchainReactor(
	state sm.State,
	blockExec *sm.BlockExecutor,
	store *store.BlockStore,
	fastSync bool,
	switchToConsensusFn SwitchToConsensusFn,
) *BlockchainReactor {
	// Allow the case where InitialHeight > 1: after InitChain, the Handshaker sets
	// state.LastBlockHeight = InitialHeight - 1, but the block store is still empty
	// (Height() == 0). A non-empty store must always match state.
	if store.Height() != 0 && state.LastBlockHeight != store.Height() {
		panic(fmt.Sprintf("state (%v) and store (%v) height mismatch", state.LastBlockHeight,
			store.Height()))
	}

	requestsCh := make(chan BlockRequest, maxTotalRequesters)

	const capacity = 1000                      // must be bigger than peers count
	errorsCh := make(chan peerError, capacity) // so we don't block in #Receive#pool.AddBlock

	// When the store is empty (fresh chain) and InitialHeight > 1, the
	// Handshaker has set state.LastBlockHeight = InitialHeight - 1 but the
	// store is still at height 0. Use the state height so the pool starts
	// syncing at InitialHeight, not at 1.
	startHeight := store.Height() + 1
	if store.Height() == 0 && state.LastBlockHeight > 0 {
		startHeight = state.LastBlockHeight + 1
	}

	pool := NewBlockPool(
		startHeight,
		requestsCh,
		errorsCh,
	)

	bcR := &BlockchainReactor{
		initialState:        state,
		blockExec:           blockExec,
		store:               store,
		pool:                pool,
		fastSync:            fastSync,
		requestsCh:          requestsCh,
		errorsCh:            errorsCh,
		switchToConsensusFn: switchToConsensusFn,
	}
	bcR.BaseReactor = *p2p.NewBaseReactor("BlockchainReactor", bcR)
	return bcR
}

// SetLogger implements cmn.Service by setting the logger on reactor and pool.
func (bcR *BlockchainReactor) SetLogger(l *slog.Logger) {
	bcR.BaseService.Logger = l
	bcR.pool.Logger = l
}

// SetDeferredWrites makes block sync hold the writes of the block store, of
// the state and of app in memory, and write all of them at flush points, if
// their databases are HoldingDBs and app agrees to defer its commits (see
// DeferredWrites). Call it before the reactor starts.
func (bcR *BlockchainReactor) SetDeferredWrites(app abci.DeferredCommitter) {
	bcR.deferred = NewDeferredWrites(bcR.store, bcR.blockExec, app)
}

// OnStart implements cmn.Service.
func (bcR *BlockchainReactor) OnStart() error {
	if bcR.fastSync {
		err := bcR.pool.Start()
		if err != nil {
			return err
		}
		bcR.routineDone = make(chan struct{})
		go func() {
			defer close(bcR.routineDone)
			bcR.poolRoutine()
		}()
	}
	return nil
}

// OnStop implements cmn.Service. It waits for the block-sync routine, which
// returns once the pool is stopped, at the end of its current block or flush
// point: whoever stops the node may then close the application's database.
func (bcR *BlockchainReactor) OnStop() {
	bcR.pool.Stop()
	if bcR.routineDone != nil {
		<-bcR.routineDone
	}
}

// GetChannels implements Reactor
func (bcR *BlockchainReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  BlockchainChannel,
			Priority:            10,
			SendQueueCapacity:   1000,
			RecvBufferCapacity:  50 * 4096,
			RecvMessageCapacity: maxMsgSize,
		},
	}
}

// InitPeer implements Reactor by creating the peer's block request queue.
// The queue exists before the peer starts, so a request arriving before
// AddPeer waits in it.
func (bcR *BlockchainReactor) InitPeer(peer p2p.PeerConn) p2p.PeerConn {
	peer.Set(blockRequestsKey, make(chan int64, maxPendingRequestsPerPeer))
	return peer
}

// AddPeer implements Reactor by sending our state to peer and starting the
// peer's respondRoutine.
func (bcR *BlockchainReactor) AddPeer(peer p2p.PeerConn) {
	msgBytes := amino.MustMarshalAny(&bcStatusResponseMessage{bcR.store.Height()})
	peer.Send(BlockchainChannel, msgBytes)
	// it's OK if send fails. will try later in poolRoutine

	if requests, ok := peer.Get(blockRequestsKey).(chan int64); ok {
		go bcR.respondRoutine(peer, requests)
	}

	// peer is added to the pool once we receive the first
	// bcStatusResponseMessage from the peer and call pool.SetPeerHeight
}

// RemovePeer implements Reactor by removing peer from the pool.
func (bcR *BlockchainReactor) RemovePeer(peer p2p.PeerConn, reason any) {
	bcR.pool.RemovePeer(peer.ID())
}

// respondToPeer queues the requested height for the peer's respondRoutine.
// Receive runs on the peer's recvRoutine, so it must not wait for send
// budget: the peer's pings would go unread, and the wait would be credited
// to the pong deadline. A request beyond maxPendingRequestsPerPeer, more than
// an honest requester keeps with us, is dropped.
func (bcR *BlockchainReactor) respondToPeer(msg *bcBlockRequestMessage, src p2p.PeerConn) {
	requests, _ := src.Get(blockRequestsKey).(chan int64)
	select {
	case requests <- msg.Height:
	default:
		bcR.Logger.Debug("Too many block requests in flight, drop one", "src", src, "height", msg.Height)
	}
}

// respondRoutine answers the peer's requests in order until the peer or the
// reactor stops, sending again a response Send fails to queue. The response
// it waits on is held outside the send budget.
func (bcR *BlockchainReactor) respondRoutine(peer p2p.PeerConn, requests <-chan int64) {
	// LoadBlock panics on a block it cannot read. Stop the peer that asked
	// for it, as recvRoutine's recover did when Receive answered requests,
	// rather than the node.
	defer func() {
		if r := recover(); r != nil {
			bcR.Logger.Error("Failed to answer a block request", "peer", peer, "err", r, "stack", string(debug.Stack()))
			bcR.Switch.StopPeerForError(peer, fmt.Errorf("answering a block request: %v", r))
		}
	}()

	for {
		var height int64
		select {
		case <-peer.Quit():
			return
		case <-bcR.Quit():
			return
		case height = <-requests:
		}

		msgBytes := bcR.blockResponse(peer, height)
		for !peer.Send(BlockchainChannel, msgBytes) {
			select {
			case <-peer.Quit():
				return
			case <-bcR.Quit():
				return
			case <-time.After(sendRetryDelay):
			}
		}
	}
}

// blockResponse returns the response to a request for height: the block if
// we have it, a bcNoBlockResponseMessage otherwise. According to the
// Tendermint spec, if all nodes are honest, no node should be requesting a
// block that's non-existent.
func (bcR *BlockchainReactor) blockResponse(peer p2p.PeerConn, height int64) []byte {
	if block := bcR.store.LoadBlock(height); block != nil {
		return amino.MustMarshalAny(&bcBlockResponseMessage{Block: block})
	}

	bcR.Logger.Info("Peer asking for a block we don't have", "src", peer, "height", height)

	return amino.MustMarshalAny(&bcNoBlockResponseMessage{Height: height})
}

// Receive implements Reactor by handling 4 types of messages (look below).
func (bcR *BlockchainReactor) Receive(chID byte, src p2p.PeerConn, msgBytes []byte) {
	// Decoding a block costs work per tx, about 170 bytes of allocation for an
	// empty tx that takes 2 bytes on the wire. So a message over
	// maxSmallMsgSize is decoded only from a peer the pool is waiting on.
	if len(msgBytes) > maxSmallMsgSize && !bcR.pool.IsWaitingOn(src.ID()) {
		bcR.Logger.Debug("Dropping a large message from a peer we asked nothing", "src", src, "bytes", len(msgBytes))
		return
	}

	msg, err := decodeMsg(msgBytes)
	if err != nil {
		bcR.Logger.Error("Error decoding message", "src", src, "chId", chID, "msg", msg, "err", err, "bytes", msgBytes)
		bcR.Switch.StopPeerForError(src, err)
		return
	}

	if err = msg.ValidateBasic(); err != nil {
		bcR.Logger.Error("Peer sent us invalid msg", "peer", src, "msg", msg, "err", err)
		bcR.Switch.StopPeerForError(src, err)
		return
	}

	bcR.Logger.Debug("Receive", "src", src, "chID", chID, "msg", msg)

	switch msg := msg.(type) {
	case *bcBlockRequestMessage:
		bcR.respondToPeer(msg, src)
	case *bcBlockResponseMessage:
		bcR.pool.AddBlock(src.ID(), msg.Block, len(msgBytes))
	case *bcStatusRequestMessage:
		// Send peer our state.
		msgBytes := amino.MustMarshalAny(&bcStatusResponseMessage{bcR.store.Height()})
		src.TrySend(BlockchainChannel, msgBytes)
	case *bcStatusResponseMessage:
		// Got a peer status. Unverified.
		bcR.pool.SetPeerHeight(src.ID(), msg.Height)
	default:
		bcR.Logger.Error(fmt.Sprintf("Unknown message type %v", reflect.TypeOf(msg)))
	}
}

// peerRequest is a block request and the connection it was made for.
type peerRequest struct {
	BlockRequest
	peer p2p.PeerConn
}

// trySendRequest sends a block request to the connection it was made for
// without blocking. It reports false only when that connection still holds
// the peer's ID but had no room for it. A request whose connection is gone
// is dropped, even when a new connection holds the ID: the pool redoes the
// peer's requests when it removes it, and stops a peer whose block it did
// not ask that peer for.
func (bcR *BlockchainReactor) trySendRequest(request peerRequest) bool {
	if request.peer == nil || bcR.Switch.Peers().Get(request.PeerID) != request.peer {
		return true
	}

	return request.peer.TrySend(BlockchainChannel, amino.MustMarshalAny(&bcBlockRequestMessage{request.Height}))
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or otherwise slow it down!
func (bcR *BlockchainReactor) poolRoutine() {
	trySyncTicker := time.NewTicker(trySyncIntervalMS * time.Millisecond)
	statusUpdateTicker := time.NewTicker(statusUpdateIntervalSeconds * time.Second)
	switchToConsensusTicker := time.NewTicker(switchToConsensusIntervalSeconds * time.Second)

	blocksSynced := 0

	chainID := bcR.initialState.ChainID
	state := bcR.initialState

	lastHundred := time.Now()
	lastRate := 0.0

	didProcessCh := make(chan struct{}, 1)

	// deferred holds the writes since the last flush point; nil while not
	// deferring.
	var deferred *DeferredWrites
	if bcR.deferred != nil && bcR.deferred.Start(state.LastBlockHeight) {
		deferred = bcR.deferred
	}

	go func() {
		// This goroutine serves every peer, so it keeps the requests TrySend
		// could not queue instead of blocking on one peer.
		var unsent []peerRequest
		retryTicker := time.NewTicker(sendRetryDelay)
		defer retryTicker.Stop()

		for {
			select {
			case <-bcR.Quit():
				return
			case <-bcR.pool.Quit():
				return
			case req := <-bcR.requestsCh:
				request := peerRequest{req, bcR.Switch.Peers().Get(req.PeerID)}
				if !bcR.trySendRequest(request) {
					unsent = append(unsent, request)
				}
			case <-retryTicker.C:
				kept := unsent[:0]
				for _, request := range unsent {
					if !bcR.trySendRequest(request) {
						kept = append(kept, request)
					}
				}
				unsent = kept
			case err := <-bcR.errorsCh:
				peer := bcR.Switch.Peers().Get(err.peerID)
				if peer != nil {
					bcR.Switch.StopPeerForError(peer, err)
				}

			case <-statusUpdateTicker.C:
				// ask for status updates
				go bcR.BroadcastStatusRequest() //nolint: errcheck
			}
		}
	}()

FOR_LOOP:
	for {
		select {
		case <-switchToConsensusTicker.C:
			height, numPending, lenRequesters := bcR.pool.GetStatus()

			bcR.Logger.Debug("Consensus ticker", "numPending", numPending, "total", lenRequesters)
			if deferred != nil && deferred.NeedsFlushPoint() {
				// Also when no block arrives, so a stalled window is written.
				deferred.FlushPoint(state.LastBlockHeight)
			}
			if bcR.pool.IsCaughtUp() {
				// No consensus block may land in a held database, and a stop
				// meanwhile must not start consensus: the pool stops only
				// once, so its Stop fails if OnStop stopped it first.
				if deferred != nil {
					deferred.FlushPoint(state.LastBlockHeight)
					deferred.Stop()
					deferred = nil
				}
				if err := bcR.pool.Stop(); err != nil {
					break FOR_LOOP // stopped meanwhile, by OnStop
				}
				bcR.Logger.Info("Time to switch to consensus reactor!", "height", height)

				// switchToConsensusFn may be nil under test harnesses that
				// construct a reactor in isolation; production wiring always
				// supplies it.
				if bcR.switchToConsensusFn != nil {
					bcR.switchToConsensusFn(state, blocksSynced)
				}

				break FOR_LOOP
			}

		case <-trySyncTicker.C: // chan time
			select {
			case didProcessCh <- struct{}{}:
			default:
			}

		case <-didProcessCh:
			// NOTE: It is a subtle mistake to process more than a single block
			// at a time (e.g. 10) here, because we only TrySend 1 request per
			// loop.  The ratio mismatch can result in starving of blocks, a
			// sudden burst of requests and responses, and repeat.
			// Consequently, it is better to split these routines rather than
			// coupling them as it's written here.  TODO uncouple from request
			// routine.

			// See if there are any blocks to sync.
			first, second := bcR.pool.PeekTwoBlocks()
			// bcR.Logger.Info("TrySync peeked", "first", first, "second", second)
			if first == nil || second == nil {
				// We need both to sync the first block.
				continue FOR_LOOP
			} else {
				// Try again quickly next loop.
				didProcessCh <- struct{}{}
			}

			firstParts := first.MakePartSet(types.BlockPartSizeBytes)
			firstPartsHeader := firstParts.Header()
			firstID := types.BlockID{Hash: first.Hash(), PartsHeader: firstPartsHeader}
			// Finally, verify the first block using the second's commit
			// NOTE: we can probably make this more efficient, but note that calling
			// first.Hash() doesn't verify the tx contents, so MakePartSet() is
			// currently necessary.
			err := state.Validators.VerifyCommit(
				chainID, firstID, first.Height, second.LastCommit)
			if err == nil {
				err = first.ValidateBasic()
			}
			if err != nil {
				bcR.Logger.Error("Error in validation", "err", err)
				peerID := bcR.pool.RedoRequest(first.Height)
				peer := bcR.Switch.Peers().Get(peerID)
				if peer != nil {
					// NOTE: we've already removed the peer's request, but we
					// still need to clean up the rest.
					bcR.Switch.StopPeerForError(peer, fmt.Errorf("BlockchainReactor validation error: %w", err))
				}
				peerID2 := bcR.pool.RedoRequest(second.Height)
				peer2 := bcR.Switch.Peers().Get(peerID2)
				if peer2 != nil && peer2 != peer {
					// NOTE: we've already removed the peer's request, but we
					// still need to clean up the rest.
					bcR.Switch.StopPeerForError(peer2, fmt.Errorf("BlockchainReactor validation error: %w", err))
				}
				continue FOR_LOOP
			} else {
				bcR.pool.PopRequest()

				// TODO: batch saves so we dont persist to disk every block
				bcR.store.SaveBlock(first, firstParts, second.LastCommit)

				// TODO: same thing for app - but we would need a way to
				// get the hash without persisting the state
				var err error
				state, err = bcR.blockExec.ApplyBlock(state, firstID, first)
				if err != nil {
					// TODO This is bad, are we zombie?
					panic(fmt.Sprintf("Failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err))
				}
				blocksSynced++

				if deferred != nil && deferred.NeedsFlushPoint() {
					deferred.FlushPoint(state.LastBlockHeight)
				}

				if blocksSynced%100 == 0 {
					lastRate = 0.9*lastRate + 0.1*(100/time.Since(lastHundred).Seconds())
					bcR.Logger.Info("Fast Sync Rate", "height", bcR.pool.height,
						"max_peer_height", bcR.pool.MaxPeerHeight(), "blocks/s", lastRate)
					lastHundred = time.Now()
				}
			}
			continue FOR_LOOP

		case <-bcR.pool.Quit():
			// Stopping (OnStop stops the pool, then waits for this routine,
			// so the reactor's own Quit closes only after it returns): what
			// is held is dropped, as on a crash, and fetched again after the
			// restart.
			break FOR_LOOP
		}
	}
}

// BroadcastStatusRequest broadcasts `BlockStore` height.
func (bcR *BlockchainReactor) BroadcastStatusRequest() error {
	msgBytes := amino.MustMarshalAny(&bcStatusRequestMessage{bcR.store.Height()})
	bcR.Switch.Broadcast(BlockchainChannel, msgBytes)
	return nil
}

// -----------------------------------------------------------------------------
// Messages

// BlockchainMessage is a generic message for this reactor.
type BlockchainMessage interface {
	ValidateBasic() error
}

func decodeMsg(bz []byte) (msg BlockchainMessage, err error) {
	if len(bz) > maxMsgSize {
		return msg, fmt.Errorf("msg exceeds max size (%d > %d)", len(bz), maxMsgSize)
	}
	err = amino.Unmarshal(bz, &msg)
	return
}

// -------------------------------------

type bcBlockRequestMessage struct {
	Height int64
}

// ValidateBasic performs basic validation.
func (m *bcBlockRequestMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative height")
	}
	return nil
}

func (m *bcBlockRequestMessage) String() string {
	return fmt.Sprintf("[bcBlockRequestMessage %v]", m.Height)
}

type bcNoBlockResponseMessage struct {
	Height int64
}

// ValidateBasic performs basic validation.
func (m *bcNoBlockResponseMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative height")
	}
	return nil
}

func (m *bcNoBlockResponseMessage) String() string {
	return fmt.Sprintf("[bcNoBlockResponseMessage %d]", m.Height)
}

// -------------------------------------

type bcBlockResponseMessage struct {
	Block *types.Block
}

// ValidateBasic performs basic validation.
func (m *bcBlockResponseMessage) ValidateBasic() error {
	// The body is hashed only once a commit has signed the header, in poolRoutine.
	return m.Block.ValidateHeaderBasic()
}

func (m *bcBlockResponseMessage) String() string {
	return fmt.Sprintf("[bcBlockResponseMessage %v]", m.Block.Height)
}

// -------------------------------------

type bcStatusRequestMessage struct {
	Height int64
}

// ValidateBasic performs basic validation.
func (m *bcStatusRequestMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative height")
	}
	return nil
}

func (m *bcStatusRequestMessage) String() string {
	return fmt.Sprintf("[bcStatusRequestMessage %v]", m.Height)
}

// -------------------------------------

type bcStatusResponseMessage struct {
	Height int64
}

// ValidateBasic performs basic validation.
func (m *bcStatusResponseMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative height")
	}
	return nil
}

func (m *bcStatusResponseMessage) String() string {
	return fmt.Sprintf("[bcStatusResponseMessage %v]", m.Height)
}
