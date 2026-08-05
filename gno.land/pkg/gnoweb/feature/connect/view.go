package connect

import (
	"html/template"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

const WalletsViewType components.ViewType = "wallets-view"

// WalletGroup is one platform's wallets, as the install page lists them.
type WalletGroup struct {
	Platform string // registry key, e.g. "ios"
	Label    string // display label, e.g. "iOS"
	Wallets  []components.Wallet
}

// WalletsData is the render payload for templates/wallets.html.
type WalletsData struct {
	Groups []WalletGroup
}

// GroupWallets buckets the registry by platform in components.WalletPlatforms
// order, dropping platforms nothing supports. A wallet appears in every
// platform it declares.
func GroupWallets(ws []components.Wallet) []WalletGroup {
	groups := make([]WalletGroup, 0, len(components.WalletPlatforms))
	for _, platform := range components.WalletPlatforms {
		var members []components.Wallet
		for _, w := range ws {
			for _, p := range w.Platforms {
				if p == platform {
					members = append(members, w)
					break
				}
			}
		}
		if len(members) == 0 {
			continue
		}
		groups = append(groups, WalletGroup{
			Platform: platform,
			Label:    components.WalletPlatformLabel(platform),
			Wallets:  members,
		})
	}
	return groups
}

// walletsComponent renders from the feature's own template set, so
// components.IndexLayout can wrap the result in the standard chrome without
// the components package knowing these templates.
type walletsComponent struct {
	tmpl *template.Template
	name string
	data any
}

func (c *walletsComponent) Render(w io.Writer) error {
	return c.tmpl.ExecuteTemplate(w, c.name, c.data)
}

// NewWalletsView wraps the install page so callers compose it inside
// IndexLayout through the standard *components.View shape.
func NewWalletsView(data WalletsData) *components.View {
	return &components.View{
		Type:      WalletsViewType,
		Component: &walletsComponent{tmpl: WalletsTemplate, name: "renderWallets", data: data},
	}
}
