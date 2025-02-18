package p2p

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type (
	netAddressDelegate func() types.NetAddress
	acceptDelegate     func(context.Context, PeerBehavior) (PeerConn, error)
	dialDelegate       func(context.Context, types.NetAddress, PeerBehavior) (PeerConn, error)
	removeDelegate     func(PeerConn)
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

func (m *mockTransport) Accept(ctx context.Context, behavior PeerBehavior) (PeerConn, error) {
	if m.acceptFn != nil {
		return m.acceptFn(ctx, behavior)
	}

	return nil, nil
}

func (m *mockTransport) Dial(ctx context.Context, address types.NetAddress, behavior PeerBehavior) (PeerConn, error) {
	if m.dialFn != nil {
		return m.dialFn(ctx, address, behavior)
	}

	return nil, nil
}

func (m *mockTransport) Remove(p PeerConn) {
	if m.removeFn != nil {
		m.removeFn(p)
	}
}

type (
	addDelegate         func(PeerConn)
	removePeerDelegate  func(types.ID) bool
	hasDelegate         func(types.ID) bool
	hasIPDelegate       func(net.IP) bool
	getDelegate         func(types.ID) PeerConn
	listDelegate        func() []PeerConn
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

func (m *mockSet) Add(peer PeerConn) {
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

func (m *mockSet) Get(key types.ID) PeerConn {
	if m.getFn != nil {
		return m.getFn(key)
	}

	return nil
}

func (m *mockSet) List() []PeerConn {
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

type (
	listenerAcceptDelegate func() (net.Conn, error)
	closeDelegate          func() error
	addrDelegate           func() net.Addr
)

type mockListener struct {
	acceptFn listenerAcceptDelegate
	closeFn  closeDelegate
	addrFn   addrDelegate
}

func (m *mockListener) Accept() (net.Conn, error) {
	if m.acceptFn != nil {
		return m.acceptFn()
	}

	return nil, nil
}

func (m *mockListener) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}

	return nil
}

func (m *mockListener) Addr() net.Addr {
	if m.addrFn != nil {
		return m.addrFn()
	}

	return nil
}

type (
	readDelegate        func([]byte) (int, error)
	writeDelegate       func([]byte) (int, error)
	localAddrDelegate   func() net.Addr
	remoteAddrDelegate  func() net.Addr
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

func (m *mockConn) Read(buff []byte) (int, error) {
	if m.readFn != nil {
		return m.readFn(buff)
	}

	return 0, nil
}

func (m *mockConn) Write(buff []byte) (int, error) {
	if m.writeFn != nil {
		return m.writeFn(buff)
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
	startDelegate            func() error
	onStartDelegate          func() error
	stopDelegate             func() error
	onStopDelegate           func()
	resetDelegate            func() error
	onResetDelegate          func() error
	isRunningDelegate        func() bool
	quitDelegate             func() <-chan struct{}
	stringDelegate           func() string
	setLoggerDelegate        func(*slog.Logger)
	setSwitchDelegate        func(Switch)
	getChannelsDelegate      func() []*conn.ChannelDescriptor
	initPeerDelegate         func(PeerConn)
	addPeerDelegate          func(PeerConn)
	removeSwitchPeerDelegate func(PeerConn, any)
	receiveDelegate          func(byte, PeerConn, []byte)
)

type mockReactor struct {
	startFn       startDelegate
	onStartFn     onStartDelegate
	stopFn        stopDelegate
	onStopFn      onStopDelegate
	resetFn       resetDelegate
	onResetFn     onResetDelegate
	isRunningFn   isRunningDelegate
	quitFn        quitDelegate
	stringFn      stringDelegate
	setLoggerFn   setLoggerDelegate
	setSwitchFn   setSwitchDelegate
	getChannelsFn getChannelsDelegate
	initPeerFn    initPeerDelegate
	addPeerFn     addPeerDelegate
	removePeerFn  removeSwitchPeerDelegate
	receiveFn     receiveDelegate
}

func (m *mockReactor) Start() error {
	if m.startFn != nil {
		return m.startFn()
	}

	return nil
}

func (m *mockReactor) OnStart() error {
	if m.onStartFn != nil {
		return m.onStartFn()
	}

	return nil
}

func (m *mockReactor) Stop() error {
	if m.stopFn != nil {
		return m.stopFn()
	}

	return nil
}

func (m *mockReactor) OnStop() {
	if m.onStopFn != nil {
		m.onStopFn()
	}
}

func (m *mockReactor) Reset() error {
	if m.resetFn != nil {
		return m.resetFn()
	}

	return nil
}

func (m *mockReactor) OnReset() error {
	if m.onResetFn != nil {
		return m.onResetFn()
	}

	return nil
}

func (m *mockReactor) IsRunning() bool {
	if m.isRunningFn != nil {
		return m.isRunningFn()
	}

	return false
}

func (m *mockReactor) Quit() <-chan struct{} {
	if m.quitFn != nil {
		return m.quitFn()
	}

	return nil
}

func (m *mockReactor) String() string {
	if m.stringFn != nil {
		return m.stringFn()
	}

	return ""
}

func (m *mockReactor) SetLogger(logger *slog.Logger) {
	if m.setLoggerFn != nil {
		m.setLoggerFn(logger)
	}
}

func (m *mockReactor) SetSwitch(s Switch) {
	if m.setSwitchFn != nil {
		m.setSwitchFn(s)
	}
}

func (m *mockReactor) GetChannels() []*conn.ChannelDescriptor {
	if m.getChannelsFn != nil {
		return m.getChannelsFn()
	}

	return nil
}

func (m *mockReactor) InitPeer(peer PeerConn) PeerConn {
	if m.initPeerFn != nil {
		m.initPeerFn(peer)
	}

	return nil
}

func (m *mockReactor) AddPeer(peer PeerConn) {
	if m.addPeerFn != nil {
		m.addPeerFn(peer)
	}
}

func (m *mockReactor) RemovePeer(peer PeerConn, reason any) {
	if m.removePeerFn != nil {
		m.removePeerFn(peer, reason)
	}
}

func (m *mockReactor) Receive(chID byte, peer PeerConn, msgBytes []byte) {
	if m.receiveFn != nil {
		m.receiveFn(chID, peer, msgBytes)
	}
}
