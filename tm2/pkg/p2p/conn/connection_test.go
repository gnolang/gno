package conn

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p/config"
)

const maxPingPongPacketSize = 1024 // bytes

func createTestMConnection(t *testing.T, conn net.Conn) *MConnection {
	t.Helper()

	onReceive := func(chID byte, msgBytes []byte) {
	}
	onError := func(r error) {
	}
	c := createMConnectionWithCallbacks(t, conn, onReceive, onError)
	c.SetLogger(log.NewTestingLogger(t))
	return c
}

func createMConnectionWithCallbacks(
	t *testing.T,
	conn net.Conn,
	onReceive func(chID byte, msgBytes []byte),
	onError func(r error),
) *MConnection {
	t.Helper()

	cfg := DefaultMConnConfig()
	cfg.PingInterval = 90 * time.Millisecond
	cfg.PongTimeout = 45 * time.Millisecond
	chDescs := []*ChannelDescriptor{{ID: 0x01, Priority: 1, SendQueueCapacity: 1}}
	c := NewMConnectionWithConfig(conn, chDescs, onReceive, onError, cfg)
	c.SetLogger(log.NewTestingLogger(t))
	return c
}

func TestMConnectionSendFlushStop(t *testing.T) {
	t.Parallel()

	server, client := NetPipe()
	defer server.Close() //nolint: errcheck
	defer client.Close() //nolint: errcheck

	clientConn := createTestMConnection(t, client)
	err := clientConn.Start()
	require.Nil(t, err)
	defer clientConn.Stop()

	msg := []byte("abc")
	assert.True(t, clientConn.Send(0x01, msg))

	// start the reader in a new routine, so we can flush
	errCh := make(chan error)
	go func() {
		msgBuf := make([]byte, 1024) // sufficiently big for "abc"
		_, err := server.Read(msgBuf)
		if err != nil {
			t.Error(err)
			return
		}
		bz, _, err := amino.DecodeByteSlice(msgBuf)
		if err != nil {
			t.Error(err)
		}
		var msg2 Packet
		err = amino.Unmarshal(bz, &msg2)
		if err != nil {
			t.Error(err)
		}
		errCh <- err
	}()

	// stop the conn - it should flush all conns
	clientConn.FlushStop()

	timer := time.NewTimer(3 * time.Second)
	select {
	case <-errCh:
	case <-timer.C:
		t.Error("timed out waiting for msgs to be read")
	}
}

func TestMConnectionSend(t *testing.T) {
	t.Parallel()

	server, client := NetPipe()
	defer server.Close() //nolint: errcheck
	defer client.Close() //nolint: errcheck

	mconn := createTestMConnection(t, client)
	err := mconn.Start()
	require.Nil(t, err)
	defer mconn.Stop()

	msg := []byte("Ant-Man")
	assert.True(t, mconn.Send(0x01, msg))
	// Note: subsequent Send/TrySend calls could pass because we are reading from
	// the send queue in a separate goroutine.
	_, err = server.Read(make([]byte, len(msg)))
	if err != nil {
		t.Error(err)
	}
	assert.True(t, mconn.CanSend(0x01))

	msg = []byte("Spider-Man")
	assert.True(t, mconn.TrySend(0x01, msg))
	_, err = server.Read(make([]byte, len(msg)))
	if err != nil {
		t.Error(err)
	}

	assert.False(t, mconn.CanSend(0x05), "CanSend should return false because channel is unknown")
	assert.False(t, mconn.Send(0x05, []byte("Absorbing Man")), "Send should return false because channel is unknown")
}

func TestMConnectionReceive(t *testing.T) {
	t.Parallel()

	server, client := NetPipe()
	defer server.Close() //nolint: errcheck
	defer client.Close() //nolint: errcheck

	receivedCh := make(chan []byte)
	errorsCh := make(chan any)
	onReceive := func(chID byte, msgBytes []byte) {
		receivedCh <- msgBytes
	}
	onError := func(r error) {
		errorsCh <- r
	}
	mconn1 := createMConnectionWithCallbacks(t, client, onReceive, onError)
	err := mconn1.Start()
	require.Nil(t, err)
	defer mconn1.Stop()

	mconn2 := createTestMConnection(t, server)
	err = mconn2.Start()
	require.Nil(t, err)
	defer mconn2.Stop()

	msg := []byte("Cyclops")
	assert.True(t, mconn2.Send(0x01, msg))

	select {
	case receivedBytes := <-receivedCh:
		assert.Equal(t, msg, receivedBytes)
	case err := <-errorsCh:
		t.Fatalf("Expected %s, got %+v", msg, err)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Did not receive %s message in 500ms", msg)
	}
}

func TestMConnectionStatus(t *testing.T) {
	t.Parallel()

	server, client := NetPipe()
	defer server.Close() //nolint: errcheck
	defer client.Close() //nolint: errcheck

	mconn := createTestMConnection(t, client)
	err := mconn.Start()
	require.Nil(t, err)
	defer mconn.Stop()

	status := mconn.Status()
	assert.NotNil(t, status)
	assert.Zero(t, status.Channels[0].SendQueueSize)
}

func TestMConnectionPongTimeoutResultsInError(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	receivedCh := make(chan []byte)
	errorsCh := make(chan any)
	onReceive := func(chID byte, msgBytes []byte) {
		receivedCh <- msgBytes
	}
	onError := func(r error) {
		errorsCh <- r
	}
	mconn := createMConnectionWithCallbacks(t, client, onReceive, onError)
	err := mconn.Start()
	require.Nil(t, err)
	defer mconn.Stop()

	serverGotPing := make(chan struct{})
	go func() {
		// read ping (Any-wrapped on the wire, decode as interface)
		var pkt Packet
		_, err = amino.UnmarshalSizedReader(server, &pkt, maxPingPongPacketSize)
		assert.Nil(t, err)
		assert.IsType(t, PacketPing{}, pkt)
		serverGotPing <- struct{}{}
	}()
	<-serverGotPing

	pongTimerExpired := mconn.config.PongTimeout + 20*time.Millisecond
	select {
	case msgBytes := <-receivedCh:
		t.Fatalf("Expected error, but got %v", msgBytes)
	case err := <-errorsCh:
		assert.NotNil(t, err)
	case <-time.After(pongTimerExpired):
		t.Fatalf("Expected to receive error after %v", pongTimerExpired)
	}
}

func TestMConnectionMultiplePongsInTheBeginning(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	receivedCh := make(chan []byte)
	errorsCh := make(chan any)
	onReceive := func(chID byte, msgBytes []byte) {
		receivedCh <- msgBytes
	}
	onError := func(r error) {
		errorsCh <- r
	}
	mconn := createMConnectionWithCallbacks(t, client, onReceive, onError)
	err := mconn.Start()
	require.Nil(t, err)
	defer mconn.Stop()

	// sending 3 pongs in a row (abuse)
	_, err = server.Write(amino.MustMarshalAnySized(PacketPong{}))
	require.Nil(t, err)
	_, err = server.Write(amino.MustMarshalAnySized(PacketPong{}))
	require.Nil(t, err)
	_, err = server.Write(amino.MustMarshalAnySized(PacketPong{}))
	require.Nil(t, err)

	serverGotPing := make(chan struct{})
	go func() {
		// read ping (one byte)
		var (
			packet Packet
			err    error
		)
		_, err = amino.UnmarshalSizedReader(server, &packet, maxPingPongPacketSize)
		require.Nil(t, err)
		serverGotPing <- struct{}{}
		// respond with pong
		_, err = server.Write(amino.MustMarshalAnySized(PacketPong{}))
		require.Nil(t, err)
	}()
	<-serverGotPing

	pongTimerExpired := mconn.config.PongTimeout + 20*time.Millisecond
	select {
	case msgBytes := <-receivedCh:
		t.Fatalf("Expected no data, but got %v", msgBytes)
	case err := <-errorsCh:
		t.Fatalf("Expected no error, but got %v", err)
	case <-time.After(pongTimerExpired):
		assert.True(t, mconn.IsRunning())
	}
}

func TestMConnectionMultiplePings(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	receivedCh := make(chan []byte)
	errorsCh := make(chan any)
	onReceive := func(chID byte, msgBytes []byte) {
		receivedCh <- msgBytes
	}
	onError := func(r error) {
		errorsCh <- r
	}
	mconn := createMConnectionWithCallbacks(t, client, onReceive, onError)
	err := mconn.Start()
	require.Nil(t, err)
	defer mconn.Stop()

	// sending 3 pings in a row (abuse)
	// see https://github.com/tendermint/tendermint/issues/1190
	for range 3 {
		var pkt Packet
		_, err = server.Write(amino.MustMarshalAnySized(PacketPing{}))
		require.Nil(t, err)
		_, err = amino.UnmarshalSizedReader(server, &pkt, maxPingPongPacketSize)
		require.Nil(t, err)
		require.IsType(t, PacketPong{}, pkt)
	}

	assert.True(t, mconn.IsRunning())
}

func TestMConnectionPingPongs(t *testing.T) {
	t.Parallel()

	// check that we are not leaking any go-routines
	defer leaktest.CheckTimeout(t, 10*time.Second)()

	server, client := net.Pipe()

	defer server.Close()
	defer client.Close()

	receivedCh := make(chan []byte)
	errorsCh := make(chan any)
	onReceive := func(chID byte, msgBytes []byte) {
		receivedCh <- msgBytes
	}
	onError := func(r error) {
		errorsCh <- r
	}
	mconn := createMConnectionWithCallbacks(t, client, onReceive, onError)
	err := mconn.Start()
	require.Nil(t, err)
	defer mconn.Stop()

	serverGotPing := make(chan struct{})
	go func() {
		// read ping (Any-wrapped on the wire, decode as interface)
		var pkt Packet
		_, err = amino.UnmarshalSizedReader(server, &pkt, maxPingPongPacketSize)
		require.Nil(t, err)
		require.IsType(t, PacketPing{}, pkt)
		serverGotPing <- struct{}{}
		// respond with pong
		_, err = server.Write(amino.MustMarshalAnySized(PacketPong{}))
		require.Nil(t, err)

		time.Sleep(mconn.config.PingInterval)

		// read ping (fresh variable — interface must be nil for amino)
		var pkt2 Packet
		_, err = amino.UnmarshalSizedReader(server, &pkt2, maxPingPongPacketSize)
		require.Nil(t, err)
		require.IsType(t, PacketPing{}, pkt2)
		// respond with pong
		_, err = server.Write(amino.MustMarshalAnySized(PacketPong{}))
		require.Nil(t, err)
	}()
	<-serverGotPing

	pongTimerExpired := (mconn.config.PongTimeout + 20*time.Millisecond) * 2
	select {
	case msgBytes := <-receivedCh:
		t.Fatalf("Expected no data, but got %v", msgBytes)
	case err := <-errorsCh:
		t.Fatalf("Expected no error, but got %v", err)
	case <-time.After(2 * pongTimerExpired):
		assert.True(t, mconn.IsRunning())
	}
}

func TestMConnectionStopsAndReturnsError(t *testing.T) {
	t.Parallel()

	server, client := NetPipe()
	defer server.Close() //nolint: errcheck
	defer client.Close() //nolint: errcheck

	receivedCh := make(chan []byte)
	errorsCh := make(chan any)
	onReceive := func(chID byte, msgBytes []byte) {
		receivedCh <- msgBytes
	}
	onError := func(r error) {
		errorsCh <- r
	}
	mconn := createMConnectionWithCallbacks(t, client, onReceive, onError)
	err := mconn.Start()
	require.Nil(t, err)
	defer mconn.Stop()

	if err := client.Close(); err != nil {
		t.Error(err)
	}

	select {
	case receivedBytes := <-receivedCh:
		t.Fatalf("Expected error, got %v", receivedBytes)
	case err := <-errorsCh:
		assert.NotNil(t, err)
		assert.False(t, mconn.IsRunning())
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Did not receive error in 500ms")
	}
}

func newClientAndServerConnsForReadErrors(t *testing.T, chOnErr chan struct{}) (*MConnection, *MConnection) {
	t.Helper()

	server, client := NetPipe()

	onReceive := func(chID byte, msgBytes []byte) {}
	onError := func(r error) {}

	// create client conn with two channels
	chDescs := []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1},
		{ID: 0x02, Priority: 1, SendQueueCapacity: 1},
	}
	mconnClient := NewMConnection(client, chDescs, onReceive, onError)
	mconnClient.SetLogger(log.NewNoopLogger().With("module", "client"))
	err := mconnClient.Start()
	require.Nil(t, err)

	// create server conn with 1 channel
	// it fires on chOnErr when there's an error
	serverLogger := log.NewNoopLogger().With("module", "server")
	onError = func(_ error) {
		chOnErr <- struct{}{}
	}
	mconnServer := createMConnectionWithCallbacks(t, server, onReceive, onError)
	mconnServer.SetLogger(serverLogger)
	err = mconnServer.Start()
	require.Nil(t, err)
	return mconnClient, mconnServer
}

func expectSend(ch chan struct{}) bool {
	after := time.After(time.Second * 5)
	select {
	case <-ch:
		return true
	case <-after:
		return false
	}
}

func TestMConnectionReadErrorBadEncoding(t *testing.T) {
	t.Parallel()

	chOnErr := make(chan struct{})
	mconnClient, mconnServer := newClientAndServerConnsForReadErrors(t, chOnErr)
	defer mconnClient.Stop()
	defer mconnServer.Stop()

	client := mconnClient.conn

	// send badly encoded msgPacket
	bz := amino.MustMarshalAnySized(PacketMsg{})
	bz[4] += 0x01 // Invalid prefix bytes.

	// Write it.
	_, err := client.Write(bz)
	assert.Nil(t, err)
	assert.True(t, expectSend(chOnErr), "badly encoded msgPacket")
}

func TestMConnectionReadErrorUnknownChannel(t *testing.T) {
	t.Parallel()

	chOnErr := make(chan struct{})
	mconnClient, mconnServer := newClientAndServerConnsForReadErrors(t, chOnErr)
	defer mconnClient.Stop()
	defer mconnServer.Stop()

	msg := []byte("Ant-Man")

	// fail to send msg on channel unknown by client
	assert.False(t, mconnClient.Send(0x03, msg))

	// send msg on channel unknown by the server.
	// should cause an error
	assert.True(t, mconnClient.Send(0x02, msg))
	assert.True(t, expectSend(chOnErr), "unknown channel")
}

func TestMConnectionReadErrorLongMessage(t *testing.T) {
	t.Parallel()

	chOnErr := make(chan struct{})
	chOnRcv := make(chan struct{})

	mconnClient, mconnServer := newClientAndServerConnsForReadErrors(t, chOnErr)
	defer mconnClient.Stop()
	defer mconnServer.Stop()

	mconnServer.onReceive = func(chID byte, msgBytes []byte) {
		chOnRcv <- struct{}{}
	}

	client := mconnClient.conn

	// send msg thats just right
	var err error
	buf := new(bytes.Buffer)
	packet := PacketMsg{
		ChannelID: 0x01,
		EOF:       1,
		Bytes:     make([]byte, mconnClient.config.MaxPacketMsgPayloadSize),
	}
	_, err = amino.MarshalAnySizedWriter(buf, packet)
	assert.Nil(t, err)
	_, err = client.Write(buf.Bytes())
	assert.Nil(t, err)
	assert.True(t, expectSend(chOnRcv), "msg just right")

	// send msg thats too long
	buf = new(bytes.Buffer)
	packet = PacketMsg{
		ChannelID: 0x01,
		EOF:       1,
		Bytes:     make([]byte, mconnClient.config.MaxPacketMsgPayloadSize+100),
	}
	_, err = amino.MarshalAnySizedWriter(buf, packet)
	assert.Nil(t, err)
	_, err = client.Write(buf.Bytes())
	assert.NotNil(t, err)
	assert.True(t, expectSend(chOnErr), "msg too long")
}

func TestMConnectionReadErrorUnknownMsgType(t *testing.T) {
	t.Parallel()

	chOnErr := make(chan struct{})
	mconnClient, mconnServer := newClientAndServerConnsForReadErrors(t, chOnErr)
	defer mconnClient.Stop()
	defer mconnServer.Stop()

	// send msg with unknown msg type
	err := amino.EncodeUvarint(mconnClient.conn, 4)
	assert.Nil(t, err)
	_, err = mconnClient.conn.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	assert.Nil(t, err)
	assert.True(t, expectSend(chOnErr), "unknown msg type")
}

func TestMConnectionTrySend(t *testing.T) {
	t.Parallel()

	server, client := NetPipe()
	defer server.Close()
	defer client.Close()

	mconn := createTestMConnection(t, client)
	err := mconn.Start()
	require.Nil(t, err)
	defer mconn.Stop()

	msg := []byte("Semicolon-Woman")
	resultCh := make(chan string, 2)
	assert.True(t, mconn.TrySend(0x01, msg))
	server.Read(make([]byte, len(msg)))
	assert.True(t, mconn.CanSend(0x01))
	assert.True(t, mconn.TrySend(0x01, msg))
	assert.False(t, mconn.CanSend(0x01))
	go func() {
		mconn.TrySend(0x01, msg)
		resultCh <- "TrySend"
	}()
	assert.False(t, mconn.CanSend(0x01))
	assert.False(t, mconn.TrySend(0x01, msg))
	assert.Equal(t, "TrySend", <-resultCh)
}

func TestMaxSendBufferBudgetAppliesBeforeQueueCapacity(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	cfg := quietPingConfig()
	cfg.MaxSendBufferBytes = 256 << 10
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 100},
	}, func(byte, []byte) {}, func(error) {}, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	msg := make([]byte, 128<<10)
	require.True(t, mconn.TrySend(0x01, msg))
	require.True(t, mconn.TrySend(0x01, msg))
	assert.False(t, mconn.TrySend(0x01, msg))
	// Nothing reads the pipe, so the first message never finishes and both
	// stay reserved. The budget, not the queue, refused the third.
	assert.Equal(t, int64(256<<10), sendBufferBytes(mconn))
}

func TestSendRejectsMessageLargerThanBudget(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	cfg := quietPingConfig()
	cfg.MaxSendBufferBytes = 64 << 10
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 100},
	}, func(byte, []byte) {}, func(error) {}, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	// The message can never fit, so Send must refuse it at once instead of
	// waiting out its timeout for room that will never come.
	start := time.Now()
	assert.False(t, mconn.Send(0x01, make([]byte, 64<<10+1)))
	assert.Less(t, time.Since(start), time.Second)
	assert.Zero(t, sendBufferBytes(mconn))
}

func TestSendWaitsForSendBufferBudget(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	const messageSize = 128 << 10
	cfg := quietPingConfig()
	cfg.MaxSendBufferBytes = messageSize
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 100},
	}, func(byte, []byte) {}, func(error) {}, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	msg := make([]byte, messageSize)
	require.True(t, mconn.TrySend(0x01, msg))

	sent := make(chan bool, 1)
	go func() { sent <- mconn.Send(0x01, msg) }()

	select {
	case <-sent:
		t.Fatal("Send returned before byte budget became available")
	case <-time.After(100 * time.Millisecond):
	}

	go func() { _, _ = io.Copy(io.Discard, server) }()
	select {
	case ok := <-sent:
		assert.True(t, ok)
	case <-time.After(5 * time.Second):
		t.Fatal("Send did not resume after byte budget became available")
	}
}

func TestSendBufferBudgetReleasedAfterMultiPacketDrain(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	const messageSize = 128*1024 + 17
	cfg := quietPingConfig()
	cfg.MaxSendBufferBytes = messageSize
	received := make(chan struct{}, 1)
	descs := []*ChannelDescriptor{{ID: 0x01, Priority: 1, SendQueueCapacity: 1}}
	sender := NewMConnectionWithConfig(client, descs, func(byte, []byte) {}, func(error) {}, cfg)
	receiver := NewMConnectionWithConfig(server, descs, func(byte, []byte) { received <- struct{}{} }, func(error) {}, quietPingConfig())
	sender.SetLogger(log.NewTestingLogger(t))
	receiver.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, sender.Start())
	defer sender.Stop()

	msg := make([]byte, messageSize)
	require.True(t, sender.TrySend(0x01, msg))
	sender.sendBufferMtx.Lock()
	buffered := sender.sendBufferBytes
	sender.sendBufferMtx.Unlock()
	require.Equal(t, int64(messageSize), buffered)

	require.NoError(t, receiver.Start())
	defer receiver.Stop()

	for i := range 10 {
		if i > 0 {
			require.True(t, sender.TrySend(0x01, msg))
		}
		select {
		case <-received:
		case <-time.After(5 * time.Second):
			t.Fatal("message was not received")
		}
		require.Eventually(t, func() bool {
			sender.sendBufferMtx.Lock()
			defer sender.sendBufferMtx.Unlock()
			return sender.sendBufferBytes == 0
		}, time.Second, time.Millisecond)
	}
}

func TestSendBufferBudgetSharedAcrossChannels(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	cfg := quietPingConfig()
	cfg.MaxSendBufferBytes = 192 << 10
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 100},
		{ID: 0x02, Priority: 1, SendQueueCapacity: 100},
	}, func(byte, []byte) {}, func(error) {}, cfg)
	mconn.SetLogger(log.NewTestingLogger(t))
	require.NoError(t, mconn.Start())
	defer mconn.Stop()

	msg := make([]byte, 128<<10)
	require.True(t, mconn.TrySend(0x01, msg))
	assert.False(t, mconn.TrySend(0x02, msg))
}

func TestMConfigFromP2PMaxSendBufferBytes(t *testing.T) {
	t.Parallel()

	p2pCfg := config.DefaultP2PConfig()
	assert.Equal(t, p2pCfg.MaxSendBufferBytes, DefaultMConnConfig().MaxSendBufferBytes)

	p2pCfg.MaxSendBufferBytes = 12345
	assert.Equal(t, 12345, MConfigFromP2P(p2pCfg).MaxSendBufferBytes)
}

func sendBufferBytes(c *MConnection) int64 {
	c.sendBufferMtx.Lock()
	defer c.sendBufferMtx.Unlock()
	return c.sendBufferBytes
}

func TestFailedEnqueueReleasesSendBuffer(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	cfg := quietPingConfig()
	cfg.MaxSendBufferBytes = 1 << 20
	// Not started, so nothing drains the one queue slot.
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1},
	}, func(byte, []byte) {}, func(error) {}, cfg)
	ch := mconn.channelsIdx[0x01]

	msg := make([]byte, 1024)
	require.True(t, ch.trySendBytes(msg))
	require.Equal(t, int64(len(msg)), sendBufferBytes(mconn))

	assert.False(t, ch.trySendBytes(msg), "the queue is full")
	assert.Equal(t, int64(len(msg)), sendBufferBytes(mconn), "TrySend must release what it reserved")

	assert.False(t, ch.sendBytes(msg, 10*time.Millisecond), "the queue is full")
	assert.Equal(t, int64(len(msg)), sendBufferBytes(mconn), "Send must release what it reserved")
}

func TestSendBufferReleasedOnlyAtMessageEnd(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	cfg := quietPingConfig()
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1},
	}, func(byte, []byte) {}, func(error) {}, cfg)
	ch := mconn.channelsIdx[0x01]

	msg := make([]byte, 3*cfg.MaxPacketMsgPayloadSize)
	require.True(t, ch.trySendBytes(msg))
	require.True(t, ch.isSendPending())

	for i := range 3 {
		_, completed, err := ch.writePacketMsgTo(io.Discard)
		require.NoError(t, err)
		if i < 2 {
			assert.Zero(t, completed, "packet %d is not the last one", i)
		} else {
			assert.Equal(t, len(msg), completed, "the last packet completes the whole message")
		}
	}
}

func TestSendGivesUpWhenConnectionStops(t *testing.T) {
	t.Parallel()

	const messageSize = 128 << 10
	for _, tc := range []struct {
		name     string
		budget   int
		capacity int
	}{
		{"waiting for budget", messageSize, 100},
		{"waiting for queue", 1 << 20, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server, client := net.Pipe()
			defer server.Close()
			defer client.Close()

			cfg := quietPingConfig()
			cfg.MaxSendBufferBytes = tc.budget
			mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
				{ID: 0x01, Priority: 1, SendQueueCapacity: tc.capacity},
			}, func(byte, []byte) {}, func(error) {}, cfg)
			mconn.SetLogger(log.NewTestingLogger(t))
			require.NoError(t, mconn.Start())

			// The peer never reads, so fill the budget or the queue. sendRoutine
			// takes at most one message off the queue before the write blocks,
			// so fill again once it has.
			msg := make([]byte, messageSize)
			for mconn.TrySend(0x01, msg) {
			}
			time.Sleep(50 * time.Millisecond)
			for mconn.TrySend(0x01, msg) {
			}

			sent := make(chan bool, 1)
			go func() { sent <- mconn.Send(0x01, msg) }()
			select {
			case <-sent:
				t.Fatal("Send returned while the connection was full")
			case <-time.After(100 * time.Millisecond):
			}

			require.NoError(t, mconn.Stop())
			select {
			case ok := <-sent:
				assert.False(t, ok)
			case <-time.After(time.Second):
				t.Fatal("Send kept waiting after the connection stopped")
			}
		})
	}
}

func TestSendEnqueuesWhenTimerAndQueueAreBothReady(t *testing.T) {
	t.Parallel()

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	cfg := quietPingConfig()
	cfg.MaxSendBufferBytes = 1 << 20
	mconn := NewMConnectionWithConfig(client, []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 100},
	}, func(byte, []byte) {}, func(error) {}, cfg)
	ch := mconn.channelsIdx[0x01]

	// A zero timeout has fired by the time the queue is tried, as when the
	// budget frees up right at the deadline. The queue has room, so every
	// send must succeed.
	msg := make([]byte, 1024)
	for i := range 50 {
		require.True(t, ch.sendBytes(msg, 0), "send %d", i)
	}
}
