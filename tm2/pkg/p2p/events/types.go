package events

import (
	"net"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type EventType string

const (
	PeerConnected    EventType = "PeerConnected"    // emitted when a fresh peer connects
	PeerDisconnected EventType = "PeerDisconnected" // emitted when a peer disconnects
)

// Event is a generic p2p event
type Event interface {
	// Type returns the type information for the event
	Type() EventType
}

type PeerConnectedEvent struct {
	PeerID  types.ID // the ID of the peer
	Address net.Addr // the remote address of the peer
}

func (p PeerConnectedEvent) Type() EventType {
	return PeerConnected
}

type PeerDisconnectedEvent struct {
	PeerID  types.ID // the ID of the peer
	Address net.Addr // the remote address of the peer
	Reason  error    // the disconnect reason, if any
}

func (p PeerDisconnectedEvent) Type() EventType {
	return PeerDisconnected
}
