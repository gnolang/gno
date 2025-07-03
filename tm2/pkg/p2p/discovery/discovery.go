package discovery

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

const (
	// Channel is the unique channel for the peer discovery protocol
	Channel = byte(0x50)

	// discoveryInterval is the peer discovery interval, for random peers
	discoveryInterval = time.Second * 3

	// maxPeersShared is the maximum number of peers shared in the discovery request
	maxPeersShared = 30
)

// descriptor is the constant peer discovery protocol descriptor
var descriptor = &conn.ChannelDescriptor{
	ID:                  Channel,
	Priority:            1,       // peer discovery is high priority
	SendQueueCapacity:   20,      // more than enough active conns
	RecvMessageCapacity: 5242880, // 5MB
}

// Reactor wraps the logic for the peer exchange protocol
type Reactor struct {
	// This embed and the usage of "services"
	// like the peer discovery reactor highlight the
	// flipped design of the p2p package.
	// The peer exchange service needs to be instantiated _outside_
	// the p2p module, because of this flipped design.
	// Peers communicate with each other through Reactor channels,
	// which are instantiated outside the p2p module
	p2p.BaseReactor

	ctx      context.Context
	cancelFn context.CancelFunc

	discoveryInterval time.Duration
}

// NewReactor creates a new peer discovery reactor
func NewReactor(opts ...Option) *Reactor {
	ctx, cancelFn := context.WithCancel(context.Background())

	r := &Reactor{
		ctx:               ctx,
		cancelFn:          cancelFn,
		discoveryInterval: discoveryInterval,
	}

	r.BaseReactor = *p2p.NewBaseReactor("Reactor", r)

	// Apply the options
	for _, opt := range opts {
		opt(r)
	}

	return r
}

// OnStart runs the peer discovery protocol
func (r *Reactor) OnStart() error {
	go func() {
		ticker := time.NewTicker(r.discoveryInterval)
		defer ticker.Stop()

		for {
			select {
			case <-r.ctx.Done():
				r.Logger.Debug("discovery service stopped")

				return
			case <-ticker.C:
				// Run the discovery protocol //

				// Grab a random peer, and engage
				// them for peer discovery
				peers := r.Switch.Peers().List()

				if len(peers) == 0 {
					// No discovery to run
					continue
				}

				// Generate a random peer index
				randomPeer, _ := rand.Int(
					rand.Reader,
					big.NewInt(int64(len(peers))),
				)

				// Request peers, async
				go r.requestPeers(peers[randomPeer.Int64()])
			}
		}
	}()

	return nil
}

// OnStop stops the peer discovery protocol
func (r *Reactor) OnStop() {
	r.cancelFn()
}

// requestPeers requests the peer set from the given peer
func (r *Reactor) requestPeers(peer p2p.PeerConn) {
	// Initiate peer discovery
	r.Logger.Debug("running peer discovery", "peer", peer.ID())

	// Prepare the request
	// (empty, as it's a notification)
	req := &Request{}

	reqBytes, err := amino.MarshalAny(req)
	if err != nil {
		r.Logger.Error("unable to marshal discovery request", "err", err)

		return
	}

	// Send the request
	if !peer.Send(Channel, reqBytes) {
		r.Logger.Warn("unable to send discovery request", "peer", peer.ID())
	}
}

// GetChannels returns the channels associated with peer discovery
func (r *Reactor) GetChannels() []*conn.ChannelDescriptor {
	return []*conn.ChannelDescriptor{descriptor}
}

// Receive handles incoming messages for the peer discovery reactor
func (r *Reactor) Receive(chID byte, peer p2p.PeerConn, msgBytes []byte) {
	r.Logger.Debug(
		"received message",
		"peerID", peer.ID(),
		"chID", chID,
	)

	// Unmarshal the message
	var msg Message

	if err := amino.UnmarshalAny(msgBytes, &msg); err != nil {
		r.Logger.Error("unable to unmarshal discovery message", "err", err)

		return
	}

	// Validate the message
	if err := msg.ValidateBasic(); err != nil {
		r.Logger.Warn("unable to validate discovery message", "err", err)

		return
	}

	switch msg := msg.(type) {
	case *Request:
		if err := r.handleDiscoveryRequest(peer); err != nil {
			r.Logger.Warn("unable to handle discovery request", "err", err)
		}
	case *Response:
		// Make the peers available for dialing on the switch
		r.Switch.DialPeers(msg.Peers...)
	default:
		r.Logger.Warn("invalid message received", "msg", msgBytes)
	}
}

// handleDiscoveryRequest prepares a peer list that can be shared
// with the peer requesting discovery
func (r *Reactor) handleDiscoveryRequest(peer p2p.PeerConn) error {
	var (
		localPeers = r.Switch.Peers().List()
		peers      = make([]*types.NetAddress, 0, len(localPeers))
	)

	// Exclude the private peers from being shared,
	// as well as peers who are not dialable
	localPeers = slices.DeleteFunc(localPeers, func(p p2p.PeerConn) bool {
		var (
			// Private peers are peers whose information is kept private to the node
			privatePeer = p.IsPrivate()
			// The reason we don't validate the net address with .Routable()
			// is because of legacy logic that supports local loopbacks as advertised
			// peer addresses. Introducing a .Routable() constraint will filter all
			// local loopback addresses shared by peers, and will cause local deployments
			// (and unit test deployments) to break and require additional setup
			invalidDialAddress = p.NodeInfo().DialAddress().Validate() != nil
		)

		return privatePeer || invalidDialAddress
	})

	// Check if there is anything to share,
	// to avoid useless traffic
	if len(localPeers) == 0 {
		r.Logger.Warn("no peers to share in discovery request")

		return nil
	}

	// Shuffle and limit the peers shared
	shufflePeers(localPeers)

	if len(localPeers) > maxPeersShared {
		localPeers = localPeers[:maxPeersShared]
	}

	for _, p := range localPeers {
		// Make sure only routable peers are shared
		peers = append(peers, p.NodeInfo().DialAddress())
	}

	// Create the response, and marshal
	// it to Amino binary
	resp := &Response{
		Peers: peers,
	}

	preparedResp, err := amino.MarshalAny(resp)
	if err != nil {
		return fmt.Errorf("unable to marshal discovery response, %w", err)
	}

	// Send the response to the peer
	if !peer.Send(Channel, preparedResp) {
		return fmt.Errorf("unable to send discovery response to peer %s", peer.ID())
	}

	return nil
}

// shufflePeers shuffles the peer list in-place
func shufflePeers(peers []p2p.PeerConn) {
	for i := len(peers) - 1; i > 0; i-- {
		jBig, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))

		j := int(jBig.Int64())

		// Swap elements
		peers[i], peers[j] = peers[j], peers[i]
	}
}
