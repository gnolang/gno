package discovery

import (
	"net"

	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type (
	broadcastDelegate        func(byte, []byte)
	peersDelegate            func() p2p.PeerSet
	stopPeerForErrorDelegate func(p2p.Peer, error)
	dialPeersDelegate        func(...*types.NetAddress)
)

type mockSwitch struct {
	broadcastFn        broadcastDelegate
	peersFn            peersDelegate
	stopPeerForErrorFn stopPeerForErrorDelegate
	dialPeersFn        dialPeersDelegate
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

func (m *mockSwitch) StopPeerForError(peer p2p.Peer, err error) {
	if m.stopPeerForErrorFn != nil {
		m.stopPeerForErrorFn(peer, err)
	}
}

func (m *mockSwitch) DialPeers(peerAddrs ...*types.NetAddress) {
	if m.dialPeersFn != nil {
		m.dialPeersFn(peerAddrs...)
	}
}

type (
	addDelegate         func(p2p.Peer)
	removeDelegate      func(types.ID) bool
	hasDelegate         func(types.ID) bool
	hasIPDelegate       func(net.IP) bool
	getPeerDelegate     func(types.ID) p2p.Peer
	listDelegate        func() []p2p.Peer
	sizeDelegate        func() int
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
	sizeFn        sizeDelegate
	numInboundFn  numInboundDelegate
	numOutboundFn numOutboundDelegate
}

func (m *mockPeerSet) Add(peer p2p.Peer) {
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

func (m *mockPeerSet) HasIP(ip net.IP) bool {
	if m.hasIPFn != nil {
		return m.hasIPFn(ip)
	}

	return false
}

func (m *mockPeerSet) Get(key types.ID) p2p.Peer {
	if m.getFn != nil {
		return m.getFn(key)
	}

	return nil
}

func (m *mockPeerSet) List() []p2p.Peer {
	if m.listFn != nil {
		return m.listFn()
	}

	return nil
}

func (m *mockPeerSet) Size() int {
	if m.sizeFn != nil {
		return m.sizeFn()
	}

	return 0
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
