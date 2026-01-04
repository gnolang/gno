package mock

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/stretchr/testify/require"
)

type (
	flushStopDelegate    func()
	idDelegate           func() types.ID
	remoteIPDelegate     func() net.IP
	remoteAddrDelegate   func() net.Addr
	isOutboundDelegate   func() bool
	isPersistentDelegate func() bool
	isPrivateDelegate    func() bool
	closeConnDelegate    func() error
	nodeInfoDelegate     func() types.NodeInfo
	statusDelegate       func() conn.ConnectionStatus
	socketAddrDelegate   func() *types.NetAddress
	sendDelegate         func(byte, []byte) bool
	trySendDelegate      func(byte, []byte) bool
	setDelegate          func(string, any)
	getDelegate          func(string) any
	stopDelegate         func() error
)

// GeneratePeers generates random peers
func GeneratePeers(t *testing.T, count int) []*Peer {
	t.Helper()

	peers := make([]*Peer, count)

	for i := range count {
		var (
			key     = types.GenerateNodeKey()
			address = "127.0.0.1:8080"
		)

		tcpAddr, err := net.ResolveTCPAddr("tcp", address)
		require.NoError(t, err)

		addr, err := types.NewNetAddress(key.ID(), tcpAddr)
		require.NoError(t, err)

		p := &Peer{
			IDFn: func() types.ID {
				return key.ID()
			},
			NodeInfoFn: func() types.NodeInfo {
				return types.NodeInfo{
					NetAddress: addr,
				}
			},
			SocketAddrFn: func() *types.NetAddress {
				return addr
			},
		}

		p.BaseService = *service.NewBaseService(
			slog.New(slog.NewTextHandler(io.Discard, nil)),
			fmt.Sprintf("peer-%d", i),
			p,
		)

		peers[i] = p
	}

	return peers
}

type Peer struct {
	service.BaseService

	FlushStopFn    flushStopDelegate
	IDFn           idDelegate
	RemoteIPFn     remoteIPDelegate
	RemoteAddrFn   remoteAddrDelegate
	IsOutboundFn   isOutboundDelegate
	IsPersistentFn isPersistentDelegate
	IsPrivateFn    isPrivateDelegate
	CloseConnFn    closeConnDelegate
	NodeInfoFn     nodeInfoDelegate
	StopFn         stopDelegate
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

func (m *Peer) Stop() error {
	if m.StopFn != nil {
		return m.StopFn()
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

func (m *Peer) IsPrivate() bool {
	if m.IsPrivateFn != nil {
		return m.IsPrivateFn()
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

type Conn struct {
	ReadFn             readDelegate
	WriteFn            writeDelegate
	CloseFn            closeDelegate
	LocalAddrFn        localAddrDelegate
	RemoteAddrFn       remoteAddrDelegate
	SetDeadlineFn      setDeadlineDelegate
	SetReadDeadlineFn  setDeadlineDelegate
	SetWriteDeadlineFn setDeadlineDelegate
}

func (m *Conn) Read(b []byte) (int, error) {
	if m.ReadFn != nil {
		return m.ReadFn(b)
	}

	return 0, nil
}

func (m *Conn) Write(b []byte) (int, error) {
	if m.WriteFn != nil {
		return m.WriteFn(b)
	}

	return 0, nil
}

func (m *Conn) Close() error {
	if m.CloseFn != nil {
		return m.CloseFn()
	}

	return nil
}

func (m *Conn) LocalAddr() net.Addr {
	if m.LocalAddrFn != nil {
		return m.LocalAddrFn()
	}

	return nil
}

func (m *Conn) RemoteAddr() net.Addr {
	if m.RemoteAddrFn != nil {
		return m.RemoteAddrFn()
	}

	return nil
}

func (m *Conn) SetDeadline(t time.Time) error {
	if m.SetDeadlineFn != nil {
		return m.SetDeadlineFn(t)
	}

	return nil
}

func (m *Conn) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFn != nil {
		return m.SetReadDeadlineFn(t)
	}

	return nil
}

func (m *Conn) SetWriteDeadline(t time.Time) error {
	if m.SetWriteDeadlineFn != nil {
		return m.SetWriteDeadlineFn(t)
	}

	return nil
}

type (
	startDelegate  func() error
	stringDelegate func() string
)

type MConn struct {
	FlushFn   flushStopDelegate
	StartFn   startDelegate
	StopFn    stopDelegate
	SendFn    sendDelegate
	TrySendFn trySendDelegate
	StatusFn  statusDelegate
	StringFn  stringDelegate
}

func (m *MConn) FlushStop() {
	if m.FlushFn != nil {
		m.FlushFn()
	}
}

func (m *MConn) Start() error {
	if m.StartFn != nil {
		return m.StartFn()
	}

	return nil
}

func (m *MConn) Stop() error {
	if m.StopFn != nil {
		return m.StopFn()
	}

	return nil
}

func (m *MConn) Send(ch byte, data []byte) bool {
	if m.SendFn != nil {
		return m.SendFn(ch, data)
	}

	return false
}

func (m *MConn) TrySend(ch byte, data []byte) bool {
	if m.TrySendFn != nil {
		return m.TrySendFn(ch, data)
	}

	return false
}

func (m *MConn) SetLogger(_ *slog.Logger) {}

func (m *MConn) Status() conn.ConnectionStatus {
	if m.StatusFn != nil {
		return m.StatusFn()
	}

	return conn.ConnectionStatus{}
}

func (m *MConn) String() string {
	if m.StringFn != nil {
		return m.StringFn()
	}

	return ""
}
