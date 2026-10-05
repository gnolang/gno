package gnoweb

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/feature/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestOrigin(t *testing.T) {
	t.Parallel()

	trusted, err := state.ParseTrustedProxies([]string{"10.0.0.0/8"})
	require.NoError(t, err)
	const proxy = "10.0.0.2:4242"

	cases := []struct {
		name     string
		remote   string
		host     string
		tls      bool
		fwdProto string
		fwdHost  string
		want     string
	}{
		{
			name: "plain http no proxy",
			host: "127.0.0.1:8888",
			want: "http://127.0.0.1:8888",
		},
		{
			name: "plain https no proxy",
			host: "gno.land",
			tls:  true,
			want: "https://gno.land",
		},
		{
			name:     "behind proxy with X-Forwarded-Proto https",
			remote:   proxy,
			host:     "internal-pod-1",
			fwdProto: "https",
			fwdHost:  "gno.land",
			want:     "https://gno.land",
		},
		{
			name:     "behind proxy with X-Forwarded-Proto http",
			remote:   proxy,
			host:     "internal-pod-1",
			fwdProto: "http",
			fwdHost:  "preview.gno.land",
			want:     "http://preview.gno.land",
		},
		{
			name:     "X-Forwarded-Host with chain takes leftmost",
			remote:   proxy,
			host:     "backend",
			fwdProto: "https",
			fwdHost:  "gno.land, edge-1, edge-2",
			want:     "https://gno.land",
		},
		{
			// A cache in front of gnoweb does not key on this header, so
			// believing it from anyone would let one request poison the
			// links and AI prompts served to the next visitor.
			name:     "ignores X-Forwarded-Host from an untrusted peer",
			remote:   "203.0.113.7:4242",
			host:     "gno.land",
			fwdProto: "https",
			fwdHost:  "evil.example",
			want:     "https://gno.land",
		},
		{
			name:     "ignores invalid X-Forwarded-Proto",
			host:     "gno.land",
			tls:      true,
			fwdProto: "javascript",
			want:     "https://gno.land",
		},
		{
			name:    "empty X-Forwarded-Host falls back to Host",
			host:    "gno.land",
			tls:     true,
			fwdHost: "",
			want:    "https://gno.land",
		},
		{
			name:    "whitespace-only X-Forwarded-Host falls back to Host",
			host:    "gno.land",
			tls:     true,
			fwdHost: "   ",
			want:    "https://gno.land",
		},
		{
			name: "ipv6 host",
			host: "[::1]:8888",
			want: "http://[::1]:8888",
		},
		{
			name: "empty host returns empty (path-relative fallback)",
			host: "",
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &http.Request{
				RemoteAddr: tc.remote,
				Host:       tc.host,
				Header:     http.Header{},
			}
			if tc.tls {
				r.TLS = &tls.ConnectionState{}
			}
			if tc.fwdProto != "" {
				r.Header.Set("X-Forwarded-Proto", tc.fwdProto)
			}
			if tc.fwdHost != "" {
				r.Header.Set("X-Forwarded-Host", tc.fwdHost)
			}

			assert.Equal(t, tc.want, requestOrigin(r, trusted))
		})
	}
}
