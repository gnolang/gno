package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRemoteURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"tcp protocol", "tcp://127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"http protocol", "http://127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"https protocol", "https://rpc.gno.land:443", "https://rpc.gno.land:443"},
		{"no protocol", "127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"no protocol with domain", "rpc.gno.land:443", "http://rpc.gno.land:443"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := normalizeRemoteURL(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNormalizeRemoteURL_UnsupportedProtocol(t *testing.T) {
	t.Parallel()

	unsupportedCases := []struct {
		name  string
		input string
	}{
		{"unix protocol", "unix://var/run/gno.sock"},
		{"file protocol", "file:///path/to/file"},
		{"ftp protocol", "ftp://example.com"},
		{"ws protocol", "ws://example.com"},
	}

	for _, tc := range unsupportedCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Panics(t, func() {
				normalizeRemoteURL(tc.input)
			}, "Expected panic for unsupported protocol: %s", tc.input)
		})
	}
}

func TestSetupWeb(t *testing.T) {
	opts := defaultWebOptions
	opts.bind = "127.0.0.1:0"
	stdio := commands.NewDefaultIO()

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0o644)
	require.NoError(t, err)
	defer devNull.Close()

	stdio.SetOut(devNull)

	_, err = setupWeb(&opts, []string{}, stdio)
	require.NoError(t, err)

	opts.trustedProxies = "10.0.0.0/8, 192.0.2.1"
	_, err = setupWeb(&opts, []string{}, stdio)
	require.NoError(t, err)

	// A bad entry fails at startup instead of quietly trusting less.
	opts.trustedProxies = "10.0.0.0/8,not-an-ip"
	_, err = setupWeb(&opts, []string{}, stdio)
	require.ErrorContains(t, err, "not-an-ip")
}

// A realm notice text that renders to nothing must refuse to start, unless the
// notice is disabled.
func TestSetupWeb_RealmNoticeText(t *testing.T) {
	t.Setenv("GNOWEB_REALM_NOTICE_TEXT", "# heading only")
	stdio := commands.NewDefaultIO()
	stdio.SetOut(commands.WriteNopCloser(io.Discard))

	opts := defaultWebOptions
	opts.bind = "127.0.0.1:0"
	_, err := setupWeb(&opts, []string{}, stdio)
	require.ErrorContains(t, err, "invalid GNOWEB_REALM_NOTICE_TEXT")

	opts.noRealmNotice = true
	_, err = setupWeb(&opts, []string{}, stdio)
	require.NoError(t, err)
}

// TestSetupTrust checks that the trusted paths reach the app with the realm
// notice on or off: they also decide what a page's head may repeat.
func TestSetupTrust(t *testing.T) {
	t.Setenv("GNOWEB_REALM_NOTICE_TEXT", "")

	for _, noNotice := range []bool{false, true} {
		t.Run(fmt.Sprintf("no-realm-notice=%v", noNotice), func(t *testing.T) {
			cfg := defaultWebOptions
			cfg.noRealmNotice = noNotice
			appcfg := gnoweb.NewDefaultAppConfig()

			require.NoError(t, setupTrust(&cfg, appcfg, slog.New(slog.DiscardHandler)))
			assert.Equal(t, strings.Split(gnoweb.DefaultTrustedPaths, ","), appcfg.TrustedPaths)
			assert.Equal(t, !noNotice, appcfg.RealmNotice.Enabled())
		})
	}
}

// TestIndexCommunityFlag checks the flag reaches the config and defaults to
// none; the values themselves are TestCommunityIndexText's.
func TestIndexCommunityFlag(t *testing.T) {
	t.Parallel()

	cfg := defaultWebOptions
	fs := flag.NewFlagSet("gnoweb", flag.ContinueOnError)
	cfg.RegisterFlags(fs)
	require.NoError(t, fs.Parse(nil))
	assert.Equal(t, gnoweb.IndexNoCommunity, cfg.indexCommunity, "names are free: community pages stay out of search by default")
	require.NoError(t, fs.Parse([]string{"-index-community=registered"}))
	assert.Equal(t, gnoweb.IndexRegisteredCommunity, cfg.indexCommunity)
}

// TestNoIndexFlag checks -no-index reaches the app: a testnet that loses it
// silently stays in search. The startup log reports the policy in effect.
func TestNoIndexFlag(t *testing.T) {
	t.Setenv("GNOWEB_REALM_NOTICE_TEXT", "")

	for _, args := range [][]string{nil, {"-no-index"}} {
		cfg := defaultWebOptions
		fs := flag.NewFlagSet("gnoweb", flag.ContinueOnError)
		cfg.RegisterFlags(fs)
		require.NoError(t, fs.Parse(args))
		cfg.bind = "127.0.0.1:0"

		var out strings.Builder
		stdio := commands.NewDefaultIO()
		stdio.SetOut(commands.WriteNopCloser(&out))
		_, err := setupWeb(&cfg, nil, stdio)
		require.NoError(t, err)
		assert.Contains(t, out.String(), fmt.Sprintf(`"noindex": %v`, args != nil), args)
	}
}

// Dummy handler to simulate the processing chain.
// It now returns a more detailed message in the response body.
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func TestSecureHeadersMiddlewareIPFSGateway(t *testing.T) {
	cspFor := func(gateway string) string {
		handler := SecureHeadersMiddleware(http.HandlerFunc(dummyHandler), true, "http://example.com", gateway)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest("GET", "http://example.com", nil))
		return rec.Result().Header.Get("Content-Security-Policy")
	}

	csp := cspFor("https://ipfs.filebase.io")
	assert.Contains(t, csp, " https://ipfs.filebase.io ")
	assert.NotContains(t, csp, "https://ipfs.io")
	assert.NotContains(t, csp, "cloudflare-ipfs.com")
	assert.NotContains(t, cspFor(""), "filebase")

	// A subdomain gateway (Kubo, by default on localhost) redirects
	// /ipfs/<cid> to <cid>.ipfs.<host>, so those subdomains are allowed too.
	assert.Contains(t, cspFor("http://localhost:8080"),
		" http://localhost:8080 http://*.ipfs.localhost:8080 http://*.ipns.localhost:8080;")
	// An IP address has no subdomains.
	assert.Contains(t, cspFor("http://127.0.0.1:8080"), " http://127.0.0.1:8080;")
}

func TestSecureHeadersMiddlewareStrict(t *testing.T) {
	handler := SecureHeadersMiddleware(http.HandlerFunc(dummyHandler), true, "http://example.com", "https://ipfs.filebase.io")

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	res := rec.Result()

	// Check common headers.
	if res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options 'nosniff', got '%s'", res.Header.Get("X-Content-Type-Options"))
	}
	if res.Header.Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options 'DENY', got '%s'", res.Header.Get("X-Frame-Options"))
	}
	if res.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Errorf("Expected Referrer-Policy 'no-referrer', got '%s'", res.Header.Get("Referrer-Policy"))
	}

	// Check headers specific to strict mode.
	csp := res.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "img-src 'self' data: https://gnolang.github.io") {
		t.Errorf("Expected Content-Security-Policy img-src to carry the image host allowlist, got '%s'", csp)
	}
	// assets.gnoteam.com has no DNS record, so it must not be allowlisted as an image source.
	if strings.Contains(csp, "assets.gnoteam.com") {
		t.Errorf("Expected Content-Security-Policy not to allowlist 'assets.gnoteam.com', got '%s'", csp)
	}
	// The search dropdown fetches the same-origin /search.json, so connect-src must allow 'self'.
	if !strings.Contains(csp, "connect-src 'self'") {
		t.Errorf("Expected Content-Security-Policy connect-src to contain 'self', got '%s'", csp)
	}
	// style-src must carry a per-response nonce so CodeMirror's injected styles are allowed
	// without relaxing the policy to 'unsafe-inline'.
	if !strings.Contains(csp, "style-src 'self' 'nonce-") {
		t.Errorf("Expected Content-Security-Policy style-src to carry a nonce, got '%s'", csp)
	}
	if res.Header.Get("Strict-Transport-Security") != "max-age=31536000" {
		t.Errorf("Expected Strict-Transport-Security 'max-age=31536000', got '%s'", res.Header.Get("Strict-Transport-Security"))
	}

	// Optionally, verify the response body.
	body := rec.Body.String()
	if !strings.Contains(body, "OK") {
		t.Errorf("Unexpected response body: %s", body)
	}
}

func TestSecureHeadersMiddlewareNonceMatchesContext(t *testing.T) {
	// The nonce echoed in the CSP header must equal the one exposed to downstream
	// handlers via the request context (which they embed in <meta name="csp-nonce">),
	// and it must be freshly generated per response.
	var seen []string
	handler := SecureHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, gnoweb.CSPNonceFromContext(r.Context()))
		w.Write([]byte("OK"))
	}), true, "http://example.com", "")

	do := func() string {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest("GET", "http://example.com", nil))
		return rec.Result().Header.Get("Content-Security-Policy")
	}

	csp1, csp2 := do(), do()

	if len(seen) != 2 || seen[0] == "" || seen[1] == "" {
		t.Fatalf("expected two non-empty context nonces, got %v", seen)
	}
	if !strings.Contains(csp1, "'nonce-"+seen[0]+"'") {
		t.Errorf("CSP header nonce does not match context nonce: header=%q ctx=%q", csp1, seen[0])
	}
	if seen[0] == seen[1] {
		t.Errorf("expected a fresh nonce per response, got the same value twice: %q", seen[0])
	}
	if csp1 == csp2 {
		t.Errorf("expected CSP headers to differ per response due to nonce, both were %q", csp1)
	}
}

func TestSecureHeadersMiddlewareNonStrict(t *testing.T) {
	handler := SecureHeadersMiddleware(http.HandlerFunc(dummyHandler), false, "http://example.com", "https://ipfs.filebase.io")

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	res := rec.Result()

	// Check that the common headers are set.
	if res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options 'nosniff', got '%s'", res.Header.Get("X-Content-Type-Options"))
	}
	if res.Header.Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options 'DENY', got '%s'", res.Header.Get("X-Frame-Options"))
	}
	if res.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Errorf("Expected Referrer-Policy 'no-referrer', got '%s'", res.Header.Get("Referrer-Policy"))
	}

	// In non-strict mode, CSP and HSTS should not be defined.
	if csp := res.Header.Get("Content-Security-Policy"); csp != "" {
		t.Errorf("Did not expect Content-Security-Policy in non-strict mode, got '%s'", csp)
	}
	if hsts := res.Header.Get("Strict-Transport-Security"); hsts != "" {
		t.Errorf("Did not expect Strict-Transport-Security in non-strict mode, got '%s'", hsts)
	}

	// Optionally, verify the response body.
	body := rec.Body.String()
	if !strings.Contains(body, "OK") {
		t.Errorf("Unexpected response body: %s", body)
	}
}

func TestParseAliases(t *testing.T) {
	t.Parallel()

	var (
		existingStatic = path.Join(t.TempDir(), "existing")
		missingStatic  = path.Join(t.TempDir(), "missing")
	)

	_, err := os.Create(existingStatic)
	require.NoError(t, err)

	cases := []struct {
		name       string
		aliasesStr string
		mapSize    int
		error      bool
	}{
		{
			name:       "empty",
			aliasesStr: "",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "only whitespaces",
			aliasesStr: "    ",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "no separator",
			aliasesStr: "alias1",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "too many separators",
			aliasesStr: "alias1 = = target1",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "empty entry",
			aliasesStr: "alias1 = target1, , alias3 = target3",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "valid entry",
			aliasesStr: "alias1 = target1, alias2 = target2, alias3 = target3",
			mapSize:    3,
			error:      false,
		},
		{
			name:       "alias existing static file",
			aliasesStr: "alias1 = static:" + existingStatic,
			mapSize:    1,
			error:      false,
		},
		{
			name:       "alias multiple static files",
			aliasesStr: "alias1 = static:" + existingStatic + ", alias2 = static:" + existingStatic + ", alias3 = static:" + existingStatic,
			mapSize:    3,
			error:      false,
		},
		{
			name:       "alias missing static file",
			aliasesStr: "alias1 = static:" + missingStatic,
			mapSize:    0,
			error:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			aliases, err := parseAliases(tc.aliasesStr)
			if tc.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.mapSize, len(aliases))
		})
	}
}
