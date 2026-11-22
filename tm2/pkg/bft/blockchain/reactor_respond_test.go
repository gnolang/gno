package blockchain

import (
	"net"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	p2pmock "github.com/gnolang/gno/tm2/pkg/p2p/mock"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newServingReactor returns a reactor whose store holds height blocks of two
// txSize-byte txs. It writes the package config the parallel tests read, so
// the tests using it do not run in parallel.
func newServingReactor(t *testing.T, height int64, txSize int) *BlockchainReactor {
	t.Helper()

	config, _ = cfg.ResetTestRoot("blockchain_reactor_respond_test")
	t.Cleanup(func() { os.RemoveAll(config.RootDir) })

	genDoc, privVals := randGenesisDoc(1, false, 30)
	pair := newBlockchainReactorWithTxs(log.NewNoopLogger(), genDoc, privVals, height, func(h int64) (txs []types.Tx) {
		for i := range 2 {
			tx := make([]byte, txSize)
			tx[0], tx[1], tx[2] = byte(h), byte(h>>8), byte(i)
			txs = append(txs, tx)
		}
		return txs
	})
	t.Cleanup(func() { pair.app.Stop() })

	return pair.reactor
}

// newPeer returns a peer with a data store. The caller sets its sends.
func newPeer() *p2pmock.Peer {
	var data sync.Map

	return &p2pmock.Peer{
		IDFn:  func() p2pTypes.ID { return "peer" },
		SetFn: func(k string, v any) { data.Store(k, v) },
		GetFn: func(k string) any {
			v, _ := data.Load(k)
			return v
		},
	}
}

func TestRespondRoutineExits(t *testing.T) {
	for _, stopPeer := range []bool{true, false} {
		for _, retrying := range []bool{false, true} {
			name := "reactor stops"
			if stopPeer {
				name = "peer stops"
			}
			if retrying {
				name += " while retrying"
			} else {
				name += " while idle"
			}
			t.Run(name, func(t *testing.T) {
				bcR := newServingReactor(t, 1, 10)
				require.NoError(t, bcR.Start())
				t.Cleanup(func() { bcR.Stop() })

				peer := newPeer()
				peer.BaseService = *service.NewBaseService(nil, "peer", peer)
				require.NoError(t, peer.BaseService.Start())
				t.Cleanup(func() { peer.BaseService.Stop() })
				attempted := make(chan struct{}, 1)
				peer.SendFn = func(byte, []byte) bool {
					select {
					case attempted <- struct{}{}:
					default:
					}
					return false
				}

				requests := make(chan int64, 1)
				done := make(chan struct{})
				go func() {
					defer close(done)
					bcR.respondRoutine(peer, requests)
				}()
				if retrying {
					requests <- 1
					select {
					case <-attempted:
					case <-time.After(5 * time.Second):
						t.Fatal("respondRoutine did not attempt to send")
					}
				}

				if stopPeer {
					require.NoError(t, peer.BaseService.Stop())
				} else {
					require.NoError(t, bcR.Stop())
				}
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					t.Fatal("respondRoutine did not exit")
				}
			})
		}
	}
}

func TestBlockRequestsPastTheQueueAreDropped(t *testing.T) {
	bcR := newServingReactor(t, 1, 10)
	peer := newPeer()
	bcR.InitPeer(peer) // Nothing drains the queue without AddPeer.
	requests := peer.Get(blockRequestsKey).(chan int64)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range maxPendingRequestsPerPeer + 1 {
			bcR.Receive(BlockchainChannel, peer, amino.MustMarshalAny(&bcBlockRequestMessage{Height: 1}))
		}
	}()

	select {
	case <-done:
		require.Len(t, requests, maxPendingRequestsPerPeer)
	case <-time.After(5 * time.Second):
		<-requests // Unblock a regressed send before failing.
		<-done
		t.Fatal("Receive waited on a full request queue")
	}
}

// stopRecordingSwitch is a Switch that reports why it stops a peer.
type stopRecordingSwitch struct {
	p2p.Switch
	stopped chan error
}

func (s *stopRecordingSwitch) StopPeerForError(_ p2p.PeerConn, err error) { s.stopped <- err }

// TestUnreadableBlockStopsRequester checks that a request for a block the
// store cannot read stops the peer that asked for it instead of crashing the
// node.
func TestUnreadableBlockStopsRequester(t *testing.T) {
	bcR := newServingReactor(t, 1, 10)

	// Store height 1 with a part whose bytes no longer decode, as after a
	// torn write.
	block := bcR.store.LoadBlock(1)
	parts := block.MakePartSet(types.BlockPartSizeBytes)
	parts.GetPart(0).Bytes = []byte{0xff, 0xff, 0xff}
	unreadable := store.NewBlockStore(memdb.NewMemDB())
	unreadable.SaveBlock(block, parts, &types.Commit{})
	bcR.store = unreadable

	sw := &stopRecordingSwitch{stopped: make(chan error, 1)}
	bcR.SetSwitch(sw)
	peer := newPeer()
	peer.SendFn = func(byte, []byte) bool { return true }
	bcR.InitPeer(peer)
	bcR.AddPeer(peer)

	bcR.Receive(BlockchainChannel, peer, amino.MustMarshalAny(&bcBlockRequestMessage{Height: 1}))

	select {
	case err := <-sw.stopped:
		assert.ErrorContains(t, err, "answering a block request")
	case <-time.After(5 * time.Second):
		t.Fatal("the requesting peer was not stopped")
	}
}

// serveBlocks starts an MConnection on sock and wires it to bcR as a switch
// would.
func serveBlocks(t *testing.T, bcR *BlockchainReactor, sock net.Conn, cfg conn.MConnConfig, onError func(error)) {
	t.Helper()

	peer := newPeer()
	mc := conn.NewMConnectionWithConfig(sock, bcR.GetChannels(), func(ch byte, b []byte) {
		bcR.Receive(ch, peer, b)
	}, onError, cfg)
	mc.SetLogger(log.NewNoopLogger())
	peer.SendFn = mc.Send
	peer.TrySendFn = mc.TrySend
	// A running peer, so respondRoutine exits when the test ends.
	peer.BaseService = *service.NewBaseService(nil, "peer", peer)
	require.NoError(t, peer.BaseService.Start())
	t.Cleanup(func() { peer.BaseService.Stop() })

	bcR.InitPeer(peer)
	require.NoError(t, mc.Start())
	t.Cleanup(func() { mc.Stop() })
	bcR.AddPeer(peer)
}

// tcpPipe returns both ends of a loopback TCP connection. Unlike net.Pipe it
// buffers, so a ping can sit unread behind requests sent before it.
func tcpPipe(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		c, _ := ln.Accept()
		accepted <- c
	}()
	client, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(t, err)
	server := <-accepted
	require.NotNil(t, server)
	t.Cleanup(func() {
		client.Close()
		server.Close()
	})

	return server, client
}

// TestBlockResponsesWaitingForBudgetLeavePingsAnswered checks that a node
// whose block responses wait for send budget still reads the requester's
// pings. The upload rate and budget make the last responses wait well past
// the requester's pong deadline.
func TestBlockResponsesWaitingForBudgetLeavePingsAnswered(t *testing.T) {
	const heights = 12
	bcR := newServingReactor(t, heights, 50_000)
	descs := []*conn.ChannelDescriptor{bcR.GetChannels()[0]}
	serverSock, clientSock := tcpPipe(t)

	var failed atomic.Value
	onError := func(side string) func(error) {
		return func(err error) { failed.Store(side + ": " + err.Error()) }
	}

	serverCfg := conn.DefaultMConnConfig()
	serverCfg.SendRate = 400_000
	serverCfg.MaxSendBufferBytes = 256 << 10
	serveBlocks(t, bcR, serverSock, serverCfg, onError("server"))

	clientCfg := conn.DefaultMConnConfig()
	clientCfg.PingInterval = time.Second
	clientCfg.PongTimeout = 500 * time.Millisecond
	var received atomic.Int64
	client := conn.NewMConnectionWithConfig(clientSock, descs, func(_ byte, b []byte) {
		if msg, err := decodeMsg(b); err == nil {
			if _, ok := msg.(*bcBlockResponseMessage); ok {
				received.Add(1)
			}
		}
	}, onError("client"), clientCfg)
	client.SetLogger(log.NewNoopLogger())

	require.NoError(t, client.Start())
	defer client.Stop()

	for h := int64(1); h <= heights; h++ {
		require.True(t, client.Send(BlockchainChannel, amino.MustMarshalAny(&bcBlockRequestMessage{Height: h})))
	}

	assert.Eventually(t, func() bool {
		return received.Load() == heights || failed.Load() != nil
	}, 15*time.Second, 20*time.Millisecond)
	assert.Nil(t, failed.Load(), "connection dropped")
	assert.Equal(t, int64(heights), received.Load())
}

// TestBlockResponseSentAgainAfterSendFails checks that a response Send fails
// to queue is sent again: the requester never asks for it twice.
func TestBlockResponseSentAgainAfterSendFails(t *testing.T) {
	bcR := newServingReactor(t, 2, 10)
	require.NoError(t, bcR.Start())
	defer bcR.Stop()

	var (
		calls atomic.Int64
		sent  = make(chan []byte, 1)
	)
	peer := newPeer()
	peer.SendFn = func(_ byte, b []byte) bool {
		if msg, _ := decodeMsg(b); reflect.TypeOf(msg) != reflect.TypeFor[*bcBlockResponseMessage]() {
			return true // the status AddPeer sends
		}
		if calls.Add(1) == 1 {
			return false // as when Send's timeout passes before the budget has room
		}
		sent <- b
		return true
	}
	bcR.InitPeer(peer)
	bcR.AddPeer(peer)

	bcR.Receive(BlockchainChannel, peer, amino.MustMarshalAny(&bcBlockRequestMessage{Height: 1}))

	select {
	case b := <-sent:
		msg, err := decodeMsg(b)
		require.NoError(t, err)
		require.IsType(t, &bcBlockResponseMessage{}, msg)
		assert.Equal(t, int64(1), msg.(*bcBlockResponseMessage).Block.Height)
	case <-time.After(5 * time.Second):
		t.Fatal("the response was dropped after one failed Send")
	}
}

// TestRequesterThatNeverPongsIsDropped checks that a peer which keeps asking
// for blocks while reading slowly cannot keep its connection without
// answering pings: serving it must not count as our own recv stall.
func TestRequesterThatNeverPongsIsDropped(t *testing.T) {
	const heights = 12
	bcR := newServingReactor(t, heights, 50_000)
	serverSock, clientSock := net.Pipe()
	defer serverSock.Close()
	defer clientSock.Close()

	cfg := conn.DefaultMConnConfig()
	cfg.PingInterval = time.Second
	cfg.PongTimeout = 500 * time.Millisecond
	cfg.MaxSendBufferBytes = 256 << 10
	dropped := make(chan error, 1)
	serveBlocks(t, bcR, serverSock, cfg, func(err error) {
		select {
		case dropped <- err:
		default:
		}
	})

	// The raw client reads slowly and never answers a ping.
	go func() {
		buf := make([]byte, 4096)
		for {
			if _, err := clientSock.Read(buf); err != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	go func() {
		for h := int64(1); ; h = h%heights + 1 {
			msg := amino.MustMarshalAny(&bcBlockRequestMessage{Height: h})
			pkt := conn.PacketMsg{ChannelID: BlockchainChannel, EOF: 1, Bytes: msg}
			if _, err := amino.MarshalAnySizedWriter(clientSock, pkt); err != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case err := <-dropped:
		assert.ErrorContains(t, err, "pong timeout")
	case <-time.After(10 * time.Second):
		t.Fatal("a peer that never answers pings stayed connected")
	}
}

// onePeerSwitch is a Switch whose peer set holds one peer.
type onePeerSwitch struct {
	p2p.Switch
	p2p.PeerSet
	peer p2p.PeerConn
}

func (s *onePeerSwitch) Peers() p2p.PeerSet { return s }

func (s *onePeerSwitch) Get(id p2pTypes.ID) p2p.PeerConn {
	if id == s.peer.ID() {
		return s.peer
	}
	return nil
}

func (s *onePeerSwitch) Broadcast(byte, []byte) {}

func (s *onePeerSwitch) StopPeerForError(p2p.PeerConn, error) {}

// TestBlockRequestSentAgainAfterTrySendFails checks that a block request
// TrySend could not queue is sent again: the requester waits for its block
// and never asks twice.
func TestBlockRequestSentAgainAfterTrySendFails(t *testing.T) {
	bcR := newServingReactor(t, 0, 10)

	var (
		failed atomic.Int64 // the height of the request TrySend refused
		resent = make(chan struct{})
		once   sync.Once
	)
	peer := newPeer()
	peer.TrySendFn = func(_ byte, b []byte) bool {
		msg, err := decodeMsg(b)
		require.NoError(t, err)
		height := msg.(*bcBlockRequestMessage).Height
		if failed.CompareAndSwap(0, height) {
			return false // as when the connection's send budget is full
		}
		if height == failed.Load() {
			once.Do(func() { close(resent) })
		}
		return true
	}
	bcR.SetSwitch(&onePeerSwitch{peer: peer})
	require.NoError(t, bcR.Start())
	defer bcR.Stop()

	bcR.pool.SetPeerHeight(peer.ID(), 5)

	select {
	case <-resent:
	case <-time.After(5 * time.Second):
		t.Fatalf("the request for height %d was dropped after one failed TrySend", failed.Load())
	}
}

// swappingSwitch is a Switch whose peer set holds one connection, which a
// test replaces with another under the same peer ID, as registerPeer does.
type swappingSwitch struct {
	p2p.Switch
	p2p.PeerSet
	peer atomic.Pointer[p2pmock.Peer]
}

func (s *swappingSwitch) Peers() p2p.PeerSet { return s }

func (s *swappingSwitch) Get(id p2pTypes.ID) p2p.PeerConn {
	if p := s.peer.Load(); p != nil && p.ID() == id {
		return p
	}
	return nil
}

func (s *swappingSwitch) Broadcast(byte, []byte) {}

func (s *swappingSwitch) StopPeerForError(p2p.PeerConn, error) {}

// TestBlockRequestNotSentOnLaterConnection checks that a block request
// TrySend could not queue on one connection is not sent on a later
// connection with the same peer ID: the pool would stop that peer for
// answering a request it did not make to that connection.
func TestBlockRequestNotSentOnLaterConnection(t *testing.T) {
	const peerHeight = 5
	bcR := newServingReactor(t, 0, 10)

	var refused sync.Map // heights the old connection could not queue
	oldConn := newPeer()
	oldConn.TrySendFn = func(_ byte, b []byte) bool {
		msg, err := decodeMsg(b)
		require.NoError(t, err)
		refused.Store(msg.(*bcBlockRequestMessage).Height, true)
		return false // as when the connection is stopping
	}
	var sent atomic.Int64
	newConn := newPeer()
	newConn.TrySendFn = func(byte, []byte) bool {
		sent.Add(1)
		return true
	}

	sw := &swappingSwitch{}
	sw.peer.Store(oldConn)
	bcR.SetSwitch(sw)
	require.NoError(t, bcR.Start())
	defer bcR.Stop()

	bcR.pool.SetPeerHeight(oldConn.ID(), peerHeight)
	require.Eventually(t, func() bool {
		n := 0
		refused.Range(func(any, any) bool { n++; return true })
		return n == peerHeight
	}, 5*time.Second, time.Millisecond, "the pool did not request every height")

	sw.peer.Store(newConn)
	time.Sleep(3 * sendRetryDelay)
	assert.Zero(t, sent.Load(), "requests made for the old connection went out on the new one")
}
