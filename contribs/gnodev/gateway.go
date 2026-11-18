package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// publicRPCMethods is every tm2 RPC route a staging node serves to the
// network. Anything else is refused at the gateway, so a route added to tm2
// later stays private until someone decides otherwise. The unsafe_* routes,
// dial_* and the websocket are deliberately absent.
var publicRPCMethods = map[string]bool{
	"health":               true,
	"status":               true,
	"net_info":             true,
	"blockchain":           true,
	"genesis":              true,
	"block":                true,
	"block_results":        true,
	"commit":               true,
	"tx":                   true,
	"validators":           true,
	"dump_consensus_state": true,
	"consensus_state":      true,
	"consensus_params":     true,
	"unconfirmed_txs":      true,
	"num_unconfirmed_txs":  true,
	"broadcast_tx_commit":  true,
	"broadcast_tx_sync":    true,
	"broadcast_tx_async":   true,
	"abci_query":           true,
	"abci_info":            true,
}

// rpcGateway serves the node's JSON-RPC on the web listener, so a staging
// deployment is one port, one hostname and one certificate, and the node's own
// RPC listener never has to leave the loopback.
//
// It answers both shapes a client sends: URI requests (GET /status) and
// JSON-RPC POSTs, single or batched. Each named method is checked against
// publicRPCMethods before anything reaches the node.
type rpcGateway struct {
	proxy   *httputil.ReverseProxy
	maxBody int64
}

// newRPCGateway proxies to the node's RPC listen address. The address arrives
// in tm2 form ("tcp://127.0.0.1:26657" or "unix:///path/to.sock"), which is
// not a URL an HTTP client accepts, so the scheme is rewritten, and a unix
// socket is dialled through the transport. A listener bound to 0.0.0.0 is
// dialled on the loopback: same socket, and dialling the wildcard is not
// portable.
func newRPCGateway(remote string, maxBody int64) (*rpcGateway, error) {
	if sock, ok := strings.CutPrefix(remote, "unix://"); ok {
		if sock == "" {
			return nil, errors.New("empty rpc socket path")
		}
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "unix"})
		proxy.Transport = &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", sock)
			},
		}
		return &rpcGateway{proxy: proxy, maxBody: maxBody}, nil
	}

	addr := remote
	for _, scheme := range []string{"tcp://", "http://", "https://"} {
		addr = strings.TrimPrefix(addr, scheme)
	}
	if addr == "" {
		return nil, errors.New("empty rpc address")
	}
	if strings.Contains(addr, "://") {
		return nil, fmt.Errorf("unsupported rpc address %q: want tcp:// or unix://", remote)
	}
	addr = strings.Replace(addr, "0.0.0.0:", "127.0.0.1:", 1)

	target, err := url.Parse("http://" + addr)
	if err != nil {
		return nil, fmt.Errorf("parse rpc address %q: %w", remote, err)
	}

	return &rpcGateway{
		proxy:   httputil.NewSingleHostReverseProxy(target),
		maxBody: maxBody,
	}, nil
}

// ServeHTTP expects a path already relative to the RPC root: "/" for
// JSON-RPC, "/<method>" for a URI request.
func (g *rpcGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := strings.Trim(r.URL.Path, "/")

	switch {
	case method == "" && r.Method == http.MethodPost:
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, g.maxBody))
		if err != nil {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		methods, err := jsonRPCMethods(body)
		if err != nil {
			http.Error(w, "invalid JSON-RPC request", http.StatusBadRequest)
			return
		}
		for _, m := range methods {
			if !publicRPCMethods[m] {
				denyRPC(w, m)
				return
			}
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))

	case method == "":
		// GET: the node's route listing. OPTIONS: a CORS preflight, which
		// names no method and is answered by the node's CORS handler. Either
		// way the body is dropped: tm2 runs a JSON-RPC body on "/" whatever
		// the HTTP method, and only a POST body has been checked above.
		if r.Method != http.MethodGet && r.Method != http.MethodOptions {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body, r.ContentLength = http.NoBody, 0

	case !publicRPCMethods[method]:
		denyRPC(w, method)
		return

	default:
		// A URI request carries its arguments in the query string.
		r.Body, r.ContentLength = http.NoBody, 0
	}

	r.URL.Path = "/" + method
	r.URL.RawPath = ""
	g.proxy.ServeHTTP(w, r)
}

// jsonRPCMethods returns the method of every request in a JSON-RPC body,
// accepting both a single request and a batch, the two shapes tm2 serves.
func jsonRPCMethods(body []byte) ([]string, error) {
	type request struct {
		Method string `json:"method"`
	}

	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var batch []request
		if err := json.Unmarshal(trimmed, &batch); err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			return nil, errors.New("empty batch")
		}
		methods := make([]string, len(batch))
		for i, req := range batch {
			methods[i] = req.Method
		}
		return methods, nil
	}

	var single request
	if err := json.Unmarshal(trimmed, &single); err != nil {
		return nil, err
	}
	return []string{single.Method}, nil
}

func denyRPC(w http.ResponseWriter, method string) {
	http.Error(w, fmt.Sprintf("rpc method %q is not served by this gateway", method), http.StatusForbidden)
}

// isJSONRPCPost reports whether r is a JSON-RPC call addressed to the web
// root, which is where gnokey, gnoclient and wallets send one when given
// https://<host> as their remote. gnoweb's own POSTs are form submissions,
// so the content type is enough to tell them apart. A browser's CORS
// preflight for such a POST counts too, since gnoweb does not answer OPTIONS.
func isJSONRPCPost(r *http.Request) bool {
	if r.URL.Path != "/" && r.URL.Path != "" {
		return false
	}
	switch r.Method {
	case http.MethodPost:
		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		return err == nil && mt == "application/json"
	case http.MethodOptions:
		return r.Header.Get("Access-Control-Request-Method") == http.MethodPost
	}
	return false
}
