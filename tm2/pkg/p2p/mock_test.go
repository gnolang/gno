package p2p

import (
	"context"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/errors"
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

// newDialRecorder returns a transport whose dials always fail, along with
// the channel every dialed address is sent to, buffered with the given size
func newDialRecorder(size int) (*mockTransport, <-chan types.NetAddress) {
	dialed := make(chan types.NetAddress, size)

	return &mockTransport{
		dialFn: func(_ context.Context, addr types.NetAddress, _ PeerBehavior) (PeerConn, error) {
			dialed <- addr

			return nil, errors.New("unable to dial")
		},
	}, dialed
}

type (
	addDelegate         func(PeerConn) error
	removePeerDelegate  func(types.ID) bool
	hasDelegate         func(types.ID) bool
	getDelegate         func(types.ID) PeerConn
	listDelegate        func() []PeerConn
	numInboundDelegate  func() uint64
	numOutboundDelegate func() uint64
)

type mockSet struct {
	addFn         addDelegate
	removeFn      removePeerDelegate
	hasFn         hasDelegate
	listFn        listDelegate
	getFn         getDelegate
	numInboundFn  numInboundDelegate
	numOutboundFn numOutboundDelegate
}

func (m *mockSet) Add(peer PeerConn) error {
	if m.addFn != nil {
		return m.addFn(peer)
	}

	return nil
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
	initPeerDelegate         func(PeerConn) PeerConn
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
		return m.initPeerFn(peer)
	}

	return peer
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
