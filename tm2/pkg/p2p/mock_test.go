package p2p

import (
	"log/slog"
	"net"
	"time"

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

type (
	readDelegate        func([]byte) (int, error)
	writeDelegate       func([]byte) (int, error)
	closeDelegate       func() error
	localAddrDelegate   func() net.Addr
	setDeadlineDelegate func(time.Time) error
)

type mockConn struct {
	readFn             readDelegate
	writeFn            writeDelegate
	closeFn            closeDelegate
	localAddrFn        localAddrDelegate
	remoteAddrFn       remoteAddrDelegate
	setDeadlineFn      setDeadlineDelegate
	setReadDeadlineFn  setDeadlineDelegate
	setWriteDeadlineFn setDeadlineDelegate
}

func (m *mockConn) Read(b []byte) (int, error) {
	if m.readFn != nil {
		return m.readFn(b)
	}

	return 0, nil
}

func (m *mockConn) Write(b []byte) (int, error) {
	if m.writeFn != nil {
		return m.writeFn(b)
	}

	return 0, nil
}

func (m *mockConn) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}

	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	if m.localAddrFn != nil {
		return m.localAddrFn()
	}

	return nil
}

func (m *mockConn) RemoteAddr() net.Addr {
	if m.remoteAddrFn != nil {
		return m.remoteAddrFn()
	}

	return nil
}

func (m *mockConn) SetDeadline(t time.Time) error {
	if m.setDeadlineFn != nil {
		return m.setDeadlineFn(t)
	}

	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	if m.setReadDeadlineFn != nil {
		return m.setReadDeadlineFn(t)
	}

	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	if m.setWriteDeadlineFn != nil {
		return m.setWriteDeadlineFn(t)
	}

	return nil
}

type (
	startDelegate  func() error
	stopDelegate   func() error
	stringDelegate func() string
)

type mockMConn struct {
	flushFn   flushStopDelegate
	startFn   startDelegate
	stopFn    stopDelegate
	sendFn    sendDelegate
	trySendFn trySendDelegate
	statusFn  statusDelegate
	stringFn  stringDelegate
}

func (m *mockMConn) FlushStop() {
	if m.flushFn != nil {
		m.flushFn()
	}
}

func (m *mockMConn) Start() error {
	if m.startFn != nil {
		return m.startFn()
	}

	return nil
}

func (m *mockMConn) Stop() error {
	if m.stopFn != nil {
		return m.stopFn()
	}

	return nil
}

func (m *mockMConn) Send(ch byte, data []byte) bool {
	if m.sendFn != nil {
		return m.sendFn(ch, data)
	}

	return false
}

func (m *mockMConn) TrySend(ch byte, data []byte) bool {
	if m.trySendFn != nil {
		return m.trySendFn(ch, data)
	}

	return false
}

func (m *mockMConn) SetLogger(_ *slog.Logger) {}

func (m *mockMConn) Status() conn.ConnectionStatus {
	if m.statusFn != nil {
		return m.statusFn()
	}

	return conn.ConnectionStatus{}
}

func (m *mockMConn) String() string {
	if m.stringFn != nil {
		return m.stringFn()
	}

	return ""
}
