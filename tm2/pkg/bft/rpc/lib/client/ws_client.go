package rpcclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gorilla/websocket"
)

const (
	defaultMaxReconnectAttempts = 25
	defaultWriteWait            = 0
	defaultReadWait             = 0
	defaultPingPeriod           = 0
)

var _ BatchClient = (*WSClient)(nil)

// WSClient is a WebSocket client. The methods of WSClient are safe for use by
// multiple goroutines.
type WSClient struct {
	service.BaseService

	conn *websocket.Conn

	Address  string // IP:PORT or /path/to/socket
	Endpoint string // /websocket/url/endpoint
	Dialer   func(string, string) (net.Conn, error)

	// Single user facing channel to read RPCResponses from, closed only when the client is being stopped.
	ResponsesCh chan types.RPCResponses

	// Callback, which will be called each time after successful reconnect.
	onReconnect func()

	// internal channels
	send            chan wrappedRPCRequests // user requests
	backlog         chan wrappedRPCRequests // stores user requests received during a conn failure
	reconnectAfter  chan error              // reconnect requests
	readRoutineQuit chan struct{}           // a way for readRoutine to close writeRoutine

	wg sync.WaitGroup

	mtx            sync.RWMutex
	sentLastPingAt time.Time
	reconnecting   bool

	// Maximum reconnect attempts (0 or greater; default: 25).
	maxReconnectAttempts int

	// Time allowed to write a message to the server. 0 means block until operation succeeds.
	writeWait time.Duration

	// Time allowed to read the next message from the server. 0 means block until operation succeeds.
	readWait time.Duration

	// Send pings to server with this period. Must be less than readWait. If 0, no pings will be sent.
	pingPeriod time.Duration

	// Support both ws and wss protocols
	protocol string

	idPrefix types.JSONRPCID

	// Since requests are sent in, and responses
	// are parsed asynchronously in this implementation,
	// the thread that is parsing the response result
	// needs to perform Amino decoding.
	// The thread that spawn the request, and the thread that parses
	// the request do not live in the same context.
	// This information (the object type) needs to be available at this moment,
	// so a map is used to temporarily store those types
	requestResultsMap map[types.JSONRPCID]any
}

func (c *WSClient) SendBatch(ctx context.Context, requests wrappedRPCRequests) (types.RPCResponses, error) {
	if err := c.Send(ctx, requests...); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, errors.New("timed out")
	case responses := <-c.ResponsesCh:
		return responses, nil
	}
}

// NewWSClient returns a new client. See the commentary on the func(*WSClient)
// functions for a detailed description of how to configure ping period and
// pong wait time. The endpoint argument must begin with a `/`.
// The function panics if the provided address is invalid.
func NewWSClient(remoteAddr, endpoint string, options ...func(*WSClient)) *WSClient {
	protocol, addr, err := toClientAddrAndParse(remoteAddr)
	if err != nil {
		panic(fmt.Sprintf("invalid remote %s: %s", remoteAddr, err))
	}
	// default to ws protocol, unless wss is explicitly specified
	if protocol != "wss" {
		protocol = "ws"
	}

	c := &WSClient{
		Address:  addr,
		Dialer:   makeHTTPDialer(remoteAddr),
		Endpoint: endpoint,

		maxReconnectAttempts: defaultMaxReconnectAttempts,
		readWait:             defaultReadWait,
		writeWait:            defaultWriteWait,
		pingPeriod:           defaultPingPeriod,
		protocol:             protocol,
		idPrefix:             types.JSONRPCStringID("ws-client-" + random.RandStr(8)),
		requestResultsMap:    make(map[types.JSONRPCID]any),
	}
	c.BaseService = *service.NewBaseService(nil, "WSClient", c)
	for _, option := range options {
		option(c)
	}
	return c
}

// MaxReconnectAttempts sets the maximum number of reconnect attempts before returning an error.
// It should only be used in the constructor and is not Goroutine-safe.
func MaxReconnectAttempts(max int) func(*WSClient) {
	return func(c *WSClient) {
		c.maxReconnectAttempts = max
	}
}

// ReadWait sets the amount of time to wait before a websocket read times out.
// It should only be used in the constructor and is not Goroutine-safe.
func ReadWait(readWait time.Duration) func(*WSClient) {
	return func(c *WSClient) {
		c.readWait = readWait
	}
}

// WriteWait sets the amount of time to wait before a websocket write times out.
// It should only be used in the constructor and is not Goroutine-safe.
func WriteWait(writeWait time.Duration) func(*WSClient) {
	return func(c *WSClient) {
		c.writeWait = writeWait
	}
}

// PingPeriod sets the duration for sending websocket pings.
// It should only be used in the constructor - not Goroutine-safe.
func PingPeriod(pingPeriod time.Duration) func(*WSClient) {
	return func(c *WSClient) {
		c.pingPeriod = pingPeriod
	}
}

// OnReconnect sets the callback, which will be called every time after
// successful reconnect.
func OnReconnect(cb func()) func(*WSClient) {
	return func(c *WSClient) {
		c.onReconnect = cb
	}
}

// String returns WS client full address.
func (c *WSClient) String() string {
	return fmt.Sprintf("%s (%s)", c.Address, c.Endpoint)
}

func (c *WSClient) GetIDPrefix() types.JSONRPCID {
	return c.idPrefix
}

// OnStart implements service.Service by dialing a server and creating read and
// write routines.
func (c *WSClient) OnStart() error {
	err := c.dial()
	if err != nil {
		return err
	}

	c.ResponsesCh = make(chan types.RPCResponses)

	c.send = make(chan wrappedRPCRequests)
	// 1 additional error may come from the read/write
	// goroutine depending on which failed first.
	c.reconnectAfter = make(chan error, 1)
	// capacity for 1 request. a user won't be able to send more because the send
	// channel is unbuffered.
	c.backlog = make(chan wrappedRPCRequests, 1)

	c.startReadWriteRoutines()
	go c.reconnectRoutine()

	return nil
}

// Stop overrides service.Service#Stop. There is no other way to wait until Quit
// channel is closed.
func (c *WSClient) Stop() error {
	if err := c.BaseService.Stop(); err != nil {
		return err
	}
	// only close user-facing channels when we can't write to them
	c.wg.Wait()
	close(c.ResponsesCh)

	return nil
}

// IsReconnecting returns true if the client is reconnecting right now.
func (c *WSClient) IsReconnecting() bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.reconnecting
}

// IsActive returns true if the client is running and not reconnecting.
func (c *WSClient) IsActive() bool {
	return c.IsRunning() && !c.IsReconnecting()
}

// Send the given RPC request to the server. Results will be available on
// ResponsesCh, errors, if any, on ErrorsCh. Will block until send succeeds or
// ctx.Done is closed.
func (c *WSClient) Send(ctx context.Context, requests ...*wrappedRPCRequest) error {
	select {
	case c.send <- requests:
		c.Logger.Info("sent requests", "num", len(requests))
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Call the given method. See Send description.
func (c *WSClient) Call(ctx context.Context, method string, params map[string]any, result any) error {
	id := generateRequestID(c.idPrefix)

	request, err := types.MapToRequest(id, method, params)
	if err != nil {
		return err
	}

	return c.Send(
		ctx,
		&wrappedRPCRequest{
			request: request,
			result:  result,
		},
	)
}

// CallWithArrayParams the given method with params in a form of array. See
// Send description.
func (c *WSClient) CallWithArrayParams(ctx context.Context, method string, params []any, result any) error {
	id := generateRequestID(c.idPrefix)

	request, err := types.ArrayToRequest(id, method, params)
	if err != nil {
		return err
	}

	return c.Send(
		ctx,
		&wrappedRPCRequest{
			request: request,
			result:  result,
		},
	)
}

// -----------
// Private methods

func (c *WSClient) dial() error {
	dialer := &websocket.Dialer{
		NetDial: c.Dialer,
		Proxy:   http.ProxyFromEnvironment,
	}
	rHeader := http.Header{}
	conn, _, err := dialer.Dial(c.protocol+"://"+c.Address+c.Endpoint, rHeader)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// reconnect tries to redial up to maxReconnectAttempts with exponential
// backoff.
func (c *WSClient) reconnect() error {
	attempt := 0

	c.mtx.Lock()
	c.reconnecting = true
	c.mtx.Unlock()
	defer func() {
		c.mtx.Lock()
		c.reconnecting = false
		c.mtx.Unlock()
	}()

	for {
		jitter := time.Duration(random.RandFloat64() * float64(time.Second)) // 1s == (1e9 ns)
		backoffDuration := jitter + ((1 << uint(attempt)) * time.Second)

		c.Logger.Info("reconnecting", "attempt", attempt+1, "backoff_duration", backoffDuration)
		time.Sleep(backoffDuration)

		err := c.dial()
		if err != nil {
			c.Logger.Error("failed to redial", "err", err)
		} else {
			c.Logger.Info("reconnected")
			if c.onReconnect != nil {
				go c.onReconnect()
			}
			return nil
		}

		attempt++

		if attempt > c.maxReconnectAttempts {
			return errors.Wrap(err, "reached maximum reconnect attempts")
		}
	}
}

func (c *WSClient) startReadWriteRoutines() {
	c.wg.Add(2)
	c.readRoutineQuit = make(chan struct{})
	go c.readRoutine()
	go c.writeRoutine()
}

func (c *WSClient) processBacklog() error {
	select {
	case wrappedRequests := <-c.backlog:
		if c.writeWait > 0 {
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
				c.Logger.Error("failed to set write deadline", "err", err)
			}
		}

		writeData := extractWriteData(wrappedRequests)

		if err := c.conn.WriteJSON(writeData); err != nil {
			c.Logger.Error("failed to resend request", "err", err)
			c.reconnectAfter <- err
			// requeue request
			c.backlog <- wrappedRequests
			return err
		}

		c.Logger.Info("resend a request", "req", writeData)
	default:
	}
	return nil
}

func (c *WSClient) reconnectRoutine() {
	for {
		select {
		case originalError := <-c.reconnectAfter:
			// wait until writeRoutine and readRoutine finish
			c.wg.Wait()
			if err := c.reconnect(); err != nil {
				c.Logger.Error("failed to reconnect", "err", err, "original_err", originalError)
				c.Stop()
				return
			}
			// drain reconnectAfter
		LOOP:
			for {
				select {
				case <-c.reconnectAfter:
				default:
					break LOOP
				}
			}
			err := c.processBacklog()
			if err == nil {
				c.startReadWriteRoutines()
			}

		case <-c.Quit():
			return
		}
	}
}

// The client ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *WSClient) writeRoutine() {
	var ticker *time.Ticker
	if c.pingPeriod > 0 {
		// ticker with a predefined period
		ticker = time.NewTicker(c.pingPeriod)
	} else {
		// ticker that never fires
		ticker = &time.Ticker{C: make(<-chan time.Time)}
	}

	defer func() {
		ticker.Stop()
		c.conn.Close()
		// err != nil {
		// ignore error; it will trigger in tests
		// likely because it's closing an already closed connection
		// }
		c.wg.Done()
	}()

	for {
		select {
		case wrappedRequests := <-c.send:
			if c.writeWait > 0 {
				if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
					c.Logger.Error("failed to set write deadline", "err", err)
				}
			}

			// Save the results for later lookups
			for _, req := range wrappedRequests {
				c.requestResultsMap[req.request.ID] = req.result
			}

			writeData := extractWriteData(wrappedRequests)

			if err := c.conn.WriteJSON(writeData); err != nil {
				c.Logger.Error("failed to send request", "err", err)
				c.reconnectAfter <- err
				// add request to the backlog, so we don't lose it
				c.backlog <- wrappedRequests
				return
			}
		case <-ticker.C:
			if c.writeWait > 0 {
				if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
					c.Logger.Error("failed to set write deadline", "err", err)
				}
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.Logger.Error("failed to write ping", "err", err)
				c.reconnectAfter <- err
				return
			}
			c.mtx.Lock()
			c.sentLastPingAt = time.Now()
			c.mtx.Unlock()
			c.Logger.Debug("sent ping")
		case <-c.readRoutineQuit:
			return
		case <-c.Quit():
			if err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				c.Logger.Error("failed to write message", "err", err)
			}
			return
		}
	}
}

// The client ensures that there is at most one reader to a connection by
// executing all reads from this goroutine.
func (c *WSClient) readRoutine() {
	defer func() {
		c.conn.Close()
		// err != nil {
		// ignore error; it will trigger in tests
		// likely because it's closing an already closed connection
		// }
		c.wg.Done()
	}()

	c.conn.SetPongHandler(func(string) error {
		/*
			TODO latency metrics
			// gather latency stats
			c.mtx.RLock()
			t := c.sentLastPingAt
			c.mtx.RUnlock()
		*/

		c.Logger.Debug("got pong")
		return nil
	})

	for {
		// reset deadline for every message type (control or data)
		if c.readWait > 0 {
			if err := c.conn.SetReadDeadline(time.Now().Add(c.readWait)); err != nil {
				c.Logger.Error("failed to set read deadline", "err", err)
			}
		}
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				return
			}

			c.Logger.Error("failed to read response", "err", err)
			close(c.readRoutineQuit)
			c.reconnectAfter <- err
			return
		}

		var responses types.RPCResponses

		// Try to unmarshal as a batch of responses first
		if err := json.Unmarshal(data, &responses); err != nil {
			// Try to unmarshal as a single response
			var response types.RPCResponse

			if err := json.Unmarshal(data, &response); err != nil {
				c.Logger.Error("failed to parse response", "err", err, "data", string(data))

				continue
			}

			result, ok := c.requestResultsMap[response.ID]
			if !ok {
				c.Logger.Error("response result not set", "id", response.ID, "response", response)

				continue
			}

			// Clear the expected result for this ID
			delete(c.requestResultsMap, response.ID)

			if err := unmarshalResponseIntoResult(&response, response.ID, result); err != nil {
				c.Logger.Error("failed to parse response", "err", err, "data", string(data))

				continue
			}

			responses = types.RPCResponses{response}
		}

		for _, userResponse := range responses {
			c.Logger.Info("got response", "resp", userResponse.Result)
		}

		// Combine a non-blocking read on BaseService.Quit with a non-blocking write on ResponsesCh to avoid blocking
		// c.wg.Wait() in c.Stop(). Note we rely on Quit being closed so that it sends unlimited Quit signals to stop
		// both readRoutine and writeRoutine
		select {
		case <-c.Quit():
		case c.ResponsesCh <- responses:
		}
	}
}

// extractWriteData extracts write data (single request or multiple requests)
// from the given wrapped requests set
func extractWriteData(wrappedRequests wrappedRPCRequests) any {
	if len(wrappedRequests) == 1 {
		return wrappedRequests[0].extractRPCRequest()
	}

	return wrappedRequests.extractRPCRequests()
}
