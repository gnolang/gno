package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/cmap"
	"github.com/gnolang/gno/tm2/pkg/p2p/config"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"golang.org/x/sync/errgroup"
)

const (
	// wait a random amount of time from this interval
	// before dialing peers or reconnecting to help prevent DoS
	dialRandomizerIntervalMilliseconds = 3000

	// repeatedly try to reconnect for a few minutes
	// ie. 5 * 20 = 100s
	reconnectAttempts = 20
	reconnectInterval = 5 * time.Second

	// then move into exponential backoff mode for ~1day
	// ie. 3**10 = 16hrs
	reconnectBackOffAttempts    = 10
	reconnectBackOffBaseSeconds = 3
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

// PeerFilterFunc to be implemented by filter hooks after a new Peer has been
// fully setup.
type PeerFilterFunc func(PeerSet, Peer) error

// Switch handles peer connections and exposes an API to receive incoming messages
// on `Reactors`.  Each `Reactor` is responsible for handling incoming messages of one
// or more `Channels`.  So while sending outgoing messages is typically performed on the peer,
// incoming messages are received on the reactor.
type Switch struct {
	service.BaseService

	config       *config.P2PConfig // TODO remove this dependency
	reactors     map[string]Reactor
	chDescs      []*conn.ChannelDescriptor
	reactorsByCh map[byte]Reactor

	dialing      *cmap.CMap
	reconnecting *cmap.CMap

	nodeInfo NodeInfo // our node info
	nodeKey  *NodeKey // our node privkey

	peers           PeerSet  // currently active peer set
	persistentPeers sync.Map // ID -> *NetAddress; peers whose connections are constant
	transport       Transport

	filterTimeout time.Duration
	peerFilters   []PeerFilterFunc
}

// NetAddress returns the address the switch is listening on.
func (sw *Switch) NetAddress() *NetAddress {
	addr := sw.transport.NetAddress()

	return &addr
}

// SwitchOption sets an optional parameter on the Switch.
type SwitchOption func(*Switch)

// NewSwitch creates a new Switch with the given config.
func NewSwitch(
	cfg *config.P2PConfig,
	transport Transport,
	options ...SwitchOption,
) *Switch {
	sw := &Switch{
		config:        cfg,
		reactors:      make(map[string]Reactor),
		chDescs:       make([]*conn.ChannelDescriptor, 0),
		reactorsByCh:  make(map[byte]Reactor),
		peers:         NewSet(),
		dialing:       cmap.NewCMap(),
		reconnecting:  cmap.NewCMap(),
		transport:     transport,
		filterTimeout: defaultFilterTimeout,
	}

	sw.BaseService = *service.NewBaseService(nil, "P2P Switch", sw)

	for _, option := range options {
		option(sw)
	}

	return sw
}

// AddReactor adds the given reactor to the switch.
// NOTE: Not goroutine safe.
func (sw *Switch) AddReactor(name string, reactor Reactor) Reactor {
	for _, chDesc := range reactor.GetChannels() {
		chID := chDesc.ID
		// No two reactors can share the same channel.
		if sw.reactorsByCh[chID] != nil {
			panic(fmt.Sprintf("Channel %X has multiple reactors %v & %v", chID, sw.reactorsByCh[chID], reactor))
		}

		sw.chDescs = append(sw.chDescs, chDesc)
		sw.reactorsByCh[chID] = reactor
	}

	sw.reactors[name] = reactor

	reactor.SetSwitch(sw)

	return reactor
}

// RemoveReactor removes the given Reactor from the Switch.
// NOTE: Not goroutine safe.
func (sw *Switch) RemoveReactor(name string, reactor Reactor) {
	for _, chDesc := range reactor.GetChannels() {
		// remove channel description
		for i := 0; i < len(sw.chDescs); i++ {
			if chDesc.ID == sw.chDescs[i].ID {
				sw.chDescs = append(sw.chDescs[:i], sw.chDescs[i+1:]...)
				break
			}
		}

		delete(sw.reactorsByCh, chDesc.ID)
	}

	delete(sw.reactors, name)

	reactor.SetSwitch(nil)
}

// Reactor returns the reactor with the given name.
// NOTE: Not goroutine safe.
func (sw *Switch) Reactor(name string) Reactor {
	return sw.reactors[name]
}

// NodeInfo returns the switch's NodeInfo.
// NOTE: Not goroutine safe.
func (sw *Switch) NodeInfo() NodeInfo {
	return sw.nodeInfo
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
func (sw *Switch) StopPeerForError(peer Peer, reason interface{}) {
	sw.Logger.Error("Stopping peer for error", "peer", peer, "err", reason)
	sw.stopAndRemovePeer(peer, reason)

	if peer.IsPersistent() {
		var addr *NetAddress
		if peer.IsOutbound() { // socket address for outbound peers
			addr = peer.SocketAddr()
		} else { // self-reported address for inbound peers
			addr = peer.NodeInfo().NetAddress
		}
		go sw.reconnectToPeer(addr)
	}
}

func (sw *Switch) stopAndRemovePeer(peer Peer, reason interface{}) {
	sw.transport.Cleanup(peer)
	peer.Stop()

	for _, reactor := range sw.reactors {
		reactor.RemovePeer(peer, reason)
	}

	// Removing a peer should go last to avoid a situation where a peer
	// reconnect to our node and the switch calls InitPeer before
	// RemovePeer is finished.
	// https://github.com/tendermint/classic/issues/3338
	sw.peers.Remove(peer.ID())
}

// reconnectToPeer tries to reconnect to the addr, first repeatedly
// with a fixed interval, then with exponential backoff.
// If no success after all that, it stops trying.
// NOTE: this will keep trying even if the handshake or auth fails.
// TODO: be more explicit with error types so we only retry on certain failures
//   - ie. if we're getting ErrDuplicatePeer we can stop
func (sw *Switch) reconnectToPeer(addr *NetAddress) {
	if sw.reconnecting.Has(addr.ID.String()) {
		return
	}
	sw.reconnecting.Set(addr.ID.String(), addr)
	defer sw.reconnecting.Delete(addr.ID.String())

	start := time.Now()
	sw.Logger.Info("Reconnecting to peer", "addr", addr)
	for i := 0; i < reconnectAttempts; i++ {
		if !sw.IsRunning() {
			return
		}

		err := sw.dialPeerWithAddress(addr)
		if err == nil {
			return // success
		} else if _, ok := err.(CurrentlyDialingOrExistingAddressError); ok {
			return
		}

		sw.Logger.Info("Error reconnecting to peer. Trying again", "tries", i, "err", err, "addr", addr)
		// sleep a set amount
		sw.randomSleep(reconnectInterval)
		continue
	}

	sw.Logger.Error("Failed to reconnect to peer. Beginning exponential backoff",
		"addr", addr, "elapsed", time.Since(start))
	for i := 0; i < reconnectBackOffAttempts; i++ {
		if !sw.IsRunning() {
			return
		}

		// sleep an exponentially increasing amount
		sleepIntervalSeconds := math.Pow(reconnectBackOffBaseSeconds, float64(i))
		sw.randomSleep(time.Duration(sleepIntervalSeconds) * time.Second)

		err := sw.dialPeerWithAddress(addr)
		if err == nil {
			return // success
		} else if _, ok := err.(CurrentlyDialingOrExistingAddressError); ok {
			return
		}
		sw.Logger.Info("Error reconnecting to peer. Trying again", "tries", i, "err", err, "addr", addr)
	}
	sw.Logger.Error("Failed to reconnect to peer. Giving up", "addr", addr, "elapsed", time.Since(start))
}

// ---------------------------------------------------------------------
// Dialing

// DialPeersAsync dials a list of peers asynchronously in random order.
// Used to dial peers from config on startup or from unsafe-RPC (trusted sources).
// It ignores NetAddressLookupError. However, if there are other errors, first
// encounter is returned.
// Nop if there are no peers.
func (sw *Switch) DialPeersAsync(peers []string) error {
	netAddrs, errs := NewNetAddressFromStrings(peers)
	// report all the errors
	for _, err := range errs {
		sw.Logger.Error("Error in peer's address", "err", err)
	}
	// return first non-NetAddressLookupError error
	for _, err := range errs {
		if _, ok := err.(NetAddressLookupError); ok {
			continue
		}
		return err
	}
	sw.dialPeersAsync(netAddrs)
	return nil
}

func (sw *Switch) dialPeersAsync(netAddrs []*NetAddress) {
	var (
		ourAddr = sw.NetAddress()

		wg sync.WaitGroup
	)

	wg.Add(len(netAddrs))

	for _, peerAddr := range netAddrs {
		go func(addr *NetAddress) {
			defer wg.Done()

			if addr.Same(ourAddr) {
				sw.Logger.Debug(
					"ignoring self-dial attempt",
					"addr",
					addr,
				)

				return
			}

			sw.randomSleep(0) // TODO remove this

			if err := sw.dialPeerWithAddress(addr); err != nil {
				sw.Logger.Debug("Error dialing peer", "err", err)
			}
		}(peerAddr)
	}

	wg.Wait()
}

// dialPeerWithAddress dials the given peer and runs sw.addPeer if it connects
// and authenticates successfully.
// If we're currently dialing this address or it belongs to an existing peer,
// CurrentlyDialingOrExistingAddressError is returned.
func (sw *Switch) dialPeerWithAddress(addr *NetAddress) error {
	if sw.isDialingOrExistingAddress(addr) {
		return CurrentlyDialingOrExistingAddressError{addr.String()}
	}

	// TODO clean up
	sw.dialing.Set(addr.ID.String(), addr)
	defer sw.dialing.Delete(addr.ID.String())

	return sw.addOutboundPeerWithConfig(addr)
}

// sleep for interval plus some random amount of ms on [0, dialRandomizerIntervalMilliseconds]
func (sw *Switch) randomSleep(interval time.Duration) {
	r, err := rand.Int(rand.Reader, big.NewInt(dialRandomizerIntervalMilliseconds))
	if err != nil {
		sw.Logger.Error("unable to generate random sleep value", "err", err)

		return
	}

	duration := time.Duration(r.Uint64()) * time.Millisecond

	time.Sleep(duration + interval)
}

// isDialingOrExistingAddress returns true if switch has a peer with the given
// address or dialing it at the moment.
func (sw *Switch) isDialingOrExistingAddress(addr *NetAddress) bool {
	return sw.dialing.Has(addr.ID.String()) ||
		sw.peers.Has(addr.ID) ||
		(!sw.config.AllowDuplicateIP && sw.peers.HasIP(addr.IP))
}

// AddPersistentPeers allows you to set persistent peers. It ignores
// NetAddressLookupError. However, if there are other errors, first encounter is
// returned.
// TODO change to net addresses
// TODO make an option
func (sw *Switch) AddPersistentPeers(addrs []string) error {
	sw.Logger.Info("Adding persistent peers", "addrs", addrs)
	netAddrs, errs := NewNetAddressFromStrings(addrs)
	// report all the errors
	for _, err := range errs {
		sw.Logger.Error("Error in peer's address", "err", err)
	}
	// return first non-NetAddressLookupError error
	for _, err := range errs {
		if _, ok := err.(NetAddressLookupError); ok {
			continue
		}
		return err
	}

	// Set the persistent peers
	for _, peerAddr := range netAddrs {
		sw.persistentPeers.Store(peerAddr.ID, peerAddr)
	}

	return nil
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

// dial the peer; make secret connection; authenticate against the dialed ID;
// add the peer.
// if dialing fails, start the reconnect loop. If handshake fails, it's over.
// If peer is started successfully, reconnectLoop will start when
// StopPeerForError is called.
func (sw *Switch) addOutboundPeerWithConfig(addr *NetAddress) error {
	sw.Logger.Info("Dialing peer", "address", addr)

	p, err := sw.transport.Dial(*addr, peerConfig{
		chDescs:     sw.chDescs,
		onPeerError: sw.StopPeerForError,
		isPersistent: func(address *NetAddress) bool {
			return sw.isPersistentPeer(address.ID)
		},
		reactorsByCh: sw.reactorsByCh,
	})
	if err != nil {
		if e, ok := err.(RejectedError); ok {
			if e.IsSelf() {
				// TODO: warn?
				return err
			}
		}

		// retry persistent peers after
		// any dial error besides IsSelf()
		if sw.isPersistentPeer(addr.ID) {
			go sw.reconnectToPeer(addr)
		}

		return err
	}

	if err := sw.addPeer(p); err != nil {
		sw.transport.Cleanup(p)
		if p.IsRunning() {
			_ = p.Stop()
		}
		return err
	}

	return nil
}

// TODO remove this entirely
func (sw *Switch) filterPeer(p Peer) error {
	// Avoid duplicate
	if sw.peers.Has(p.ID()) {
		return RejectedError{
			id:          p.ID(),
			isDuplicate: true,
		}
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), sw.filterTimeout)
	defer cancelFn()

	g, _ := errgroup.WithContext(ctx)

	for _, filterFn := range sw.peerFilters {
		g.Go(func() error {
			return filterFn(sw.peers, p)
		})
	}

	if err := g.Wait(); err != nil {
		return RejectedError{id: p.ID(), err: err, isFiltered: true}
	}

	return nil
}

// addPeer starts up the Peer and adds it to the Switch. Error is returned if
// the peer is filtered out or failed to start or can't be added.
func (sw *Switch) addPeer(p Peer) error {
	if err := sw.filterPeer(p); err != nil {
		return err
	}

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

	// Update the telemetry data
	sw.logTelemetry()

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

	// Log the dialing peer count
	metrics.DialingPeers.Record(context.Background(), int64(sw.dialing.Size()))
}
