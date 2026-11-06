package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// newRPCProxy forwards /rpc to the node's own JSON-RPC listener.
//
// The listen address arrives as a tm2 address ("tcp://127.0.0.1:26657"), which
// is not a URL any HTTP client accepts, so the scheme is rewritten. A listener
// bound to 0.0.0.0 is dialled on the loopback instead: it is the same socket,
// and connecting to the wildcard address is not portable.
// rpcTargetURL turns the node's listen address into a URL the proxy can dial.
//
// GetRemoteAddress reports a tm2 listen address ("tcp://host:port"), which no
// HTTP client accepts, so the scheme is rewritten. A listener bound to 0.0.0.0
// is dialled on the loopback instead: it is the same socket, and the wildcard
// address is not a portable dial target.
func rpcTargetURL(remote string) (*url.URL, error) {
	addr := remote
	for _, scheme := range []string{"tcp://", "http://", "https://"} {
		addr = strings.TrimPrefix(addr, scheme)
	}
	if addr == "" {
		return nil, fmt.Errorf("empty rpc address")
	}
	addr = strings.Replace(addr, "0.0.0.0:", "127.0.0.1:", 1)

	target, err := url.Parse("http://" + addr)
	if err != nil {
		return nil, fmt.Errorf("parse rpc address %q: %w", remote, err)
	}
	return target, nil
}

// newRPCProxy forwards /rpc to the node's own JSON-RPC listener.
func newRPCProxy(remote string) (http.Handler, error) {
	target, err := rpcTargetURL(remote)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// StripPrefix so /rpc and /rpc/status reach the node as / and /status.
	// Both are registered by the caller: ServeMux treats "/rpc" and "/rpc/"
	// as different patterns, and a client may send either.
	return http.StripPrefix("/rpc", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		proxy.ServeHTTP(w, r)
	})), nil
}
