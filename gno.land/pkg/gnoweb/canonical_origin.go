package gnoweb

import (
	"fmt"
	"net/url"
	"strconv"
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
	// Parse checks only what follows the last colon, so a stray one before it
	// (gno.land:x:443) has to be refused here; a bracketed IPv6 host is fine.
	strayColon := strings.Contains(u.Hostname(), ":") && !strings.HasPrefix(u.Host, "[")
	// Parse checks only that the port is digits, so 0 or 99999 gets through.
	badPort := false
	if p := u.Port(); p != "" {
		n, err := strconv.Atoi(p)
		badPort = err != nil || n < 1 || n > 65535
	}
	if (scheme != "http" && scheme != "https") || host == "" || strings.HasSuffix(host, ":") || strayColon || badPort ||
		!strings.EqualFold((&url.URL{Scheme: u.Scheme, Host: u.Host}).String(), origin) {
		return "", fmt.Errorf("invalid canonical origin %q: want scheme://host[:port]", origin)
	}
	return scheme + "://" + host, nil
}
