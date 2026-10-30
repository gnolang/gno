package gnoweb

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
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
		{"https://xn--bcher-kva.example", "https://xn--bcher-kva.example"}, // punycode
		{"http://ipfs.localhost:8080", "http://ipfs.localhost:8080"},
	}
	for _, tc := range valid {
		got, err := normalizeIPFSGateway(tc.in, "gno.land", "")
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
		"https://gno.land",                   // the chain domain
		"https://ipfs.GNO.land",              // a subdomain of it
		"https://gno.land.",                  // the chain domain, trailing dot
		"https://*",                          // a CSP wildcard
		"https://*.example",                  // a CSP wildcard
		"https://gw,example",                 // not a host
		`https://gw"x`,                       // not a host
		"https://ipfs.filebase.io:",          // empty port
		"https://gw.example:0",               // port out of range
		"https://gw.example:99999",           // port out of range
		"http://[::1]:8080",                  // IPv6: no CSP host-source for it
		"https://[2001:db8::1]",              // IPv6
	}
	for _, in := range invalid {
		_, err := normalizeIPFSGateway(in, "gno.land", "")
		assert.Error(t, err, "input %q", in)
	}

	// A deployment's canonical host is off-limits, like the chain domain.
	for _, in := range []string{"https://web.example.org", "https://ipfs.WEB.example.org"} {
		_, err := normalizeIPFSGateway(in, "gno.land", "web.example.org")
		assert.ErrorContains(t, err, "canonical origin", "input %q", in)
	}
	_, err := normalizeIPFSGateway("https://ipfs.gno.example.org", "gno.land", "gno.example.org.")
	assert.ErrorContains(t, err, "canonical origin", "a trailing dot on the canonical host")
	got, err := normalizeIPFSGateway("https://ipfs.filebase.io", "gno.land", "web.example.org")
	require.NoError(t, err)
	assert.Equal(t, "https://ipfs.filebase.io", got)
}

// secretGateways are invalid gateway settings carrying a secret, which no
// error or log line may echo.
var secretGateways = []string{
	"https://user:s3cret@gw.example",
	"https://x.mypinata.cloud?pinataGatewayToken=s3cret",
	"https://x.mypinata.cloud/#s3cret",
	"https://user:s3cret^x@gw.example", // url.Parse fails, and its error quotes the value
	"https://user:s3cret/x@gw.example", // or part of it, as a port
	"https://gw.example/s3cret",        // a token in the path
	"https://gw.example;token=s3cret",
	`https://gw.example\s3cret`,
}

func TestNewRouterIPFSGateway(t *testing.T) {
	t.Parallel()

	const cid = "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"
	render := func(gateway, page string) (string, *AppConfig) {
		cfg := NewDefaultAppConfig()
		cfg.ChainID = "dev" // skips the chain-id lookup: no node needed
		cfg.RenderConfig.IPFSGateway = gateway
		cfg.Aliases["/ipfs-page"] = AliasTarget{Value: page, Kind: StaticMarkdown}

		router, err := NewRouter(log.NewTestingLogger(t), cfg)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ipfs-page", nil))
		require.Equal(t, http.StatusOK, rec.Code)
		return rec.Body.String(), cfg
	}

	out, cfg := render("https://gw.example/", "![x](ipfs://"+cid+")")
	assert.Equal(t, "https://gw.example", cfg.RenderConfig.IPFSGateway, "normalized in place for the CSP")
	assert.Contains(t, out, `src="https://gw.example/ipfs/`+cid+`"`)

	out, _ = render("", "![x](ipfs://"+cid+")")
	assert.Contains(t, out, `src="ipfs://`+cid+`"`)

	// A URL stored on the default gateway follows an operator who moves to
	// another one, so the CSP never has to keep the old default.
	out, _ = render("https://gw.example", "![x]("+DefaultIPFSGateway+"/ipfs/"+cid+")")
	assert.Contains(t, out, `src="https://gw.example/ipfs/`+cid+`"`)

	bad := NewDefaultAppConfig()
	bad.ChainID = "dev"
	bad.RenderConfig.IPFSGateway = "http://gw.example"
	_, err := NewRouter(log.NewTestingLogger(t), bad)
	assert.ErrorContains(t, err, "invalid IPFS gateway")

	// The startup error never echoes the value, which may hold a secret.
	for _, secret := range secretGateways {
		bad.RenderConfig.IPFSGateway = secret
		_, err = NewRouter(log.NewTestingLogger(t), bad)
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "s3cret")
	}

	onCanonical := NewDefaultAppConfig()
	onCanonical.ChainID = "dev"
	onCanonical.CanonicalOrigin = "https://web.example.org"
	onCanonical.RenderConfig.IPFSGateway = "https://ipfs.web.example.org"
	_, err = NewRouter(log.NewTestingLogger(t), onCanonical)
	assert.ErrorContains(t, err, "canonical origin")

	// Only the gnoweb and gnodev flags default to a gateway.
	assert.Empty(t, NewDefaultAppConfig().RenderConfig.IPFSGateway)
}

func TestNewHTMLRendererIPFSGateway(t *testing.T) {
	t.Parallel()

	const cid = "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"
	render := func(gateway string) string {
		cfg := NewDefaultRenderConfig()
		cfg.IPFSGateway = gateway
		r := NewHTMLRenderer(log.NewTestingLogger(t), cfg, nil)
		var out bytes.Buffer
		_, err := r.RenderRealm(&out, &weburl.GnoURL{Path: "/r/test"}, []byte("![x](ipfs://"+cid+")"), RealmRenderContext{})
		require.NoError(t, err)
		return out.String()
	}

	assert.Contains(t, render("https://gw.example/"), `src="https://gw.example/ipfs/`+cid+`"`)
	// Not rendered as a relative URL: an invalid gateway is ignored.
	assert.Contains(t, render("gw.example"), `src="ipfs://`+cid+`"`)
}

func TestNewHTMLRendererIPFSGatewayLog(t *testing.T) {
	t.Parallel()

	for _, secret := range secretGateways {
		var logs bytes.Buffer
		cfg := NewDefaultRenderConfig()
		cfg.IPFSGateway = secret
		NewHTMLRenderer(slog.New(slog.NewTextHandler(&logs, nil)), cfg, nil)
		assert.Contains(t, logs.String(), "ignoring invalid IPFS gateway", secret)
		assert.NotContains(t, logs.String(), "s3cret")
	}
}
