package discovery

import (
	"net"

	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/events"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type (
	broadcastDelegate        func(byte, []byte)
	peersDelegate            func() p2p.PeerSet
	stopPeerForErrorDelegate func(p2p.PeerConn, error)
	dialPeersDelegate        func(...*types.NetAddress)
	subscribeDelegate        func(events.EventFilter) (<-chan events.Event, func())
)

type mockSwitch struct {
	broadcastFn        broadcastDelegate
	peersFn            peersDelegate
	stopPeerForErrorFn stopPeerForErrorDelegate
	dialPeersFn        dialPeersDelegate
	subscribeFn        subscribeDelegate
}

func (m *mockSwitch) Broadcast(chID byte, data []byte) {
	if m.broadcastFn != nil {
		m.broadcastFn(chID, data)
	}
}

func (m *mockSwitch) Peers() p2p.PeerSet {
	if m.peersFn != nil {
		return m.peersFn()
	}

	return nil
}

func (m *mockSwitch) StopPeerForError(peer p2p.PeerConn, err error) {
	if m.stopPeerForErrorFn != nil {
		m.stopPeerForErrorFn(peer, err)
	}
}

func (m *mockSwitch) DialPeers(peerAddrs ...*types.NetAddress) {
	if m.dialPeersFn != nil {
		m.dialPeersFn(peerAddrs...)
	}
}

func (m *mockSwitch) Subscribe(filter events.EventFilter) (<-chan events.Event, func()) {
	if m.subscribeFn != nil {
		m.subscribeFn(filter)
	}

	return nil, func() {}
}

type (
	addDelegate         func(p2p.PeerConn)
	removeDelegate      func(types.ID) bool
	hasDelegate         func(types.ID) bool
	hasIPDelegate       func(net.IP) bool
	getPeerDelegate     func(types.ID) p2p.PeerConn
	listDelegate        func() []p2p.PeerConn
	numInboundDelegate  func() uint64
	numOutboundDelegate func() uint64
)

type mockPeerSet struct {
	addFn         addDelegate
	removeFn      removeDelegate
	hasFn         hasDelegate
	hasIPFn       hasIPDelegate
	getFn         getPeerDelegate
	listFn        listDelegate
	numInboundFn  numInboundDelegate
	numOutboundFn numOutboundDelegate
}

func (m *mockPeerSet) Add(peer p2p.PeerConn) {
	if m.addFn != nil {
		m.addFn(peer)
	}
}

func (m *mockPeerSet) Remove(key types.ID) bool {
	if m.removeFn != nil {
		m.removeFn(key)
	}

	return false
}

func (m *mockPeerSet) Has(key types.ID) bool {
	if m.hasFn != nil {
		return m.hasFn(key)
	}

	return false
}

func (m *mockPeerSet) Get(key types.ID) p2p.PeerConn {
	if m.getFn != nil {
		return m.getFn(key)
	}

	return nil
}

func (m *mockPeerSet) List() []p2p.PeerConn {
	if m.listFn != nil {
		return m.listFn()
	}

	return nil
}

func (m *mockPeerSet) NumInbound() uint64 {
	if m.numInboundFn != nil {
		return m.numInboundFn()
	}

	return 0
}

func (m *mockPeerSet) NumOutbound() uint64 {
	if m.numOutboundFn != nil {
		return m.numOutboundFn()
	}

	return 0
}
