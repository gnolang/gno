package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/config"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/dial"
	"github.com/gnolang/gno/tm2/pkg/p2p/events"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
)

// MultiplexConfigFromP2P returns a multiplex connection configuration
// with fields updated from the P2PConfig
func MultiplexConfigFromP2P(cfg *config.P2PConfig) conn.MConnConfig {
	mConfig := conn.DefaultMConnConfig()
	mConfig.FlushThrottle = cfg.FlushThrottleTimeout
	mConfig.SendRate = cfg.SendRate
	mConfig.RecvRate = cfg.RecvRate
	mConfig.MaxPacketMsgPayloadSize = cfg.MaxPacketMsgPayloadSize
	return mConfig
}

// Switch handles peer connections and exposes an API to receive incoming messages
// on `Reactors`.  Each `Reactor` is responsible for handling incoming messages of one
// or more `Channels`.  So while sending outgoing messages is typically performed on the peer,
// incoming messages are received on the reactor.
type Switch struct {
	service.BaseService

	config *config.P2PConfig // TODO remove this dependency

	reactors     map[string]Reactor        // TODO wrap
	chDescs      []*conn.ChannelDescriptor // TODO wrap
	reactorsByCh map[byte]Reactor          // TODO wrap

	peers           PeerSet  // currently active peer set
	persistentPeers sync.Map // ID -> *NetAddress; peers whose connections are constant
	transport       Transport

	dialQueue *dial.Queue
	events    *events.Events
}

// NewSwitch creates a new Switch with the given config.
func NewSwitch(
	cfg *config.P2PConfig,
	transport Transport,
	options ...SwitchOption,
) *Switch {
	sw := &Switch{
		config:       cfg,
		reactors:     make(map[string]Reactor),
		chDescs:      make([]*conn.ChannelDescriptor, 0),
		reactorsByCh: make(map[byte]Reactor),
		peers:        NewSet(),
		transport:    transport,
		dialQueue:    dial.NewQueue(),
		events:       events.New(),
	}

	sw.BaseService = *service.NewBaseService(nil, "P2P Switch", sw)

	for _, option := range options {
		option(sw)
	}

	return sw
}

// NetAddress returns the address the switch is listening on.
func (sw *Switch) NetAddress() *NetAddress {
	addr := sw.transport.NetAddress()

	return &addr
}

// Subscribe registers to live events happening on the p2p Switch.
// Returns the notification channel, along with an unsubscribe method
func (sw *Switch) Subscribe(filterFn events.EventFilter) (<-chan events.Event, func()) {
	return sw.events.Subscribe(filterFn)
}

// ---------------------------------------------------------------------
// Service start/stop

// OnStart implements BaseService. It starts all the reactors and peers.
func (sw *Switch) OnStart() error {
	// Start reactors
	for _, reactor := range sw.reactors {
		if err := reactor.Start(); err != nil {
			return fmt.Errorf("unable to start reactor %w", err)
		}
	}

	// Run the peer accept routine
	// TODO propagate ctx down
	go sw.runAcceptLoop(context.Background())

	// Run the dial routine
	// TODO propagate ctx down
	go sw.runDialLoop(context.Background())

	// Run the redial routine
	// TODO propagate ctx down
	go sw.runRedialLoop(context.Background())

	return nil
}

// OnStop implements BaseService. It stops all peers and reactors.
func (sw *Switch) OnStop() {
	// Stop transport
	if err := sw.transport.Close(); err != nil {
		sw.Logger.Error("unable to gracefully close transport", "err", err)
	}

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

// Broadcast broadcasts the given data to the given channel, across the
// entire switch peer set
func (sw *Switch) Broadcast(chID byte, data []byte) {
	var wg sync.WaitGroup

	for _, p := range sw.peers.List() {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// TODO propagate the context, instead of relying
			// on the underlying multiplex conn
			if !p.Send(chID, data) {
				sw.Logger.Error(
					"unable to perform broadcast, channel ID %X, peer ID %s",
					chID, p.ID(),
				)
			}
		}()
	}

	// Wait for all the sends to complete,
	// at the mercy of the multiplex connection
	// send routine :)
	// TODO: I'm not sure Broadcast should be blocking, at all
	wg.Wait()
}

// Peers returns the set of peers that are connected to the switch.
func (sw *Switch) Peers() PeerSet {
	return sw.peers
}

// StopPeerForError disconnects from a peer due to external error.
// If the peer is persistent, it will attempt to reconnect.
// TODO: make record depending on reason.
func (sw *Switch) StopPeerForError(peer Peer, err error) {
	sw.Logger.Error("Stopping peer for error", "peer", peer, "err", err)

	sw.stopAndRemovePeer(peer, err)

	if !peer.IsPersistent() {
		return
	}

	// socket address for outbound peers
	addr := peer.SocketAddr()

	if !peer.IsOutbound() {
		// self-reported address for inbound peers
		addr = peer.NodeInfo().NetAddress
	}

	// Add the peer to the dial queue
	sw.DialPeers(addr)
}

func (sw *Switch) stopAndRemovePeer(peer Peer, err error) {
	// Remove the peer from the transport
	sw.transport.Cleanup(peer)

	// Stop the peer connection multiplexing
	if stopErr := peer.Stop(); stopErr != nil {
		sw.Logger.Error(
			"unable to gracefully stop peer",
			"peer", peer,
			"err", err,
		)
	}

	// Alert the reactors of a peer removal
	for _, reactor := range sw.reactors {
		reactor.RemovePeer(peer, err)
	}

	// Removing a peer should go last to avoid a situation where a peer
	// reconnect to our node and the switch calls InitPeer before
	// RemovePeer is finished.
	// https://github.com/tendermint/classic/issues/3338
	sw.peers.Remove(peer.ID())
}

// ---------------------------------------------------------------------
// Dialing

func (sw *Switch) runDialLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			sw.Logger.Debug("dial context canceled")

			return
		default:
			// Grab a dial item
			item := sw.dialQueue.Peek()
			if item == nil {
				// Nothing to dial
				continue
			}

			// Check if the dial time is right
			// for the item
			if time.Now().Before(item.Time) {
				// Nothing to dial
				continue
			}

			// Dial the peer
			sw.Logger.Info(
				"dialing peer",
				"address", item.Address.String(),
			)

			peerAddr := item.Address

			// TODO pass context to dial
			p, err := sw.transport.Dial(*peerAddr, peerConfig{
				chDescs:     sw.chDescs,
				onPeerError: sw.StopPeerForError,
				isPersistent: func(address *NetAddress) bool {
					return sw.isPersistentPeer(address.ID)
				},
				reactorsByCh: sw.reactorsByCh,
			})
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

				sw.transport.Cleanup(p)

				if !p.IsRunning() {
					// TODO check if this check is even required
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
func (sw *Switch) runRedialLoop(ctx context.Context) {
	// Set up the event subscription for persistent peer disconnects
	subCh, unsubFn := sw.Subscribe(func(event events.Event) bool {
		// Make sure the peer event relates to a peer disconnect
		if event.Type() != events.PeerDisconnected {
			return false
		}

		disconnectEv, ok := event.(*events.PeerDisconnectedEvent)
		if !ok {
			return false
		}

		return sw.isPersistentPeer(disconnectEv.PeerID)
	})
	defer unsubFn()

	for {
		select {
		case <-ctx.Done():
			sw.Logger.Debug("redial context canceled")

			return
		case ev := <-subCh:
			disconnectEv, ok := ev.(*events.PeerDisconnectedEvent)
			if !ok {
				continue
			}

			// Dial the disconnected peer
			// TODO add backoff mechanism
			sw.DialPeers(&disconnectEv.Address)
		}
	}
}

// DialPeers adds the peers to the dial queue for async dialing.
// To monitor dial progress, subscribe to adequate p2p Switch events
func (sw *Switch) DialPeers(peerAddrs ...*NetAddress) {
	for _, peerAddr := range peerAddrs {
		item := dial.Item{
			Time:    time.Now(),
			Address: peerAddr,
		}

		sw.dialQueue.Push(item)
	}
}

// isPersistentPeer returns a flag indicating if a peer
// is present in the persistent peer set
func (sw *Switch) isPersistentPeer(id ID) bool {
	_, persistent := sw.persistentPeers.Load(id)

	return persistent
}

// runAcceptLoop is the main powerhouse method
// for accepting incoming peer connections, filtering them,
// and persisting them
func (sw *Switch) runAcceptLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			sw.Logger.Debug("switch context close received")

			return
		default:
			p, err := sw.transport.Accept(peerConfig{
				chDescs:      sw.chDescs,
				onPeerError:  sw.StopPeerForError,
				reactorsByCh: sw.reactorsByCh,
				isPersistent: func(address *NetAddress) bool {
					return sw.isPersistentPeer(address.ID)
				},
			})
			if err != nil {
				sw.Logger.Error(
					"error encountered during peer connection accept",
					"err", err,
				)

				continue
			}

			// Ignore connection if we already have enough peers.
			if in := sw.Peers().NumInbound(); in >= sw.config.MaxNumInboundPeers {
				sw.Logger.Info(
					"Ignoring inbound connection: already have enough inbound peers",
					"address", p.SocketAddr(),
					"have", in,
					"max", sw.config.MaxNumInboundPeers,
				)

				sw.transport.Cleanup(p)

				continue
			}

			// There are open peer slots, add peers
			if err := sw.addPeer(p); err != nil {
				sw.transport.Cleanup(p)

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
}

// addPeer starts up the Peer and adds it to the Switch. Error is returned if
// the peer is filtered out or failed to start or can't be added.
func (sw *Switch) addPeer(p Peer) error {
	p.SetLogger(sw.Logger.With("peer", p.SocketAddr()))

	// Handle the shut down case where the switch has stopped, but we're
	// concurrently trying to add a peer.
	if !sw.IsRunning() {
		// XXX should this return an error or just log and terminate?
		sw.Logger.Error("Won't start a peer - switch is not running", "peer", p)
		return nil
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

	return nil
}

// logTelemetry logs the switch telemetry data
// to global metrics funnels
func (sw *Switch) logTelemetry() {
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
