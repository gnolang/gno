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

	for _, wallet := range w {
		assert.NotEmpty(t, wallet.Name, "wallet name")
		assert.NotEmpty(t, wallet.ID, "wallet id")
		assert.NotEmpty(t, wallet.RDNS, "wallet rdns")
		assert.Contains(t, []WalletKind{WalletKindExtension, WalletKindApp}, wallet.Kind)
		assert.NotEmpty(t, wallet.Platforms, "wallet platforms")
		assert.NotEmpty(t, wallet.Install, "wallet install links")
		assert.True(t, strings.HasPrefix(wallet.Icon, "data:image/"),
			"icon must be a self-contained data:image/ URI, got %q", wallet.Icon)
		if wallet.Kind == WalletKindApp {
			assert.Regexp(t, walletSchemeRe, wallet.Scheme, "an app must carry a launch scheme")
		}
	}
}

func TestWallets_ContainsGnokeyAndAdena(t *testing.T) {
	t.Parallel()

	byID := map[string]Wallet{}
	for _, w := range Wallets() {
		byID[w.ID] = w
	}

	gnokey, ok := byID["gnokey"]
	require.True(t, ok, "registry should contain the gnokey entry")
	assert.Equal(t, WalletKindApp, gnokey.Kind)
	assert.Equal(t, "land.gno.gnokey", gnokey.Scheme)
	assert.Equal(t, "land.gno.gnokey", gnokey.RDNS)
	assert.Empty(t, gnokey.Global, "an app has no window global")
	assert.ElementsMatch(t, []string{"ios", "android"}, gnokey.Platforms)

	adena, ok := byID["adena"]
	require.True(t, ok, "registry should contain the adena entry")
	assert.Equal(t, WalletKindExtension, adena.Kind)
	assert.Empty(t, adena.Scheme, "an extension is reached in-page, not by scheme")
	assert.Equal(t, "land.gno.adena", adena.RDNS)
	assert.Equal(t, "adena", adena.Global)
}

func TestValidateWallets(t *testing.T) {
	t.Parallel()

	app := func() Wallet {
		return Wallet{
			Name: "W", ID: "w", RDNS: "land.gno.w", Kind: WalletKindApp,
			Scheme: "land.gno.w", Icon: "data:image/svg+xml;base64,x",
			Platforms: []string{"ios"},
			Install:   []InstallLink{{Label: "GitHub", URL: "https://example.com/w"}},
		}
	}
	ext := func() Wallet {
		w := app()
		w.Kind, w.Scheme, w.Global, w.Platforms = WalletKindExtension, "", "w", []string{"chrome"}
		return w
	}

	cases := []struct {
		name    string
		mutate  func(*Wallet) // applied to the second of two entries
		wantErr string
	}{
		{"missing name", func(w *Wallet) { w.Name = "" }, "name, id and rdns are required"},
		{"missing id", func(w *Wallet) { w.ID = "" }, "name, id and rdns are required"},
		{"missing rdns", func(w *Wallet) { w.RDNS = "" }, "name, id and rdns are required"},
		{"unknown kind", func(w *Wallet) { w.Kind = "plugin" }, "kind must be"},
		{"empty kind", func(w *Wallet) { w.Kind = "" }, "kind must be"},
		{"app without scheme", func(w *Wallet) { w.Scheme = "" }, "an app entry requires a scheme"},
		{"scheme with host", func(w *Wallet) { w.Scheme = "land.gno.w2://tx" }, "not a valid URL scheme"},
		// Passes a bare "no ://" check but executes if it ever reaches
		// window.location; the grammar must reject any ":".
		{"javascript payload", func(w *Wallet) { w.Scheme = "javascript:alert(1)//" }, "not a valid URL scheme"},
		{"non-image icon", func(w *Wallet) { w.Icon = "data:text/html,x" }, "data:image/ URI"},
		{"remote icon", func(w *Wallet) { w.Icon = "https://example.com/i.svg" }, "data:image/ URI"},
		{"no platforms", func(w *Wallet) { w.Platforms = nil }, "at least one platform"},
		{"unknown platform", func(w *Wallet) { w.Platforms = []string{"symbian"} }, "unknown platform"},
		{"no install link", func(w *Wallet) { w.Install = nil }, "at least one install link"},
		{"install link without label", func(w *Wallet) { w.Install[0].Label = "" }, "install link needs a label"},
		{"insecure install link", func(w *Wallet) { w.Install[0].URL = "http://example.com" }, "install URL must be https"},
		{"duplicate id", func(w *Wallet) { w.ID = "w" }, "duplicate wallet id"},
		{"duplicate rdns", func(w *Wallet) { w.RDNS = "land.gno.w" }, "duplicate wallet rdns"},
		{"duplicate scheme", func(w *Wallet) { w.Scheme = "land.gno.w" }, "duplicate wallet scheme"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first, second := app(), app()
			second.ID, second.RDNS, second.Scheme = "w2", "land.gno.w2", "land.gno.w2"
			tc.mutate(&second)
			err := validateWallets([]Wallet{first, second})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}

	t.Run("global on an app", func(t *testing.T) {
		t.Parallel()
		w := app()
		w.Global = "gnokey"
		err := validateWallets([]Wallet{w})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only an extension may declare a global")
	})

	t.Run("malformed global", func(t *testing.T) {
		t.Parallel()
		w := ext()
		w.Global = "window.adena"
		err := validateWallets([]Wallet{w})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a valid global name")
	})

	t.Run("two scheme-less extensions do not collide", func(t *testing.T) {
		t.Parallel()
		a, b := ext(), ext()
		b.ID, b.RDNS, b.Global = "w2", "land.gno.w2", "w2"
		assert.NoError(t, validateWallets([]Wallet{a, b}))
	})

	extW := ext()
	extW.ID, extW.RDNS = "w-ext", "land.gno.w-ext" // distinct from app(), which also uses "w"/"land.gno.w"
	assert.NoError(t, validateWallets([]Wallet{app(), extW}))
	assert.NoError(t, validateWallets(nil))
}

func TestWalletPlatformLabel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "iOS", WalletPlatformLabel("ios"))
	assert.Equal(t, "Android", WalletPlatformLabel("android"))
	assert.Equal(t, "Chrome", WalletPlatformLabel("chrome"))
	assert.Equal(t, "symbian", WalletPlatformLabel("symbian"))
	// Every accepted platform must have a label defined.
	for _, p := range WalletPlatforms {
		assert.NotEqual(t, p, WalletPlatformLabel(p), "platform %q has no label", p)
	}
}

func TestWalletsJSON_RoundTrips(t *testing.T) {
	t.Parallel()

	var roundtrip []Wallet
	require.NoError(t, json.Unmarshal([]byte(WalletsJSON()), &roundtrip))
	assert.Equal(t, Wallets(), roundtrip)
}

// The registry must survive html/template escaping verbatim so the browser can
// JSON.parse it.
func TestHelpView_StillRenders(t *testing.T) {
	t.Parallel()

	view := HelpView(HelpData{RealmName: "test"})
	var buf bytes.Buffer
	require.NoError(t, view.Render(&buf))
	assert.Contains(t, buf.String(), "test")
}

// The chooser now lives in the header layout, so the help view must not embed
// a second copy of the registry.
func TestHelpView_NoLongerEmbedsRegistry(t *testing.T) {
	t.Parallel()

	view := HelpView(HelpData{RealmName: "test"})
	var buf bytes.Buffer
	require.NoError(t, view.Render(&buf))

	out := buf.String()
	assert.NotContains(t, out, `data-wallet-launch-target="wallet-registry"`)
	assert.NotContains(t, out, `data-wallet-launch-target="chooser"`)
}
