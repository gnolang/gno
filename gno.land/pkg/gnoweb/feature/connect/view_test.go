package connect

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupWallets_OrderAndMembership(t *testing.T) {
	t.Parallel()

	ws := []components.Wallet{
		{Name: "Ext", ID: "ext", RDNS: "x.ext", Kind: components.WalletKindExtension,
			Platforms: []string{"chrome", "firefox"},
			Install:   []components.InstallLink{{Label: "Store", URL: "https://example.com/e"}}},
		{Name: "App", ID: "app", RDNS: "x.app", Kind: components.WalletKindApp, Scheme: "x.app",
			Platforms: []string{"ios", "android"},
			Install:   []components.InstallLink{{Label: "GitHub", URL: "https://example.com/a"}}},
	}

	groups := GroupWallets(ws)

	// Groups follow components.WalletPlatforms order, and empty ones are dropped.
	var got []string
	for _, g := range groups {
		got = append(got, g.Platform)
	}
	assert.Equal(t, []string{"ios", "android", "chrome", "firefox"}, got)

	assert.Equal(t, "iOS", groups[0].Label)
	require.Len(t, groups[0].Wallets, 1)
	assert.Equal(t, "App", groups[0].Wallets[0].Name)
	require.Len(t, groups[2].Wallets, 1)
	assert.Equal(t, "Ext", groups[2].Wallets[0].Name)
}

func TestGroupWallets_Empty(t *testing.T) {
	t.Parallel()

	assert.Empty(t, GroupWallets(nil))
}

func TestNewWalletsView_Renders(t *testing.T) {
	t.Parallel()

	view := NewWalletsView(WalletsData{Groups: GroupWallets(components.Wallets())})
	require.NotNil(t, view)
	assert.Equal(t, WalletsViewType, view.Type)

	var buf bytes.Buffer
	require.NoError(t, view.Render(&buf))

	out := buf.String()
	assert.Contains(t, out, "Gnokey")
	assert.Contains(t, out, "Adena")
	assert.Contains(t, out, `data-platform="ios"`)
	assert.Contains(t, out, `data-platform="chrome"`)
	assert.Contains(t, out, "https://github.com/gnolang/gnokey-mobile")
	// Kind is shown so a user knows what they are installing.
	assert.Contains(t, out, "Extension")
	assert.Contains(t, out, "App")
}
