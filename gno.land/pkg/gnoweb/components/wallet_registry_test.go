package components

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWallets_EmbeddedRegistry(t *testing.T) {
	t.Parallel()

	w := Wallets()
	require.NotEmpty(t, w, "embedded wallet registry should not be empty")

	// Every entry must carry the fields the frontend chooser relies on.
	for _, wallet := range w {
		assert.NotEmpty(t, wallet.Name, "wallet name")
		assert.NotEmpty(t, wallet.ID, "wallet id")
		assert.NotEmpty(t, wallet.Scheme, "wallet scheme")
		assert.NotContains(t, wallet.Scheme, "://", "scheme must be stored bare")
		assert.True(t, strings.HasPrefix(wallet.Icon, "data:"),
			"icon must be a self-contained data: URI, got %q", wallet.Icon)
	}
}

func TestWallets_ContainsGnokey(t *testing.T) {
	t.Parallel()

	var found *Wallet
	for _, wallet := range Wallets() {
		if wallet.ID == "gnokey" {
			found = &wallet
			break
		}
	}

	require.NotNil(t, found, "registry should contain the gnokey entry")
	assert.Equal(t, "land.gno.gnokey", found.Scheme)
	assert.Contains(t, found.Platforms, "ios")
	assert.Contains(t, found.Platforms, "android")
}

func TestHelpView_PopulatesWallets(t *testing.T) {
	t.Parallel()

	view := HelpView(HelpData{RealmName: "test"})
	require.NotNil(t, view)

	tc, ok := view.Component.(*TemplateComponent)
	require.True(t, ok, "unexpected component type %T", view.Component)
	params, ok := tc.data.(helpViewParams)
	require.True(t, ok, "unexpected view data type %T", tc.data)

	// WalletsJSON must round-trip to the registry so the frontend can parse it.
	var roundtrip []Wallet
	require.NoError(t, json.Unmarshal([]byte(params.WalletsJSON), &roundtrip))
	assert.Equal(t, Wallets(), roundtrip)
}

// The registry must survive html/template escaping verbatim so the browser can
// JSON.parse it.
func TestHelpView_RendersWalletRegistry(t *testing.T) {
	t.Parallel()

	view := HelpView(HelpData{RealmName: "test"})
	var buf bytes.Buffer
	require.NoError(t, view.Render(&buf))

	out := buf.String()
	assert.Contains(t, out, `data-wallet-launch-target="wallet-registry"`)
	assert.Contains(t, out, "land.gno.gnokey")
	assert.Contains(t, out, "data:image/svg+xml;base64,")
}
