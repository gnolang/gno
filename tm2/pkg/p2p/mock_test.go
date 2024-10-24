package p2p

import (
	"context"
	"net"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type (
	netAddressDelegate func() types.NetAddress
	acceptDelegate     func(context.Context, PeerBehavior) (Peer, error)
	dialDelegate       func(context.Context, types.NetAddress, PeerBehavior) (Peer, error)
	removeDelegate     func(Peer)
)

type mockTransport struct {
	netAddressFn netAddressDelegate
	acceptFn     acceptDelegate
	dialFn       dialDelegate
	removeFn     removeDelegate
}

func (m *mockTransport) NetAddress() types.NetAddress {
	if m.netAddressFn != nil {
		return m.netAddressFn()
	}

	return types.NetAddress{}
}

func (m *mockTransport) Accept(ctx context.Context, behavior PeerBehavior) (Peer, error) {
	if m.acceptFn != nil {
		return m.acceptFn(ctx, behavior)
	}

	return nil, nil
}

func (m *mockTransport) Dial(ctx context.Context, address types.NetAddress, behavior PeerBehavior) (Peer, error) {
	if m.dialFn != nil {
		return m.dialFn(ctx, address, behavior)
	}

	return nil, nil
}

func (m *mockTransport) Remove(p Peer) {
	if m.removeFn != nil {
		m.removeFn(p)
	}
}

type (
	addDelegate         func(Peer)
	removePeerDelegate  func(types.ID) bool
	hasDelegate         func(types.ID) bool
	hasIPDelegate       func(net.IP) bool
	getDelegate         func(types.ID) Peer
	listDelegate        func() []Peer
	numInboundDelegate  func() uint64
	numOutboundDelegate func() uint64
)

type mockSet struct {
	addFn         addDelegate
	removeFn      removePeerDelegate
	hasFn         hasDelegate
	hasIPFn       hasIPDelegate
	listFn        listDelegate
	getFn         getDelegate
	numInboundFn  numInboundDelegate
	numOutboundFn numOutboundDelegate
}

func (m *mockSet) Add(peer Peer) {
	if m.addFn != nil {
		m.addFn(peer)
	}
}

func (m *mockSet) Remove(key types.ID) bool {
	if m.removeFn != nil {
		m.removeFn(key)
	}

	return false
}

func (m *mockSet) Has(key types.ID) bool {
	if m.hasFn != nil {
		return m.hasFn(key)
	}

	return false
}

func (m *mockSet) HasIP(ip net.IP) bool {
	if m.hasIPFn != nil {
		return m.hasIPFn(ip)
	}

	return false
}

func (m *mockSet) Get(key types.ID) Peer {
	if m.getFn != nil {
		return m.getFn(key)
	}

	return nil
}

func (m *mockSet) List() []Peer {
	if m.listFn != nil {
		return m.listFn()
	}

	return nil
}

func (m *mockSet) NumInbound() uint64 {
	if m.numInboundFn != nil {
		return m.numInboundFn()
	}

	return 0
}

func (m *mockSet) NumOutbound() uint64 {
	if m.numOutboundFn != nil {
		return m.numOutboundFn()
	}

	return 0
}
