package p2p

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
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

var (
	// defaultDialTimeout is the default wait time for a dial to succeed
	defaultDialTimeout = 3 * time.Second

	// seedDialInterval is the minimum wait time between two seed dial rounds
	seedDialInterval = 30 * time.Second

	// defaultRedialInterval is the period at which the redial loop looks for
	// missing persistent peers
	defaultRedialInterval = 5 * time.Second

	// persistentRedialMaxBackoff is the ceiling of the backoff between two dials
	// of a missing persistent peer, before its jitter
	persistentRedialMaxBackoff = 30 * time.Second
)

var (
	// errDuplicatePeer is returned when a connection carries a peer ID the
	// peer set already holds
	errDuplicatePeer = errors.New("duplicate peer")

	// errMaxOutboundPeers is returned when a dialed connection would exceed
	// the outbound peer limit
	errMaxOutboundPeers = errors.New("already have max outbound peers")

	// errPeerStopped is returned when a peer is stopped while being added
	errPeerStopped = errors.New("peer stopped while being added")

	// errSimultaneousOpen is the reason a connection is closed when another
	// connection to the same peer, in the opposite direction, is kept over it
	errSimultaneousOpen = errors.New("replaced by the connection kept for a simultaneous open")
)

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

	// redialInterval is the period at which the redial loop looks for missing
	// persistent peers
	redialInterval time.Duration

	// allowDuplicateIP disables the guard that stops a single remote IP from
	// occupying more than one inbound peer slot
	allowDuplicateIP bool

	reactors     map[string]Reactor
	peerBehavior *reactorPeerBehavior

	peers PeerSet // currently active peer set (live connections)

	// registry serializes every step that reads the peer set entry of a peer
	// ID and then changes it, so no connection is added, kept or removed on a
	// stale view of that entry
	registry        sync.Mutex
	persistentPeers sync.Map // ID -> *NetAddress; peers whose connections are constant
	seeds           sync.Map // ID -> *NetAddress; bootstrap peers, not kept alive
	privatePeers    sync.Map // ID -> nothing; lookup table of peers who are not shared
	transport       Transport

	dialQueue           *dial.Queue // dials of discovered peers and seeds
	persistentDialQueue *dial.Queue // dials of persistent peers, fed by the redial loop
	dialNotify          chan struct{}
	events              *events.Events
}

// NewMultiplexSwitch creates a new MultiplexSwitch with the given config.
func NewMultiplexSwitch(
	transport Transport,
	opts ...SwitchOption,
) *MultiplexSwitch {
	defaultCfg := config.DefaultP2PConfig()

	sw := &MultiplexSwitch{
		reactors:            make(map[string]Reactor),
		peers:               newSet(),
		transport:           transport,
		dialQueue:           dial.NewQueue(),
		persistentDialQueue: dial.NewQueue(),
		dialNotify:          make(chan struct{}, 1),
		events:              events.New(),
		maxInboundPeers:     defaultCfg.MaxNumInboundPeers,
		maxOutboundPeers:    defaultCfg.MaxNumOutboundPeers,
		redialInterval:      defaultRedialInterval,
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

	// Run the seed dial routine.
	// The seed dial routine falls back to the seed nodes
	// whenever the switch has run out of peers to dial
	go sw.runSeedDialLoop(sw.ctx)

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
// A persistent peer is redialed by the redial loop, on its configured address
func (sw *MultiplexSwitch) StopPeerForError(peer PeerConn, err error) {
	sw.Logger.Error("Stopping peer for error", "peer", peer, "err", err)

	sw.stopAndRemovePeer(peer, err)
}

// isSuperseded reports whether a different connection is registered under this
// peer's ID. Two connections hold one peer ID while a reconnect races the
// teardown of the connection it supersedes, or while a replacement resolves a
// simultaneous open, and the peer set entry under that ID belongs to whichever
// won
func (sw *MultiplexSwitch) isSuperseded(peer PeerConn) bool {
	registered := sw.peers.Get(peer.ID())

	return registered != nil && registered != peer
}

// removeReactorPeerState walks the reactors' RemovePeer so the state their
// InitPeer created is given back. Without it, an addPeer path that returns an
// error after InitPeer has run leaves that state held for the lifetime of the
// process.
//
// This is unconditional, including for a connection another has superseded: a
// reactor keying its state on the connection, as mempoolIDs does, gives back
// only what this connection took, and giving nothing back is a leak. A reactor
// keying on the peer ID instead cannot tell the two apart either way
func (sw *MultiplexSwitch) removeReactorPeerState(peer PeerConn, err error) {
	for _, reactor := range sw.reactors {
		reactor.RemovePeer(peer, err)
	}
}

func (sw *MultiplexSwitch) stopAndRemovePeer(peer PeerConn, err error) {
	// Remove the peer from the transport
	sw.transport.Remove(peer)

	// Stop the peer connection multiplexing before closing the socket. Stopping
	// closes the recv routine's quit channel first, so the recv routine does
	// not report the close as an error and start a second teardown of this
	// connection. A peer already stopped is left alone: whichever stopped it
	// owns its teardown
	if stopErr := peer.Stop(); stopErr != nil {
		if errors.Is(stopErr, service.ErrAlreadyStopped) {
			return
		}

		sw.Logger.Error(
			"unable to gracefully stop peer",
			"peer", peer,
			"err", stopErr,
		)
	}

	// Close the (original) peer connection. Stopping a started peer already
	// closed it, so net.ErrClosed is the expected outcome
	if closeErr := peer.CloseConn(); closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
		sw.Logger.Error(
			"unable to gracefully close peer connection",
			"peer", peer,
			"err", closeErr,
		)
	}

	// Alert the reactors of a peer removal
	sw.removeReactorPeerState(peer, err)

	// Removing a peer should go last to avoid a situation where a peer
	// reconnect to our node and the switch calls InitPeer before
	// RemovePeer is finished.
	// https://github.com/tendermint/tendermint/issues/3338
	//
	// A connection superseded by another sharing its peer ID, after losing the
	// race for the peer set or being replaced to resolve a simultaneous open,
	// has its own socket closed above and its own reactor state given back,
	// but the entry under that ID is the live connection's, and this one does
	// not announce a disconnect for a peer that stays connected
	if !sw.removeUnlessSuperseded(peer) {
		sw.Logger.Debug(
			"not removing the peer set entry of a superseded connection",
			"peer", peer,
			"err", err,
		)

		return
	}

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
			// Grab the next dial item
			item, queue := sw.peekDialItem()
			if item == nil {
				// Nothing to dial, wait until something is
				// added to a queue
				sw.waitForPeersToDial(ctx)
				continue
			}

			// Check if the dial time is right
			// for the item
			if wait := time.Until(item.Time); wait > 0 {
				// Nothing to dial yet, wait until the item is due
				sw.waitForDialTime(ctx, wait)

				continue
			}

			// Pop the item from its dial queue. The dial loop is the only
			// consumer, so a push since the peek can only have put an earlier,
			// also due, item at the head
			item = queue.Pop()
			peerAddr := item.Address

			// Check if the peer is already connected
			ps := sw.Peers()
			if ps.Has(peerAddr.ID) {
				sw.Logger.Debug(
					"skipping dial, peer already connected",
					"address", peerAddr.String(),
				)

				continue
			}

			// Dial the peer
			sw.Logger.Info(
				"dialing peer",
				"address", item.Address.String(),
			)

			sw.dialPeer(ctx, peerAddr)
		}
	}
}

// peekDialItem returns the next item to dial, along with the queue holding it.
// A persistent peer due for dialing goes first, then a due discovered peer.
// With nothing due, it returns the item due first, so the dial loop can wait
// for it. The returned item is nil when both queues are empty
func (sw *MultiplexSwitch) peekDialItem() (*dial.Item, *dial.Queue) {
	var (
		now        = time.Now()
		persistent = sw.persistentDialQueue.Peek()
		general    = sw.dialQueue.Peek()
	)

	switch {
	case persistent != nil && !now.Before(persistent.Time):
		// A due persistent peer goes first
		return persistent, sw.persistentDialQueue
	case general != nil && (persistent == nil || !general.Time.After(persistent.Time)):
		// Otherwise the general head, when it is due or due first
		return general, sw.dialQueue
	default:
		// The persistent head is due first, or both queues are empty
		return persistent, sw.persistentDialQueue
	}
}

// dialPeer dials the given peer address, and registers the resulting
// connection with the switch.
// It is a separate method so the dial context is released when the dial
// completes. Deferring it inside the dial loop instead would pile every
// dial's cancel onto the loop, to be released only at shutdown
func (sw *MultiplexSwitch) dialPeer(ctx context.Context, peerAddr *types.NetAddress) {
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

		return
	}

	// Register the peer with the switch
	if err = sw.addPeer(p); err != nil {
		// A connection refused as a duplicate logs at Info: it is the
		// tie-break's designed outcome on one node of every simultaneous open
		logFn := sw.Logger.Error
		if errors.Is(err, errDuplicatePeer) {
			logFn = sw.Logger.Info
		}

		logFn(
			"unable to add peer",
			"peer", p,
			"err", err,
		)

		if !p.IsRunning() {
			sw.rejectConn(p)

			return
		}

		sw.transport.Remove(p)

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

// runRedialLoop starts the persistent peer redial loop.
// It is the only producer of persistent peer dials
func (sw *MultiplexSwitch) runRedialLoop(ctx context.Context) {
	ticker := time.NewTicker(sw.redialInterval)
	defer ticker.Stop()

	// Dial attempts of each persistent peer since it last connected.
	// Only this goroutine reads or writes it
	attempts := make(map[types.ID]uint)

	subCh, unsubFn := sw.Subscribe(func(event events.Event) bool {
		if event.Type() != events.PeerConnected {
			return false
		}

		ev := event.(events.PeerConnectedEvent)

		return sw.isPersistentPeer(ev.PeerID)
	})
	defer unsubFn()

	// Run the initial redial pass on start,
	// in case persistent peer connections are not
	// active
	sw.queueMissingPersistentPeers(attempts, time.Now())

	for {
		select {
		case <-ctx.Done():
			sw.Logger.Debug("redial crawl context canceled")

			return
		case <-ticker.C:
			sw.queueMissingPersistentPeers(attempts, time.Now())
		case event := <-subCh:
			// A persistent peer reconnected, clear its backoff.
			// A peer that connects and drops at once is then redialed at most
			// once per tick: the tick rate-limits it
			ev := event.(events.PeerConnectedEvent)

			delete(attempts, ev.PeerID)
		}
	}
}

// queueMissingPersistentPeers queues a dial for every persistent peer that is
// neither connected nor already queued, on its configured address. The first
// dial is due right away, and every later one waits for a backoff that doubles
// with each attempt, up to persistentRedialMaxBackoff. Persistent peers are
// exempt from the outbound peer limit, as in addPeer
func (sw *MultiplexSwitch) queueMissingPersistentPeers(attempts map[types.ID]uint, now time.Time) {
	peers := sw.Peers()

	sw.persistentPeers.Range(func(key, value any) bool {
		var (
			id   = key.(types.ID)
			addr = value.(*types.NetAddress)
		)

		// Skip peers that are connected or already queued, and our own
		// address, which a shared persistent peer list can contain
		if peers.Has(id) ||
			sw.persistentDialQueue.Has(addr) ||
			addr.Same(sw.transport.NetAddress()) {
			return true
		}

		dialTime := now

		if n, attempted := attempts[id]; attempted {
			// Subsequent attempt: apply backoff
			dialTime = now.Add(
				calculateBackoff(
					n,
					time.Second,
					persistentRedialMaxBackoff,
				),
			)

			attempts[id] = n + 1
		} else {
			// First attempt
			attempts[id] = 0
		}

		sw.persistentDialQueue.Push(dial.Item{
			Time:    dialTime,
			Address: addr,
		})
		sw.notifyAddPeerToDial()

		return true
	})
}

// runSeedDialLoop starts the seed node dial loop.
// Seeds are bootstrap peers: they are dialed on node start, and afterwards
// only when the switch has run out of peers to dial. The loop ticks on a fixed
// interval, which doubles as the minimum delay between two dial rounds
func (sw *MultiplexSwitch) runSeedDialLoop(ctx context.Context) {
	ticker := time.NewTicker(seedDialInterval)
	defer ticker.Stop()

	// Run the initial seed dial round on start, so a fresh node has an entry
	// point into the network. Bootstrap and fallback share a single path, and
	// the same outbound slot accounting
	sw.dialSeed()

	for {
		select {
		case <-ctx.Done():
			sw.Logger.Debug("seed dial context canceled")

			return
		case <-ticker.C:
			sw.dialSeed()
		}
	}
}

// hasDialableItem returns a flag indicating if either dial queue holds an item
// that can be dialed right now. peekDialItem returns a due item whenever there
// is one, so a returned item scheduled in the future means every queued item is
// currently backing off
func (sw *MultiplexSwitch) hasDialableItem() bool {
	item, _ := sw.peekDialItem()

	return item != nil && !time.Now().Before(item.Time)
}

// dialSeed dials a single seed node, picked at random, if the switch has
// nothing left to dial. Seeds go through the regular outbound dial path, so
// they are subject to the maximum outbound peer limit like any other peer
func (sw *MultiplexSwitch) dialSeed() {
	peers := sw.Peers()

	// Seeds exist to fill open outbound slots. With none available,
	// there is nothing a seed could contribute
	if peers.NumOutbound() >= sw.maxOutboundPeers {
		return
	}

	// Check if there is anything left to dial.
	// As long as the switch has dialable peers, the seeds are not needed
	if sw.hasDialableItem() {
		return
	}

	// Gather the seeds that are neither connected nor already queued. A seed
	// that is also a persistent peer is left to the redial loop
	candidates := make([]*types.NetAddress, 0)

	sw.seeds.Range(func(key, value any) bool {
		var (
			id   = key.(types.ID)
			addr = value.(*types.NetAddress)
		)

		if !peers.Has(id) && !sw.dialQueue.Has(addr) && !sw.isPersistentPeer(id) {
			candidates = append(candidates, addr)
		}

		return true
	})

	if len(candidates) == 0 {
		// No seed is worth dialing
		return
	}

	// Dial a single seed. Queuing every seed at once would fill the outbound
	// peer slots with bootstrap connections; if this one turns out to be
	// unreachable, the next round picks another candidate
	addr := candidates[randomIndex(len(candidates))]

	sw.Logger.Info(
		"dialing seed node",
		"address", addr.String(),
	)

	sw.DialPeers(addr)
}

// randomIndex returns a random index within [0, n).
// It falls back to the first index if the random source is unavailable
func randomIndex(n int) int {
	if n <= 1 {
		return 0
	}

	index, err := rand.Int(
		rand.Reader,
		big.NewInt(int64(n)),
	)
	if err != nil {
		return 0
	}

	return int(index.Int64())
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

	// Calculate the interval by exponentiating the base interval by the number of attempts,
	// capped at maxInterval. The shift is only taken when its result fits under maxInterval:
	// past 63 bits it wraps around, to a negative or a meaningless interval
	interval := maxInterval
	if baseInterval <= maxInterval>>attempts {
		interval = baseInterval << attempts
	}

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
// Persistent peers are left to the redial loop, which dials them on their
// configured address.
// To monitor dial progress, subscribe to adequate p2p MultiplexSwitch events
func (sw *MultiplexSwitch) DialPeers(peerAddrs ...*types.NetAddress) {
	peers := sw.Peers()

	for _, peerAddr := range peerAddrs {
		// Check if this is our address
		if peerAddr.Same(sw.transport.NetAddress()) {
			continue
		}

		// Check if this is a persistent peer
		if sw.isPersistentPeer(peerAddr.ID) {
			continue
		}

		// Check if the peer is already connected, or the address already
		// queued. Peer exchange shares the same addresses over and over
		if peers.Has(peerAddr.ID) || sw.dialQueue.Has(peerAddr) {
			continue
		}

		// Ignore dial if the limit is reached
		if out := peers.NumOutbound(); out >= sw.maxOutboundPeers {
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

// keepsRegistered reports whether the connection already registered for a peer
// is kept over an incoming connection to the same peer. Two connections in
// opposite directions are resolved the same way on both ends: the one dialed
// by the node with the lower ID survives. Two connections in the same
// direction keep the registered one, and so does a switch that does not know
// its own ID yet
func (sw *MultiplexSwitch) keepsRegistered(registered, incoming PeerConn) bool {
	if registered.IsOutbound() == incoming.IsOutbound() {
		return true
	}

	var (
		local  = sw.transport.NetAddress().ID
		remote = registered.ID()
	)

	if local == "" || local == remote {
		return true
	}

	// The node with the lower ID dialed the connection both ends keep
	keepOutbound := local < remote

	return registered.IsOutbound() == keepOutbound
}

// direction names the direction of a peer connection, for logs
func direction(p PeerConn) string {
	if p.IsOutbound() {
		return "outbound"
	}

	return "inbound"
}

// resolveDuplicate returns the connection registered for the incoming
// connection's peer, if any, and whether the tie-break keeps it over the
// incoming one. kept is false when nothing is registered
func (sw *MultiplexSwitch) resolveDuplicate(incoming PeerConn) (registered PeerConn, kept bool) {
	registered = sw.peers.Get(incoming.ID())
	if registered == nil {
		return nil, false
	}

	return registered, sw.keepsRegistered(registered, incoming)
}

// registerPeer adds the peer to the peer set. When a connection to the same
// peer is already registered, the tie-break decides: either the registered
// connection is kept and the peer is refused with errDuplicatePeer, or the
// peer takes its place and the replaced connection is returned, for the
// caller to tear down
func (sw *MultiplexSwitch) registerPeer(p PeerConn) (PeerConn, error) {
	sw.registry.Lock()
	defer sw.registry.Unlock()

	registered, kept := sw.resolveDuplicate(p)
	if registered == nil {
		return nil, sw.peers.Add(p)
	}

	if kept {
		return nil, errDuplicatePeer
	}

	sw.peers.Remove(p.ID())

	return registered, sw.peers.Add(p)
}

// removeUnlessSuperseded removes the peer set entry of the peer's ID, unless a
// different connection holds it. It reports false only when a different
// connection holds the entry, and true otherwise, including when no entry
// exists. Checking and removing in one step keeps a connection that replaced
// this one between the two from losing its entry
func (sw *MultiplexSwitch) removeUnlessSuperseded(p PeerConn) bool {
	sw.registry.Lock()
	defer sw.registry.Unlock()

	if sw.isSuperseded(p) {
		return false
	}

	sw.peers.Remove(p.ID())

	return true
}

// hasPeerFromIP returns a flag indicating if the active peer set already
// contains a peer connected from the given IP, other than the peer with the
// given ID. A connection of that peer does not count, since registering a
// second connection of one ID either replaces the first or refuses the second
func (sw *MultiplexSwitch) hasPeerFromIP(ip net.IP, except types.ID) bool {
	if ip == nil {
		return false
	}

	for _, p := range sw.peers.List() {
		if p.ID() != except && ip.Equal(p.RemoteIP()) {
			return true
		}
	}

	return false
}

// rejectConn drops a connection the switch has decided not to keep: an inbound
// one the accept loop refuses, or a dialed one addPeer refuses before starting it.
//
// transport.Remove only forgets the connection; the socket the STS handshake
// established has to be closed explicitly, or -- since a rejected peer was never
// started, so no Stop() path runs -- it lingers until the netFD finalizer does
// it. That lets a host open connections faster than the GC reclaims them.
//
// It also covers a dialed peer whose Start failed, which leaks the same way,
// and one stopped while being added (errPeerStopped), whose connection is
// already closed: the second close then only yields the Debug line below.
func (sw *MultiplexSwitch) rejectConn(p PeerConn) {
	sw.transport.Remove(p)

	if err := p.CloseConn(); err != nil {
		sw.Logger.Debug(
			"unable to close rejected peer connection",
			"peer", p,
			"err", err,
		)
	}
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
		case errors.Is(err, errTransportClosed):
			// Underlaying transport as been closed
			sw.Logger.Warn("cannot accept connection on closed transport, exiting")
			return // exit
		default:
			// An error occurred during accept, report and continue
			sw.Logger.Error("error encountered during peer connection accept", "err", err)
			continue
		}

		// A second connection to an already connected peer is refused, unless
		// it wins the tie-break of a simultaneous open and replaces the
		// registered one
		registered, kept := sw.resolveDuplicate(p)
		if kept {
			sw.Logger.Info(
				"Ignoring inbound connection: already connected",
				"address", p.SocketAddr(),
				"id", p.ID(),
				"kept", direction(registered),
			)

			sw.rejectConn(p)
			continue
		}

		// A replacement swaps one of our outbound connections for an inbound
		// one, so it takes an inbound slot like any other connection. Only a
		// persistent peer's replacement is exempt, so two persistent peers
		// dialing each other at once can still converge on a node whose
		// inbound slots are full. The exemption holds while the connection
		// being replaced is still registered when this check runs; if it is
		// already gone, the new connection goes through the inbound limit
		// like a new peer and the redial loop recovers. Persistent peers are
		// configured by the operator, so no remote party can use the exemption
		exempt := registered != nil && sw.isPersistentPeer(p.ID())

		// Ignore connection if we already have enough peers.
		if in := sw.Peers().NumInbound(); !exempt && in >= sw.maxInboundPeers {
			sw.Logger.Info(
				"Ignoring inbound connection: already have enough inbound peers",
				"address", p.SocketAddr(),
				"have", in,
				"max", sw.maxInboundPeers,
			)

			sw.rejectConn(p)
			continue
		}

		// Reject a second connection from an IP that already holds a peer slot.
		// Peer IDs are self-generated node keys, so without this a single host
		// can mint fresh identities and occupy every inbound slot. The
		// connection of the arriving peer's own ID does not count: registration
		// resolves two connections of one ID, either replacing the registered
		// one, which then leaves the peer set, or refusing the new one. Our dial
		// may also have reached the peer on another IP than its inbound
		// connection (behind a NAT, or on a multi-homed host)
		if !sw.allowDuplicateIP && sw.hasPeerFromIP(p.RemoteIP(), p.ID()) {
			sw.Logger.Info(
				"Ignoring inbound connection: peer from this IP already connected",
				"address", p.SocketAddr(),
				"id", p.ID(),
			)

			sw.rejectConn(p)
			continue
		}

		// There are open peer slots, add peers
		if err := sw.addPeer(p); err != nil {
			sw.rejectConn(p)

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

	// Refuse a connection registerPeer would refuse anyway before any reactor
	// sees it, so it neither starts nor leaves reactor state behind. The dial
	// loop's own Has check races the dial it guards, so this is the first
	// point where the check is worth anything. registerPeer stays the
	// authoritative one
	_, kept := sw.resolveDuplicate(p)
	if kept {
		return errDuplicatePeer
	}

	// Enforce the outbound limit where the peer is actually added. DialPeers
	// only checks it when an address is queued, and NumOutbound cannot change
	// while that loop runs, so a single batch of queued dials would otherwise
	// overshoot the limit without bound. A connection replacing a registered
	// inbound one takes an outbound slot like any other. Persistent peers are
	// exempt, as MaxNumOutboundPeers documents
	if p.IsOutbound() && !sw.isPersistentPeer(p.ID()) {
		if out := sw.peers.NumOutbound(); out >= sw.maxOutboundPeers {
			sw.Logger.Info(
				"Ignoring outbound connection: already have max outbound peers",
				"have", out,
				"max", sw.maxOutboundPeers,
				"id", p.ID(),
			)

			return errMaxOutboundPeers
		}
	}

	// Add some data to the peer, which is required by reactors.
	for _, reactor := range sw.reactors {
		p = reactor.InitPeer(p)
	}

	// Start the peer's send/recv routines.
	// Must start it before adding it to the peer set
	// to prevent Start and Stop from being called concurrently.
	if err := p.Start(); err != nil {
		sw.Logger.Error("Error starting peer", "err", err, "peer", p)

		sw.removeReactorPeerState(p, err)

		return err
	}

	// Add the peer to the peer set. Do this before starting the reactors
	// so that if Receive errors, we will find the peer and remove it.
	replaced, err := sw.registerPeer(p)

	// Each connection gets its RemovePeer exactly once. The connection p
	// replaced gets it here, before p's AddPeer, unless its own teardown,
	// after the remote closed it, began first, in which case that RemovePeer
	// may run concurrently with p's AddPeer. p holds the entry, so the
	// teardown leaves that entry alone
	if replaced != nil {
		sw.Logger.Info(
			"replacing connection to resolve a simultaneous open",
			"peer", p,
			"kept", direction(p),
		)

		sw.stopAndRemovePeer(replaced, errSimultaneousOpen)
	}

	if err != nil {
		sw.removeReactorPeerState(p, err)

		return err
	}

	// The peer can have been stopped while it was being added: the recv
	// routine p.Start() spawned reports an error to stopAndRemovePeer, which
	// removes from the peer set last, so its Remove can have run before the
	// registration above. Adding a stopped peer would hold its slot and its ID
	// for the lifetime of the process, since nothing removes a peer twice.
	//
	// Its reactor state needs no unwinding here: whatever stopped the peer
	// walked the reactors' RemovePeer on the way
	if !p.IsRunning() {
		sw.removeUnlessSuperseded(p)

		return errPeerStopped
	}

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

// waitForDialTime waits for the given duration to elapse, for a new item to be
// queued for dialing, or for the context to be canceled, whichever comes first.
// The dial notification cannot be ignored, since a newly queued item can be due
// before the one currently at the head of the queue
func (sw *MultiplexSwitch) waitForDialTime(ctx context.Context, wait time.Duration) {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	case <-sw.dialNotify:
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
