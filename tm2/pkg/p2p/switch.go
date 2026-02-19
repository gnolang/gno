package p2p

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/config"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/dial"
	"github.com/gnolang/gno/tm2/pkg/p2p/events"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
)

// defaultDialTimeout is the default wait time for a dial to succeed
var defaultDialTimeout = 3 * time.Second

type reactorPeerBehavior struct {
	chDescs      []*conn.ChannelDescriptor
	reactorsByCh map[byte]Reactor

	handlePeerErrFn    func(PeerConn, error)
	isPersistentPeerFn func(types.ID) bool
	isPrivatePeerFn    func(types.ID) bool
}

func (r *reactorPeerBehavior) ReactorChDescriptors() []*conn.ChannelDescriptor {
	return r.chDescs
}

func (r *reactorPeerBehavior) Reactors() map[byte]Reactor {
	return r.reactorsByCh
}

func (r *reactorPeerBehavior) HandlePeerError(p PeerConn, err error) {
	r.handlePeerErrFn(p, err)
}

func (r *reactorPeerBehavior) IsPersistentPeer(id types.ID) bool {
	return r.isPersistentPeerFn(id)
}

func (r *reactorPeerBehavior) IsPrivatePeer(id types.ID) bool {
	return r.isPrivatePeerFn(id)
}

// MultiplexSwitch handles peer connections and exposes an API to receive incoming messages
// on `Reactors`.  Each `Reactor` is responsible for handling incoming messages of one
// or more `Channels`.  So while sending outgoing messages is typically performed on the peer,
// incoming messages are received on the reactor.
type MultiplexSwitch struct {
	service.BaseService

	ctx      context.Context
	cancelFn context.CancelFunc

	maxInboundPeers  uint64
	maxOutboundPeers uint64

	reactors     map[string]Reactor
	peerBehavior *reactorPeerBehavior

	peers           PeerSet  // currently active peer set (live connections)
	persistentPeers sync.Map // ID -> *NetAddress; peers whose connections are constant
	privatePeers    sync.Map // ID -> nothing; lookup table of peers who are not shared
	transport       Transport

	dialQueue  *dial.Queue
	dialNotify chan struct{}
	events     *events.Events
}

// NewMultiplexSwitch creates a new MultiplexSwitch with the given config.
func NewMultiplexSwitch(
	transport Transport,
	opts ...SwitchOption,
) *MultiplexSwitch {
	defaultCfg := config.DefaultP2PConfig()

	sw := &MultiplexSwitch{
		reactors:         make(map[string]Reactor),
		peers:            newSet(),
		transport:        transport,
		dialQueue:        dial.NewQueue(),
		dialNotify:       make(chan struct{}, 1),
		events:           events.New(),
		maxInboundPeers:  defaultCfg.MaxNumInboundPeers,
		maxOutboundPeers: defaultCfg.MaxNumOutboundPeers,
	}

	// Set up the peer dial behavior
	sw.peerBehavior = &reactorPeerBehavior{
		chDescs:         make([]*conn.ChannelDescriptor, 0),
		reactorsByCh:    make(map[byte]Reactor),
		handlePeerErrFn: sw.StopPeerForError,
		isPersistentPeerFn: func(id types.ID) bool {
			return sw.isPersistentPeer(id)
		},
		isPrivatePeerFn: func(id types.ID) bool {
			return sw.isPrivatePeer(id)
		},
	}

	sw.BaseService = *service.NewBaseService(nil, "P2P MultiplexSwitch", sw)

	// Set up the context
	sw.ctx, sw.cancelFn = context.WithCancel(context.Background())

	// Apply the options
	for _, opt := range opts {
		opt(sw)
	}

	return sw
}

// Subscribe registers to live events happening on the p2p Switch.
// Returns the notification channel, along with an unsubscribe method
func (sw *MultiplexSwitch) Subscribe(filterFn events.EventFilter) (<-chan events.Event, func()) {
	return sw.events.Subscribe(filterFn)
}

// ---------------------------------------------------------------------
// Service start/stop

// OnStart implements BaseService. It starts all the reactors and peers.
func (sw *MultiplexSwitch) OnStart() error {
	// Start reactors
	for _, reactor := range sw.reactors {
		if err := reactor.Start(); err != nil {
			return fmt.Errorf("unable to start reactor %w", err)
		}
	}

	// Run the peer accept routine.
	// The accept routine asynchronously accepts
	// and processes incoming peer connections
	go sw.runAcceptLoop(sw.ctx)

	// Run the dial routine.
	// The dial routine parses items in the dial queue
	// and initiates outbound peer connections
	go sw.runDialLoop(sw.ctx)

	// Run the redial routine.
	// The redial routine monitors for important
	// peer disconnects, and attempts to reconnect
	// to them
	go sw.runRedialLoop(sw.ctx)

	return nil
}

// OnStop implements BaseService. It stops all peers and reactors.
func (sw *MultiplexSwitch) OnStop() {
	// Close all hanging threads
	sw.cancelFn()

	// Stop peers
	for _, p := range sw.peers.List() {
		sw.stopAndRemovePeer(p, nil)
	}

	// Stop reactors
	for _, reactor := range sw.reactors {
		if err := reactor.Stop(); err != nil {
			sw.Logger.Error("unable to gracefully stop reactor", "err", err)
		}
	}
}

// Broadcast broadcasts the given data to the given channel,
// across the entire switch peer set, without blocking
func (sw *MultiplexSwitch) Broadcast(chID byte, data []byte) {
	for _, p := range sw.peers.List() {
		go func() {
			// This send context is managed internally
			// by the Peer's underlying connection implementation
			if !p.Send(chID, data) {
				sw.Logger.Error(
					"unable to perform broadcast",
					"chID", chID,
					"peerID", p.ID(),
				)
			}
		}()
	}
}

// Peers returns the set of peers that are connected to the switch.
func (sw *MultiplexSwitch) Peers() PeerSet {
	return sw.peers
}

// StopPeerForError disconnects from a peer due to external error.
// If the peer is persistent, it will attempt to reconnect
func (sw *MultiplexSwitch) StopPeerForError(peer PeerConn, err error) {
	sw.Logger.Error("Stopping peer for error", "peer", peer, "err", err)

	sw.stopAndRemovePeer(peer, err)

	if !peer.IsPersistent() {
		// Peer is not a persistent peer,
		// no need to initiate a redial
		return
	}

	// Add the peer to the dial queue
	sw.DialPeers(peer.SocketAddr())
}

func (sw *MultiplexSwitch) stopAndRemovePeer(peer PeerConn, err error) {
	// Remove the peer from the transport
	sw.transport.Remove(peer)

	// Close the (original) peer connection
	if closeErr := peer.CloseConn(); closeErr != nil {
		sw.Logger.Error(
			"unable to gracefully close peer connection",
			"peer", peer,
			"err", closeErr,
		)
	}

	// Stop the peer connection multiplexing
	if stopErr := peer.Stop(); stopErr != nil {
		sw.Logger.Error(
			"unable to gracefully stop peer",
			"peer", peer,
			"err", stopErr,
		)
	}

	// Alert the reactors of a peer removal
	for _, reactor := range sw.reactors {
		reactor.RemovePeer(peer, err)
	}

	// Removing a peer should go last to avoid a situation where a peer
	// reconnect to our node and the switch calls InitPeer before
	// RemovePeer is finished.
	// https://github.com/tendermint/tendermint/issues/3338
	sw.peers.Remove(peer.ID())

	sw.events.Notify(events.PeerDisconnectedEvent{
		Address: peer.RemoteAddr(),
		PeerID:  peer.ID(),
		Reason:  err,
	})
}

// ---------------------------------------------------------------------
// Dialing

func (sw *MultiplexSwitch) runDialLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			sw.Logger.Debug("dial context canceled")
			return

		default:
			// Grab a dial item
			item := sw.dialQueue.Peek()
			if item == nil {
				// Nothing to dial, wait until something is
				// added to the queue
				sw.waitForPeersToDial(ctx)
				continue
			}

			// Check if the dial time is right
			// for the item
			if time.Now().Before(item.Time) {
				// Nothing to dial
				continue
			}

			// Pop the item from the dial queue
			item = sw.dialQueue.Pop()

			// Dial the peer
			sw.Logger.Info(
				"dialing peer",
				"address", item.Address.String(),
			)

			peerAddr := item.Address

			// Check if the peer is already connected
			ps := sw.Peers()
			if ps.Has(peerAddr.ID) {
				sw.Logger.Warn(
					"ignoring dial request for existing peer",
					"id", peerAddr.ID,
				)

				continue
			}

			// Create a dial context
			dialCtx, cancelFn := context.WithTimeout(ctx, defaultDialTimeout)
			defer cancelFn()

			p, err := sw.transport.Dial(dialCtx, *peerAddr, sw.peerBehavior)
			if err != nil {
				sw.Logger.Error(
					"unable to dial peer",
					"peer", peerAddr,
					"err", err,
				)

				continue
			}

			// Register the peer with the switch
			if err = sw.addPeer(p); err != nil {
				sw.Logger.Error(
					"unable to add peer",
					"peer", p,
					"err", err,
				)

				sw.transport.Remove(p)

				if !p.IsRunning() {
					continue
				}

				if stopErr := p.Stop(); stopErr != nil {
					sw.Logger.Error(
						"unable to gracefully stop peer",
						"peer", p,
						"err", stopErr,
					)
				}
			}

			// Log the telemetry
			sw.logTelemetry()
		}
	}
}

// runRedialLoop starts the persistent peer redial loop
func (sw *MultiplexSwitch) runRedialLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	type backoffItem struct {
		lastDialTime time.Time
		attempts     uint
	}

	var (
		backoffMap = make(map[types.ID]*backoffItem)

		mux sync.RWMutex
	)

	setBackoffItem := func(id types.ID, item *backoffItem) {
		mux.Lock()
		defer mux.Unlock()

		backoffMap[id] = item
	}

	getBackoffItem := func(id types.ID) *backoffItem {
		mux.RLock()
		defer mux.RUnlock()

		return backoffMap[id]
	}

	clearBackoffItem := func(id types.ID) {
		mux.Lock()
		defer mux.Unlock()

		delete(backoffMap, id)
	}

	subCh, unsubFn := sw.Subscribe(func(event events.Event) bool {
		if event.Type() != events.PeerConnected {
			return false
		}

		ev := event.(events.PeerConnectedEvent)

		return sw.isPersistentPeer(ev.PeerID)
	})
	defer unsubFn()

	// redialFn goes through the persistent peer list
	// and dials missing peers
	redialFn := func() {
		var (
			peers       = sw.Peers()
			peersToDial = make([]*types.NetAddress, 0)
		)

		// Gather addresses of persistent peers that are missing or
		// not already in the dial queue
		sw.persistentPeers.Range(func(key, value any) bool {
			var (
				id   = key.(types.ID)
				addr = value.(*types.NetAddress)
			)

			if !peers.Has(id) && !sw.dialQueue.Has(addr) {
				peersToDial = append(peersToDial, addr)
			}

			return true
		})

		if len(peersToDial) == 0 {
			// No persistent peers need dialing
			return
		}

		// Prepare dial items with the appropriate backoff
		dialItems := make([]dial.Item, 0, len(peersToDial))
		for _, addr := range peersToDial {
			item := getBackoffItem(addr.ID)

			if item == nil {
				// First attempt
				now := time.Now()

				dialItems = append(dialItems,
					dial.Item{
						Time:    now,
						Address: addr,
					},
				)

				setBackoffItem(addr.ID, &backoffItem{
					lastDialTime: now,
					attempts:     0,
				})

				continue
			}

			// Subsequent attempt: apply backoff
			var (
				attempts = item.attempts + 1
				dialTime = time.Now().Add(
					calculateBackoff(
						item.attempts,
						time.Second,
						10*time.Minute,
					),
				)
			)

			dialItems = append(dialItems,
				dial.Item{
					Time:    dialTime,
					Address: addr,
				},
			)

			setBackoffItem(addr.ID, &backoffItem{
				lastDialTime: dialTime,
				attempts:     attempts,
			})
		}

		// Add these items to the dial queue
		sw.dialItems(dialItems...)
	}

	// Run the initial redial loop on start,
	// in case persistent peer connections are not
	// active
	redialFn()

	for {
		select {
		case <-ctx.Done():
			sw.Logger.Debug("redial crawl context canceled")

			return
		case <-ticker.C:
			redialFn()
		case event := <-subCh:
			// A persistent peer reconnected,
			// clear their redial queue
			ev := event.(events.PeerConnectedEvent)

			clearBackoffItem(ev.PeerID)
		}
	}
}

// calculateBackoff calculates the backoff interval by exponentiating the base interval
// by the number of attempts. The returned interval is capped at maxInterval and has a
// jitter factor applied to it (+/- 10% of interval, max 10 sec).
func calculateBackoff(
	attempts uint,
	baseInterval time.Duration,
	maxInterval time.Duration,
) time.Duration {
	const (
		defaultBaseInterval = time.Second * 1
		defaultMaxInterval  = time.Second * 60
	)

	// Sanitize base interval parameter.
	if baseInterval <= 0 {
		baseInterval = defaultBaseInterval
	}

	// Sanitize max interval parameter.
	if maxInterval <= 0 {
		maxInterval = defaultMaxInterval
	}

	// Calculate the interval by exponentiating the base interval by the number of attempts.
	interval := min(baseInterval<<attempts, maxInterval)

	// Below is the code to add a jitter factor to the interval.
	// Read random bytes into an 8 bytes buffer (size of an int64).
	var randBytes [8]byte
	if _, err := rand.Read(randBytes[:]); err != nil {
		return interval
	}

	// Convert the random bytes to an int64.
	var randInt64 int64
	_ = binary.Read(bytes.NewReader(randBytes[:]), binary.NativeEndian, &randInt64)

	// Calculate the random jitter multiplier (float between -1 and 1).
	jitterMultiplier := float64(randInt64) / float64(math.MaxInt64)

	const (
		maxJitterDuration   = 10 * time.Second
		maxJitterPercentage = 10 // 10%
	)

	// Calculate the maximum jitter based on interval percentage.
	maxJitter := min(interval*maxJitterPercentage/100, maxJitterDuration)

	// Calculate the jitter.
	jitter := time.Duration(float64(maxJitter) * jitterMultiplier)

	return interval + jitter
}

// DialPeers adds the peers to the dial queue for async dialing.
// To monitor dial progress, subscribe to adequate p2p MultiplexSwitch events
func (sw *MultiplexSwitch) DialPeers(peerAddrs ...*types.NetAddress) {
	for _, peerAddr := range peerAddrs {
		// Check if this is our address
		if peerAddr.Same(sw.transport.NetAddress()) {
			continue
		}

		// Ignore dial if the limit is reached
		if out := sw.Peers().NumOutbound(); out >= sw.maxOutboundPeers {
			sw.Logger.Warn(
				"ignoring dial request: already have max outbound peers",
				"have", out,
				"max", sw.maxOutboundPeers,
			)

			continue
		}

		item := dial.Item{
			Time:    time.Now(),
			Address: peerAddr,
		}

		sw.dialQueue.Push(item)
		sw.notifyAddPeerToDial()
	}
}

// dialItems adds custom dial items for the multiplex switch
func (sw *MultiplexSwitch) dialItems(dialItems ...dial.Item) {
	for _, dialItem := range dialItems {
		// Check if this is our address
		if dialItem.Address.Same(sw.transport.NetAddress()) {
			continue
		}

		// Ignore dial if the limit is reached
		if out := sw.Peers().NumOutbound(); out >= sw.maxOutboundPeers {
			sw.Logger.Warn(
				"ignoring dial request: already have max outbound peers",
				"have", out,
				"max", sw.maxOutboundPeers,
			)

			continue
		}

		sw.dialQueue.Push(dialItem)
		sw.notifyAddPeerToDial()
	}
}

// isPersistentPeer returns a flag indicating if a peer
// is present in the persistent peer set
func (sw *MultiplexSwitch) isPersistentPeer(id types.ID) bool {
	_, persistent := sw.persistentPeers.Load(id)

	return persistent
}

// isPrivatePeer returns a flag indicating if a peer
// is present in the private peer set
func (sw *MultiplexSwitch) isPrivatePeer(id types.ID) bool {
	_, persistent := sw.privatePeers.Load(id)

	return persistent
}

// runAcceptLoop is the main powerhouse method
// for accepting incoming peer connections, filtering them,
// and persisting them
func (sw *MultiplexSwitch) runAcceptLoop(ctx context.Context) {
	for {
		p, err := sw.transport.Accept(ctx, sw.peerBehavior)

		switch {
		case err == nil: // ok
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			// Upper context as been canceled/timeout
			sw.Logger.Debug("switch context close received")
			return // exit
		case errors.As(err, &errTransportClosed):
			// Underlaying transport as been closed
			sw.Logger.Warn("cannot accept connection on closed transport, exiting")
			return // exit
		default:
			// An error occurred during accept, report and continue
			sw.Logger.Error("error encountered during peer connection accept", "err", err)
			continue
		}

		// Ignore connection if we already have enough peers.
		if in := sw.Peers().NumInbound(); in >= sw.maxInboundPeers {
			sw.Logger.Info(
				"Ignoring inbound connection: already have enough inbound peers",
				"address", p.SocketAddr(),
				"have", in,
				"max", sw.maxInboundPeers,
			)

			sw.transport.Remove(p)
			continue
		}

		// There are open peer slots, add peers
		if err := sw.addPeer(p); err != nil {
			sw.transport.Remove(p)

			if p.IsRunning() {
				_ = p.Stop()
			}

			sw.Logger.Info(
				"Ignoring inbound connection: error while adding peer",
				"err", err,
				"id", p.ID(),
			)
		}
	}
}

// addPeer starts up the Peer and adds it to the MultiplexSwitch. Error is returned if
// the peer is filtered out or failed to start or can't be added.
func (sw *MultiplexSwitch) addPeer(p PeerConn) error {
	p.SetLogger(sw.Logger.With("peer", p.SocketAddr()))

	// Add some data to the peer, which is required by reactors.
	for _, reactor := range sw.reactors {
		p = reactor.InitPeer(p)
	}

	// Start the peer's send/recv routines.
	// Must start it before adding it to the peer set
	// to prevent Start and Stop from being called concurrently.
	if err := p.Start(); err != nil {
		sw.Logger.Error("Error starting peer", "err", err, "peer", p)

		return err
	}

	// Add the peer to the peer set. Do this before starting the reactors
	// so that if Receive errors, we will find the peer and remove it.
	sw.peers.Add(p)

	// Start all the reactor protocols on the peer.
	for _, reactor := range sw.reactors {
		reactor.AddPeer(p)
	}

	sw.Logger.Info("Added peer", "peer", p)

	sw.events.Notify(events.PeerConnectedEvent{
		Address: p.RemoteAddr(),
		PeerID:  p.ID(),
	})

	return nil
}

func (sw *MultiplexSwitch) notifyAddPeerToDial() {
	select {
	case sw.dialNotify <- struct{}{}:
	default:
	}
}

func (sw *MultiplexSwitch) waitForPeersToDial(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-sw.dialNotify:
	}
}

// logTelemetry logs the switch telemetry data
// to global metrics funnels
func (sw *MultiplexSwitch) logTelemetry() {
	// Update the telemetry data
	if !telemetry.MetricsEnabled() {
		return
	}

	// Fetch the number of peers
	outbound, inbound := sw.peers.NumOutbound(), sw.peers.NumInbound()

	// Log the outbound peer count
	metrics.OutboundPeers.Record(context.Background(), int64(outbound))

	// Log the inbound peer count
	metrics.InboundPeers.Record(context.Background(), int64(inbound))
}
