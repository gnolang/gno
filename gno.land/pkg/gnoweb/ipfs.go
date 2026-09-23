package gnoweb

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// DefaultIPFSGateway is the public gateway that ipfs:// URLs and URLs on
// retired public gateways (ipfs.io, dweb.link, ...) are rewritten to. See
// gno.land/adr/prxxxx_gnoweb_ipfs_gateway.md for how it was chosen.
const DefaultIPFSGateway = "https://ipfs.filebase.io"

// normalizeIPFSGateway validates an IPFS gateway setting and returns its
// origin, "scheme://host[:port]". An empty value disables IPFS rewriting.
//
// The gateway must be https (http only on loopback, for local dev), must
// carry no credentials, path, query or fragment, and must not be on the
// gnoweb domain: links there would be classified as internal, and a path
// gateway would serve any author's HTML from the same site.
func normalizeIPFSGateway(raw, domain string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	host := strings.ToLower(u.Hostname())
	domain = strings.ToLower(domain)
	switch {
	case u.Scheme != "https" && (u.Scheme != "http" || !isLoopbackHost(host)):
		return "", errors.New("must use https (http is only allowed on localhost)")
	case host == "":
		return "", errors.New("missing host")
	case u.User != nil:
		return "", errors.New("must not contain credentials")
	case (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "":
		return "", errors.New("must be an origin, without path, query or fragment")
	case domain != "" && (host == domain || strings.HasSuffix(host, "."+domain)):
		return "", fmt.Errorf("must not be on the gnoweb domain %q", domain)
	}

	return u.Scheme + "://" + strings.ToLower(u.Host), nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
