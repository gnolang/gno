package gnoweb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// renderHome serves "/" through a freshly built router and returns the body.
func renderHome(t *testing.T, mutate func(*AppConfig)) string {
	t.Helper()

	cfg := NewDefaultAppConfig()
	cfg.NodeRemote = sharedNodeRemote(t)
	mutate(cfg)

	router, err := NewRouter(log.NewTestingLogger(t), cfg)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, response.Code)

	return response.Body.String()
}

// TestNetworkBannerShownOffMainnet is the point of the feature: gnoweb
// otherwise names the chain only inside a popup behind a toggle, so a gnoweb
// pointed at the wrong chain renders identically to one pointed at the right
// chain.
func TestNetworkBannerShownOffMainnet(t *testing.T) {
	t.Parallel()

	body := renderHome(t, func(cfg *AppConfig) { cfg.ChainID = "mystaging" })

	assert.Contains(t, body, "b-banner")
	assert.Contains(t, body, "b-banner--warning")
	assert.Contains(t, body, "mystaging")
	assert.Contains(t, body, "Not gno.land mainnet")
}

func TestNetworkBannerHiddenOnMainnet(t *testing.T) {
	t.Parallel()

	body := renderHome(t, func(cfg *AppConfig) { cfg.ChainID = MainnetChainID })

	assert.NotContains(t, body, "b-banner")
}

func TestNetworkBannerDisabled(t *testing.T) {
	t.Parallel()

	body := renderHome(t, func(cfg *AppConfig) {
		cfg.ChainID = "mystaging"
		cfg.NoNetworkBanner = true
	})

	assert.NotContains(t, body, "b-banner")
}

// TestExplicitBannerWins covers precedence: an operator who configured a banner
// gets theirs, not the automatic one, even off mainnet.
func TestExplicitBannerWins(t *testing.T) {
	t.Parallel()

	body := renderHome(t, func(cfg *AppConfig) {
		cfg.ChainID = "mystaging"
		banner, err := components.NewBannerData("ACME internal staging", components.BannerOptions{
			Variant: components.BannerCaution,
			Color:   "#ff8800",
		})
		require.NoError(t, err)
		cfg.Banner = banner
	})

	assert.Contains(t, body, "ACME internal staging")
	assert.Contains(t, body, "b-banner--caution")
	assert.Contains(t, body, `style="--banner-bg:#ff8800"`)
	assert.NotContains(t, body, "Not gno.land mainnet")
}

func TestNetworkBannerText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		chainID  string
		remote   string
		contains []string
	}{
		{"chain and rpc", "dev", "127.0.0.1:26657", []string{"dev", "127.0.0.1:26657"}},
		{"chain only", "dev", "", []string{"dev"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			banner, err := NetworkBanner(tc.chainID, tc.remote)
			require.NoError(t, err)
			require.True(t, banner.Enabled())

			var sb strings.Builder
			require.NoError(t, banner.Render(&sb))
			for _, want := range tc.contains {
				assert.Contains(t, sb.String(), want)
			}
			assert.Equal(t, "b-banner--warning", banner.VariantClass())
		})
	}
}
