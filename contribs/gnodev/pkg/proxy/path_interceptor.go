package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"log/slog"
	"net"
	"net/http"
	gopath "path"
	"strconv"
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

// NewPathInterceptor creates a new path proxy interceptor.
func NewPathInterceptor(logger *slog.Logger, target net.Addr) (*PathInterceptor, error) {
	// Create a listener on the target address
	proxyListener, err := net.Listen(target.Network(), target.String())
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s://%s", target.Network(), target.String())
	}

	// Find on a new random port for the target
	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on tcp://127.0.0.1:0")
	}
	proxyAddr := targetListener.Addr()
	// Immediately close this listener after proxy initialization
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

		proxy.logger.Debug("new connection", "remote", conn.RemoteAddr())
		go proxy.handleConnection(conn)
	}
}

// handleConnection processes a single connection between client and target.
func (proxy *PathInterceptor) handleConnection(inConn net.Conn) {
	logger := proxy.logger.With(
		slog.String("in", inConn.RemoteAddr().String()),
	)
	defer inConn.Close()

	var buffer bytes.Buffer
	tee := io.TeeReader(inConn, &buffer)
	reader := bufio.NewReader(tee)

	// First, read and process the HTTP request (this may trigger a reload)
	request, err := http.ReadRequest(reader)
	if err != nil {
		logger.Debug("read request failed", "error", err)
		return
	}

	// Check for websocket upgrade - handle differently
	if isWebSocket(request) {
		proxy.handleWebSocketConnection(inConn, &buffer, logger)
		return
	}

	// Read and process the request body
	body, err := io.ReadAll(request.Body)
	request.Body.Close()
	if err != nil {
		logger.Debug("body read failed", "error", err)
		return
	}

	// Call handlers BEFORE establishing target connection
	// This allows handlers to reload the node if needed
	if err := proxy.handleRequest(body); err != nil {
		proxy.logger.Debug("request handler warning", "error", err)
	}

	// NOW establish connection to the target (after any reload has completed)
	outConn, err := net.Dial(proxy.proxyAddr.Network(), proxy.proxyAddr.String())
	if err != nil {
		logger.Error("target connection failed", "target", proxy.proxyAddr.String(), "error", err)
		return
	}
	defer outConn.Close()
	logger = logger.With(slog.String("out", outConn.RemoteAddr().String()))

	// Forward the buffered request
	if _, err := outConn.Write(buffer.Bytes()); err != nil {
		logger.Debug("request forward failed", "error", err)
		return
	}

	// Setup bidirectional copying for the rest of the connection
	var wg sync.WaitGroup
	wg.Add(2)

	// Response path (target -> client)
	go func() {
		defer wg.Done()
		_, err := io.Copy(inConn, outConn)
		if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
			logger.Debug("response copy error", "error", err)
		}
	}()

	// Request path (client -> target) - forward any remaining data
	go func() {
		defer wg.Done()
		_, err := io.Copy(outConn, inConn)
		if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
			logger.Debug("request copy error", "error", err)
		}
	}()

	wg.Wait()
	logger.Debug("connection closed")
}

// handleWebSocketConnection handles WebSocket upgrade requests
func (proxy *PathInterceptor) handleWebSocketConnection(inConn net.Conn, buffer *bytes.Buffer, logger *slog.Logger) {
	// For WebSocket, establish connection first then forward everything
	outConn, err := net.Dial(proxy.proxyAddr.Network(), proxy.proxyAddr.String())
	if err != nil {
		logger.Error("target connection failed for websocket", "target", proxy.proxyAddr.String(), "error", err)
		return
	}
	defer outConn.Close()

	// Forward the buffered request
	if _, err := outConn.Write(buffer.Bytes()); err != nil {
		logger.Debug("websocket request forward failed", "error", err)
		return
	}

	// Bidirectional copy for WebSocket
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(inConn, outConn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(outConn, inConn)
	}()

	wg.Wait()
}

// processHTTPRequests handles the HTTP request/response cycle.
func (proxy *PathInterceptor) processHTTPRequests(reader *bufio.Reader, buffer *bytes.Buffer, outConn net.Conn) error {
	for {
		request, err := http.ReadRequest(reader)
		if err != nil {
			return fmt.Errorf("read request failed: %w", err)
		}

		// Check for websocket upgrade
		if isWebSocket(request) {
			return errors.New("websocket upgrade requested")
		}

		// Read and process the request body
		body, err := io.ReadAll(request.Body)
		request.Body.Close()
		if err != nil {
			return fmt.Errorf("body read failed: %w", err)
		}

		if err := proxy.handleRequest(body); err != nil {
			proxy.logger.Debug("request handler warning", "error", err)
		}

		// Forward the original request bytes
		if _, err := outConn.Write(buffer.Bytes()); err != nil {
			return fmt.Errorf("request forward failed: %w", err)
		}

		buffer.Reset() // Prepare for the next request
	}
}

func isWebSocket(req *http.Request) bool {
	return strings.EqualFold(req.Header.Get("Upgrade"), "websocket")
}

type uniqPaths map[string]struct{}

func (upaths uniqPaths) list() []string {
	paths := make([]string, 0, len(upaths))
	for p := range upaths {
		paths = append(paths, p)
	}
	return paths
}

// Add a path to
func (upaths uniqPaths) addPath(path string) {
	path = cleanupPath(path)
	upaths[path] = struct{}{}
}

func (upaths uniqPaths) addPackageDeps(pkg *std.MemPackage) {
	fset := token.NewFileSet()
	for _, file := range pkg.Files {
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}

		f, err := parser.ParseFile(fset, file.Name, file.Body, parser.ImportsOnly)
		if err != nil {
			continue
		}

		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			upaths.addPath(path)
		}
	}
}

// handleRequest parses and processes the RPC request body.
func (proxy *PathInterceptor) handleRequest(body []byte) error {
	ps := make(uniqPaths)
	if err := parseRPCRequest(body, ps); err != nil {
		return fmt.Errorf("unable to parse RPC request: %w", err)
	}

	paths := ps.list()
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

// parseRPCRequest unmarshals and processes RPC requests, returning paths.
func parseRPCRequest(body []byte, upaths uniqPaths) error {
	var req rpctypes.RPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return fmt.Errorf("unable to unmarshal RPC request: %w", err)
	}

	switch req.Method {
	case "abci_query":
		var squery struct {
			Path string `json:"path"`
			Data []byte `json:"data,omitempty"`
		}
		if err := json.Unmarshal(req.Params, &squery); err != nil {
			return fmt.Errorf("unable to unmarshal params: %w", err)
		}

		return handleQuery(squery.Path, squery.Data, upaths)

	case "broadcast_tx_commit":
		var stx struct {
			Tx []byte `json:"tx"`
		}
		if err := json.Unmarshal(req.Params, &stx); err != nil {
			return fmt.Errorf("unable to unmarshal params: %w", err)
		}

		return handleTx(stx.Tx, upaths)
	}

	return fmt.Errorf("unhandled method: %q", req.Method)
}

// handleTx processes the transaction and returns relevant paths.
func handleTx(bz []byte, upaths uniqPaths) error {
	var tx std.Tx
	if err := amino.Unmarshal(bz, &tx); err != nil {
		return fmt.Errorf("unable to unmarshal tx: %w", err)
	}

	for _, msg := range tx.Msgs {
		switch msg := msg.(type) {
		case vm.MsgAddPackage:
			// NOTE: Do not add the package itself to avoid conflict.
			if msg.Package != nil {
				upaths.addPackageDeps(msg.Package)
			}
		case vm.MsgRun:
			// NOTE: Do not add the package itself to avoid conflict.
			if msg.Package != nil {
				upaths.addPackageDeps(msg.Package)
			}
		case vm.MsgCall:
			upaths.addPath(msg.PkgPath)
		}
	}

	return nil
}

// handleQuery processes the query and returns relevant paths.
func handleQuery(path string, data []byte, upaths uniqPaths) error {
	switch path {
	case ".app/simulate":
		return handleTx(data, upaths)

	case "vm/qrender", "vm/qfile", "vm/qfuncs", "vm/qeval":
		path, _, _ := strings.Cut(string(data), ":") // Cut arguments out
		upaths.addPath(path)
		return nil

	default:
		return fmt.Errorf("unhandled: %q", path)
	}

	// XXX: handle more cases
}

func cleanupPath(path string) string {
	path = gopath.Clean(path)
	// If path is a file, grab the directory instead
	if ext := gopath.Ext(path); ext != "" {
		path = gopath.Dir(path)
	}

	return path
}
