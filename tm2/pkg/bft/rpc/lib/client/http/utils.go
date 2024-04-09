package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

const (
	protoHTTP  = "http"
	protoHTTPS = "https"
	protoWSS   = "wss"
	protoWS    = "ws"
	protoTCP   = "tcp"
)

// DefaultHTTPClient is used to create an http client with some default parameters.
// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://)
func defaultHTTPClient(remoteAddr string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			// Set to true to prevent GZIP-bomb DoS attacks
			DisableCompression: true,
			DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
				return makeHTTPDialer(remoteAddr)(network, addr)
			},
		},
	}
}

func makeHTTPDialer(remoteAddr string) func(string, string) (net.Conn, error) {
	protocol, address, err := parseRemoteAddr(remoteAddr)
	if err != nil {
		return func(_ string, _ string) (net.Conn, error) {
			return nil, err
		}
	}

	// net.Dial doesn't understand http/https, so change it to TCP
	switch protocol {
	case protoHTTP, protoHTTPS:
		protocol = protoTCP
	}

	return func(proto, addr string) (net.Conn, error) {
		return net.Dial(protocol, address)
	}
}

// protocol - client's protocol (for example, "http", "https", "wss", "ws", "tcp")
// trimmedS - rest of the address (for example, "192.0.2.1:25", "[2001:db8::1]:80") with "/" replaced with "."
func toClientAddrAndParse(remoteAddr string) (string, string, error) {
	protocol, address, err := parseRemoteAddr(remoteAddr)
	if err != nil {
		return "", "", err
	}

	// protocol to use for http operations, to support both http and https
	var clientProtocol string
	// default to http for unknown protocols (ex. tcp)
	switch protocol {
	case protoHTTP, protoHTTPS, protoWS, protoWSS:
		clientProtocol = protocol
	default:
		clientProtocol = protoHTTP
	}

	// replace / with . for http requests (kvstore domain)
	trimmedAddress := strings.Replace(address, "/", ".", -1)

	return clientProtocol, trimmedAddress, nil
}

func toClientAddress(remoteAddr string) (string, error) {
	clientProtocol, trimmedAddress, err := toClientAddrAndParse(remoteAddr)
	if err != nil {
		return "", err
	}

	return clientProtocol + "://" + trimmedAddress, nil
}

// network - name of the network (for example, "tcp", "unix")
// s - rest of the address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
// TODO: Deprecate support for IP:PORT or /path/to/socket
func parseRemoteAddr(remoteAddr string) (network string, s string, err error) {
	parts := strings.SplitN(remoteAddr, "://", 2)
	var protocol, address string
	switch {
	case len(parts) == 1:
		// default to tcp if nothing specified
		protocol, address = protoTCP, remoteAddr
	case len(parts) == 2:
		protocol, address = parts[0], parts[1]
	default:
		return "", "", fmt.Errorf("invalid addr: %s", remoteAddr)
	}

	return protocol, address, nil
}

// isOKStatus returns a boolean indicating if the response
// status code is between 200 and 299 (inclusive)
func isOKStatus(code int) bool { return code >= 200 && code <= 299 }
