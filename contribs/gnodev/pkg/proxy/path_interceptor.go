package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
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

// NewPathInterceptor creates a proxy loader with a logger and target address.
func NewPathInterceptor(logger *slog.Logger, target net.Addr) (*PathInterceptor, error) {
	// Create a lisnener with target addr
	porxyListener, err := net.Listen(target.Network(), target.String())
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s://%s", target.Network(), target.String())
	}

	// Find a new random port for the target
	targertListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s://%s", target.Network(), target.String())
	}
	proxyAddr := targertListener.Addr()
	// Immedialty close this listener after proxy init
	defer targertListener.Close()

	proxy := &PathInterceptor{
		listener:   porxyListener,
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

// ProxyAddress returns the network address of the proxy.
func (proxy *PathInterceptor) TargetAddress() string {
	return fmt.Sprintf("%s://%s", proxy.targetAddr.Network(), proxy.targetAddr.String())
}

// handleConnections manages incoming connections to the proxy.
func (proxy *PathInterceptor) handleConnections() {
	defer proxy.listener.Close()

	for {
		conn, err := proxy.listener.Accept()
		if err != nil {
			proxy.logger.Debug("failed to accept connection", "error", err)
			continue
		}
		go proxy.handleConnection(conn)
	}
}

// handleConnection processes a single connection.
func (proxy *PathInterceptor) handleConnection(inConn net.Conn) {
	defer inConn.Close()

	var buffer bytes.Buffer
	teeReader := io.TeeReader(inConn, &buffer)

	defer func() { proxy.forwardRequest(&buffer, inConn) }()

	request, err := http.ReadRequest(bufio.NewReader(teeReader))
	if err != nil {
		proxy.logger.Debug("failed to read HTTP request", "error", err)
		return
	}

	if request.Header.Get("Upgrade") == "websocket" {
		proxy.logger.Debug("WebSocket connection detected, forwarding directly")
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		proxy.logger.Warn("failed to read request body", "error", err)
		return
	}
	defer request.Body.Close()

	if err := proxy.handleRequest(body); err != nil {
		proxy.logger.Debug("error handling request", "error", err)
	}
}

// forwardRequest forwards the buffered request to the target address.
func (proxy *PathInterceptor) forwardRequest(buffer *bytes.Buffer, inConn net.Conn) {
	outConn, err := net.Dial(proxy.proxyAddr.Network(), proxy.proxyAddr.String())
	if err != nil {
		proxy.logger.Error("failed to connect to remote socket", "address", proxy.proxyAddr, "error", err)
		return
	}
	defer outConn.Close()

	if buffer.Len() > 0 {
		if _, err := outConn.Write(buffer.Bytes()); err != nil {
			proxy.logger.Error("unable to write to socket", "error", err)
			return
		}
	}

	go io.Copy(outConn, inConn)
	io.Copy(inConn, outConn)
}

// handleRequest parses and processes the RPC request body.
func (proxy *PathInterceptor) handleRequest(body []byte) error {
	paths, err := parseRPCRequest(body)
	if err != nil {
		return fmt.Errorf("unable to parse rpc request: %w", err)
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
		return nil, fmt.Errorf("unable to unmarshal rpc request: %w", err)
	}

	if req.Method != "abci_query" {
		return nil, fmt.Errorf("not an abci query")
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
