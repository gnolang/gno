package p2p

import (
	"net"

	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/service"
)

type (
	flushStopDelegate    func()
	idDelegate           func() ID
	remoteIPDelegate     func() net.IP
	remoteAddrDelegate   func() net.Addr
	isOutboundDelegate   func() bool
	isPersistentDelegate func() bool
	closeConnDelegate    func() error
	nodeInfoDelegate     func() NodeInfo
	statusDelegate       func() conn.ConnectionStatus
	socketAddrDelegate   func() *NetAddress
	sendDelegate         func(byte, []byte) bool
	trySendDelegate      func(byte, []byte) bool
	setDelegate          func(string, any)
	getDelegate          func(string) any
)

type mockPeer struct {
	service.BaseService

	flushStopFn    flushStopDelegate
	idFn           idDelegate
	remoteIPFn     remoteIPDelegate
	remoteAddrFn   remoteAddrDelegate
	isOutboundFn   isOutboundDelegate
	isPersistentFn isPersistentDelegate
	closeConnFn    closeConnDelegate
	nodeInfoFn     nodeInfoDelegate
	statusFn       statusDelegate
	socketAddrFn   socketAddrDelegate
	sendFn         sendDelegate
	trySendFn      trySendDelegate
	setFn          setDelegate
	getFn          getDelegate
}

func (m *mockPeer) FlushStop() {
	if m.flushStopFn != nil {
		m.flushStopFn()
	}
}

func (m *mockPeer) ID() ID {
	if m.idFn != nil {
		return m.idFn()
	}

	return ""
}

func (m *mockPeer) RemoteIP() net.IP {
	if m.remoteIPFn != nil {
		return m.remoteIPFn()
	}

	return nil
}

func (m *mockPeer) RemoteAddr() net.Addr {
	if m.remoteAddrFn != nil {
		return m.remoteAddrFn()
	}

	return nil
}

func (m *mockPeer) IsOutbound() bool {
	if m.isOutboundFn != nil {
		return m.isOutboundFn()
	}

	return false
}

func (m *mockPeer) IsPersistent() bool {
	if m.isPersistentFn != nil {
		return m.isPersistentFn()
	}

	return false
}

func (m *mockPeer) CloseConn() error {
	if m.closeConnFn != nil {
		return m.closeConnFn()
	}

	return nil
}

func (m *mockPeer) NodeInfo() NodeInfo {
	if m.nodeInfoFn != nil {
		return m.nodeInfoFn()
	}

	return NodeInfo{}
}

func (m *mockPeer) Status() conn.ConnectionStatus {
	if m.statusFn != nil {
		return m.statusFn()
	}

	return conn.ConnectionStatus{}
}

func (m *mockPeer) SocketAddr() *NetAddress {
	if m.socketAddrFn != nil {
		return m.socketAddrFn()
	}

	return nil
}

func (m *mockPeer) Send(classifier byte, data []byte) bool {
	if m.sendFn != nil {
		return m.sendFn(classifier, data)
	}

	return false
}

func (m *mockPeer) TrySend(classifier byte, data []byte) bool {
	if m.trySendFn != nil {
		return m.trySendFn(classifier, data)
	}

	return false
}

func (m *mockPeer) Set(key string, data any) {
	if m.setFn != nil {
		m.setFn(key, data)
	}
}

func (m *mockPeer) Get(key string) any {
	if m.getFn != nil {
		return m.getFn(key)
	}

	return nil
}
