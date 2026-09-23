package gnoweb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeIPFSGateway(t *testing.T) {
	t.Parallel()

	valid := []struct{ in, want string }{
		{"", ""},
		{"  ", ""},
		{DefaultIPFSGateway, DefaultIPFSGateway}, // the default is already normalized
		{"https://ipfs.filebase.io/", "https://ipfs.filebase.io"},
		{" https://IPFS.Filebase.io ", "https://ipfs.filebase.io"},
		{"https://gw.example:8443", "https://gw.example:8443"},
		{"http://localhost:8080", "http://localhost:8080"},
		{"http://127.0.0.1:8080/", "http://127.0.0.1:8080"},
		{"http://[::1]:8080", "http://[::1]:8080"},
		{"http://ipfs.localhost:8080", "http://ipfs.localhost:8080"},
	}
	for _, tc := range valid {
		got, err := normalizeIPFSGateway(tc.in, "gno.land")
		require.NoError(t, err, "input %q", tc.in)
		assert.Equal(t, tc.want, got, "input %q", tc.in)
	}

	invalid := []string{
		"ipfs.filebase.io",                   // no scheme
		"http://ipfs.filebase.io",            // http off loopback
		"javascript:alert(1)",                // not a gateway at all
		"https://",                           // no host
		"https://user:pass@ipfs.filebase.io", // credentials would land in every page
		"https://ipfs.filebase.io/ipfs",      // path
		"https://ipfs.filebase.io?x=1",       // query
		"https://ipfs.filebase.io?",          // empty query
		"https://ipfs.filebase.io#x",         // fragment
		"https://ipfs .filebase.io",          // whitespace in host
		"https://gno.land",                   // the gnoweb domain
		"https://ipfs.GNO.land",              // a subdomain of it
	}
	for _, in := range invalid {
		_, err := normalizeIPFSGateway(in, "gno.land")
		assert.Error(t, err, "input %q", in)
	}
}

func TestNewRouterIPFSGateway(t *testing.T) {
	t.Parallel()

	const cid = "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"
	render := func(gateway string) (string, *AppConfig) {
		cfg := NewDefaultAppConfig()
		cfg.ChainID = "dev" // skips the chain-id lookup: no node needed
		cfg.IPFSGateway = gateway
		cfg.Aliases["/ipfs-page"] = AliasTarget{Value: "![x](ipfs://" + cid + ")", Kind: StaticMarkdown}

		router, err := NewRouter(log.NewTestingLogger(t), cfg)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ipfs-page", nil))
		require.Equal(t, http.StatusOK, rec.Code)
		return rec.Body.String(), cfg
	}

	out, cfg := render("https://gw.example/")
	assert.Equal(t, "https://gw.example", cfg.IPFSGateway, "normalized in place for the CSP")
	assert.Contains(t, out, `src="https://gw.example/ipfs/`+cid+`"`)

	out, _ = render("")
	assert.Contains(t, out, `src="ipfs://`+cid+`"`)

	bad := NewDefaultAppConfig()
	bad.ChainID = "dev"
	bad.IPFSGateway = "http://gw.example"
	_, err := NewRouter(log.NewTestingLogger(t), bad)
	assert.ErrorContains(t, err, "invalid IPFS gateway")
}
