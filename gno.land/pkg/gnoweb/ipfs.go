package gnoweb

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// DefaultIPFSGateway is the default of the gnoweb and gnodev gateway flags:
// the public gateway that ipfs:// URLs and URLs on the gateways in
// markdown's ipfsRewrittenGateways are rewritten to. See
// gno.land/adr/pr6270_gnoweb_ipfs_gateway.md for how it was chosen.
const DefaultIPFSGateway = "https://ipfs.filebase.io"

// gatewayHost matches a lowercase DNS name or IPv4 address: the hosts a CSP
// host-source can name. Wildcards, IPv6 literals and trailing dots fail.
var gatewayHost = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$`)

// normalizeIPFSGateway validates an IPFS gateway setting and returns its
// origin, "scheme://host[:port]". An empty value disables IPFS rewriting.
//
// The origin goes into the CSP, so it must be a plain DNS name or IPv4
// address with an optional valid port. The gateway must be https (http only
// on loopback, for local dev) and must carry no credentials, path, query or
// fragment. It must also not be on a host gnoweb is served from, nor a
// subdomain of one, where a path gateway would serve any author's HTML as
// that host: domain, the chain domain (gno.land), where gno.land's own gnoweb
// runs, and canonicalHost, the host of AppConfig.CanonicalOrigin when a
// deployment declares one. Either may be empty. A sibling host under the
// same registrable domain is not detected; that would need the public
// suffix list.
func normalizeIPFSGateway(raw, domain, canonicalHost string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		// url.Parse's errors quote the value or parts of it, credentials
		// included, so none of it is passed on.
		return "", errors.New("not a valid URL")
	}

	host := strings.ToLower(u.Hostname())
	switch {
	case !gatewayHost.MatchString(host):
		return "", errors.New("host must be a DNS name or an IPv4 address")
	case strings.HasSuffix(u.Host, ":") || (u.Port() != "" && !validPort(u.Port())):
		return "", errors.New("invalid port")
	case u.Scheme != "https" && (u.Scheme != "http" || !isLoopbackHost(host)):
		return "", errors.New("must use https (http is only allowed on localhost)")
	case u.User != nil:
		return "", errors.New("must not contain credentials")
	case (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "":
		return "", errors.New("must be an origin, without path, query or fragment")
	case onHost(host, domain):
		return "", fmt.Errorf("must not be on the chain domain %q", domain)
	case onHost(host, canonicalHost):
		return "", fmt.Errorf("must not be on the canonical origin's host %q", canonicalHost)
	}

	return u.Scheme + "://" + strings.ToLower(u.Host), nil
}

// originHost returns the host of an origin such as normalizeCanonicalOrigin
// returns, or "" for an empty one.
func originHost(origin string) string {
	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// onHost reports whether host is base or a subdomain of it. An empty base
// matches nothing; a trailing dot on base is ignored.
func onHost(host, base string) bool {
	base = strings.TrimSuffix(strings.ToLower(base), ".")
	return base != "" && (host == base || strings.HasSuffix(host, "."+base))
}

func validPort(port string) bool {
	n, err := strconv.Atoi(port)
	return err == nil && n >= 1 && n <= 65535
}

func isLoopbackHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
