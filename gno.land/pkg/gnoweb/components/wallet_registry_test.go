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
		assert.Regexp(t, walletSchemeRe, wallet.Scheme, "scheme must be a bare URL scheme")
		assert.True(t, strings.HasPrefix(wallet.Icon, "data:image/"),
			"icon must be a self-contained data:image/ URI, got %q", wallet.Icon)
	}
}

func TestValidateWallets(t *testing.T) {
	t.Parallel()

	valid := func() Wallet {
		return Wallet{Name: "W", ID: "w", Scheme: "land.gno.w", Icon: "data:image/svg+xml;base64,x"}
	}

	cases := []struct {
		name    string
		mutate  func(*Wallet) // applied to the second of two entries
		wantErr string
	}{
		{"missing name", func(w *Wallet) { w.Name = "" }, "name and id are required"},
		{"missing id", func(w *Wallet) { w.ID = "" }, "name and id are required"},
		{"scheme with host", func(w *Wallet) { w.Scheme = "land.gno.w2://tx" }, "not a valid URL scheme"},
		// Passes a bare "no ://" check but executes if it ever reaches
		// window.location; the grammar must reject any ":".
		{"javascript payload", func(w *Wallet) { w.Scheme = "javascript:alert(1)//" }, "not a valid URL scheme"},
		{"empty scheme", func(w *Wallet) { w.Scheme = "" }, "not a valid URL scheme"},
		{"non-image icon", func(w *Wallet) { w.Icon = "data:text/html,x" }, "data:image/ URI"},
		{"remote icon", func(w *Wallet) { w.Icon = "https://example.com/i.svg" }, "data:image/ URI"},
		{"duplicate id", func(w *Wallet) { w.ID = "w" }, "duplicate wallet id"},
		{"duplicate scheme", func(w *Wallet) { w.Scheme = "land.gno.w" }, "duplicate wallet scheme"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first, second := valid(), valid()
			second.ID, second.Scheme = "w2", "land.gno.w2"
			tc.mutate(&second)
			err := validateWallets([]Wallet{first, second})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}

	assert.NoError(t, validateWallets([]Wallet{valid()}))
	assert.NoError(t, validateWallets(nil))
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
	assert.Contains(t, out, "data:image/png;base64,")
}
