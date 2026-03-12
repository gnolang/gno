package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
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

	logger       *slog.Logger
	server       *http.Server
	reverseProxy *httputil.ReverseProxy
	handlers     []PathHandler
	muHandlers   sync.RWMutex
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

	targetHost := proxyAddr.String()

	proxy := &PathInterceptor{
		logger:     logger,
		targetAddr: target,
		proxyAddr:  proxyAddr,
		reverseProxy: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = targetHost
				req.Host = targetHost
			},
			// Disable keep-alive so each request gets a fresh connection.
			// The target node may restart at any time (during lazy-load reload),
			// which would leave pooled connections dead.
			// Cost is negligible since the target is always localhost.
			Transport: &http.Transport{DisableKeepAlives: true},
		},
	}

	proxy.server = &http.Server{
		Handler: proxy,
	}

	go proxy.server.Serve(proxyListener)

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

// ServeHTTP intercepts HTTP requests, extracts package paths, and forwards to the target.
func (proxy *PathInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle WebSocket upgrades via raw TCP pipe
	if isWebSocket(r) {
		proxy.handleWebSocket(w, r)
		return
	}

	// Read body for path interception
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		proxy.logger.Debug("body read failed", "error", err)
		http.Error(w, "failed to read body", http.StatusBadGateway)
		return
	}

	// Intercept paths — this may trigger a synchronous node reload
	if err := proxy.handleRequest(body); err != nil {
		proxy.logger.Debug("request handler warning", "error", err)
	}

	// Restore body for forwarding
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))

	// Forward to the target node (fresh connection per request)
	proxy.reverseProxy.ServeHTTP(w, r)
}

// handleWebSocket hijacks the client connection and pipes data to the target.
func (proxy *PathInterceptor) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Dial the target
	targetConn, err := net.Dial(proxy.proxyAddr.Network(), proxy.proxyAddr.String())
	if err != nil {
		proxy.logger.Debug("websocket upstream dial failed", "error", err)
		http.Error(w, "upstream dial failed", http.StatusBadGateway)
		return
	}

	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		proxy.logger.Debug("hijacking not supported")
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		targetConn.Close()
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		proxy.logger.Debug("hijack failed", "error", err)
		http.Error(w, "hijack failed", http.StatusInternalServerError)
		targetConn.Close()
		return
	}

	// Forward the original upgrade request to the target
	if err := r.Write(targetConn); err != nil {
		clientConn.Close()
		targetConn.Close()
		return
	}

	// Bidirectional copy
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(targetConn, clientConn)
		targetConn.Close()
	}()
	go func() {
		defer wg.Done()
		io.Copy(clientConn, targetConn)
		clientConn.Close()
	}()
	wg.Wait()
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

// Close closes the proxy server and listener.
func (proxy *PathInterceptor) Close() error {
	return proxy.server.Close()
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
	case "vm/qobject": // ignore
	default:
		return fmt.Errorf("unhandled: %q", path)
	}
	return nil
}

func cleanupPath(path string) string {
	path = gopath.Clean(path)
	// If path is a file, grab the directory instead
	if ext := gopath.Ext(path); ext != "" {
		path = gopath.Dir(path)
	}

	return path
}
