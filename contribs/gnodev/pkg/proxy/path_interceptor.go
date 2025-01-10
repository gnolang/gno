package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type PathHandler func(path ...string)

type PathInterceptor struct {
	proxyAddr, targetAddr net.Addr

	logger     *slog.Logger
	listener   net.Listener
	handlers   []PathHandler
	muHandlers sync.RWMutex
}

// NewPathInterceptor creates a path proxy interceptor.
func NewPathInterceptor(logger *slog.Logger, target net.Addr) (*PathInterceptor, error) {
	// Create a listener with the target address
	proxyListener, err := net.Listen(target.Network(), target.String())
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s://%s", target.Network(), target.String())
	}

	// Find a new random port for the target
	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on tcp://127.0.0.1:0")
	}
	proxyAddr := targetListener.Addr()
	// Immediately close this listener after proxy init
	defer targetListener.Close()

	proxy := &PathInterceptor{
		listener:   proxyListener,
		logger:     logger,
		targetAddr: target,
		proxyAddr:  proxyAddr,
	}

	go proxy.handleConnections()

	return proxy, nil
}

// HandlePath adds a new path handler to the interceptor.
func (proxy *PathInterceptor) HandlePath(fn PathHandler) {
	proxy.muHandlers.Lock()
	defer proxy.muHandlers.Unlock()
	proxy.handlers = append(proxy.handlers, fn)
}

// ProxyAddress returns the network address of the proxy.
func (proxy *PathInterceptor) ProxyAddress() string {
	return fmt.Sprintf("%s://%s", proxy.proxyAddr.Network(), proxy.proxyAddr.String())
}

// TargetAddress returns the network address of the target.
func (proxy *PathInterceptor) TargetAddress() string {
	return fmt.Sprintf("%s://%s", proxy.targetAddr.Network(), proxy.targetAddr.String())
}

// handleConnections manages incoming connections to the proxy.
func (proxy *PathInterceptor) handleConnections() {
	defer proxy.listener.Close()

	for {
		conn, err := proxy.listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				proxy.logger.Debug("failed to accept connection", "error", err)
			}

			return
		}

		proxy.logger.Debug("new connection", "from", conn.RemoteAddr())
		go proxy.handleConnection(conn)
	}
}

// handleConnection processes a single connection.
func (proxy *PathInterceptor) handleConnection(inConn net.Conn) {
	defer inConn.Close()

	outConn, err := net.Dial(proxy.proxyAddr.Network(), proxy.proxyAddr.String())
	if err != nil {
		proxy.logger.Error("failed to connect to remote socket", "address", proxy.proxyAddr, "error", err)
		return
	}
	defer outConn.Close()

	// Redirect all the response directly to the incoming connection
	go func() {
		io.Copy(inConn, outConn)
		proxy.logger.Debug("incoming connection is closing", "from", inConn.RemoteAddr())
		inConn.Close()
	}()

	var buffer bytes.Buffer
	teeReader := io.TeeReader(inConn, &buffer)

	readerio := bufio.NewReader(teeReader)
	for {
		proxy.logger.Debug("reading request", "from", inConn.RemoteAddr())
		request, err := http.ReadRequest(readerio)
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				proxy.logger.Debug("connection closed", "from", inConn.RemoteAddr())
				return
			}

			proxy.logger.Debug("failed to read HTTP request", "error", err)
			// not an actual HTTP request, forward connection directly
			break
		}

		if request.Header.Get("Upgrade") == "websocket" {
			proxy.logger.Debug("WebSocket connection detected, forwarding directly")
			// WebSocket connection not supported (yet), forward connection directly
			break
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			proxy.logger.Warn("failed to read request body", "error", err)
			break
		}

		if err := proxy.handleRequest(body); err != nil {
			proxy.logger.Debug("error handling request", "error", err)
		}
		request.Body.Close()

		proxy.logger.Debug("forwarding request", "from", inConn.RemoteAddr(), "to", outConn.RemoteAddr())
		if _, err := outConn.Write(buffer.Bytes()); err != nil {
			proxy.logger.Error("unable to write to socket", "error", err)
			return // stoping here as we are not able to write to outgoing conn
		}

		// Reset buffer
		buffer.Reset()
	}

	// Forward what we read to the outbound connection
	proxy.logger.Debug("forwarding remaining connection", "from", inConn.RemoteAddr(), "to", outConn.RemoteAddr())
	if _, err := outConn.Write(buffer.Bytes()); err != nil {
		proxy.logger.Error("unable to write to socket", "error", err)
	}

	// Forward directly incoming conn to outgoing conn
	io.Copy(outConn, inConn)

	proxy.logger.Debug("connection ended", "from", inConn.RemoteAddr())
}

// handleRequest parses and processes the RPC request body.
func (proxy *PathInterceptor) handleRequest(body []byte) error {
	paths, err := parseRPCRequest(body)
	if err != nil {
		return fmt.Errorf("unable to parse RPC request: %w", err)
	}

	if len(paths) == 0 {
		return nil
	}
	proxy.logger.Debug("parsed request paths", "paths", paths)

	proxy.muHandlers.RLock()
	defer proxy.muHandlers.RUnlock()

	for _, handle := range proxy.handlers {
		handle(paths...)
	}

	return nil
}

// Close closes the proxy listener.
func (proxy *PathInterceptor) Close() error {
	return proxy.listener.Close()
}

type sDataQuery struct {
	Path string `json:"path"`
	Data []byte `json:"data,omitempty"`
}

// parseRPCRequest unmarshals and processes RPC requests, returning paths.
func parseRPCRequest(body []byte) ([]string, error) {
	var req rpctypes.RPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("unable to unmarshal RPC request: %w", err)
	}

	if req.Method != "abci_query" {
		return nil, fmt.Errorf("not an ABCI query")
	}

	var query sDataQuery
	if err := json.Unmarshal(req.Params, &query); err != nil {
		return nil, fmt.Errorf("unable to unmarshal params: %w", err)
	}

	return handleQuery(query)
}

// handleQuery processes the query and returns relevant paths.
func handleQuery(query sDataQuery) ([]string, error) {
	var paths []string

	switch query.Path {
	case ".app/simulate":
		var tx std.Tx
		if err := amino.Unmarshal(query.Data, &tx); err != nil {
			return nil, fmt.Errorf("unable to unmarshal tx: %w", err)
		}

		for _, msg := range tx.Msgs {
			switch msg := msg.(type) {
			case vm.MsgCall:
				paths = append(paths, msg.PkgPath)
			case vm.MsgRun:
				paths = append(paths, msg.Package.Path)
			}
		}
		return paths, nil

	case "vm/qrender", "vm/qfile":
		path, _, _ := strings.Cut(string(query.Data), ":")
		u, err := url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("unable to parse path: %w", err)
		}
		return []string{strings.TrimSpace(u.Path)}, nil

	default:
		return nil, fmt.Errorf("unhandled: %q", query.Path)
	}

	// XXX: handle more cases
}
