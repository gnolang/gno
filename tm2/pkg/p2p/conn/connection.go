package conn

import (
	"bufio"
	goerrors "errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/flow"
	"github.com/gnolang/gno/tm2/pkg/p2p/config"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/timer"
)

const (
	defaultMaxPacketMsgPayloadSize = 1024

	numBatchPacketMsgs = 10
	minReadBufferSize  = 1024
	minWriteBufferSize = 65536
	updateStats        = 2 * time.Second

	// some of these defaults are written in the user config
	// flushThrottle, sendRate, recvRate
	defaultFlushThrottle = 100 * time.Millisecond

	defaultSendQueueCapacity   = 1
	defaultRecvBufferCapacity  = 4096
	defaultRecvMessageCapacity = 22020096      // 21MB
	defaultSendRate            = int64(512000) // 500KB/s
	defaultRecvRate            = int64(512000) // 500KB/s
	defaultSendTimeout         = 10 * time.Second
	defaultPingInterval        = 60 * time.Second
	defaultPongTimeout         = 45 * time.Second

	// defaultRecvAssemblyTimeout bounds how long a single incomplete message
	// may be assembled from partial PacketMsgs (EOF=0) in a channel's recving
	// buffer. The deadline is anchored to the first partial packet of a message
	// and is not extended by subsequent packets, so a peer cannot pin memory in
	// the recving buffer indefinitely by dribbling partial packets while
	// answering pings. At the default 5MB/s recv rate any legitimate message
	// (blocks are bounded by MaxBlockDataBytes, ~2MB) completes in well under a
	// second -- measured, 0.4s for 2MB and ~1.6s at the 8MB MaxDataBytes ceiling
	// -- so this is orders of magnitude of headroom.
	//
	// It only measures time the *peer* had to make progress: time recvRoutine
	// spends inside a reactor callback is credited back, since nothing on the
	// connection is being read then and the delay is not the peer's doing.
	defaultRecvAssemblyTimeout = 30 * time.Second

	// defaultMaxRecvBufferBytes caps the total bytes buffered across all of a
	// connection's channels' recving buffers at any instant. Without it the
	// exposure is the sum of every channel's RecvMessageCapacity (~38MB on a
	// full node), letting a peer pin that much per connection.
	//
	// Because it is the only bound on that exposure, it is what determines a
	// node's aggregate worst case (this budget times the number of peer slots),
	// so it is sized against real traffic rather than against the channel caps.
	// A connection assembles at most one incomplete message per channel, and the
	// largest legitimate ones are:
	//
	//	blockchain      8MB   a committed block, bounded by MaxDataBytes,
	//	                      which is capped at MaxBlockDataBytesLimit
	//	consensus       4MB   4 channels x 1MB (the consensus maxMsgSize)
	//	mempool      8MB-128K a single tx, bounded by MaxTxBytes, which
	//	                      consensus-param validation requires to leave
	//	                      MaxBlockOverheadBytes inside MaxDataBytes
	//	discovery      ~KBs   at most maxPeersShared (30) addresses
	//
	// The worst case a *legal* chain configuration can reach is therefore
	// 8MB + 4MB + (8MB - 128KB) ~= 19.9MB, so 20MB covers it -- but only just,
	// and only because MaxTxBytes cannot reach MaxDataBytes. A chain at the 2MB
	// default sits at ~7MB. Anything that raises MaxBlockDataBytesLimit, adds a
	// channel, or loosens the MaxTxBytes bound has to raise this budget too.
	// (See TestDefaultBudgetCoversWorstLegalConfig, which fails if that drifts.)
	defaultMaxRecvBufferBytes = 20 << 20 // 20MB

	// recvAssemblyStallGrace is how long recvRoutine may spend in a reactor
	// callback before that time is credited back to the assembly deadlines. It
	// only exists to keep the per-message path free of the channel walk; any
	// stall long enough to matter against a 30s deadline is orders of magnitude
	// above it.
	recvAssemblyStallGrace = 50 * time.Millisecond
)

type (
	receiveCbFunc func(chID byte, msgBytes []byte)
	errorCbFunc   func(error)
)

/*
Each peer has one `MConnection` (multiplex connection) instance.

__multiplex__ *noun* a system or signal involving simultaneous transmission of
several messages along a single channel of communication.

Each `MConnection` handles message transmission on multiple abstract communication
`Channel`s.  Each channel has a globally unique byte id.
The byte id and the relative priorities of each `Channel` are configured upon
initialization of the connection.

There are two methods for sending messages:

	func (m MConnection) Send(chID byte, msgBytes []byte) bool {}
	func (m MConnection) TrySend(chID byte, msgBytes []byte}) bool {}

`Send(chID, msgBytes)` is a blocking call that waits until `msg` is
successfully queued for the channel with the given id byte `chID`, or until the
request times out.  The message `msg` is serialized using Go-Amino.

`TrySend(chID, msgBytes)` is a nonblocking call that returns false if the
channel's queue is full.

Inbound message bytes are handled with an onReceive callback function.
*/
type MConnection struct {
	service.BaseService

	conn          net.Conn
	bufConnReader *bufio.Reader
	bufConnWriter *bufio.Writer
	sendMonitor   *flow.Monitor
	recvMonitor   *flow.Monitor
	send          chan struct{}
	pong          chan struct{}
	channels      []*Channel
	channelsIdx   map[byte]*Channel
	onReceive     receiveCbFunc
	onError       errorCbFunc
	errored       uint32
	config        MConnConfig

	// Closing quitSendRoutine will cause the sendRoutine to eventually quit.
	// doneSendRoutine is closed when the sendRoutine actually quits.
	quitSendRoutine chan struct{}
	doneSendRoutine chan struct{}

	// Closing quitRecvRouting will cause the recvRouting to eventually quit.
	quitRecvRoutine chan struct{}

	// used to ensure FlushStop and OnStop
	// are safe to call concurrently.
	stopMtx sync.Mutex

	flushTimer *timer.ThrottleTimer // flush writes as necessary but throttled.
	pingTimer  *time.Ticker         // send pings periodically

	// close conn if pong is not received in pongTimeout
	pongTimer     *time.Timer
	pongTimeoutCh chan bool // true - timeout, false - peer sent pong

	chStatsTimer *time.Ticker // update channel stats periodically

	created time.Time // time of creation

	// recvBufferBytes is the total number of bytes currently buffered across all
	// channels' recving buffers (incomplete, partially-assembled messages). It
	// is only ever read or written from recvRoutine, so it needs no locking.
	recvBufferBytes int

	// recvStallSince is the unix-nano time at which recvRoutine entered a
	// reactor callback, or 0 when it is reading. Written by recvRoutine, read by
	// the channels' assembly-timer goroutines and by sendRoutine, neither of
	// which may charge a peer for time this end spent not reading.
	recvStallSince atomic.Int64

	// recvStalledTotal is the cumulative nanoseconds recvRoutine has spent in
	// reactor callbacks over the life of the connection. sendRoutine diffs it
	// across a ping/pong window to find how much of that window this end was
	// not reading for. Written by recvRoutine.
	recvStalledTotal atomic.Int64

	// lastPongAt is the unix-nano time the most recent pong was read. Written by
	// recvRoutine, read by sendRoutine to tell a genuinely unanswered ping from
	// one whose pong lost the race for pongTimeoutCh.
	lastPongAt atomic.Int64

	// pingSentAt and stalledAtPing snapshot the start of the current pong
	// window. Only touched from sendRoutine.
	pingSentAt    time.Time
	stalledAtPing int64

	_maxPacketMsgSize int
}

// MConnConfig is a MConnection configuration.
type MConnConfig struct {
	SendRate int64 `toml:"send_rate"`
	RecvRate int64 `toml:"recv_rate"`

	// Maximum payload size
	MaxPacketMsgPayloadSize int `toml:"max_packet_msg_payload_size"`

	// Interval to flush writes (throttled)
	FlushThrottle time.Duration `toml:"flush_throttle"`

	// Interval to send pings
	PingInterval time.Duration `toml:"ping_interval"`

	// Maximum wait time for pongs
	PongTimeout time.Duration `toml:"pong_timeout"`

	// RecvAssemblyTimeout bounds how long a single incomplete message may be
	// assembled from partial PacketMsgs in a channel's recving buffer. The
	// deadline is anchored to the first partial packet of a message and is not
	// extended by later packets, but time spent inside a reactor callback is
	// credited back. When <= 0 no assembly deadline is enforced.
	RecvAssemblyTimeout time.Duration `toml:"recv_assembly_timeout"`

	// MaxRecvBufferBytes caps the total bytes buffered across all of the
	// connection's channels' recving buffers at any instant. When <= 0 no total
	// budget is enforced (only the per-channel RecvMessageCapacity applies).
	MaxRecvBufferBytes int `toml:"max_recv_buffer_bytes"`
}

// DefaultMConnConfig returns the default config.
func DefaultMConnConfig() MConnConfig {
	return MConnConfig{
		SendRate:                defaultSendRate,
		RecvRate:                defaultRecvRate,
		MaxPacketMsgPayloadSize: defaultMaxPacketMsgPayloadSize,
		FlushThrottle:           defaultFlushThrottle,
		PingInterval:            defaultPingInterval,
		PongTimeout:             defaultPongTimeout,
		RecvAssemblyTimeout:     defaultRecvAssemblyTimeout,
		MaxRecvBufferBytes:      defaultMaxRecvBufferBytes,
	}
}

// MConfigFromP2P returns a multiplex connection configuration
// with fields updated from the P2PConfig
func MConfigFromP2P(cfg *config.P2PConfig) MConnConfig {
	mConfig := DefaultMConnConfig()
	mConfig.FlushThrottle = cfg.FlushThrottleTimeout
	mConfig.SendRate = cfg.SendRate
	mConfig.RecvRate = cfg.RecvRate
	mConfig.MaxPacketMsgPayloadSize = cfg.MaxPacketMsgPayloadSize
	mConfig.RecvAssemblyTimeout = cfg.RecvAssemblyTimeout
	mConfig.MaxRecvBufferBytes = cfg.MaxRecvBufferBytes

	return mConfig
}

// NewMConnection wraps net.Conn and creates multiplex connection
func NewMConnection(conn net.Conn, chDescs []*ChannelDescriptor, onReceive receiveCbFunc, onError errorCbFunc) *MConnection {
	return NewMConnectionWithConfig(
		conn,
		chDescs,
		onReceive,
		onError,
		DefaultMConnConfig())
}

// NewMConnectionWithConfig wraps net.Conn and creates multiplex connection with a config
func NewMConnectionWithConfig(conn net.Conn, chDescs []*ChannelDescriptor, onReceive receiveCbFunc, onError errorCbFunc, config MConnConfig) *MConnection {
	if config.PongTimeout >= config.PingInterval {
		panic("pongTimeout must be less than pingInterval (otherwise, next ping will reset pong timer)")
	}
	mconn := &MConnection{
		conn:          conn,
		bufConnReader: bufio.NewReaderSize(conn, minReadBufferSize),
		bufConnWriter: bufio.NewWriterSize(conn, minWriteBufferSize),
		sendMonitor:   flow.New(0, 0),
		recvMonitor:   flow.New(0, 0),
		send:          make(chan struct{}, 1),
		pong:          make(chan struct{}, 1),
		onReceive:     onReceive,
		onError:       onError,
		config:        config,
		created:       time.Now(),
	}

	// Create channels
	channelsIdx := map[byte]*Channel{}
	channels := []*Channel{}

	for _, desc := range chDescs {
		channel := newChannel(mconn, *desc)
		channelsIdx[channel.desc.ID] = channel
		channels = append(channels, channel)
	}
	mconn.channels = channels
	mconn.channelsIdx = channelsIdx

	mconn.BaseService = *service.NewBaseService(nil, "MConnection", mconn)

	// maxPacketMsgSize() is a bit heavy, so call just once
	mconn._maxPacketMsgSize = mconn.maxPacketMsgSize()

	return mconn
}

func (c *MConnection) SetLogger(l *slog.Logger) {
	c.BaseService.SetLogger(l)
	for _, ch := range c.channels {
		ch.SetLogger(l)
	}
}

// OnStart implements BaseService
func (c *MConnection) OnStart() error {
	if err := c.BaseService.OnStart(); err != nil {
		return err
	}
	c.flushTimer = timer.NewThrottleTimer("flush", c.config.FlushThrottle)
	c.pingTimer = time.NewTicker(c.config.PingInterval)
	c.pongTimeoutCh = make(chan bool, 1)
	c.chStatsTimer = time.NewTicker(updateStats)
	c.quitSendRoutine = make(chan struct{})
	c.doneSendRoutine = make(chan struct{})
	c.quitRecvRoutine = make(chan struct{})
	go c.sendRoutine()
	go c.recvRoutine()
	return nil
}

// stopServices stops the BaseService and timers and closes the quitSendRoutine.
// if the quitSendRoutine was already closed, it returns true, otherwise it returns false.
// It uses the stopMtx to ensure only one of FlushStop and OnStop can do this at a time.
func (c *MConnection) stopServices() (alreadyStopped bool) {
	c.stopMtx.Lock()
	defer c.stopMtx.Unlock()

	select {
	case <-c.quitSendRoutine:
		// already quit
		return true
	default:
	}

	select {
	case <-c.quitRecvRoutine:
		// already quit
		return true
	default:
	}

	c.BaseService.OnStop()
	c.flushTimer.Stop()
	c.pingTimer.Stop()
	c.chStatsTimer.Stop()

	// inform the recvRouting that we are shutting down
	close(c.quitRecvRoutine)
	close(c.quitSendRoutine)
	return false
}

// FlushStop replicates the logic of OnStop.
// It additionally ensures that all successful
// .Send() calls will get flushed before closing
// the connection.
func (c *MConnection) FlushStop() {
	if c.stopServices() {
		return
	}

	// this block is unique to FlushStop
	{
		// wait until the sendRoutine exits
		// so we dont race on calling sendSomePacketMsgs
		<-c.doneSendRoutine

		// Send and flush all pending msgs.
		// Since sendRoutine has exited, we can call this
		// safely
		eof := c.sendSomePacketMsgs()
		for !eof {
			eof = c.sendSomePacketMsgs()
		}
		c.flush()

		// Now we can close the connection
	}

	c.conn.Close() //nolint: errcheck

	// We can't close pong safely here because
	// recvRoutine may write to it after we've stopped.
	// Though it doesn't need to get closed at all,
	// we close it @ recvRoutine.

	// c.Stop()
}

// OnStop implements BaseService
func (c *MConnection) OnStop() {
	if c.stopServices() {
		return
	}

	c.conn.Close() //nolint: errcheck

	// We can't close pong safely here because
	// recvRoutine may write to it after we've stopped.
	// Though it doesn't need to get closed at all,
	// we close it @ recvRoutine.
}

func (c *MConnection) String() string {
	return fmt.Sprintf("MConn{%v}", c.conn.RemoteAddr())
}

func (c *MConnection) flush() {
	c.Logger.Debug("Flush", "conn", c)
	err := c.bufConnWriter.Flush()
	if err != nil {
		c.Logger.Error("MConnection flush failed", "err", err)
	}
}

// Catch panics, usually caused by remote disconnects.
func (c *MConnection) _recover() {
	if r := recover(); r != nil {
		c.Logger.Error("MConnection panicked", "err", r, "stack", string(debug.Stack()))
		c.stopForError(errors.New("recovered from panic: %v", r))
	}
}

func (c *MConnection) stopForError(r error) {
	c.Stop()
	if atomic.CompareAndSwapUint32(&c.errored, 0, 1) {
		if c.onError != nil {
			c.onError(r)
		}
	}
}

// Queues a message to be sent to channel.
func (c *MConnection) Send(chID byte, msgBytes []byte) bool {
	if !c.IsRunning() {
		return false
	}

	// Send message to channel.
	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error(fmt.Sprintf("Cannot send bytes, unknown channel %X", chID))
		return false
	}

	success := channel.sendBytes(msgBytes)
	if success {
		// Wake up sendRoutine if necessary
		select {
		case c.send <- struct{}{}:
		default:
		}
	} else {
		c.Logger.Debug("Send failed", "channel", chID, "conn", c, "msgBytes", fmt.Sprintf("%X", msgBytes))
	}
	return success
}

// Queues a message to be sent to channel.
// Nonblocking, returns true if successful.
func (c *MConnection) TrySend(chID byte, msgBytes []byte) bool {
	if !c.IsRunning() {
		return false
	}

	// Send message to channel.
	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error(fmt.Sprintf("Cannot send bytes, unknown channel %X", chID))
		return false
	}

	ok = channel.trySendBytes(msgBytes)
	if ok {
		// Wake up sendRoutine if necessary
		select {
		case c.send <- struct{}{}:
		default:
		}
	}

	return ok
}

// CanSend returns true if you can send more data onto the chID, false
// otherwise. Use only as a heuristic.
func (c *MConnection) CanSend(chID byte) bool {
	if !c.IsRunning() {
		return false
	}

	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error(fmt.Sprintf("Unknown channel %X", chID))
		return false
	}
	return channel.canSend()
}

// sendRoutine polls for packets to send from channels.
func (c *MConnection) sendRoutine() {
	defer c._recover()

FOR_LOOP:
	for {
		var _n int64
		var err error
	SELECTION:
		select {
		case <-c.flushTimer.Ch:
			// NOTE: flushTimer.Set() must be called every time
			// something is written to .bufConnWriter.
			c.flush()
		case <-c.chStatsTimer.C:
			for _, channel := range c.channels {
				channel.updateStats()
			}
		case <-c.pingTimer.C:
			c.Logger.Debug("Send Ping")
			_n, err = amino.MarshalAnySizedWriter(c.bufConnWriter, PacketPing{})
			if err != nil {
				break SELECTION
			}
			c.sendMonitor.Update(int(_n))
			c.Logger.Debug("Starting pong timer", "dur", c.config.PongTimeout)
			c.pingSentAt = time.Now()
			c.stalledAtPing = c.recvStalledTotal.Load()
			c.pongTimer = time.AfterFunc(c.config.PongTimeout, c.signalPongTimeout)
			c.flush()
		case timeout := <-c.pongTimeoutCh:
			switch {
			case !timeout:
				c.stopPongTimer()
			case c.lastPongAt.Load() > c.pingSentAt.UnixNano():
				// The peer did answer this ping; its signal just lost the race
				// for the single slot in pongTimeoutCh.
				c.Logger.Debug("Pong timer fired for an answered ping")
				c.stopPongTimer()
			case c.pongRemaining() > 0:
				// The pong may well be sitting in the socket unread: recvRoutine
				// was parked in a reactor callback for part of this window, and
				// that is our doing, not the peer's. Wait out the remainder.
				c.Logger.Debug("Pong window extended past a local recv stall")
				c.stopPongTimer()
				c.pongTimer = time.AfterFunc(c.pongRemaining(), c.signalPongTimeout)
			default:
				c.Logger.Debug("Pong timeout")
				err = errors.New("pong timeout")
			}
		case <-c.pong:
			c.Logger.Debug("Send Pong")
			_n, err = amino.MarshalAnySizedWriter(c.bufConnWriter, PacketPong{})
			if err != nil {
				break SELECTION
			}
			c.sendMonitor.Update(int(_n))
			c.flush()
		case <-c.quitSendRoutine:
			break FOR_LOOP
		case <-c.send:
			// Send some PacketMsgs
			eof := c.sendSomePacketMsgs()
			if !eof {
				// Keep sendRoutine awake.
				select {
				case c.send <- struct{}{}:
				default:
				}
			}
		}

		if !c.IsRunning() {
			break FOR_LOOP
		}
		if err != nil {
			c.Logger.Error("Connection failed @ sendRoutine", "conn", c, "err", err)
			c.stopForError(err)
			break FOR_LOOP
		}
	}

	// Cleanup
	c.stopPongTimer()
	close(c.doneSendRoutine)
}

// Returns true if messages from channels were exhausted.
// Blocks in accordance to .sendMonitor throttling.
func (c *MConnection) sendSomePacketMsgs() bool {
	// Block until .sendMonitor says we can write.
	// Once we're ready we send more than we asked for,
	// but amortized it should even out.
	c.sendMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.SendRate), true)

	// Now send some PacketMsgs.
	for range numBatchPacketMsgs {
		if c.sendPacketMsg() {
			return true
		}
	}
	return false
}

// Returns true if messages from channels were exhausted.
func (c *MConnection) sendPacketMsg() bool {
	// Choose a channel to create a PacketMsg from.
	// The chosen channel will be the one whose recentlySent/priority is the least.
	var leastRatio float32 = math.MaxFloat32
	var leastChannel *Channel
	for _, channel := range c.channels {
		// If nothing to send, skip this channel
		if !channel.isSendPending() {
			continue
		}
		// Get ratio, and keep track of lowest ratio.
		ratio := float32(channel.recentlySent) / float32(channel.desc.Priority)
		if ratio < leastRatio {
			leastRatio = ratio
			leastChannel = channel
		}
	}

	// Nothing to send?
	if leastChannel == nil {
		return true
	}
	// c.Logger.Info("Found a msgPacket to send")

	// Make & send a PacketMsg from this channel
	_n, err := leastChannel.writePacketMsgTo(c.bufConnWriter)
	if err != nil {
		c.Logger.Error("Failed to write PacketMsg", "err", err)
		c.stopForError(err)
		return true
	}
	c.sendMonitor.Update(int(_n))
	c.flushTimer.Set()
	return false
}

// recvRoutine reads PacketMsgs and reconstructs the message using the channels' "recving" buffer.
// After a whole message has been assembled, it's pushed to onReceive().
// Blocks depending on how the connection is throttled.
// Otherwise, it never blocks.
func (c *MConnection) recvRoutine() {
	defer c._recover()

FOR_LOOP:
	for {
		// Block until .recvMonitor says we can read.
		c.recvMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.RecvRate), true)

		// Peek into bufConnReader for debugging
		/*
			if numBytes := c.bufConnReader.Buffered(); numBytes > 0 {
				bz, err := c.bufConnReader.Peek(min(numBytes, 100))
				if err == nil {
					// return
				} else {
					c.Logger.Debug("Error peeking connection buffer", "err", err)
					// return nil
				}
				c.Logger.Info("Peek connection buffer", "numBytes", numBytes, "bz", bz)
			}
		*/

		// Read packet type
		var packet Packet
		var _n int64
		var err error
		_n, err = amino.UnmarshalSizedReader(c.bufConnReader, &packet, int64(c._maxPacketMsgSize))
		c.recvMonitor.Update(int(_n))

		if err != nil {
			// stopServices was invoked and we are shutting down
			// receiving is expected to fail since we will close the connection
			select {
			case <-c.quitRecvRoutine:
				break FOR_LOOP
			default:
			}

			if c.IsRunning() {
				if goerrors.Is(err, io.EOF) {
					c.Logger.Info("Connection is closed @ recvRoutine (likely by the other side)", "conn", c)
				} else {
					c.Logger.Error("Connection failed @ recvRoutine (reading byte)", "conn", c, "err", err)
				}
				c.stopForError(err)
			}
			break FOR_LOOP
		}

		// Read more depending on packet type.
		switch pkt := packet.(type) {
		case PacketPing:
			// TODO: prevent abuse, as they cause flush()'s.
			// https://github.com/tendermint/tendermint/issues/1190
			c.Logger.Debug("Receive Ping")
			select {
			case c.pong <- struct{}{}:
			default:
				// never block
			}
		case PacketPong:
			c.Logger.Debug("Receive Pong")
			// Record the arrival as well as signalling it. pongTimeoutCh holds
			// one value and both sends are non-blocking, so a timer firing first
			// masks the pong that answered it; sendRoutine cross-checks this.
			c.lastPongAt.Store(time.Now().UnixNano())

			select {
			case c.pongTimeoutCh <- false:
			default:
				// never block
			}
		case PacketMsg:
			channel, ok := c.channelsIdx[pkt.ChannelID]
			if !ok || channel == nil {
				err := fmt.Errorf("unknown channel %X", pkt.ChannelID)
				c.Logger.Error("Connection failed @ recvRoutine", "conn", c, "err", err)
				c.stopForError(err)
				break FOR_LOOP
			}

			msgBytes, err := channel.recvPacketMsg(pkt)
			if err != nil {
				if c.IsRunning() {
					c.Logger.Error("Connection failed @ recvRoutine", "conn", c, "err", err)
					c.stopForError(err)
				}
				break FOR_LOOP
			}
			if msgBytes != nil {
				// NOTE: This means the reactor.Receive runs in the same thread as the p2p recv routine
				//
				// Some of those reactors block for a long time -- the mempool's
				// Receive waits on the mutex ApplyBlock holds across commit and
				// recheck, and the consensus reactor's does a blocking send into
				// peerMsgQueue. Nothing on this connection is read meanwhile, so
				// any channel mid-assembly would have our stall counted against
				// its deadline, and the peer's pong would sit unread in the
				// socket past its own deadline -- dropping an innocent peer
				// either way. Mark the stall for deadlines that expire during
				// it, and credit it to them afterwards.
				stallStart := time.Now()
				c.recvStallSince.Store(stallStart.UnixNano())

				c.onReceive(pkt.ChannelID, msgBytes)

				// Accumulate before clearing: a reader in between over-credits
				// this stall, which errs towards keeping the peer. The other
				// order would under-credit and drop it.
				stalled := time.Since(stallStart)
				c.recvStalledTotal.Add(int64(stalled))
				c.recvStallSince.Store(0)

				if stalled >= recvAssemblyStallGrace {
					for _, channel := range c.channels {
						channel.extendRecvAssemblyDeadline(stalled)
					}
				}
			}
		default:
			err := fmt.Errorf("unknown message type %v", reflect.TypeOf(packet))
			c.Logger.Error("Connection failed @ recvRoutine", "conn", c, "err", err)
			c.stopForError(err)
			break FOR_LOOP
		}
	}

	// Cleanup
	// Stop any per-channel assembly timers still pending. recvRoutine owns the
	// channels' recving state, so this is the safe place to release them.
	for _, channel := range c.channels {
		channel.stopRecvAssemblyTimer()
	}
	close(c.pong)
	for range c.pong {
		// Drain
	}
}

// signalPongTimeout tells sendRoutine the pong window elapsed. sendRoutine
// decides what that means; see pongRemaining.
func (c *MConnection) signalPongTimeout() {
	select {
	case c.pongTimeoutCh <- true:
	default:
	}
}

// pongRemaining reports how much of the current pong window is left once time
// recvRoutine spent inside reactor callbacks is discounted. A pong cannot be
// read while recvRoutine is parked in one, so charging that time to the peer
// drops a healthy connection for a stall on our side -- the same reasoning as
// the recv assembly deadline, on the other end of the same read loop.
// Only called from sendRoutine.
func (c *MConnection) pongRemaining() time.Duration {
	stalled := time.Duration(c.recvStalledTotal.Load() - c.stalledAtPing)

	if since := c.recvStallSince.Load(); since != 0 {
		// A stall in progress is not in recvStalledTotal yet.
		stalled += time.Since(time.Unix(0, since))
	}

	return c.config.PongTimeout - (time.Since(c.pingSentAt) - stalled)
}

// not goroutine-safe
func (c *MConnection) stopPongTimer() {
	if c.pongTimer != nil {
		_ = c.pongTimer.Stop()
		c.pongTimer = nil
	}
}

// maxPacketMsgSize returns a maximum size of PacketMsg, including the overhead
// of amino encoding.
func (c *MConnection) maxPacketMsgSize() int {
	return len(amino.MustMarshalAnySized(PacketMsg{
		ChannelID: 0x01,
		EOF:       1,
		Bytes:     make([]byte, c.config.MaxPacketMsgPayloadSize),
	})) + 10 // leave room for changes in amino
}

type ConnectionStatus struct {
	Duration    time.Duration
	SendMonitor flow.Status
	RecvMonitor flow.Status
	Channels    []ChannelStatus
}

type ChannelStatus struct {
	ID                byte
	SendQueueCapacity int
	SendQueueSize     int
	Priority          int
	RecentlySent      int64
}

func (c *MConnection) Status() ConnectionStatus {
	var status ConnectionStatus
	status.Duration = time.Since(c.created)
	status.SendMonitor = c.sendMonitor.Status()
	status.RecvMonitor = c.recvMonitor.Status()
	status.Channels = make([]ChannelStatus, len(c.channels))
	for i, ch := range c.channels {
		channel := ch
		status.Channels[i] = ChannelStatus{
			ID:                channel.desc.ID,
			SendQueueCapacity: cap(channel.sendQueue),
			SendQueueSize:     int(atomic.LoadInt32(&channel.sendQueueSize)),
			Priority:          channel.desc.Priority,
			RecentlySent:      atomic.LoadInt64(&channel.recentlySent),
		}
	}
	return status
}

// -----------------------------------------------------------------------------

type ChannelDescriptor struct {
	ID                  byte
	Priority            int
	SendQueueCapacity   int
	RecvBufferCapacity  int
	RecvMessageCapacity int
}

func (chDesc ChannelDescriptor) FillDefaults() (filled ChannelDescriptor) {
	if chDesc.SendQueueCapacity == 0 {
		chDesc.SendQueueCapacity = defaultSendQueueCapacity
	}
	if chDesc.RecvBufferCapacity == 0 {
		chDesc.RecvBufferCapacity = defaultRecvBufferCapacity
	}
	if chDesc.RecvMessageCapacity == 0 {
		chDesc.RecvMessageCapacity = defaultRecvMessageCapacity
	}
	filled = chDesc
	return
}

// TODO: lowercase.
// NOTE: not goroutine-safe.
type Channel struct {
	conn          *MConnection
	desc          ChannelDescriptor
	sendQueue     chan []byte
	sendQueueSize int32 // atomic.
	recving       []byte
	sending       []byte
	recentlySent  int64 // exponential moving average

	// recvAssemblyMtx guards the assembly deadline for the message currently
	// being built in recving. Unlike the rest of the recv state this cannot be
	// recvRoutine-only: the timer callback runs on its own goroutine and re-arms
	// itself when the deadline has moved.
	recvAssemblyMtx sync.Mutex
	// recvAssemblyTimer enforces RecvAssemblyTimeout for the message currently
	// being assembled in recving. It is started on the first partial packet of a
	// message and stopped on completion; it is deliberately never reset by
	// subsequent partial packets.
	recvAssemblyTimer *time.Timer
	// recvAssemblyDeadline is when that message stops getting the benefit of the
	// doubt. It is anchored to the first partial packet and only ever moves
	// forward by time this end spent not reading -- never by anything the peer
	// does. See extendRecvAssemblyDeadline.
	recvAssemblyDeadline time.Time

	maxPacketMsgPayloadSize int

	Logger *slog.Logger
}

func newChannel(conn *MConnection, desc ChannelDescriptor) *Channel {
	desc = desc.FillDefaults()
	if desc.Priority <= 0 {
		panic("Channel default priority must be a positive integer")
	}
	return &Channel{
		conn:                    conn,
		desc:                    desc,
		sendQueue:               make(chan []byte, desc.SendQueueCapacity),
		recving:                 make([]byte, 0, desc.RecvBufferCapacity),
		maxPacketMsgPayloadSize: conn.config.MaxPacketMsgPayloadSize,
	}
}

func (ch *Channel) SetLogger(l *slog.Logger) {
	ch.Logger = l
}

// Queues message to send to this channel.
// Goroutine-safe
// Times out (and returns false) after defaultSendTimeout
func (ch *Channel) sendBytes(bytes []byte) bool {
	select {
	case ch.sendQueue <- bytes:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return true
	case <-time.After(defaultSendTimeout):
		return false
	}
}

// Queues message to send to this channel.
// Nonblocking, returns true if successful.
// Goroutine-safe
func (ch *Channel) trySendBytes(bytes []byte) bool {
	select {
	case ch.sendQueue <- bytes:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return true
	default:
		return false
	}
}

// Goroutine-safe
func (ch *Channel) loadSendQueueSize() (size int) {
	return int(atomic.LoadInt32(&ch.sendQueueSize))
}

// Goroutine-safe
// Use only as a heuristic.
func (ch *Channel) canSend() bool {
	return ch.loadSendQueueSize() < defaultSendQueueCapacity
}

// Returns true if any PacketMsgs are pending to be sent.
// Call before calling nextPacketMsg()
// Goroutine-safe
func (ch *Channel) isSendPending() bool {
	if len(ch.sending) == 0 {
		if len(ch.sendQueue) == 0 {
			return false
		}
		ch.sending = <-ch.sendQueue
	}
	return true
}

// Creates a new PacketMsg to send.
// Not goroutine-safe
func (ch *Channel) nextPacketMsg() PacketMsg {
	packet := PacketMsg{}
	packet.ChannelID = ch.desc.ID
	maxSize := ch.maxPacketMsgPayloadSize
	packet.Bytes = ch.sending[:min(maxSize, len(ch.sending))]
	if len(ch.sending) <= maxSize {
		packet.EOF = byte(0x01)
		ch.sending = nil
		atomic.AddInt32(&ch.sendQueueSize, -1) // decrement sendQueueSize
	} else {
		packet.EOF = byte(0x00)
		ch.sending = ch.sending[min(maxSize, len(ch.sending)):]
	}
	return packet
}

// Writes next PacketMsg to w and updates c.recentlySent.
// Not goroutine-safe
func (ch *Channel) writePacketMsgTo(w io.Writer) (n int64, err error) {
	packet := ch.nextPacketMsg()
	n, err = amino.MarshalAnySizedWriter(w, packet)
	atomic.AddInt64(&ch.recentlySent, n)
	return
}

// Handles incoming PacketMsgs. It returns a message bytes if message is
// complete. NOTE message bytes may change on next call to recvPacketMsg.
// Not goroutine-safe
func (ch *Channel) recvPacketMsg(packet PacketMsg) ([]byte, error) {
	recvCap, recvReceived := ch.desc.RecvMessageCapacity, len(ch.recving)+len(packet.Bytes)
	if recvCap < recvReceived {
		return nil, fmt.Errorf("received message exceeds available capacity: %v < %v", recvCap, recvReceived)
	}

	// Enforce the total per-connection recving budget across all channels. The
	// per-channel RecvMessageCapacity check above only bounds a single channel;
	// without this, a peer can fill every channel's buffer at once (the sum of
	// all RecvMessageCapacity values, ~38MB on a full node -- it was ~130MB
	// before the blockchain reactor's envelope was right-sized).
	if budget := ch.conn.config.MaxRecvBufferBytes; budget > 0 {
		if total := ch.conn.recvBufferBytes + len(packet.Bytes); total > budget {
			return nil, fmt.Errorf("total recving buffer budget exceeded: %v > %v", total, budget)
		}
	}

	ch.recving = append(ch.recving, packet.Bytes...)
	ch.conn.recvBufferBytes += len(packet.Bytes)

	if packet.EOF == byte(0x01) {
		msgBytes := ch.recving

		// The message is complete: stop its assembly deadline and release the
		// buffered bytes from the total budget.
		ch.stopRecvAssemblyTimer()
		ch.conn.recvBufferBytes -= len(msgBytes)

		// Release the buffer. Reslicing (recving[:0]) alone would retain a grown
		// backing array for the lifetime of the connection, so a single large
		// message would pin that memory indefinitely. Re-allocating whenever the
		// array merely grew past RecvBufferCapacity has the opposite problem: it
		// puts a realloc-and-regrow on the path of *every* message larger than
		// that capacity, which is the common case on the channels carrying the
		// largest messages -- blockchain and consensus data configure 200KB
		// against multi-MB blocks, and the mempool channel leaves it at the 4KB
		// default against MaxTxBytes-sized txs. Measured on a 2MB message with a
		// 200KB capacity that costs 2.50ms, 9.07MB and 10 allocs, against 47.6us
		// and no allocations when the array is reused.
		//
		// So keep the array while the traffic on this channel is still using it,
		// and hand it back on the first message that is not: a channel steadily
		// carrying large messages stays allocation-free, while one that saw a
		// single outsized message releases the memory on its next ordinary
		// message. Reuse is safe: amino copies byte slices out while decoding, so
		// nothing downstream aliases recving.
		if cap(ch.recving) > ch.desc.RecvBufferCapacity && len(msgBytes)*2 < cap(ch.recving) {
			ch.recving = make([]byte, 0, ch.desc.RecvBufferCapacity)
		} else {
			ch.recving = ch.recving[:0]
		}
		return msgBytes, nil
	}

	// Partial packet: this message is still being assembled. Start the assembly
	// deadline on the first partial packet so a peer cannot pin the buffer
	// indefinitely by never sending EOF (interleaving pongs to stay alive).
	ch.startRecvAssemblyTimer()
	return nil, nil
}

// startRecvAssemblyTimer starts the assembly deadline for the message currently
// being assembled in recving, if it is not already running. The deadline is
// anchored to the first partial packet and is intentionally NOT reset by later
// partial packets, so a peer cannot keep an incomplete message buffered forever
// by dribbling packets. On expiry the whole connection is torn down.
func (ch *Channel) startRecvAssemblyTimer() {
	timeout := ch.conn.config.RecvAssemblyTimeout
	if timeout <= 0 {
		return
	}

	ch.recvAssemblyMtx.Lock()
	defer ch.recvAssemblyMtx.Unlock()

	if ch.recvAssemblyTimer != nil {
		return
	}

	ch.recvAssemblyDeadline = time.Now().Add(timeout)
	ch.recvAssemblyTimer = time.AfterFunc(timeout, ch.onRecvAssemblyTimeout)
}

// onRecvAssemblyTimeout runs on the timer goroutine when the assembly timer
// fires. The timer is scheduled against the deadline as it stood when it was
// armed, so it can fire early: extendRecvAssemblyDeadline moves the deadline
// without rescheduling, and a stall still in progress has not been credited at
// all yet. Either way the answer is to re-arm for what is left rather than tear
// the connection down.
func (ch *Channel) onRecvAssemblyTimeout() {
	ch.recvAssemblyMtx.Lock()

	if ch.recvAssemblyTimer == nil {
		// The message completed between the timer firing and this lock.
		ch.recvAssemblyMtx.Unlock()

		return
	}

	deadline := ch.recvAssemblyDeadline
	if stalledSince := ch.conn.recvStallSince.Load(); stalledSince != 0 {
		// recvRoutine is parked in a reactor callback right now. Nothing has
		// been read since -- from this peer or any other -- so that time is not
		// the peer's to answer for.
		deadline = deadline.Add(time.Since(time.Unix(0, stalledSince)))
	}

	if remaining := time.Until(deadline); remaining > 0 {
		ch.recvAssemblyTimer.Reset(remaining)
		ch.recvAssemblyMtx.Unlock()

		return
	}

	ch.recvAssemblyMtx.Unlock()

	// Outside the lock: stopForError tears the whole connection down, which ends
	// up back in recvRoutine's cleanup calling stopRecvAssemblyTimer.
	ch.conn.stopForError(fmt.Errorf(
		"recv assembly timeout: channel %X did not complete message within %v",
		ch.desc.ID, ch.conn.config.RecvAssemblyTimeout,
	))
}

// extendRecvAssemblyDeadline pushes an in-progress assembly deadline back by
// time recvRoutine spent inside a reactor callback. The timer is left scheduled
// where it is; when it fires it notices the deadline moved and re-arms, which
// costs one spurious wakeup and avoids racing Reset against the callback.
func (ch *Channel) extendRecvAssemblyDeadline(by time.Duration) {
	ch.recvAssemblyMtx.Lock()
	defer ch.recvAssemblyMtx.Unlock()

	if ch.recvAssemblyTimer == nil {
		return
	}

	ch.recvAssemblyDeadline = ch.recvAssemblyDeadline.Add(by)
}

// stopRecvAssemblyTimer stops and clears the assembly deadline if running.
func (ch *Channel) stopRecvAssemblyTimer() {
	ch.recvAssemblyMtx.Lock()
	defer ch.recvAssemblyMtx.Unlock()

	if ch.recvAssemblyTimer != nil {
		ch.recvAssemblyTimer.Stop()
		ch.recvAssemblyTimer = nil
	}
}

// Call this periodically to update stats for throttling purposes.
// Not goroutine-safe
func (ch *Channel) updateStats() {
	// Exponential decay of stats.
	// TODO: optimize.
	atomic.StoreInt64(&ch.recentlySent, int64(float64(atomic.LoadInt64(&ch.recentlySent))*0.8))
}

// ----------------------------------------
// Packet

type Packet interface {
	AssertPacket()
}

func (PacketPing) AssertPacket() {}
func (PacketPong) AssertPacket() {}
func (PacketMsg) AssertPacket()  {}

type PacketPing struct{}

type PacketPong struct{}

type PacketMsg struct {
	ChannelID byte
	EOF       byte // 1 means message ends here.
	Bytes     []byte
}

func (mp PacketMsg) String() string {
	return fmt.Sprintf("PacketMsg{%X:%X T:%X}", mp.ChannelID, mp.Bytes, mp.EOF)
}
