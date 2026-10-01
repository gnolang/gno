package gnoweb

import (
	"fmt"
	"net/url"
	"strings"
)

// normalizeCanonicalOrigin lowercases origin and drops a default port and
// trailing slashes. Anything but a bare http(s) origin is an error.
func normalizeCanonicalOrigin(origin string) (string, error) {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if origin == "" {
		return "", nil
	}
	u, err := url.Parse(origin)
	if err != nil {
		return "", fmt.Errorf("invalid canonical origin %q: %w", origin, err)
	}
	scheme, host := strings.ToLower(u.Scheme), strings.ToLower(u.Host)
	host = strings.TrimSuffix(host, map[string]string{"https": ":443", "http": ":80"}[scheme])
	// Rebuilding from scheme and host rejects a path, query, credentials, and
	// the empty trailing "?" or "#" that Parse accepts.
	if (scheme != "http" && scheme != "https") || host == "" || strings.HasSuffix(host, ":") ||
		!strings.EqualFold((&url.URL{Scheme: u.Scheme, Host: u.Host}).String(), origin) {
		return "", fmt.Errorf("invalid canonical origin %q: want scheme://host[:port]", origin)
	}
	return scheme + "://" + host, nil
}
