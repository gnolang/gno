package mock

import (
	"net"

	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type (
	PeersDelegate func() p2p.PeerSet
)

type Peers struct {
	PeersFn PeersDelegate
}

func (m *Peers) Peers() p2p.PeerSet {
	if m.PeersFn != nil {
		return m.PeersFn()
	}

	return nil
}

type (
	AddDelegate         func(p2p.PeerConn)
	RemoveDelegate      func(p2pTypes.ID) bool
	HasDelegate         func(p2pTypes.ID) bool
	HasIPDelegate       func(net.IP) bool
	GetPeerDelegate     func(p2pTypes.ID) p2p.PeerConn
	ListDelegate        func() []p2p.PeerConn
	NumInboundDelegate  func() uint64
	NumOutboundDelegate func() uint64
)

type PeerSet struct {
	AddFn         AddDelegate
	RemoveFn      RemoveDelegate
	HasFn         HasDelegate
	HasIPFn       HasIPDelegate
	GetFn         GetPeerDelegate
	ListFn        ListDelegate
	NumInboundFn  NumInboundDelegate
	NumOutboundFn NumOutboundDelegate
}

func (m *PeerSet) Add(peer p2p.PeerConn) {
	if m.AddFn != nil {
		m.AddFn(peer)
	}
}

func (m *PeerSet) Remove(key p2pTypes.ID) bool {
	if m.RemoveFn != nil {
		m.RemoveFn(key)
	}

	return false
}

func (m *PeerSet) Has(key p2pTypes.ID) bool {
	if m.HasFn != nil {
		return m.HasFn(key)
	}

	return false
}

func (m *PeerSet) Get(key p2pTypes.ID) p2p.PeerConn {
	if m.GetFn != nil {
		return m.GetFn(key)
	}

	return nil
}

func (m *PeerSet) List() []p2p.PeerConn {
	if m.ListFn != nil {
		return m.ListFn()
	}

	return nil
}

func (m *PeerSet) NumInbound() uint64 {
	if m.NumInboundFn != nil {
		return m.NumInboundFn()
	}

	return 0
}

func (m *PeerSet) NumOutbound() uint64 {
	if m.NumOutboundFn != nil {
		return m.NumOutboundFn()
	}

	return 0
}

type (
	ListenersDelegate   func() []string
	IsListeningDelegate func() bool
	NodeInfoDelegate    func() p2pTypes.NodeInfo
)

type Transport struct {
	ListenersFn   ListenersDelegate
	IsListeningFn IsListeningDelegate
	NodeInfoFn    NodeInfoDelegate
}

func (m *Transport) Listeners() []string {
	if m.ListenersFn != nil {
		return m.ListenersFn()
	}

	return nil
}

func (m *Transport) IsListening() bool {
	if m.IsListeningFn != nil {
		return m.IsListeningFn()
	}

	return false
}

func (m *Transport) NodeInfo() p2pTypes.NodeInfo {
	if m.NodeInfoFn != nil {
		return m.NodeInfoFn()
	}

	return p2pTypes.NodeInfo{}
}
