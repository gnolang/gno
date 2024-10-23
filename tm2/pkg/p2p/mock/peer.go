package mock

import (
	"log/slog"
	"net"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
)

type (
	flushStopDelegate    func()
	idDelegate           func() types.ID
	remoteIPDelegate     func() net.IP
	remoteAddrDelegate   func() net.Addr
	isOutboundDelegate   func() bool
	isPersistentDelegate func() bool
	closeConnDelegate    func() error
	nodeInfoDelegate     func() types.NodeInfo
	statusDelegate       func() conn.ConnectionStatus
	socketAddrDelegate   func() *types.NetAddress
	sendDelegate         func(byte, []byte) bool
	trySendDelegate      func(byte, []byte) bool
	setDelegate          func(string, any)
	getDelegate          func(string) any
)

type Peer struct {
	service.BaseService

	FlushStopFn    flushStopDelegate
	IDFn           idDelegate
	RemoteIPFn     remoteIPDelegate
	RemoteAddrFn   remoteAddrDelegate
	IsOutboundFn   isOutboundDelegate
	IsPersistentFn isPersistentDelegate
	CloseConnFn    closeConnDelegate
	NodeInfoFn     nodeInfoDelegate
	StatusFn       statusDelegate
	SocketAddrFn   socketAddrDelegate
	SendFn         sendDelegate
	TrySendFn      trySendDelegate
	SetFn          setDelegate
	GetFn          getDelegate
}

func (m *Peer) FlushStop() {
	if m.FlushStopFn != nil {
		m.FlushStopFn()
	}
}

func (m *Peer) ID() types.ID {
	if m.IDFn != nil {
		return m.IDFn()
	}

	return ""
}

func (m *Peer) RemoteIP() net.IP {
	if m.RemoteIPFn != nil {
		return m.RemoteIPFn()
	}

	return nil
}

func (m *Peer) RemoteAddr() net.Addr {
	if m.RemoteAddrFn != nil {
		return m.RemoteAddrFn()
	}

	return nil
}

func (m *Peer) IsOutbound() bool {
	if m.IsOutboundFn != nil {
		return m.IsOutboundFn()
	}

	return false
}

func (m *Peer) IsPersistent() bool {
	if m.IsPersistentFn != nil {
		return m.IsPersistentFn()
	}

	return false
}

func (m *Peer) CloseConn() error {
	if m.CloseConnFn != nil {
		return m.CloseConnFn()
	}

	return nil
}

func (m *Peer) NodeInfo() types.NodeInfo {
	if m.NodeInfoFn != nil {
		return m.NodeInfoFn()
	}

	return types.NodeInfo{}
}

func (m *Peer) Status() conn.ConnectionStatus {
	if m.StatusFn != nil {
		return m.StatusFn()
	}

	return conn.ConnectionStatus{}
}

func (m *Peer) SocketAddr() *types.NetAddress {
	if m.SocketAddrFn != nil {
		return m.SocketAddrFn()
	}

	return nil
}

func (m *Peer) Send(classifier byte, data []byte) bool {
	if m.SendFn != nil {
		return m.SendFn(classifier, data)
	}

	return false
}

func (m *Peer) TrySend(classifier byte, data []byte) bool {
	if m.TrySendFn != nil {
		return m.TrySendFn(classifier, data)
	}

	return false
}

func (m *Peer) Set(key string, data any) {
	if m.SetFn != nil {
		m.SetFn(key, data)
	}
}

func (m *Peer) Get(key string) any {
	if m.GetFn != nil {
		return m.GetFn(key)
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

type MockConn struct {
	ReadFn             readDelegate
	WriteFn            writeDelegate
	CloseFn            closeDelegate
	LocalAddrFn        localAddrDelegate
	RemoteAddrFn       remoteAddrDelegate
	SetDeadlineFn      setDeadlineDelegate
	SetReadDeadlineFn  setDeadlineDelegate
	SetWriteDeadlineFn setDeadlineDelegate
}

func (m *MockConn) Read(b []byte) (int, error) {
	if m.ReadFn != nil {
		return m.ReadFn(b)
	}

	return 0, nil
}

func (m *MockConn) Write(b []byte) (int, error) {
	if m.WriteFn != nil {
		return m.WriteFn(b)
	}

	return 0, nil
}

func (m *MockConn) Close() error {
	if m.CloseFn != nil {
		return m.CloseFn()
	}

	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	if m.LocalAddrFn != nil {
		return m.LocalAddrFn()
	}

	return nil
}

func (m *MockConn) RemoteAddr() net.Addr {
	if m.RemoteAddrFn != nil {
		return m.RemoteAddrFn()
	}

	return nil
}

func (m *MockConn) SetDeadline(t time.Time) error {
	if m.SetDeadlineFn != nil {
		return m.SetDeadlineFn(t)
	}

	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFn != nil {
		return m.SetReadDeadlineFn(t)
	}

	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	if m.SetWriteDeadlineFn != nil {
		return m.SetWriteDeadlineFn(t)
	}

	return nil
}

type (
	startDelegate  func() error
	stopDelegate   func() error
	stringDelegate func() string
)

type MockMConn struct {
	FlushFn   flushStopDelegate
	StartFn   startDelegate
	StopFn    stopDelegate
	SendFn    sendDelegate
	TrySendFn trySendDelegate
	StatusFn  statusDelegate
	StringFn  stringDelegate
}

func (m *MockMConn) FlushStop() {
	if m.FlushFn != nil {
		m.FlushFn()
	}
}

func (m *MockMConn) Start() error {
	if m.StartFn != nil {
		return m.StartFn()
	}

	return nil
}

func (m *MockMConn) Stop() error {
	if m.StopFn != nil {
		return m.StopFn()
	}

	return nil
}

func (m *MockMConn) Send(ch byte, data []byte) bool {
	if m.SendFn != nil {
		return m.SendFn(ch, data)
	}

	return false
}

func (m *MockMConn) TrySend(ch byte, data []byte) bool {
	if m.TrySendFn != nil {
		return m.TrySendFn(ch, data)
	}

	return false
}

func (m *MockMConn) SetLogger(_ *slog.Logger) {}

func (m *MockMConn) Status() conn.ConnectionStatus {
	if m.StatusFn != nil {
		return m.StatusFn()
	}

	return conn.ConnectionStatus{}
}

func (m *MockMConn) String() string {
	if m.StringFn != nil {
		return m.StringFn()
	}

	return ""
}
