package main

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveRoot builds the gnoweb handler the way gnodev does and returns the
// status and body it serves at "/".
//
// No node is needed: chainId is set, so NewRouter never queries one, and a
// static markdown alias renders without touching the chain. The status is
// returned rather than asserted, because the cases where our page is *not*
// installed fall through to the /r/gnoland/home alias and 400 with no chain
// behind them, which is itself evidence the alias was left alone.
func serveRoot(t *testing.T, cfg *AppConfig) (int, string) {
	t.Helper()

	handler, err := setupGnoWebServer(log.NewTestingLogger(t), cfg, "127.0.0.1:26657")
	require.NoError(t, err)

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))

	return res.Code, res.Body.String()
}

// TestSetupGnoWebServerStagingHome holds the wiring itself in place. Without
// it, reverting the block in setupGnoWebServer leaves every other test green:
// the markdown tests only exercise stagingHomeMarkdown, which is still called
// by nothing.
func TestSetupGnoWebServerStagingHome(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		staging  bool
		webHome  string
		wantPage bool
	}{
		{name: "staging serves the preview page", staging: true, webHome: ":none:", wantPage: true},
		{name: "staging, empty home", staging: true, webHome: "", wantPage: true},
		{name: `staging, "/" means no explicit home too`, staging: true, webHome: "/", wantPage: true},
		// An operator who asked for a home gets theirs, not ours.
		{name: "explicit home opts out", staging: true, webHome: "/r/gnoland/home", wantPage: false},
		// Local mode is untouched.
		{name: "local is unaffected", staging: false, webHome: ":none:", wantPage: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := defaultStagingOptions
			cfg.staging = tc.staging
			cfg.webHome = tc.webHome
			cfg.chainId = "wiring-test"

			code, body := serveRoot(t, &cfg)
			if tc.wantPage {
				assert.Equal(t, http.StatusOK, code)
				assert.Contains(t, body, "A gnodev preview chain")
				assert.Contains(t, body, "wiring-test")
			} else {
				assert.NotContains(t, body, "A gnodev preview chain")
			}
		})
	}
}

// TestSetupGnoWebServerLeavesDefaultAliasesAlone holds the maps.Clone in
// gnoweb.NewDefaultAppConfig in place. DefaultAliases is package-level, so
// without the clone this wiring rewrites "/" for every gnoweb in the process,
// and nothing else in the tree notices.
func TestSetupGnoWebServerLeavesDefaultAliasesAlone(t *testing.T) {
	t.Parallel()

	before := maps.Clone(gnoweb.DefaultAliases)

	cfg := defaultStagingOptions
	cfg.staging = true
	cfg.chainId = "wiring-test"
	_, _ = serveRoot(t, &cfg)

	assert.Equal(t, before, gnoweb.DefaultAliases,
		"setting the staging home must not mutate the package-level default aliases")
	assert.Equal(t, gnoweb.AliasTarget{Value: "/r/gnoland/home", Kind: gnoweb.GnowebPath}, gnoweb.DefaultAliases["/"])
}
